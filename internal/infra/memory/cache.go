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
// It's optimised for performance
type TTLCache[ID comparable, ObjT any] interface {
	GetTTL() time.Duration
	Len() int

	Get(ID) (ObjT, bool)
	Put(ID, ObjT) (newItem bool, expiresAt time.Time)
	Delete(ID) (existed bool)

	// CleanUP
	NextCleanUP() (ID, time.Time)

	// Run is requred for clean up to work
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

func (c *TTLCacheImpl[ID, Obj]) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.len
}

// unlink removes node from the cache. Doesn't update cache length
func (c *TTLCacheImpl[ID, Obj]) unlink(n *node[ID, Obj], updateItemsMap bool) {
	// unlink from list
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		// n was head
		c.head = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	} else {
		// n was tail
		c.tail = n.prev
	}

	n.prev = nil
	n.next = nil

	if updateItemsMap {
		delete(c.items, n.id)
	}
}

// unlink adds node to the cache. Doesn't update cache length
func (c *TTLCacheImpl[ID, Obj]) link(n *node[ID, Obj], updateItemsMap bool) {
	if c.tail == nil {
		c.head = n
		c.tail = n
	} else {
		n.prev = c.tail
		c.tail.next = n
		c.tail = n
	}
	if updateItemsMap {
		c.items[n.id] = n
	}
}

// Get returns value and true if present and not expired.
// If the item is expired it will be removed and (zero, false) returned.
func (c *TTLCacheImpl[ID, Obj]) Get(id ID) (Obj, bool) {
	var zero Obj

	c.lock.RLock()
	defer c.lock.RUnlock()
	item, ok := c.items[id]
	if !ok {
		return zero, false
	}
	return item.value, true
}

func (c *TTLCacheImpl[ID, Obj]) Delete(id ID) (existed bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	n, ok := c.items[id]
	if !ok {
		return false
	}
	c.unlink(n, true)
	c.len--

	return true
}

func (c *TTLCacheImpl[ID, Obj]) Put(id ID, value Obj, mode PutMode) (bool, time.Time) {
	now := time.Now()

	c.lock.Lock()
	defer c.lock.Unlock()

	n, exists := c.items[id]
	if !exists {
		c.link(&node[ID, Obj]{
			id:       id,
			value:    value,
			inserted: now,
		}, true)
		c.len++
		c.putCond.Signal()

		return true, now.Add(c.ttl)
	}

	switch mode {

	case PutModeKeepExisting:
		return false, n.inserted.Add(c.ttl)

	case PutModeReplace:
		n.value = value
		return false, n.inserted.Add(c.ttl)

	case PutModeExtendTTL:
		// renew TTL
		n.inserted = now

		if c.tail != n {
			c.unlink(n, false)
			c.link(n, false)
		}

		return false, n.inserted.Add(c.ttl)

	case PutModeNew:
		c.unlink(n, true)
		c.link(&node[ID, Obj]{
			id:       id,
			value:    value,
			inserted: now,
		}, true)
		return true, now.Add(c.ttl)
	}

	panic("unreachable")
}

func (c *TTLCacheImpl[ID, Obj]) Run(ctx context.Context) error {
	for {
		c.lock.Lock()

		// wait for items
		for c.len == 0 {
			c.putCond.Wait()
			if ctx.Err() != nil {
				c.lock.Unlock()
				return ctx.Err()
			}
		}

		// determine next expiration
		n := c.head
		expireAt := n.inserted.Add(c.ttl)
		now := time.Now()

		if now.Before(expireAt) {
			// sleep until next expiration (without holding lock)
			wait := expireAt.Sub(now)
			c.lock.Unlock()

			select {
			case <-time.After(wait):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// expired: remove head
		c.unlink(n, true)
		c.len--

		c.lock.Unlock()
	}
}
