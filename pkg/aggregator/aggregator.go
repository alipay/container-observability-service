package aggregator

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/featuregates"
	"github.com/alipay/container-observability-service/pkg/spans"
	"github.com/alipay/container-observability-service/pkg/trace"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"

	"github.com/alipay/container-observability-service/pkg/replayer"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"

	_ "github.com/alipay/container-observability-service/pkg/shares/base_processor"
	_ "github.com/alipay/container-observability-service/pkg/shares/extractor"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	k8scache "k8s.io/client-go/tools/cache"
)

var (
	OngoingTrace = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lunettes_ongoing_trace",
			Help: "traces ongoing(not completed yet)",
		},
		[]string{"trace_type"},
	)
)

type AggregatorOptions struct {
	MetricsAddr                 string
	Workers                     int
	QPS                         float32
	Burst                       int
	TraceTimeout                time.Duration
	KubeConfigFile              string
	Cluster                     string
	JaegerCollector             string
	OTLPCollector               string
	ElasticSearchEndpoint       string
	ElasticSearchUser           string
	ElasticSearchPassword       string
	ElasticSearchIndexName      string
	ElasticSearchBufferDuration time.Duration
	ElasticSearchFetchInterval  time.Duration
	FeatureGates                string
	APIServerEnabled            bool
	APIServerListenAddr         string
	APIServerTraceStatsIndex    string
	EnableTrace                 bool
}

type auditReplayer interface {
	Start(stopCh <-chan struct{})
	Stop()
}

type Aggregator struct {
	options *AggregatorOptions
	// for k8s resources
	kubeClient      clientset.Interface
	podLister       corelisters.PodLister
	podListerSynced k8scache.InformerSynced
	informerFactory informers.SharedInformerFactory
	// the rest config for the master
	kubeConfig *restclient.Config
	esConfig   *xsearch.ElasticSearchConf

	replayer auditReplayer
}

func NewAggregator(options *AggregatorOptions) (*Aggregator, error) {

	aggregator := &Aggregator{
		options: options,
	}

	var kubeClient clientset.Interface
	kubeClient, err := utils.GetClientFromIncluster(options.QPS, options.Burst)
	if err != nil {
		return nil, err
	}

	sharedInformerFactory := informers.NewSharedInformerFactory(kubeClient, 10*time.Second)
	podInformer := sharedInformerFactory.Core().V1().Pods()

	aggregator.kubeClient = kubeClient
	aggregator.informerFactory = sharedInformerFactory
	aggregator.podLister = podInformer.Lister()
	aggregator.podListerSynced = podInformer.Informer().HasSynced

	esConf := &xsearch.ElasticSearchConf{
		Endpoint: options.ElasticSearchEndpoint,
		User:     options.ElasticSearchUser,
		Password: options.ElasticSearchPassword,
		Index:    options.ElasticSearchIndexName,
	}
	aggregator.esConfig = esConf

	auditProcessor, err := replayer.NewAuditProcessor(
		esConf,
		options.ElasticSearchBufferDuration, options.ElasticSearchFetchInterval,
		options.TraceTimeout, options.Cluster, options.EnableTrace)
	if err != nil {
		return nil, err
	}
	aggregator.replayer = auditProcessor

	if featuregates.IsEnabled(spans.SpanAnalysisFeature) {
		err = spans.InitKubeSpanWatcher(options.Cluster, options.JaegerCollector)
		if err != nil {
			return nil, err
		}
	}

	if featuregates.IsEnabled(trace.TraceFeature) {
		err = trace.InitKubeTraceWatcher(options.Cluster, options.OTLPCollector)
		if err != nil {
			return nil, err
		}
	}

	return aggregator, err
}

func (a *Aggregator) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()

	a.replayer.Start(stopCh)
	klog.Infof("replayer has started")
	<-stopCh

	// close processing queue
	a.replayer.Stop()

	klog.Warning("aggregator is existing.")
	return nil
}

func init() {
	prometheus.MustRegister(OngoingTrace)
}
