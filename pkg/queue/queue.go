package queue

import (
	"github.com/prometheus/client_golang/prometheus"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	TotalEventsDroppedCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lunettes_total_events_dropped_count",
			Help: "total events dropped so far",
		},
		[]string{"name"},
	)
	inputQueueSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lunettes_input_queue_size",
			Help: "new received but not processed events in input queue",
		},
		// queue name
		[]string{"name"},
	)
)

type BoundedQueue struct {
	name                string
	capacity            int
	size                int32
	onDroppedItem       func(item interface{})
	filterItemOnProduce func(item interface{}) bool // true: item will be filtered
	itemsCh             chan interface{}
	stopCh              chan struct{}
	stopWG              sync.WaitGroup
	stopped             int32
	IsLockOSThread      bool
	IsDropEventOnFull   bool
}

// NewBoundedQueue constructs the new queue of specified capacity, and with an optional
// callback for dropped items (e.g. useful to emit metrics).
func NewBoundedQueue(name string, capacity int, onDroppedItem func(item interface{})) *BoundedQueue {
	return &BoundedQueue{
		name:              name,
		capacity:          capacity,
		onDroppedItem:     onDroppedItem,
		itemsCh:           make(chan interface{}, capacity),
		stopCh:            make(chan struct{}),
		IsLockOSThread:    false,
		IsDropEventOnFull: true,
	}
}

// StartConsumers starts a given number of goroutines consuming items from the queue
// and passing them into the consumer callback.
func (q *BoundedQueue) StartConsumers(num int, consumer func(item interface{})) {
	var startWG sync.WaitGroup
	for i := 0; i < num; i++ {
		q.stopWG.Add(1)
		startWG.Add(1)
		go func() {
			startWG.Done()
			defer q.stopWG.Done()
			if q.IsLockOSThread {
				runtime.LockOSThread()
				defer runtime.UnlockOSThread()
			}
			for {
				select {
				case item := <-q.itemsCh:
					atomic.AddInt32(&q.size, -1)
					consumer(item)
				case <-q.stopCh:
					return
				}
			}
		}()
	}
	startWG.Wait()
}

// Produce is used by the producer to submit new item to the queue. Returns false in case of queue overflow.
func (q *BoundedQueue) Produce(item interface{}) bool {
	if atomic.LoadInt32(&q.stopped) != 0 {
		TotalEventsDroppedCount.WithLabelValues(q.name).Inc()
		if q.onDroppedItem != nil {
			q.onDroppedItem(item)
		}
		return false
	}

	if q.filterItemOnProduce != nil && q.filterItemOnProduce(item) {
		return true
	}

	if q.IsDropEventOnFull {
		select {
		case q.itemsCh <- item:
			atomic.AddInt32(&q.size, 1)
			return true
		default:
			TotalEventsDroppedCount.WithLabelValues(q.name).Inc()
			if q.onDroppedItem != nil {
				q.onDroppedItem(item)
			}
			return false
		}
	} else {
		select {
		case q.itemsCh <- item:
			atomic.AddInt32(&q.size, 1)
			return true
		}
	}
}

// Stop stops all consumers, as well as the length reporter if started,
// and releases the items channel. It blocks until all consumers have stopped.
func (q *BoundedQueue) Stop() {
	atomic.StoreInt32(&q.stopped, 1) // disable producer
	close(q.stopCh)
	q.stopWG.Wait()
	close(q.itemsCh)
}

// Size returns the current size of the queue
func (q *BoundedQueue) Size() int {
	return int(atomic.LoadInt32(&q.size))
}

// Capacity returns capacity of the queue
func (q *BoundedQueue) Capacity() int {
	return q.capacity
}

// StartLengthReporting starts a timer-based gorouting that periodically reports
// current queue length to a given metrics gauge.
func (q *BoundedQueue) StartLengthReporting(reportPeriod time.Duration) {
	ticker := time.NewTicker(reportPeriod)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				size := q.Size()
				inputQueueSize.WithLabelValues(q.name).Set(float64(size))
			case <-q.stopCh:
				return
			}
		}
	}()
}

func (q *BoundedQueue) SetFilterItemFunc(filter func(item interface{}) bool) {
	q.filterItemOnProduce = filter
}

func init() {
	prometheus.MustRegister(inputQueueSize)
	prometheus.MustRegister(TotalEventsDroppedCount)
}
