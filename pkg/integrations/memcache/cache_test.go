package memcache

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_SetAndGet(t *testing.T) {
	c := New[string, int]()

	c.Set("a", 1)
	c.Set("b", 2)

	val, ok := c.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, val)

	val, ok = c.Get("b")
	assert.True(t, ok)
	assert.Equal(t, 2, val)

	_, ok = c.Get("c")
	assert.False(t, ok)
}

func TestCache_Delete(t *testing.T) {
	c := New[string, int]()

	c.Set("a", 1)
	c.Delete("a")

	_, ok := c.Get("a")
	assert.False(t, ok)
}

func TestCache_Keys(t *testing.T) {
	c := New[string, int]()

	c.Set("a", 1)
	c.Set("b", 2)

	keys := c.Keys()
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "a")
	assert.Contains(t, keys, "b")
}

func TestCache_Values(t *testing.T) {
	c := New[string, int]()

	c.Set("a", 1)
	c.Set("b", 2)

	values := c.Values()
	assert.Len(t, values, 2)
	assert.Contains(t, values, 1)
	assert.Contains(t, values, 2)
}

func TestCache_Clear(t *testing.T) {
	c := New[string, int]()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Clear()

	assert.Equal(t, 0, c.Len())
}

func TestCache_Len(t *testing.T) {
	c := New[string, int]()

	assert.Equal(t, 0, c.Len())

	c.Set("a", 1)
	assert.Equal(t, 1, c.Len())

	c.Set("b", 2)
	assert.Equal(t, 2, c.Len())
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := New[int, int]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Set(i, i*2)
		}(i)
	}

	wg.Wait()
	assert.Equal(t, 100, c.Len())

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			val, ok := c.Get(i)
			assert.True(t, ok)
			assert.Equal(t, i*2, val)
		}(i)
	}

	wg.Wait()
}
