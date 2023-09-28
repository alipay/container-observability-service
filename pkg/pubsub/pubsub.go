package pubsub

import (
	"fmt"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
)

// Subscriber is interface that who want to consume message
type Subscriber interface {
	// should be unique to used as a map key
	SubscriberName() string
	// receive message from publisher
	OnMessage(v interface{})
}

// PubSub contains channel and subscribers.
type PubSub struct {
	messageChan chan interface{}
	subscribers map[string]Subscriber
	locker      sync.Mutex
	e           chan error
	closed      bool
}

// New return new PubSub intreface.
func New(buffer int) (*PubSub, error) {

	if buffer < 1 {
		return nil, fmt.Errorf("message(%d) buffer should be greater than 0", buffer)
	}
	ps := new(PubSub)
	ps.messageChan = make(chan interface{}, buffer)
	ps.subscribers = make(map[string]Subscriber)
	go func() {
		for v := range ps.messageChan {
			ps.locker.Lock()
			t := time.Now()
			for _, subscriber := range ps.subscribers {
				go subscriber.OnMessage(v)
			}
			ps.locker.Unlock()
			metrics.DebugMethodDurationMilliSeconds.WithLabelValues("pubsub.publish").Observe(utils.TimeSinceInMilliSeconds(t))
		}
	}()
	return ps, nil
}

// Subscribe subscribe to the PubSub.
func (ps *PubSub) Subscribe(s Subscriber) error {
	ps.locker.Lock()
	defer ps.locker.Unlock()
	if _, found := ps.subscribers[s.SubscriberName()]; found {
		return fmt.Errorf("subscriber %s already exists", s.SubscriberName())
	}
	ps.subscribers[s.SubscriberName()] = s
	return nil
}

// UnSubscribe un subscribe to the PubSub.
func (ps *PubSub) UnSubscribe(s Subscriber) {
	ps.locker.Lock()
	defer ps.locker.Unlock()
	delete(ps.subscribers, s.SubscriberName())
}

// Publish will publish message to subscribers
func (ps *PubSub) Publish(v interface{}) error {
	if ps.closed {
		return fmt.Errorf("PubSub has already closed")
	}
	ps.messageChan <- v
	return nil
}

// Close closes PubSub. To inspect unbsubscribing for another subscruber, you must create message structure to notify them. After publish notifycations, Close should be called.
func (ps *PubSub) Close() error {
	ps.locker.Lock()
	defer ps.locker.Unlock()
	if ps.closed == true {
		return fmt.Errorf("PubSub has already closed")
	}
	ps.closed = true
	close(ps.messageChan)
	ps.subscribers = nil
	return nil
}
