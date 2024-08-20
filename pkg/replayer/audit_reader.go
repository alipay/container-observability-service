package replayer

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/alipay/container-observability-service/pkg/config"
	"github.com/alipay/container-observability-service/pkg/featuregates"
	"github.com/alipay/container-observability-service/pkg/nodeyaml"
	"github.com/alipay/container-observability-service/pkg/podphase"
	"github.com/alipay/container-observability-service/pkg/podyaml"
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/slo"
	"github.com/alipay/container-observability-service/pkg/spans"
	"github.com/alipay/container-observability-service/pkg/trace"

	"io"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"

	"github.com/olivere/elastic/v7"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type metaConf struct {
	LastReadTime string
}

const (
	metricNamePrefix = "lunettes_"
	metaIndexName    = "lunettes_meta"
	metaTypeName     = "meta"
	metaMapping      = `
	{
		"mappings": {
			"properties": {
				"LastReadTime": {
					"type": "long"
				}
			}
		}
	}`
)

var (
	xsearchFetedEventDurationMilliSeconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: metricNamePrefix + "xsearch_scroll_duration_milliseconds",
			Help: "how long an xsearch scroll operation to completed",
		},
	)
	xsearchSLOProDurationInMilliSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricNamePrefix + "xsearch_process_duration_milliseconds",
			Help: "time lag since the last fetch operation",
		},
		[]string{"phase"},
	)

	xsearchFetchedEventCountOneScroll = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: metricNamePrefix + "xsearch_fetched_events_one_scroll",
			Help: "how many events fetch at one scroll operation",
		},
	)

	xsearchFetchLagInMilliSeconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: metricNamePrefix + "xsearch_fetch_lag_milliseconds",
			Help: "time lag since the last fetch operation",
		},
	)

	xsearchFetchLagInMilliSecondsSum = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    metricNamePrefix + "xsearch_fetch_lag_seconds",
			Help:    "time lag since the last fetch operation",
			Buckets: []float64{0.5, 1, 2, 4, 6, 8, 10, 14, 18, 22, 26, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 90, 100, 110, 120, 130, 140, 150, 200, 250, 300, 400, 500, 600, 900, 1200},
		},
		[]string{"type"},
	)

	xsearchQueryErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: metricNamePrefix + "xsearch_query_errors",
			Help: "service operation errors",
		},
	)
)

var (
	numSlices        = 4
	queryDocNum      = 80
	lastNoEventError = false

	currentIndex = atomic.Value{}
)

func init() {
	prometheus.MustRegister(xsearchFetedEventDurationMilliSeconds)
	prometheus.MustRegister(xsearchFetchedEventCountOneScroll)
	prometheus.MustRegister(xsearchFetchLagInMilliSeconds)
	prometheus.MustRegister(xsearchFetchLagInMilliSecondsSum)
	prometheus.MustRegister(xsearchQueryErrors)
	prometheus.MustRegister(xsearchSLOProDurationInMilliSeconds)

	flag.IntVar(&numSlices, "num_scroll_slice", 4, "number of scroll slice")
	flag.IntVar(&queryDocNum, "doc_num_peer_query", 80, "enable biz process")

	currentIndex.Store("")
}

type logReader struct {
	esConf                *xsearch.ElasticSearchConf
	lastReadTime          time.Time
	auditProcessor        *AuditProcessor
	bufferDuration        time.Duration
	fetchIntervalDuration time.Duration
	cluster               string
	esClient              *elastic.Client
	lastReadTimeChan      chan time.Time
}

// todo
func newLogReader(processor *AuditProcessor,
	conf *xsearch.ElasticSearchConf,
	buffer, interval time.Duration, cluster string) (*logReader, error) {
	lr := &logReader{
		esConf:                conf,
		auditProcessor:        processor,
		bufferDuration:        buffer,
		fetchIntervalDuration: interval,
		cluster:               cluster,
	}

	//log := log.New(os.Stdout, "INFO ", log.Ltime|log.Lshortfile)
	client, err := elastic.NewClient(elastic.SetURL(lr.esConf.Endpoint),
		elastic.SetBasicAuth(lr.esConf.User, lr.esConf.Password),
		elastic.SetSniff(false),
		elastic.SetDecoder(&utils.JsoniterDecoder{}),
		//elastic.SetInfoLog(log)
	)
	if err != nil {
		return nil, err
	}

	lr.esClient = client
	lr.refreshIndex()
	return lr, nil
}

func (lr *logReader) Stop() {
	klog.Info("Stop log reader")
	lr.updateLastReadTime()
	klog.Info("Stop log reader completed")
}

func (lr *logReader) initMetaIndex() error {
	return xsearch.EnsureIndex(lr.esClient, metaIndexName, metaMapping)
}

func (lr *logReader) Run(stopCh <-chan struct{}) {
	klog.Infof("elasticsearch client is configured")

	if err := lr.initMetaIndex(); err != nil {
		panic(err)
	}

	// get last read time
	lr.lastReadTime = lr.getLastReadTime()
	klog.Infof("last read time from xsearch %+v", lr.lastReadTime)

	// only get data older than this duration, wait for not stored new event to be inserted
	fetchedUntil := time.Now().Add(-lr.bufferDuration)
	_, err := lr.recovery(fetchedUntil, stopCh)

	maxStepDuration := time.Second * 60
	if lr.fetchIntervalDuration > maxStepDuration {
		maxStepDuration = lr.fetchIntervalDuration + time.Second*10
	}

	if err == nil {
		lr.lastReadTime = fetchedUntil
	}

	// flow control window
	flowController := utils.NewFlowControl(20, 5, 0.5, 1.5, maxStepDuration.Seconds())
	var startTime time.Time

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		//isLastErr := false
		wait.Until(func() {
			startTime = time.Now()
			xsearchFetchLagInMilliSeconds.Set(utils.TimeDiffInMilliSeconds(lr.lastReadTime, startTime))
			xsearchFetchLagInMilliSecondsSum.WithLabelValues("audit_read").Observe(utils.TimeDiffInSeconds(lr.lastReadTime, startTime))

			nextEndTime := startTime.Add(-lr.bufferDuration)
			if lr.lastReadTime.After(nextEndTime) {
				klog.V(5).Info("wait for more events to be stored")
				return
			}

			//flow control for query range
			fetchTimeRange := nextEndTime.Sub(lr.lastReadTime).Seconds()
			controlledRange := time.Duration(flowController.DecideFlow(fetchTimeRange)) * time.Second

			fetchedEvent, err := lr.fetchEvents(controlledRange)
			if err != nil {
				klog.Errorf("failed to fetch events: %s", err.Error())
				//	isLastErr = true
				flowController.RecordFlow(false)
			} else {
				//	isLastErr = false
				flowController.RecordFlow(true)
			}
			xsearchFetedEventDurationMilliSeconds.Set(utils.TimeDiffInMilliSeconds(startTime, time.Now()))
			xsearchFetchedEventCountOneScroll.Set(float64(fetchedEvent))

		}, lr.fetchIntervalDuration, stopCh)
	}()

	// at most behind 10 minutues if app is killed without grace shutdown.
	go wait.Until(func() {
		lr.updateLastReadTime()
	}, 20*time.Second, stopCh)
}

func (lr *logReader) updateLastReadTime() error {
	conf := metaConf{
		LastReadTime: strconv.FormatInt(lr.lastReadTime.UnixNano(), 10),
	}
	var err error

	docID := lr.cluster
	x := utils.Dumps(conf)
	_, err = lr.esClient.Index().
		Index(metaIndexName).
		Id(docID).
		BodyString(x).
		Do(context.Background())
	if err == nil {
		klog.V(5).Infof("[%s]last read time updated to %s/%v", lr.cluster, conf.LastReadTime, lr.lastReadTime)
	}
	return err
}

func (lr *logReader) getLastReadTime() time.Time {
	now := time.Now()

	result, err := lr.esClient.Get().
		Index(metaIndexName).
		Id(lr.cluster). // use first cluster is ok. all cluster will have the same time
		Do(context.Background())
	if err != nil {
		klog.Errorf("[may be not exists] faild get last time for %s: %s", lr.cluster, err.Error())
		return now
	}

	if result.Found {
		conf := &metaConf{}
		err := json.Unmarshal(result.Source, conf)
		if err != nil {
			klog.Errorf("failed unmarshal %s: %s", string(result.Source), err.Error())
			return now
		}

		lastReadTime, err := strconv.ParseInt(conf.LastReadTime, 10, 64)
		if lastReadTime > 0 {
			t := time.Unix(0, lastReadTime)
			klog.Infof("got last read time %+v", t)
			return t
		}

		klog.Errorf("got invalid last read time %+v", conf.LastReadTime)
	}

	return now
}

func (lr *logReader) recovery(end time.Time, stopCh <-chan struct{}) (int64, error) {
	var (
		totalProcessed int64
	)

	klog.Infof("Getting all cached data from PodDeleteMileStoneMap")
	cluster := lr.cluster
	podDeleteMap, err := xsearch.GetAllPodDeleteMilestoneByScroll(cluster)
	if err != nil {
		klog.Errorf("Recovering PodDeleleteMilestone from xsearch failed, exiting... err is %v", err)
		// os.Exit(1)
	}
	slo.PodDeleteMileStoneMap = podDeleteMap
	klog.Infof("Finished getting all cached data from PodDeleteMileStoneMap, size is %d", slo.PodDeleteMileStoneMap.Size())

	klog.Infof("Deleting all cached data from PodDeleteMileStoneMap")
	xsearch.DeleteAllPodDeleteMilestone(cluster)
	klog.Infof("Finished deleting all cached data from PodDeleteMileStoneMap")

	for lr.lastReadTime.Before(end) {
		select {
		case <-stopCh:
			return totalProcessed, nil
		default:
		}
		klog.Infof("recover from %+v to %+v", lr.lastReadTime, end)
		timeRange := 30 * time.Second
		if lr.lastReadTime.Add(timeRange).After(end) {
			timeRange = end.Sub(lr.lastReadTime)
			klog.Infof("the last recovery fetch, will fetch from %+v with range %+v", lr.lastReadTime, timeRange)
		}

		processed := int64(0)
		err = utils.ReTry(func() error {
			processed, err = lr.fetchEvents(timeRange)
			return err
		}, 2*time.Second, 3)

		if err != nil {
			klog.Errorf("recovery error: %s", err)
			return totalProcessed, err
		}
		totalProcessed += processed
		lr.updateLastReadTime()
		xsearchFetchedEventCountOneScroll.Set(float64(processed))
		time.Sleep(200 * time.Millisecond)
	}
	klog.Infof("fetch %d events at recovery", totalProcessed)
	return totalProcessed, nil
}

// 获取审计日志
func (lr *logReader) fetchEvents(timeDuration time.Duration) (int64, error) {
	endTime := lr.lastReadTime.Add(timeDuration)
	startTime := lr.lastReadTime
	//daily索引生成每天凌晨00:30分
	/*m, _ := time.ParseDuration("-30m")
	indexDaily := fmt.Sprintf("append_only.%s.%s", lr.esConf.Index, startTime.Add(m).Format("2006-01-02"))*/
	indexDaily := currentIndex.Load().(string)
	if lastNoEventError {
		indexDaily = lr.esConf.Index
	}

	klog.V(6).Infof("The current index_daily to be searched: %s", indexDaily)
	var indexSliceDuration float64 = 0
	var indexSliceStart time.Time = time.Now()

	rangeQuery := elastic.NewRangeQuery("stageTimestamp").
		From(lr.lastReadTime).To(endTime).
		IncludeLower(true).
		IncludeUpper(false).
		TimeZone("UTC")
	query := elastic.NewBoolQuery().Must(rangeQuery)

	cluster := lr.cluster
	clusterQuery := elastic.NewTermQuery("annotations.cluster.keyword", cluster)
	query = query.Must(clusterQuery)

	s, _ := query.Source()
	klog.V(7).Infof("query: %s", utils.Dumps(s))

	pageSize := 10000
	totalGot := 0
	totalHits := int64(0)
	//hits := make(chan *shares.HitEvent)
	g, ctx := errgroup.WithContext(context.Background())

	//get count
	lastCount := int64(-1)
	count := int64(0)
	stableCount := 0
	for {
		curCount, err := lr.esClient.Count(indexDaily).Query(query).Do(ctx)
		if err != nil {
			return 0, err
		}

		if lastCount == curCount {
			stableCount++
		} else {
			stableCount = 0
		}
		//如果连续2次数据量相同，再进行查询
		if stableCount == 2 {
			count = curCount
			break
		}
		lastCount = curCount
		time.Sleep(200 * time.Millisecond)
	}

	indexSliceDuration = utils.TimeSinceInMilliSeconds(indexSliceStart)
	//如果count为0，直接退出等待重试，或者更新index
	if count == 0 {
		klog.Errorf("NoEventError")
		// 应对日志rotate场景
		if lastNoEventError {
			// 如果上次查询是NoEventError，则此次 index已经切换到全局（非daily模式），此次仍然没数据则大概率此时间段没数据，则查询时间需要继续向前更新
			lr.lastReadTime = endTime
		}
		lastNoEventError = true
		return 0, nil
	}
	lastNoEventError = false

	var globalErr error
	var sloScrollDuration float64 = 0
	var sloProDuration float64 = 0
	var sloUnmarshalDuration float64 = 0
	var produceFuncDuration float64 = 0

	var scrollWg sync.WaitGroup

	numPeer := count/int64(numSlices)/int64(queryDocNum) + 1
	stepDuration := timeDuration.Nanoseconds() / numPeer
	klog.Infof("data count: %d, numPeer:%d", count, numPeer)

	hitEventSlice := shares.NewHitEventSlice(10000)
	for i := 0; i < numSlices; i++ {
		// Prepare the slice
		sliceQuery := elastic.NewSliceQuery().Id(i).Max(numSlices)
		for j := 0; j < int(numPeer); j++ {
			scrollWg.Add(1)
			var curScrollDuration float64 = 0

			curQuery := elastic.NewBoolQuery().Must(clusterQuery)
			from := startTime.Add(time.Duration(stepDuration * int64(j)))
			to := startTime.Add(time.Duration(stepDuration * int64(j+1)))
			if j == int(numPeer)-1 {
				to = endTime
			}

			rangeQuery := elastic.NewRangeQuery("stageTimestamp").
				From(from).To(to).
				IncludeLower(true).
				IncludeUpper(false).
				TimeZone("UTC")
			curQuery = curQuery.Must(rangeQuery)

			go func(sliceQuery *elastic.SliceQuery, query *elastic.BoolQuery) error {
				defer func() {
					if curScrollDuration > sloScrollDuration {
						sloScrollDuration = curScrollDuration
					}
					scrollWg.Done()
				}()

				scroll := lr.esClient.Scroll(indexDaily).Query(query).Size(pageSize).Slice(sliceQuery)

				for {
					scrollStart := time.Now()
					results, err := scroll.Do(ctx)
					if results != nil && results.Error != nil {
						klog.Errorf("results failed error: %s, FailedShards len: %d", results.Error.Reason, len(results.Error.FailedShards))
					}
					curScrollDuration += utils.TimeDiffInMilliSeconds(scrollStart, time.Now())

					if err == io.EOF {
						// all results retrieved
						klog.V(8).Infof("scroll EOF, TotalHits: %d/%d, TotalGot:%d \n", results.TotalHits(), totalHits, totalGot)
						return nil
					}

					if err != nil && err != io.EOF {
						if results != nil && results.Error != nil {
							klog.Errorf("scroll result error: %s, FailedShards len: %d", results.Error.Reason, len(results.Error.FailedShards))
						}
						klog.Errorf("scroll result error: %s", err.Error())
						globalErr = err
						return err
					}

					if results != nil {
						totalHits += results.TotalHits()
					}

					// Send the hits to the hits channel
					totalGot += len(results.Hits.Hits)

					//hitEvents := shares.NewHitEventSlice(len(results.Hits.Hits))
					for _, hit := range results.Hits.Hits {
						hitEvent := shares.NewHitEvent()
						hitEvent.UnmarshalToEvent(&hit.Source)
						//hitEvents.Append(hitEvent)
						hitEventSlice.Append(hitEvent)
					}
				}
			}(sliceQuery, curQuery)
		}
	}

	// 2nd goroutine receives hits and deserializes them.
	var totalProcessed int64
	g.Go(func() error {
		start := time.Now()
		runtime.LockOSThread()
		defer func() {
			produceFuncDuration = utils.TimeSinceInMilliSeconds(start)
			runtime.UnlockOSThread()
		}()

		waitStart := time.Now()
		scrollWg.Wait()
		if globalErr != nil {
			klog.Errorf("scroll data do nothing when error found, err: %v\n", globalErr)
			return nil
		}
		if int64(totalGot) != count {
			globalErr = fmt.Errorf("scroll data not equal, target data: %d, total got:%d, from: %s, to: %s\n", count, totalGot, startTime.String(), endTime.String())
			klog.Errorf("%v", globalErr)
			return nil
		}
		klog.V(6).Infof("scroll finish, target data: %d, total got:%d, fetch from %s, range %v\n", count, totalGot, startTime.Format(time.RFC3339), timeDuration)

		hitEventSlice.Wait()
		hitEventSlice.SortByTimeStamp(true)

		sloUnmarshalDuration = utils.TimeDiffInMilliSeconds(waitStart, time.Now())

		for idx, _ := range hitEventSlice.Hits() {
			atomic.AddInt64(&totalProcessed, 1)
			hitEvent := hitEventSlice.Hits()[idx]

			if hitEvent.HasError {
				klog.Errorf("hitEvent HasError, continue")
				continue
			}

			event := *hitEvent.Event

			klog.V(8).Infof("got event %s", utils.DumpsEventKeyInfo(&event))
			//更新最近的event，防止出错断点重连
			lr.lastReadTime = event.StageTimestamp.Time

			// 此处为忽略某些 namespace 的流量，可以在 configmap 中进行配置
			// 例如：忽略压测流量 event.ObjectRef.Namespace == "cluster-loader-v3"
			if event.ObjectRef != nil &&
				utils.SliceContainsString(config.GlobalLunettesConfig().IgnoredNamespaceForAudit, event.ObjectRef.Namespace) {
				continue
			}

			shareEvent := shares.NewAuditEvent(&event)
			if event.ObjectRef.Resource == "pods" || event.ObjectRef.Resource == "events" || event.ObjectRef.Resource == "nodes" {
				shareEvent.Process()
			}

			// 将 event 放入 queue 中
			sloStart := time.Now()
			slo.Queue.Produce(shareEvent)             // 这个队列是用于 SLO 的
			podphase.WatcherQueue.Produce(shareEvent) // 这个队列是用于 pod phase 的
			podyaml.Queue.Produce(shareEvent)         // pod yaml
			nodeyaml.Queue.Produce(shareEvent)        // node yaml
			if featuregates.IsEnabled(spans.SpanAnalysisFeature) {
				spans.WatcherQueue.Produce(shareEvent)
			}
			if featuregates.IsEnabled(trace.TraceFeature) {
				trace.WatcherQueue.Produce(shareEvent)
			}

			sloProDuration += utils.TimeDiffInMilliSeconds(sloStart, time.Now())

			// Terminate early?
			select {
			default:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	// Check whether any goroutines failed.
	if err := g.Wait(); err != nil {
		globalErr = err
		klog.Errorf("failed to wait all goroutines finish when scroll elasticsearch result: %s", err.Error())
	}

	if globalErr != nil {
		xsearchQueryErrors.Inc()
		// notify tracer to check if some traces have timeout
		go func() {
			lr.lastReadTimeChan <- lr.lastReadTime
		}()

		return totalProcessed, globalErr
	}

	xsearchSLOProDurationInMilliSeconds.With(map[string]string{"phase": "slo_scroll"}).Set(sloScrollDuration)
	xsearchSLOProDurationInMilliSeconds.With(map[string]string{"phase": "slo_unmarshal"}).Set(sloUnmarshalDuration)
	xsearchSLOProDurationInMilliSeconds.With(map[string]string{"phase": "slo_delivery"}).Set(sloProDuration)
	xsearchSLOProDurationInMilliSeconds.With(map[string]string{"phase": "produce_func"}).Set(produceFuncDuration)
	xsearchSLOProDurationInMilliSeconds.With(map[string]string{"phase": "index_slice"}).Set(indexSliceDuration)

	// update last read time only when no errors found.
	// but this may leads this filed not updated if error returned continually
	lr.lastReadTime = endTime

	// notify tracer to check if some traces have timeout
	go func() {
		lr.lastReadTimeChan <- lr.lastReadTime
	}()

	klog.V(5).Infof("fetch from %s, range %v, got %d audit log", startTime.Format(time.RFC3339), timeDuration, totalProcessed)
	return totalProcessed, nil
}

func (lr *logReader) refreshIndex() {
	if currentIndex.Load().(string) == "" {
		currentIndex.Store(lr.esConf.Index)
	}

	go wait.Until(func() {
		indexDaily := fmt.Sprintf("append_only.%s.%s", lr.esConf.Index, time.Now().Format("2006-01-02"))
		if exist, _ := lr.esClient.IndexExists(indexDaily).Do(context.TODO()); exist {
			currentIndex.Store(indexDaily)
		}
	}, 5*time.Minute, make(chan struct{}))
}
