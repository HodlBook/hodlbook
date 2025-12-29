package cache

type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Delete(key K)
	Keys() []K
	Values() []V
	Clear()
	Len() int
}
