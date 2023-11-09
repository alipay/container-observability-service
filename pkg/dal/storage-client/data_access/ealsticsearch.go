package data_access

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/common"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/olivere/elastic/v7"

	// utils "github.com/alipay/container-observability-service/pkg/utils"
	"k8s.io/klog"
)

// 2. 定义一个 StorageSqlImpl struct, 该struct 包含了存储client
type StorageEsImpl struct {
	DB *elastic.Client
}

const (
	DELIVERY_ENV_PROD  = "prod"
	DELIVERY_ENV_INFRA = "infra"
	PodResource        = "Pod"
	NodeResource       = "Node"
	podYamlIndexName   = "pod_yaml"
	podYamlTypeName    = "_doc"
	nodeYamlIndexName  = "node_yaml"
	nodeYamlTypeName   = "_doc"
)

// 4. 提供一个 ProvideSqlStorate 方法, 传入一个 MysqlOptions, 返回一个 StorageInterface 和 error
func ProvideEsStorage(conf *common.ESOptions) (StorageInterface, error) {
	var err error

	esClient, err := elastic.NewClient(
		elastic.SetURL(conf.EndPoint), elastic.SetBasicAuth(conf.Username, conf.Password), elastic.SetSniff(false))
	if err != nil {
		// panic(err)
		klog.Errorf("init es client error %s", err)
		return nil, err
	}
	return &StorageEsImpl{
		DB: esClient,
	}, nil
}

// func (s *StorageEsImpl) QuerySpanWithPodUid(data model.DataModelInterface, uid string) error {
func (s *StorageEsImpl) QuerySpanWithPodUid(data interface{}, uid string) error {
	if uid == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	query := elastic.NewBoolQuery().Must(elastic.NewQueryStringQuery(fmt.Sprintf("OwnerRef.UID: \"%s\" AND OwnerRef.Resource: pods", uid)))
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(80).
		Sort("Elapsed", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil

}
func (s *StorageEsImpl) QueryLifePhaseWithPodUid(data interface{}, uid string) error {
	if uid == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryLifePhase").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podUID: \"%s\"", uid))
	query := elastic.NewBoolQuery().Must(stringQuery)

	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(200).
		Sort("startTime", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryPodYamlsWithPodUID(data interface{}, uid string) error {

	if uid == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryPodYaml").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podUID: \"%s\"", uid))
	query := elastic.NewBoolQuery().Must(stringQuery)

	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryPodYamlsWithPodName(data interface{}, podName string) error {

	if podName == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryLifePhase").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podName: \"%s\"", podName))
	query := elastic.NewBoolQuery().Must(stringQuery)

	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryPodYamlsWithHostName(data interface{}, hostName string) error {

	if hostName == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryLifePhase").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("hostname: \"%s\"", hostName))
	query := elastic.NewBoolQuery().Must(stringQuery)

	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryPodYamlsWithPodIp(data interface{}, podIp string) error {

	if podIp == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryLifePhase").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podIP: \"%s\"", podIp))
	query := elastic.NewBoolQuery().Must(stringQuery)

	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryPodListWithNodeip(data interface{}, nodeIp string, isDeleted bool) error {

	if nodeIp == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}
	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryLifePhase").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("hostIP.keyword: \"%s\"", nodeIp))
	deleteFalse := elastic.NewQueryStringQuery(fmt.Sprintf("isDeleted.keyword: \"%t\"", isDeleted))
	query := elastic.NewBoolQuery().Must(stringQuery).Must(deleteFalse)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(300).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil

}
func (s *StorageEsImpl) QueryPodUIDListByHostname(data interface{}, hostName string) error {

	if hostName == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryLifePhase").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("hostname.keyword: \"%s\"", hostName))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(300).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryPodUIDListByPodIP(data interface{}, podIp string) error {

	if podIp == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryLifePhase").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podIP.keyword: \"%s\"", podIp))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(300).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryPodUIDListByPodName(data interface{}, podName string) error {

	if podName == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryLifePhase").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podName: \"%s\"", podName))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(300).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryNodeYamlsWithNodeUid(data interface{}, nodeUid string) error {

	if nodeUid == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryNodeYaml").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("uid: \"%s\"", nodeUid))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryNodeYamlsWithNodeName(data interface{}, nodeName string) error {

	if nodeName == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryNodeYaml").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("nodeName: \"%s\"", nodeName))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryNodeYamlsWithNodeIP(data interface{}, nodeIp string) error {

	if nodeIp == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryLifePhase").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("nodeIp: \"%s\"", nodeIp))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryNodeUIDListWithNodeIp(data interface{}, nodeIp string) error {

	if nodeIp == "" {
		return fmt.Errorf("the params is error, nodeIp is nil")
	}

	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("Querynodeuid").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("nodeIp: \"%s\"", nodeIp))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryPodYamlsWithNodeIP(data interface{}, nodeIp string) error {

	returnResult, ok := data.(*[]*model.PodYaml)
	if !ok {
		return fmt.Errorf("parse error")
	}
	if nodeIp == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryPodYaml").Observe(cost)
	}()
	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("hostIP.keyword: \"%s\"", nodeIp))
	deleteFalse := elastic.NewQueryStringQuery(fmt.Sprintf("isDeleted.keyword: \"%t\"", false))
	query := elastic.NewBoolQuery().Must(stringQuery).Must(deleteFalse)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(300).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}

	dedup := make(map[string]string)
	sloMap := make(map[int][]*model.SloTraceData)
	var mutex1 sync.Mutex
	var wg sync.WaitGroup
	for _, hit := range searchResult.Hits.Hits {
		pyaml := &model.PodYaml{}
		if er := json.Unmarshal(*&hit.Source, pyaml); er == nil {
			if pyaml.Pod != nil {
				if _, ok := dedup[pyaml.PodUid]; !ok {
					podrestun := &model.PodYaml{
						AuditID:           pyaml.AuditID,
						ClusterName:       pyaml.ClusterName,
						HostIP:            pyaml.HostIP,
						PodIP:             pyaml.PodIP,
						Namespace:         pyaml.Namespace,
						PodUid:            pyaml.PodUid,
						CreationTimestamp: pyaml.Pod.CreationTimestamp.Time,
						DebugUrl:          "http://lunettes.lunettes.svc:8080/api/v1/debugpod?name=" + pyaml.PodName,
						PodName:           pyaml.PodName,
						Status:            string(pyaml.Pod.Status.Phase),
					}
					key := len(*returnResult)
					wg.Add(1)
					go func() {
						defer wg.Done()
						sloTrace := make([]*model.SloTraceData, 0)
						s.QuerySloTraceDataWithPodUID(&sloTrace, pyaml.PodUid)
						mutex1.Lock()
						sloMap[key] = sloTrace
						mutex1.Unlock()
					}()
					*returnResult = append(*returnResult, podrestun)
					dedup[pyaml.PodUid] = "true"
				}
			}
		}
	}
	wg.Wait()
	for k, v := range *returnResult {
		v.SLOType = "OutOfDate"
		v.SLOResult = "OutOfDate"
		v.SLO = "OutOfDate"
		if len(sloMap[k]) > 0 {
			v.SLO = time.Duration(sloMap[k][0].PodSLO).String()
			for i := range sloMap[k] {
				if sloMap[k][i].StartUpResultFromCreate == "success" || sloMap[k][i].DeleteResult == "success" || sloMap[k][i].UpgradeResult == "success" {
					v.SLOResult = "success"
					v.SLOType = sloMap[k][i].Type
				} else {
					v.SLOResult = "fail"
				}
			}
		}
	}

	return nil
}

func (s *StorageEsImpl) QueryPodInfoWithPodUid(data interface{}, podUid string) error {

	if podUid == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryNodeYaml").Observe(cost)
	}()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podUID: \"%s\"", podUid))
	query := elastic.NewBoolQuery().Must(stringQuery)

	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryNodephaseWithNodeUID(data interface{}, nodeUid string) error {
	if nodeUid == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryNodephase", begin)
	}()

	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("uid: \"%s\"", nodeUid))
	query := elastic.NewBoolQuery().Must(stringQuery)

	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(200).
		Sort("startTime", false).Do(context.Background())

	if err != nil {
		return fmt.Errorf("error%v", err)
	}

	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryNodephaseWithNodeName(data interface{}, nodeName string) error {
	if nodeName == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryNodephase", begin)
	}()

	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("nodeName: \"%s\"", nodeName))
	query := elastic.NewBoolQuery().Must(stringQuery)

	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(200).
		Sort("startTime", false).Do(context.Background())

	if err != nil {
		return fmt.Errorf("error%v", err)
	}

	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QuerySloTraceDataWithPodUID(data interface{}, podUid string) error {
	if podUid == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QuerySloTraceData", begin)
	}()

	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("PodUID: \"%s\"", podUid))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(10).
		Sort("CreatedTime", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}

	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryCreateSloWithResult(data interface{}, requestParams *model.SloOptions) error {
	if requestParams == nil || requestParams.Result == "" {
		return fmt.Errorf("the params is error")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QuerySlo").Observe(cost)
	}()

	env := DELIVERY_ENV_PROD
	if requestParams.Env == "infra" {
		env = DELIVERY_ENV_INFRA
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("SLOViolationReason: \"%s\"", requestParams.Result))

	query := elastic.NewBoolQuery().Must(stringQuery)

	if requestParams.Cluster != "" {
		query = query.Must(elastic.NewTermQuery("Cluster.keyword", requestParams.Cluster))
	}

	// add range query
	if !requestParams.From.IsZero() || !requestParams.To.IsZero() {
		rangeQuery := elastic.NewRangeQuery("Created").TimeZone("UTC")
		if !requestParams.From.IsZero() {
			rangeQuery = rangeQuery.From(requestParams.From)
		}
		if !requestParams.To.IsZero() {
			rangeQuery = rangeQuery.To(requestParams.To)
		}
		query = query.Must(rangeQuery)
	} else {
		rangeQuery := elastic.NewRangeQuery("Created").TimeZone("UTC").Gte("now-24h")
		query = query.Must(rangeQuery)
	}
	if requestParams.BizName != "" {
		stringQuery3 := elastic.NewQueryStringQuery(fmt.Sprintf("BizName: \"%s\"", requestParams.BizName))
		query = query.Must(stringQuery3)
	}

	if requestParams.DeliveryStatus != "" {

		stringQuery4 := elastic.NewQueryStringQuery(fmt.Sprintf("DeliveryStatusOrig: \"%s\"", requestParams.DeliveryStatus))
		if env == DELIVERY_ENV_INFRA {
			stringQuery4 = elastic.NewQueryStringQuery(fmt.Sprintf("DeliveryStatus: \"%s\"", requestParams.DeliveryStatus))
		}
		query = query.Must(stringQuery4)
	}

	if requestParams.SloTime != "" {
		sloduration, err := time.ParseDuration(requestParams.SloTime)

		if err == nil {
			stringQuery5 := elastic.NewQueryStringQuery(fmt.Sprintf("PodSLO: \"%d\"", int(sloduration)))
			if env == DELIVERY_ENV_INFRA {
				stringQuery5 = elastic.NewQueryStringQuery(fmt.Sprintf("DeliverySLO: \"%d\"", int(sloduration)))
			}
			query = query.Must(stringQuery5)
		} else {
			fmt.Printf("Error slotime format %s \n", requestParams.SloTime)
		}
	}

	querySize := 300
	if requestParams.Count != "" {
		count, err := strconv.Atoi(requestParams.Count)
		if err == nil {
			querySize = count
		}
	}
	if querySize > 500 {
		querySize = 500
	}

	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(querySize).
		Sort("Created", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}

	return nil
}
func (s *StorageEsImpl) QueryUpgradeSloWithResult(data interface{}, requestParams *model.SloOptions) error {
	if requestParams == nil || requestParams.Result == "" {
		return fmt.Errorf("the params is error")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	resultQuery := elastic.NewQueryStringQuery(fmt.Sprintf("UpgradeResult: \"%s\"", requestParams.Result))
	query := elastic.NewBoolQuery().Must(resultQuery)

	if requestParams.Cluster != "" {
		query = query.Must(elastic.NewTermQuery("Cluster.keyword", requestParams.Cluster))
	}

	if requestParams.Type != "" {
		typeQuery := elastic.NewQueryStringQuery(fmt.Sprintf("Type: \"%s\"", requestParams.Type))
		query = query.Must(typeQuery)
	}

	// add range query
	if !requestParams.From.IsZero() || !requestParams.To.IsZero() {
		rangeQuery := elastic.NewRangeQuery("CreatedTime").TimeZone("UTC")
		if !requestParams.From.IsZero() {
			rangeQuery = rangeQuery.From(requestParams.From)
		}
		if !requestParams.To.IsZero() {
			rangeQuery = rangeQuery.To(requestParams.To)
		}
		query = query.Must(rangeQuery)
	} else {
		rangeQuery := elastic.NewRangeQuery("CreatedTime").TimeZone("UTC").Gte("now-24h")
		query = query.Must(rangeQuery)
	}

	querySize := 300
	if requestParams.Count != "" {
		count, err := strconv.Atoi(requestParams.Count)
		if err == nil {
			querySize = count
		}
	}
	if querySize > 500 {
		querySize = 500
	}

	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(querySize).
		Sort("CreatedTime", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return fmt.Errorf("the params is error")
	}

	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}

	return nil
}
func (s *StorageEsImpl) QueryDeleteSloWithResult(data interface{}, requestParams *model.SloOptions) error {
	if requestParams == nil || requestParams.Result == "" {
		return fmt.Errorf("the params is error")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("DeleteResult: \"%s\"", requestParams.Result))
	query := elastic.NewBoolQuery().Must(stringQuery)

	if requestParams.Cluster != "" {
		query = query.Must(elastic.NewTermQuery("Cluster.keyword", requestParams.Cluster))
	}

	if requestParams.Type != "" {
		stringQuery4 := elastic.NewQueryStringQuery(fmt.Sprintf("Type: \"%s\"", requestParams.Type))
		query = query.Must(stringQuery4)
	}

	// add range query
	if !requestParams.From.IsZero() || !requestParams.To.IsZero() {
		rangeQuery := elastic.NewRangeQuery("CreatedTime").TimeZone("UTC")
		if !requestParams.From.IsZero() {
			rangeQuery = rangeQuery.From(requestParams.From)
		}
		if !requestParams.To.IsZero() {
			rangeQuery = rangeQuery.To(requestParams.To)
		}
		query = query.Must(rangeQuery)
	} else {
		rangeQuery := elastic.NewRangeQuery("CreatedTime").TimeZone("UTC").Gte("now-24h")
		query = query.Must(rangeQuery)
	}
	querySize := 300
	if requestParams.Count != "" {
		count, err := strconv.Atoi(requestParams.Count)
		if err == nil {
			querySize = count
		}
	}
	if querySize > 500 {
		querySize = 500
	}

	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(querySize).
		Sort("CreatedTime", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}

	return nil
}
func (s *StorageEsImpl) QueryNodeYamlWithParams(data interface{}, debugparams *model.NodeParams) error {

	if debugparams == nil {
		return fmt.Errorf("the params is error")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryLifePhase").Observe(cost)
	}()

	var stringQuery *elastic.QueryStringQuery
	if debugparams.NodeName != "" {
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("nodeName: \"%s\"", debugparams.NodeName))
	} else if debugparams.NodeUid != "" {
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("uid: \"%s\"", debugparams.NodeUid))
	} else if debugparams.NodeIp != "" {
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("nodeIp : \"%s\"", debugparams.NodeIp))
	}
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageEsImpl) QueryAuditWithAuditId(data interface{}, auditid string) error {
	if auditid == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryNodephase", begin)
	}()

	query := elastic.NewBoolQuery().Must(elastic.NewQueryStringQuery(fmt.Sprintf("auditID: \"%s\"", auditid)))
	searchReulst, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return fmt.Errorf("the query is error")
	}

	if err != nil {
		return fmt.Errorf("error%v", err)
	}

	res, ok := data.(*model.Audit)
	if !ok {
		return fmt.Errorf("the query is error")
	}
	for _, hit := range searchReulst.Hits.Hits {
		err = json.Unmarshal(*&hit.Source, &res)
		if err != nil {
			return err
		}
	}

	return nil
}
func (s *StorageEsImpl) QueryEventPodsWithPodUid(data interface{}, PodUid string) error {
	if PodUid == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryNodephase", begin)
	}()

	query := elastic.NewBoolQuery()
	query.Must(elastic.NewTermQuery("objectRef.uid.keyword", PodUid))
	query.Must(elastic.NewTermQuery("objectRef.resource", "pods"))

	searchReulst, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(10).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return fmt.Errorf("the query is error")
	}

	if err != nil {
		return fmt.Errorf("error%v", err)
	}

	res, ok := data.(*model.Audit)
	if !ok {
		return fmt.Errorf("the query is error")
	}
	for _, hit := range searchReulst.Hits.Hits {
		err = json.Unmarshal(*&hit.Source, &res)
		if err != nil {
			return err
		}
	}

	return nil
}
func (s *StorageEsImpl) QueryEventNodeWithPodUid(data interface{}, PodUid string) error {
	if PodUid == "" {
		return fmt.Errorf("the params is error, uid is nil")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryNodephase", begin)
	}()

	query := elastic.NewBoolQuery()
	query.Must(elastic.NewTermQuery("objectRef.uid.keyword", PodUid))
	query.Must(elastic.NewTermQuery("objectRef.resource", "nodes"))

	searchReulst, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(10).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return fmt.Errorf("the query is error")
	}

	if err != nil {
		return fmt.Errorf("error%v", err)
	}

	res, ok := data.(*model.Audit)
	if !ok {
		return fmt.Errorf("the query is error")
	}
	for _, hit := range searchReulst.Hits.Hits {
		err = json.Unmarshal(*&hit.Source, &res)
		if err != nil {
			return err
		}
	}

	return nil
}
func (s *StorageEsImpl) QueryEventWithTimeRange(data interface{}, from, to time.Time) error {

	query := elastic.NewBoolQuery()
	if !from.IsZero() || !to.IsZero() {
		rangeQuery := elastic.NewRangeQuery("stageTimestamp").TimeZone("UTC")
		if !from.IsZero() {
			rangeQuery = rangeQuery.From(from)
		}
		if !to.IsZero() {
			rangeQuery = rangeQuery.To(to)
		}
		query = query.Must(rangeQuery)
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryNodephase", begin)
	}()

	searchReulst, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(10).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return fmt.Errorf("the query is error")
	}

	res, ok := data.(*model.Audit)
	if !ok {
		return fmt.Errorf("the query is error")
	}
	for _, hit := range searchReulst.Hits.Hits {
		err = json.Unmarshal(*&hit.Source, &res)
		if err != nil {
			return err
		}
	}

	return nil
}
func (s *StorageEsImpl) QueryPodYamlWithParams(data interface{}, params *model.PodParams) error {

	if params == nil {
		return fmt.Errorf("the params is error")
	}
	_, esTableName, esType, err := utils.GetMetaName(data)
	if err != nil {
		return err
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryPodYamlWithParams").Observe(cost)
	}()

	var stringQuery *elastic.QueryStringQuery
	if params.Podip != "" {
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("podIP: \"%s\"", params.Podip))
	} else if params.Name != "" {
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("podName: \"%s\"", params.Name))
	} else if params.Hostname != "" {
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("hostname: \"%s\"", params.Hostname))
	}
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(esTableName).Type(esType).Query(query).Size(10).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		return fmt.Errorf("error%v", err)
	}
	var hits []*json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, &hit.Source)
	}
	hitsStr, err := json.Marshal(hits)
	if err != nil {
		return err
	}

	err = json.Unmarshal(hitsStr, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *StorageEsImpl) QueryResourceYamlWithUID(kind, uid string) (interface{}, error) {
	result := make([]interface{}, 0)
	if uid == "" {
		return result, nil
	}

	index := ""
	typ := ""
	var stringQuery *elastic.QueryStringQuery
	if kind == PodResource {
		index = podYamlIndexName
		typ = podYamlTypeName
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("podUID: \"%s\"", uid))
	} else if kind == NodeResource {
		index = nodeYamlIndexName
		typ = nodeYamlTypeName
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("uid: \"%s\"", uid))
	}

	if index == "" {
		return result, nil
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryResourceYamlWithUID").Observe(cost)
	}()

	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(index).Type(typ).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		err = fmt.Errorf("failed to get yaml of %s[%s], error: %v", kind, uid, err)
		klog.Error(err)
		return result, err
	}

	for _, hit := range searchResult.Hits.Hits {
		if kind == PodResource {
			objYaml := &model.PodYaml{}
			if er := json.Unmarshal(*&hit.Source, objYaml); er == nil {
				if objYaml.Pod != nil {
					result = append(result, objYaml)
					fmt.Println("fetch pod yaml")
				}
			}
		} else if kind == NodeResource {
			objYaml := &model.NodeYaml{}
			if er := json.Unmarshal(*&hit.Source, objYaml); er == nil {
				if objYaml.Node != nil {
					result = append(result, objYaml)
				}
			}
		}
	}

	return result, nil
}

func (s *StorageEsImpl) QueryResourceYamlWithName(kind, name string) (interface{}, error) {
	result := make([]interface{}, 0)
	if name == "" {
		return result, nil
	}

	index := ""
	typ := ""
	var stringQuery *elastic.QueryStringQuery
	if kind == PodResource {
		index = podYamlIndexName
		typ = podYamlTypeName
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("podName: \"%s\"", name))
	} else if kind == NodeResource {
		index = nodeYamlIndexName
		typ = nodeYamlTypeName
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("nodeName: \"%s\"", name))
	}

	if index == "" {
		return result, nil
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryResourceYamlWithName").Observe(cost)
	}()

	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := s.DB.Search().Index(index).Type(typ).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		err = fmt.Errorf("failed to get yaml of %s[%s], error: %v", kind, name, err)
		klog.Error(err)
		return result, err
	}

	for _, hit := range searchResult.Hits.Hits {
		if kind == PodResource {
			objYaml := &model.PodYaml{}
			if er := json.Unmarshal(*&hit.Source, objYaml); er == nil {
				if objYaml.Pod != nil {
					result = append(result, objYaml)
				}
			}
		} else if kind == NodeResource {
			objYaml := &model.NodeYaml{}
			if er := json.Unmarshal(*&hit.Source, objYaml); er == nil {
				if objYaml.Node != nil {
					result = append(result, objYaml)
				}
			}
		}
	}

	return result, nil
}
