package spans

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alipay/container-observability-service/pkg/common"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/shares"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	"k8s.io/klog/v2"
)

var (
	lunettesNs            = common.LunettesNs
	lunettesConfigMapName = "lunettes-config"
	kubeconfigPath        = "/etc/kubernetes/kubeconfig/admin.kubeconfig"
)

type SpanProcessor struct {
	Cluster   string
	writer    Writer
	config    atomic.Value
	SpanMetas *sync.Map
	now       time.Time
}

func (p *SpanProcessor) cleanMaps() {
	klog.V(5).Infof("clean up pods in processor for cluster %s", p.Cluster)
	now := p.now
	staleSpanUids := make([]interface{}, 0, 100)
	var inMemPods int = 0
	p.SpanMetas.Range(func(key, value interface{}) bool {
		inMemPods += 1
		spanMetaList := value.([]*SpanMeta)
		for idx, _ := range spanMetaList {
			spanMeta := spanMetaList[idx]
			if spanMeta != nil {
				if now.Sub(spanMeta.CreationTimestamp) > time.Minute*35 {
					staleSpanUids = append(staleSpanUids, spanMeta)
				}
			}
		}

		return true
	})
	for _, val := range staleSpanUids {
		go p.finishSpan(val.(*SpanMeta))
	}
	klog.V(5).Infof("clean up span in processor for cluster %s finished", p.Cluster)
}

func (p *SpanProcessor) Compact() {
	wait.Forever(func() {
		defer HandleCrash()
		p.cleanMaps()
	}, time.Second*30)
}

// this function should be thread safe
func (p *SpanProcessor) ProcessEvent(ev *shares.AuditEvent) {
	defer func() {
		HandleCrash()
		ev.FinishProcess(shares.SpanProcessNode)
	}()
	p.now = ev.RequestReceivedTimestamp.Time
	if ev.ObjectRef == nil {
		return
	}

	if ev.ResponseRuntimeObj == nil || ev.ResponseStatus.Code >= 300 {
		return
	}

	conf := p.getConfig()
	if conf == nil {
		return
	}
	rConfigs := conf.GetConfigByRef(ev.ObjectRef)
	for idx, _ := range rConfigs {
		if rConfigs[idx] != nil && rConfigs[idx].IsStartToTrack(ev) {
			spanMeta := NewSpanMeta(rConfigs[idx], p.Cluster, p.now, ev)

			spanMetaUID := string(spanMeta.ObjectRef.UID)
			tmpList, ok := p.SpanMetas.Load(spanMetaUID)
			if !ok || tmpList == nil {
				p.SpanMetas.Store(spanMetaUID, make([]*SpanMeta, 0))
			}

			tmpList, _ = p.SpanMetas.Load(spanMetaUID)
			spanMetaList, ok := tmpList.([]*SpanMeta)
			if !ok || spanMetaList == nil {
				continue
			}
			spanMetaList = append(spanMetaList, spanMeta)
			p.SpanMetas.Store(spanMetaUID, spanMetaList)
			metrics.SpansInMemPodsCount.WithLabelValues().Inc()
			//fmt.Printf("Store span meta list for %s, ActionType:%s\n", spanMetaUID, spanMeta.config.ActionType)
		}
	}

	go func() {
		defer ev.FinishProcess(shares.SpanProcessNode)

		uid, err := ev.GetObjectUID()
		if err != nil {
			return
		}

		tmpList, ok := p.SpanMetas.Load(string(uid))
		if !ok || tmpList == nil {
			return
		}
		spanMetaList, ok := tmpList.([]*SpanMeta)
		if !ok || spanMetaList == nil {
			return
		}

		for idx, _ := range spanMetaList {
			spanMeta := spanMetaList[idx]
			if spanMeta == nil {
				continue
			}

			spanMeta.TrackSpan(ev)
			if spanMeta.config.IsFinishToTrack(ev) {
				p.finishSpan(spanMeta)
			}
		}
	}()

}

func (p *SpanProcessor) finishSpan(spanMata *SpanMeta) {
	if _, ok := p.SpanMetas.Load(string(spanMata.ObjectRef.UID)); !ok {
		return
	}

	spanMata.mutex.Lock()
	defer spanMata.mutex.Unlock()

	if _, ok := p.SpanMetas.Load(string(spanMata.ObjectRef.UID)); !ok {
		return
	}

	spanMata.finishOpenSpanNow(p.now)
	klog.Infof("finish to track: %s", spanMata.ObjectRef.UID)
	err := p.writer.Write(spanMata)
	if err != nil {
		klog.Errorf("Finish span error, msg:%s", err.Error())
	}
	p.SpanMetas.Delete(string(spanMata.ObjectRef.UID))
	metrics.SpansInMemPodsCount.WithLabelValues().Desc()
	metrics.SpansProcessedPods.WithLabelValues().Desc()
}

func (p *SpanProcessor) getConfig() *ResourceSpanConfigList {
	if p.config.Load() == nil {
		return nil
	}
	rs, ok := p.config.Load().(*ResourceSpanConfigList)
	if !ok || rs == nil {
		return nil
	}
	return p.config.Load().(*ResourceSpanConfigList)
}

func (p *SpanProcessor) RefreshConfig() {
	cfg, err := restclient.InClusterConfig()
	if err != nil {
		klog.Errorf("failed to build config, err is %v", err)
		return
	}

	cfg.UserAgent = "lunettes"
	cs, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("failed to create clientSet: %v", err)
		return
	}

	refreshConfigMap := func() {
		lunettesConfigMap, err := cs.CoreV1().ConfigMaps(lunettesNs).Get(context.TODO(), lunettesConfigMapName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("failed to get span configmap: %v", err)
			return
		}

		spanStr := lunettesConfigMap.Data["span-config"]
		if spanStr == "" {
			klog.Errorf("span configmap data is empty")
			return
		}
		fmt.Printf("spanStr:%s \n", spanStr)

		var tmpConfig = NewResourceSpanConfigList()
		err = json.Unmarshal([]byte(spanStr), &tmpConfig)
		if err != nil {
			klog.Errorf("failed to unmarshal span configmap: %v", err)
			return
		}

		p.config.Store(&tmpConfig)
	}

	refreshConfigMap()
	stop := make(chan struct{})
	go wait.JitterUntil(refreshConfigMap, 60*time.Second, 0.0, true, stop)
}

func NewSpanProcessor(cluster string) *SpanProcessor {
	w, err := NewXSearchWriter()
	if err != nil {
		klog.Error(err)
	}

	p := &SpanProcessor{
		SpanMetas: &sync.Map{},
		Cluster:   cluster,
		writer:    w,
	}

	p.RefreshConfig()
	//define the metrics
	conf := p.getConfig()
	if conf != nil {
		metrics.DefineSpanStatistic(conf.GetExtraPropertyNames())
	}
	return p
}

func HandleCrash() {
	if r := recover(); r != nil {
		logPanic(r)
	}
}

func logPanic(r interface{}) {
	callers := getCallers(r)
	klog.Errorf("Observed a panic: %#v (%v)\n%v", r, r, callers)
}

func getCallers(r interface{}) string {
	callers := ""
	for i := 0; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		callers = callers + fmt.Sprintf("%v:%v\n", file, line)
	}
	return callers
}
