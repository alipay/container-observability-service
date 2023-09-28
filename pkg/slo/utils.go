package slo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alipay/container-observability-service/pkg/queue"
	"github.com/olivere/elastic"
	"golang.org/x/sync/errgroup"
	k8s_audit "k8s.io/apiserver/pkg/apis/audit"

	"github.com/alipay/container-observability-service/pkg/utils"

	"k8s.io/klog/v2"

	v1 "k8s.io/api/core/v1"
)

const (
	schedulingStrategyGPU        = "gpu"
	schedulingStrategyColocation = "colocation"
	resourceGPUKey               = "nvidia.com/gpu"
	resourceCPUKey               = "cpu"

	//action type
	RequestCreate  = "RequestCreate"
	RequestUpgrade = "RequestUpgrade"
	RequestStart   = "RequestStart"
	RequestStop    = "RequestStop"
	RequestDelete  = "RequestDelete"

	// Annotation request-action, operation for pod

	// Annotation delivery hosting status
)

type RequestAction struct {
	ActionType string   `json:"action-type,omitempty"`
	Containers []string `json:"containers,omitempty"`
	Timestamp  string   `json:"timestamp,omitempty"`
}

func getOwnerRefStr(pod *v1.Pod) string {
	if pod == nil || len(pod.OwnerReferences) == 0 {
		return ""
	}

	ownerKinds := make([]string, 0)
	ownerRefs := pod.OwnerReferences
	for _, ownerref := range ownerRefs {
		ownerKinds = append(ownerKinds, ownerref.Kind)
	}

	return strings.Join(ownerKinds, ",")
}

func genPodKey(clusterName, podNamespace, podName string) string {
	return fmt.Sprintf("%s/%s/%s", clusterName, podNamespace, podName)
}

func getSchedulingStrategyAndCores(pod *v1.Pod) (string, int64) {
	var cores int64
	ss := schedulingStrategyColocation
	for _, c := range pod.Spec.Containers {
		// only check requests only.
		m, err := r2m(c.Resources.Requests)
		if err != nil {
			klog.Errorf("failed to parse SchedulingStrategy: %s", err.Error())
			continue
		}

		gpu := m[resourceGPUKey]
		if gpu != "" {
			ss = schedulingStrategyGPU
		}
		cpu := m[resourceCPUKey]
		if cpu != "" {
			m := false
			if strings.Index(cpu, "m") > 0 {
				cpu = strings.Replace(cpu, "m", "", -1)
				m = true
			}

			ic, err := strconv.Atoi(cpu)
			if err != nil {
				klog.Errorf("failed parse cpu to int(%s): %s", cpu, err.Error())
			} else if ic > 0 {
				if m {
					cores += int64(ic / 1000)
				} else {
					cores += int64(ic)
				}
			}
		}
	}

	return ss, cores
}

// r2m convert resource.request/limit to map[string]string
func r2m(r v1.ResourceList) (map[string]string, error) {
	bs, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	err = json.Unmarshal(bs, &m)
	return m, err
}

func isPodWithPVC(pod *v1.Pod) bool {
	if pod == nil {
		return false
	}
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			return true
		}
	}
	return false
}

func getInitContainerImage(pod *v1.Pod) string {
	if pod == nil {
		return ""
	}

	if len(pod.Spec.InitContainers) <= 0 {
		return ""
	}

	return pod.Spec.InitContainers[0].Image
}

// start point of code for unit test
const (
	es_endpoint = "ES_ENDPOINT"
	es_username = "ES_USERNAME"
	es_password = "ES_PASSWORD"
	es_index    = "ES_INDEX"
)

// es reader for unit test
type ESReader interface {
	// eareader read k8s audit event, then send to consumer queue
	SetConsumer(int, *queue.BoundedQueue)
	// starts to read audit events from es
	Run(<-chan struct{})
	// return a chan implies whether reader has done its job
	WaitChan() <-chan struct{}
	// get total processed item
	GetProcessed() int64
}

func NewESReader() (ESReader, error) {
	startTime := time.Unix(0, 0)
	endTime := time.Now()
	return newSimpleZSearchReader(startTime, endTime)
}

func getEsConfItemFromEnv(name string) string {
	value := os.Getenv(name)
	if len(value) == 0 {
		klog.Fatalf("cannot get es config %s", name)
	}
	return value
}

func newEsClient() (*elastic.Client, error) {
	esEndPoint := getEsConfItemFromEnv(es_endpoint)
	esUserName := getEsConfItemFromEnv(es_username)
	esPassWord := getEsConfItemFromEnv(es_password)
	client, err := elastic.NewClient(elastic.SetURL(esEndPoint),
		elastic.SetBasicAuth(esUserName, esPassWord),
		elastic.SetSniff(false))
	if err != nil {
		return nil, err
	}
	return client, err
}

func newSimpleZSearchReader(startTime, endTime time.Time) (*SimpleZSearchReader, error) {
	esclient, err := newEsClient()
	if err != nil {
		return nil, err
	}
	esIndex := getEsConfItemFromEnv(es_index)
	return &SimpleZSearchReader{
		producer:    queue.NewBoundedQueue("slo-watcher", 1000000, nil),
		hasConsumer: false,
		index:       esIndex,
		esClient:    esclient,
		endTime:     endTime,
		startTime:   startTime,
	}, nil
}

// SimpleZSearchReader read from readonly index, and produce audit events to target queue.
type SimpleZSearchReader struct {
	producer       *queue.BoundedQueue
	hasConsumer    bool
	index          string
	esClient       *elastic.Client
	g              *errgroup.Group
	ctx            context.Context
	totalProcessed int64
	endTime        time.Time
	startTime      time.Time
	waitMutex      sync.Mutex
	waitChan       chan struct{}
}

func (r *SimpleZSearchReader) SetConsumer(num int, cq *queue.BoundedQueue) {
	go r.producer.StartConsumers(num, func(item interface{}) {
		event, ok := item.(*k8s_audit.Event)
		if !ok {
			return
		}
		cq.Produce(event)
	})
	r.hasConsumer = true
}

func (r *SimpleZSearchReader) Run(stop <-chan struct{}) {
	if !r.hasConsumer {
		klog.Fatal("no consumer for ESReader")
	}
	endTime := r.endTime
	startTime := r.startTime

	rangeQuery := elastic.NewRangeQuery("stageTimestamp").
		From(startTime).To(endTime).
		IncludeLower(true).
		IncludeUpper(false).
		TimeZone("UTC")
	query := elastic.NewBoolQuery().Must(rangeQuery)

	s, _ := query.Source()
	klog.V(2).Infof("query: %s", utils.Dumps(s))

	pageSize := 1000
	totalGot := 0
	hits := make(chan json.RawMessage)

	cont, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-stop:
			cancel()
		}
	}()
	r.g, r.ctx = errgroup.WithContext(cont)

	var globalErr error
	r.g.Go(func() error {
		defer close(hits)
		// Initialize scroller. Just don't call Do yet.
		scroll := r.esClient.Scroll(r.index).Query(query).Sort("stageTimestamp", true).Size(pageSize)

		for {
			results, err := scroll.Do(r.ctx)
			if err == io.EOF {
				// all results retrieved
				return nil
			}
			if err != nil {
				globalErr = err
				klog.Errorf("scroll.Do failed: %s", err.Error())
				return err
			}

			// Send the hits to the hits channel
			totalGot += len(results.Hits.Hits)
			for _, hit := range results.Hits.Hits {
				select {
				case hits <- *hit.Source:
				case <-r.ctx.Done():
					return r.ctx.Err()
				}
			}
		}
	})

	// 2nd goroutine receives hits and deserializes them.
	r.g.Go(func() error {
		for hit := range hits {
			var event k8s_audit.Event
			err := json.Unmarshal(hit, &event)
			if err != nil {
				klog.Errorf("failed unmarshal %s", err.Error())
				continue
			}

			klog.V(8).Infof("got event %s", utils.DumpsEventKeyInfo(&event))

			r.producer.Produce(&event)
			atomic.AddInt64(&r.totalProcessed, 1)
			// Terminate early?
			select {
			default:
			case <-r.ctx.Done():
				return r.ctx.Err()
			}
		}
		return nil
	})

	// Check whether any goroutines failed.
	if err := r.g.Wait(); err != nil {
		globalErr = err
	}
	if globalErr != nil {
		klog.Errorf("failed to wait all goroutines finish when scroll elasticsearch result: %s", globalErr.Error())
	}
	klog.V(2).Infof("fetch from %s, range %v, got %d audit log", startTime.Format(time.RFC3339), endTime.Sub(startTime), r.totalProcessed)
}

func (r *SimpleZSearchReader) WaitChan() <-chan struct{} {
	if r.g == nil {
		return nil
	}
	r.waitMutex.Lock()
	defer r.waitMutex.Unlock()
	if r.waitChan == nil {
		r.waitChan = make(chan struct{})
		go func() {
			r.g.Wait()
			defer close(r.waitChan)
		}()
	}
	return r.waitChan
}

func (r *SimpleZSearchReader) GetProcessed() int64 {
	return r.totalProcessed
}
