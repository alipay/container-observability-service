package utils

type SafeMap struct {
	size int64
	data ConcurrentMap
}

func NewSafeMap() *SafeMap {
	SHARD_COUNT = 32
	return &SafeMap{
		data: New(),
	}
}

func (smap *SafeMap) Set(key string, data interface{}) {
	defer IgnorePanic("safe_map_set")
	smap.data.Set(key, data)
}

func (smap *SafeMap) Get(key string) (interface{}, bool) {
	defer IgnorePanic("safe_map_get")
	value, ok := smap.data.Get(key)
	return value, ok
}

func (smap *SafeMap) Delete(key string) {
	defer IgnorePanic("safe_map_delete")
	if smap.data.Has(key) {
		smap.data.Remove(key)
	}
}

func (smap *SafeMap) Size() int {
	return smap.data.Count()
}

// IterateWithFunc iterate map
func (smap *SafeMap) IterateWithFunc(f func(interface{})) {
	defer IgnorePanic("safe_map_iterate")
	smap.data.IterCbConcurrent(func(key string, v interface{}) {
		f(v)
	})
}
