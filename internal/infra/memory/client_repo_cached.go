package memory

import (
	"context"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
)

const clientCacheTTL = 5 * time.Minute

// ClientRepoCached wraps a [goidc.ClientManager] with a local in-process TTL
// cache. Client lookups are served from cache; saves and deletes write through
// to the underlying store and invalidate the cache entry immediately.
//
// Because client IDs are always known at save/delete time, invalidation is
// exact — no background scanning needed.
type ClientRepoCached struct {
	repo  goidc.ClientManager
	cache *TTLCacheImpl[string, *goidc.Client]
}

// NewClientRepoCached wraps repo with an in-process TTL cache.
func NewClientRepoCached(repo goidc.ClientManager) *ClientRepoCached {
	return &ClientRepoCached{
		repo:  repo,
		cache: NewTTLCache[string, *goidc.Client](clientCacheTTL),
	}
}

// Run starts the cache's background cleanup loop. It blocks until ctx is
// cancelled and should be called in a goroutine.
func (r *ClientRepoCached) Run(ctx context.Context) error {
	return r.cache.Run(ctx)
}

// Save persists the client to the underlying store and refreshes the cache entry.
func (r *ClientRepoCached) Save(ctx context.Context, c *goidc.Client) error {
	if err := r.repo.Save(ctx, c); err != nil {
		return err
	}
	r.cache.Put(c.ID, c, PutModeNew)
	return nil
}

// Client returns the client with the given id, serving from cache when available.
// On a cache miss the client is fetched from the underlying store and cached.
func (r *ClientRepoCached) Client(ctx context.Context, id string) (*goidc.Client, error) {
	if c, ok := r.cache.Get(id); ok {
		return c, nil
	}

	c, err := r.repo.Client(ctx, id)
	if err != nil {
		return nil, err
	}

	r.cache.Put(id, c, PutModeNew)
	return c, nil
}

// Delete removes the client from the underlying store and evicts the cache entry.
func (r *ClientRepoCached) Delete(ctx context.Context, id string) error {
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}
	r.cache.Delete(id)
	return nil
}

var _ goidc.ClientManager = (*ClientRepoCached)(nil)
