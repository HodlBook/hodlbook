package memcache

import (
	"sync"

	"hodlbook/pkg/types/cache"
)

var _ cache.Cache[string, any] = (*Cache[string, any])(nil)

type Cache[K comparable, V any] struct {
	data  map[K]V
	mutex sync.RWMutex
}

func New[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		data: make(map[K]V),
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = value
}

func (c *Cache[K, V]) Delete(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

func (c *Cache[K, V]) Keys() []K {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	keys := make([]K, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

func (c *Cache[K, V]) Values() []V {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	values := make([]V, 0, len(c.data))
	for _, v := range c.data {
		values = append(values, v)
	}
	return values
}

func (c *Cache[K, V]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[K]V)
}

func (c *Cache[K, V]) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.data)
}
