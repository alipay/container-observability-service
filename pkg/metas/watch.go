package metas

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/alipay/container-observability-service/pkg/pubsub"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"
)

const (
	PodCreateSLO = "pod_create_slo"
	PodDeleteSLO = "pod_delete_slo"
)

var DeliveryWatchers *utils.SafeMap = utils.NewSafeMap()

func RegisterPublisher(name string) {
	if _, ok := DeliveryWatchers.Get(name); !ok {
		pb, err := pubsub.New(1000)
		if err != nil {
			klog.Errorf("cant not register publisher %s, err: %s", name, err.Error())
		}
		DeliveryWatchers.Set(name, pb)
	}
}

func GetPubLister(name string) *pubsub.PubSub {
	if pb, ok := DeliveryWatchers.Get(name); ok {
		pubSub, ok := pb.(*pubsub.PubSub)
		if ok {
			return pubSub
		}
	}
	return nil
}

type WatchSubscriber struct {
	name       string
	upper      *websocket.Upgrader
	connection *websocket.Conn
	errorChan  chan error
	conLock    sync.Mutex
}

func NewWatcher(name string, w http.ResponseWriter, r *http.Request) *WatchSubscriber {
	up := &websocket.Upgrader{} // use default options
	c, err := up.Upgrade(w, r, nil)
	if err != nil {
		klog.Errorf("new websocket upgrader error: %v", err)
		return nil
	}

	watcher := &WatchSubscriber{
		name:       name,
		upper:      up,
		connection: c,
		errorChan:  make(chan error, 100),
		conLock:    sync.Mutex{},
	}

	return watcher
}

func (c *WatchSubscriber) SubscriberName() string {
	return c.name
}

func (c *WatchSubscriber) OnMessage(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		klog.Errorf("watch publish error when json marshal, err: %v", err)
		c.errorChan <- err
	}

	c.conLock.Lock()
	err = c.connection.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s", b)))
	c.conLock.Unlock()

	if err != nil {
		klog.Errorf("watch publish write error: %v", err)
		c.errorChan <- err
	}
}

func (c *WatchSubscriber) ErrorChan() chan error {
	return c.errorChan
}

func (c *WatchSubscriber) Stop() {
	//清空chan
	for len(c.errorChan) > 0 {
		<-c.errorChan
	}

	c.connection.Close()
}
