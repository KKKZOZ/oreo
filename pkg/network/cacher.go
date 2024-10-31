package network

import (
	"fmt"

	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

type Cacher struct {
	cache        map[string]txn.GroupKeyItem
	CacheRequest int
	CacheHit     int
}

func NewCacher() *Cacher {
	return &Cacher{
		cache:        make(map[string]txn.GroupKeyItem),
		CacheRequest: 0,
		CacheHit:     0,
	}
}

func (c *Cacher) Get(key string) (txn.GroupKeyItem, bool) {
	item, ok := c.cache[key]
	c.CacheRequest++
	if ok {
		c.CacheHit++
	}
	return item, ok
}

func (c *Cacher) Set(key string, item txn.GroupKeyItem) {
	c.cache[key] = item
}

func (c *Cacher) Delete(key string) {
	delete(c.cache, key)
}

func (c *Cacher) Statistic() string {
	return fmt.Sprintf("CacheRequest: %d, CacheHit: %d, HitRate: %.2f", c.CacheRequest, c.CacheHit, float64(c.CacheHit)/float64(c.CacheRequest))
}
