package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBoundedQueue(t *testing.T) {
	dropedCount := 0
	queue := NewBoundedQueue("test-queue", 10, func(item interface{}) {
		dropedCount++
	})

	for i := 0; i < 15; i++ {
		queue.Produce(i)
	}
	assert.Equal(t, 5, dropedCount)
	assert.Equal(t, 10, queue.Size())
	assert.Equal(t, 10, queue.Capacity())

	receivedCount := 0
	queue.StartConsumers(1, func(item interface{}) {
		receivedCount++
	})
	time.Sleep(1)
	queue.Stop()
	assert.Equal(t, 10, receivedCount)
}
