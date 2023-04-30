package concurrentmap

import (
	"sync"
)

type ConcurrentMap[M ~map[K]V, K comparable, V any] struct {
	m M
	l sync.RWMutex
}

func Wrap[M ~map[K]V, K comparable, V any](m M) *ConcurrentMap[M, K, V] {
	return &ConcurrentMap[M, K, V]{m: m, l: sync.RWMutex{}}
}

func (c *ConcurrentMap[M, K, V]) WithRLock(f func(M) any) any {
	c.l.RLock()
	defer c.l.RUnlock()
	return f(c.m)
}

func (c *ConcurrentMap[M, K, V]) WithLock(f func(f M) any) any {
	c.l.Lock()
	defer c.l.Unlock()
	return f(c.m)
}

func (c *ConcurrentMap[M, K, V]) Get(key K) (V, bool) {
	c.l.RLock()
	defer c.l.RUnlock()
	val, has := c.m[key]
	return val, has
}

func (c *ConcurrentMap[M, K, V]) Put(key K, val V) {
	c.WithLock(
		func(m M) any {
			m[key] = val
			return nil
		},
	)
}

func (c *ConcurrentMap[M, K, V]) Delete(key K) {
	c.WithLock(
		func(m M) any {
			delete(m, key)
			return nil
		},
	)
}
