package pharoscache

import (
	"sync"
)

type Evicted interface {
	On(key string, value []byte)
}

type Option struct {
	MaxEntries int
	OnEvicted Evicted
}

type cache struct {
	option *Option
	lock *sync.Mutex
	buckets map[string]*bucket
}

func NewCache(option *Option) *cache {
	c := new(cache)
	if option.MaxEntries == 0 {
		option.MaxEntries = 1024
	}
	c.option = option
	c.lock = new(sync.Mutex)
	c.buckets = make(map[string]*bucket)
	return c
}

func (c *cache) Bucket(name string) *bucket {
	c.lock.Lock()
	defer c.lock.Unlock()
	if b, hit := c.buckets[name]; hit {
		return b
	}
	b := newBucket(c.option.MaxEntries, c.option.OnEvicted)
	c.buckets[name] = b
	return b
}

func (c *cache) RemoveBucket(name string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	if b, hit := c.buckets[name]; hit {
		b.stop()
		delete(c.buckets, name)
		return hit
	}
	return false
}

func (c *cache) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	keys := make([]string, 0, len(c.buckets))
	for k, b := range c.buckets {
		b.stop()
		keys = append(keys, k)
	}
	for _, k := range keys {
		delete(c.buckets, k)
	}
}