package coolcache

import (
	"github.com/patrickmn/go-cache"
	"hash/fnv"
	"time"
)

const (
	// For use with functions that take an expiration time.
	NoExpiration time.Duration = -1
)

type CoolCache struct {
	shardNumber uint64
	shards      []*cache.Cache
}

func NewCoolCache(shardNumber int, defaultExpiration, cleanupInterval time.Duration) *CoolCache {
	var cc CoolCache
	cc.shards = make([]*cache.Cache, shardNumber)
	cc.shardNumber = uint64(shardNumber)
	for i := 0; i < shardNumber; i++ {
		cc.shards[i] = cache.New(defaultExpiration, cleanupInterval)
	}
	return &cc
}

// Add an item to the cache, replacing any existing item. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (c *CoolCache) Set(key string, value interface{}, d time.Duration) {
	hashCode := c.Sum64(key)
	c.shards[hashCode%c.shardNumber].Set(key, value, d)
}

// Get an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (c *CoolCache) Get(key string) (interface{}, bool) {
	hashCode := c.Sum64(key)
	return c.shards[hashCode%c.shardNumber].Get(key)
}

func (c *CoolCache) Sum64(key string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(key))
	return hash.Sum64()
}
