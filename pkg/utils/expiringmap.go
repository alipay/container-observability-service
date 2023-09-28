package utils

import (
	"sync"
	"time"
)

func ConvertNil(value string) string {
	if len(value) == 0 {
		return "Nil"
	}

	return value
}

// expired map
type Item struct {
	value      interface{}
	expiration int64
}

type ExpiringMap struct {
	m       sync.Mutex
	items   map[string]Item
	timeout time.Duration
}

func NewExpiringMap(timeout time.Duration) *ExpiringMap {
	em := &ExpiringMap{
		items:   make(map[string]Item),
		timeout: timeout,
	}
	go em.cleanup()
	return em
}

func (em *ExpiringMap) Set(key string, value interface{}) {
	em.m.Lock()
	defer em.m.Unlock()
	em.items[key] = Item{
		value:      value,
		expiration: time.Now().Add(em.timeout).UnixNano(),
	}
}

func (em *ExpiringMap) Get(key string) (interface{}, bool) {
	em.m.Lock()
	defer em.m.Unlock()
	item, found := em.items[key]
	if !found {
		return nil, false
	}
	if time.Now().UnixNano() > item.expiration {
		return nil, false
	}
	return item.value, true
}

func (em *ExpiringMap) Delete(key string) {
	em.m.Lock()
	defer em.m.Unlock()
	delete(em.items, key)
}

func (em *ExpiringMap) deleteExpiredKeys() {
	em.m.Lock()
	defer em.m.Unlock()
	now := time.Now().UnixNano()
	for key, item := range em.items {
		if now < item.expiration {
			break
		}
		delete(em.items, key)
	}
}

func (em *ExpiringMap) cleanup() {
	for {
		time.Sleep(em.timeout)
		em.deleteExpiredKeys()
	}
}
