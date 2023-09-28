package pubsub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type subscriber struct {
	name string
	v    time.Time
}

func (s *subscriber) SubscriberName() string {
	return s.name
}

func (s *subscriber) OnMessage(v interface{}) {
	t, ok := v.(time.Time)
	if ok {
		s.v = t
	}
}

var _ Subscriber = &subscriber{}

func TestPubSubFailed(t *testing.T) {
	ps, err := New(-1)
	assert.NotNil(t, err, "error should not be nil")
	assert.Nil(t, ps, "ps should be nil")
}

func TestPubSub(t *testing.T) {
	ps, err := New(100)
	assert.Nil(t, err, "error should be nil")
	assert.NotNil(t, ps, "ps should not be nil")

	s1 := &subscriber{
		name: "s1",
	}
	s2 := &subscriber{
		name: "s2",
	}

	err = ps.Subscribe(s1)
	assert.Nil(t, err, "error should be nil")

	err = ps.Subscribe(s2)
	assert.Nil(t, err, "error should be nil")

	err = ps.Subscribe(s2)
	assert.NotNil(t, err, "error should not be nil")

	// publish message 1
	t1 := time.Now()
	err = ps.Publish(t1)
	assert.Nil(t, err, "error should be nil for publish message 1")
	// sleep to wait all subscriber received the message
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, t1, s1.v, "s1.v should be %+v", t1)
	assert.Equal(t, t1, s2.v, "s2.v should be %+v", t1)

	// unsubscribe s2
	ps.UnSubscribe(s2)

	// publish message 2
	t2 := time.Now()
	ps.Publish(t2)

	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, t2, s1.v, "s1.v should be %+v", t2)
	assert.Equal(t, t1, s2.v, "s2.v should be %+v", t1)

	// close pubsub server
	err = ps.Close()
	assert.Nil(t, err, "error should be nil for close PubSub server first time")

	err = ps.Publish(t2)
	assert.NotNil(t, err, "error should not be nil for send message to a closed PubSub server")

	err = ps.Close()
	assert.NotNil(t, err, "error should not be nil for close PubSub server second time")
}
