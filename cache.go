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

type Cache struct {
	option *Option
	lock *sync.Mutex
	buckets map[string]*Bucket
}

func NewCache(option *Option) *Cache {
	c := new(Cache)
	if option.MaxEntries == 0 {
		option.MaxEntries = 1024
	}
	c.option = option
	c.lock = new(sync.Mutex)
	c.buckets = make(map[string]*Bucket)
	return c
}

func (c *Cache) Bucket(name string) *Bucket {
	c.lock.Lock()
	defer c.lock.Unlock()
	if b, hit := c.buckets[name]; hit {
		return b
	}
	b := newBucket(c.option.MaxEntries, c.option.OnEvicted)
	c.buckets[name] = b
	return b
}