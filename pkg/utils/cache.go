package utils

import (
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type node struct {
	key        string
	val        interface{}
	prev, next *node
	lastAccess time.Time
}

type linkedLRUCache struct {
	sync.Mutex
	capacity   int
	head, tail *node
	cache      map[string]*node
	name       string
}

func LRUCache(name string, capacity int) LRU {
	lruCache := &linkedLRUCache{
		name:     name,
		capacity: capacity,
		cache:    make(map[string]*node),
	}
	//lruCache.startClearExpired()
	return lruCache
}

type LRU interface {
	Get(key string) interface{}
	Put(key string, value interface{})
}

func (c *linkedLRUCache) Get(key string) interface{} {
	c.Lock()
	defer c.Unlock()
	metrics.LRUCacheCounter.WithLabelValues(c.name, "total").Inc()

	if node, ok := c.cache[key]; ok {
		metrics.LRUCacheCounter.WithLabelValues(c.name, "hit").Inc()

		c.remove(node)
		c.add(node)
		return node.val
	}
	return nil
}

func (c *linkedLRUCache) Put(key string, val interface{}) {
	c.Lock()
	defer c.Unlock()
	klog.V(7).Infof("%s cache size: %d", c.name, len(c.cache))

	if n, ok := c.cache[key]; ok {
		c.remove(n)
		n.val = val
		c.add(n)
		return
	} else {
		n = &node{key: key, val: val}
		c.cache[key] = n
		c.add(n)
	}

	if len(c.cache) > c.capacity {
		delete(c.cache, c.tail.key)
		c.remove(c.tail)
	}
}

// add node to head
func (c *linkedLRUCache) add(n *node) {
	if c.head != nil {
		c.head.prev = n
		n.next = c.head
	}

	c.head = n
	c.head.prev = nil

	if c.tail == nil {
		c.tail = n
		c.tail.next = nil
	}

}

// remove node
func (c *linkedLRUCache) remove(n *node) {
	//remove head node
	if c.head == n {
		if n.next != nil {
			n.next.prev = nil
		}
		c.head = n.next
		return
	}

	//remove tail node
	if c.tail == n {
		c.tail = c.tail.prev
		if c.tail != nil {
			c.tail.next = nil
		}
		return
	}

	n.prev.next = n.next
	n.next.prev = n.prev
}

// clear expired node from tail
func (c *linkedLRUCache) startClearExpired() {
	expiredDur := 20 * time.Minute
	stop := make(chan struct{})
	toDelete := make([]*node, 0)
	go wait.Until(func() {
		c.Lock()
		defer c.Unlock()

		now := time.Now()
		for curNode := c.tail; curNode != c.head; curNode = curNode.prev {
			if now.Sub(curNode.lastAccess) > expiredDur {
				toDelete = append(toDelete, curNode)
			}
		}

		for idx, _ := range toDelete {
			delete(c.cache, toDelete[idx].key)
			c.remove(toDelete[idx])
		}

	}, 20*time.Second, stop)

}
