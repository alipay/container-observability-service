package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alipay/container-observability-service/pkg/metas"
	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"
)

type WatchParam struct {
	deliveryType string
}

var upgrader = websocket.Upgrader{} // use default options

func parseParam(request *http.Request) *WatchParam {
	watchParam := &WatchParam{}
	setSP(request.URL.Query(), "type", &watchParam.deliveryType)

	return watchParam
}

func validateParam(param *WatchParam) error {
	if param == nil || param.deliveryType != metas.PodCreateSLO && param.deliveryType != metas.PodDeleteSLO {
		err := fmt.Errorf("watch param can not be nil, must be one of [%s, %s]", metas.PodCreateSLO, metas.PodDeleteSLO)
		return err
	}

	return nil
}

func watch(w http.ResponseWriter, r *http.Request) {
	param := parseParam(r)
	if err := validateParam(param); err != nil {
		msg := fmt.Sprintf("failed to validatate watch delivery type, %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}

	infoType := param.deliveryType

	publisher := metas.GetPubLister(infoType)
	if publisher == nil {
		msg := fmt.Sprintf("failed to new watcher, can not find publisher by type [%s]", infoType)
		klog.Errorf(msg)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}
	watcher := metas.NewWatcher(fmt.Sprintf("%s_%s", infoType, time.Now().Format(time.RFC3339Nano)), w, r)
	err := publisher.Subscribe(watcher)
	if err != nil {
		msg := fmt.Sprintf("failed to new watcher, err: %v", err)
		klog.Errorf(msg)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}
	defer func() {
		watcher.Stop()
		publisher.UnSubscribe(watcher)
	}()

	for {
		select {
		case e := <-watcher.ErrorChan():
			klog.Errorf("error when watch, err: %v", e)
			return
		}
	}
}
