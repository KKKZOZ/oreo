package network

import (
	"fmt"
	"sync"

	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

type Cacher struct {
	mu           sync.RWMutex
	cache        map[string]txn.GroupKeyItem
	CacheRequest int
	CacheHit     int
}

func NewCacher() *Cacher {
	return &Cacher{
		mu:           sync.RWMutex{},
		cache:        make(map[string]txn.GroupKeyItem),
		CacheRequest: 0,
		CacheHit:     0,
	}
}

func (c *Cacher) Get(key string) (txn.GroupKeyItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.cache[key]
	c.CacheRequest++
	if ok {
		c.CacheHit++
	}
	return item, ok
}

func (c *Cacher) Set(key string, item txn.GroupKeyItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = item
}

func (c *Cacher) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

func (c *Cacher) Statistic() string {
	return fmt.Sprintf(
		"CacheRequest: %d, CacheHit: %d, HitRate: %.2f",
		c.CacheRequest,
		c.CacheHit,
		float64(c.CacheHit)/float64(c.CacheRequest),
	)
}

func (c *Cacher) Clear() {
	c.cache = make(map[string]txn.GroupKeyItem)
	c.CacheRequest = 0
	c.CacheHit = 0
}
