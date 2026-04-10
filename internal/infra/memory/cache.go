// Package memory provides in-memory storage implementations and interfaces
package memory

import (
	"context"
	"sync"
	"time"
)

type PutMode = int8

// All of the modes will just create a new item with a fresh TTL
// if item didn't exist. Modes cause difference only if item existed
const (
	// Replaces item, keep TTL
	PutModeReplace PutMode = iota
	// Just keep existing item (if existed). No TTL changes
	PutModeKeepExisting
	// Keep existing item (if existed), but extend it's TTL like it was just created
	PutModeExtendTTL
	// Inserts item like it's new
	// so it's TTL is also renewed
	PutModeNew
)

// TTLCache is an interface for cache where
// TTL is equal for each object in the cache
// It's optimized for performance
type TTLCache[ID comparable, ObjT any] interface {
	GetTTL() time.Duration
	Len() int

	Get(ID) (ObjT, bool)
	Put(ID, ObjT) (newItem bool, expiresAt time.Time)
	Delete(ID) (existed bool)

	// CleanUP
	NextCleanUP() (ID, time.Time)

	// Run is required for clean up to work
	// Since CleanUPs happen in order we don't
	// need to constantly scan expired elements
	Run(ctx context.Context) error
}

type node[ID comparable, Obj any] struct {
	id         ID
	value      Obj
	inserted   time.Time
	prev, next *node[ID, Obj]
}

type TTLCacheImpl[ID comparable, Obj any] struct {
	ttl time.Duration

	lock  *sync.RWMutex
	items map[ID]*node[ID, Obj]
	head  *node[ID, Obj] // oldest (next to expire)
	tail  *node[ID, Obj] // newest
	len   int
	// used to notify CleanUP look in case it's waiting for new items
	putCond *sync.Cond
}

func NewTTLCache[ID comparable, Obj any](ttl time.Duration) *TTLCacheImpl[ID, Obj] {
	lock := &sync.RWMutex{}
	return &TTLCacheImpl[ID, Obj]{
		ttl:     ttl,
		lock:    lock,
		items:   make(map[ID]*node[ID, Obj]),
		head:    nil,
		tail:    nil,
		len:     0,
		putCond: &sync.Cond{L: lock},
	}
}

func (cache *TTLCacheImpl[ID, Obj]) GetTTL() time.Duration {
	return cache.ttl
}

func (cache *TTLCacheImpl[ID, Obj]) Len() int {
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	return cache.len
}

// unlink removes node from the cache. Doesn't update cache length
func (cache *TTLCacheImpl[ID, Obj]) unlink(n *node[ID, Obj], updateItemsMap bool) {
	// unlink from list
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		// n was head
		cache.head = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	} else {
		// n was tail
		cache.tail = n.prev
	}

	n.prev = nil
	n.next = nil

	if updateItemsMap {
		delete(cache.items, n.id)
	}
}

// link adds node to the cache. Doesn't update cache length
func (cache *TTLCacheImpl[ID, Obj]) link(n *node[ID, Obj], updateItemsMap bool) {
	if cache.tail == nil {
		cache.head = n
		cache.tail = n
	} else {
		n.prev = cache.tail
		cache.tail.next = n
		cache.tail = n
	}
	if updateItemsMap {
		cache.items[n.id] = n
	}
}

// Get returns value and true if present and not expired.
// If the item is expired it will be removed and (zero, false) returned.
func (cache *TTLCacheImpl[ID, Obj]) Get(id ID) (Obj, bool) {
	var zero Obj

	cache.lock.RLock()
	defer cache.lock.RUnlock()
	item, ok := cache.items[id]
	if !ok {
		return zero, false
	}
	return item.value, true
}

func (cache *TTLCacheImpl[ID, Obj]) Delete(id ID) (existed bool) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	n, ok := cache.items[id]
	if !ok {
		return false
	}
	cache.unlink(n, true)
	cache.len--

	return true
}

func (cache *TTLCacheImpl[ID, Obj]) Put(id ID, value Obj, mode PutMode) (bool, time.Time) {
	now := time.Now()

	cache.lock.Lock()
	defer cache.lock.Unlock()

	n, exists := cache.items[id]
	if !exists {
		cache.link(&node[ID, Obj]{
			id:       id,
			value:    value,
			inserted: now,
		}, true)
		cache.len++
		cache.putCond.Signal()

		return true, now.Add(cache.ttl)
	}

	switch mode {

	case PutModeKeepExisting:
		return false, n.inserted.Add(cache.ttl)

	case PutModeReplace:
		n.value = value
		return false, n.inserted.Add(cache.ttl)

	case PutModeExtendTTL:
		// renew TTL
		n.inserted = now

		if cache.tail != n {
			cache.unlink(n, false)
			cache.link(n, false)
		}

		return false, n.inserted.Add(cache.ttl)

	case PutModeNew:
		cache.unlink(n, true)
		cache.link(&node[ID, Obj]{
			id:       id,
			value:    value,
			inserted: now,
		}, true)
		return true, now.Add(cache.ttl)
	}

	panic("unreachable")
}

func (cache *TTLCacheImpl[ID, Obj]) Run(ctx context.Context) error {
	for {
		cache.lock.Lock()

		// wait for items
		for cache.len == 0 {
			cache.putCond.Wait()
			if ctx.Err() != nil {
				cache.lock.Unlock()
				return ctx.Err()
			}
		}

		// determine next expiration
		n := cache.head
		expireAt := n.inserted.Add(cache.ttl)
		now := time.Now()

		if now.Before(expireAt) {
			// sleep until next expiration (without holding lock)
			wait := expireAt.Sub(now)
			cache.lock.Unlock()

			select {
			case <-time.After(wait):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// expired: remove head
		cache.unlink(n, true)
		cache.len--

		cache.lock.Unlock()
	}
}
