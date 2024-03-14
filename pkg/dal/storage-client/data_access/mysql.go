package data_access

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/common"
	customerrors "github.com/alipay/container-observability-service/pkg/custom-errors"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	v1 "k8s.io/api/core/v1"
	k8saudit "k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
)

// 2. 定义一个 StorageSqlImpl struct, 该struct 包含了存储client
type StorageSqlImpl struct {
	DB *gorm.DB
}

// 4. 提供一个 ProvideSqlStorate 方法, 传入一个 MysqlOptions, 返回一个 StorageInterface 和 error
func ProvideSqlStorate(conf *common.MysqlOptions) (StorageInterface, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=Local&interpolateParams=true",
		conf.Username,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DBName,
		conf.Charset,
		conf.ParseTime,
	)
	cf := &gorm.Config{
		SkipDefaultTransaction:   true,
		PrepareStmt:              false,
		DisableAutomaticPing:     true,
		DisableNestedTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		SkipInitializeWithVersion: true,
		// 模拟mysql 5.6版本
		ServerVersion: "5.6.3",
	}), cf)
	if err != nil {
		return nil, err
	}
	return &StorageSqlImpl{
		DB: db,
	}, nil
}

func (s *StorageSqlImpl) QuerySpanWithPodUid(data interface{}, uid string) error {

	if uid == "" {
		return errors.New("uid is empty")
	}

	fmt.Printf("uid is %s", uid)
	tx := s.DB.Debug().Order("span_elapsed desc").Limit(80).Where("or_uid = ?", uid).Find(data)
	if tx.Error != nil {
		klog.Error("err is", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil

}
func (s *StorageSqlImpl) QueryLifePhaseWithPodUid(data interface{}, podUID string) error {

	if podUID == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("start_time desc").Limit(200).Where("pod_uid = ?", podUID).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil
}
func (s *StorageSqlImpl) QueryPodYamlsWithPodUID(data interface{}, podUID string) error {

	if podUID == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(100).Where("pod_uid = ?", podUID).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}
	podYamls, ok := data.(*[]*model.PodYaml)
	if !ok {
		return fmt.Errorf("parse error")
	}
	pod := &v1.Pod{}
	for _, pYaml := range *podYamls {
		if er := json.Unmarshal([]byte(pYaml.SqlPod), pod); er == nil {
			pYaml.Pod = pod
		}
		klog.Infof("==pYaml.Pod=%+v", *pYaml.Pod)
	}

	return nil
}
func (s *StorageSqlImpl) QueryPodYamlsWithPodName(data interface{}, podName string) error {

	if podName == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(100).Where("pod_name = ?", podName).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}
	podYamls, ok := data.(*[]*model.PodYaml)
	if !ok {
		return fmt.Errorf("parse error")
	}
	pod := &v1.Pod{}
	for _, pYaml := range *podYamls {
		if er := json.Unmarshal([]byte(pYaml.SqlPod), pod); er == nil {
			pYaml.Pod = pod
		}
		klog.Infof("===%+v", *pYaml)
	}

	return nil
}
func (s *StorageSqlImpl) QueryPodYamlsWithHostName(data interface{}, hostName string) error {

	if hostName == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(1).Where("hostname =?", hostName).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}
	podYamls, ok := data.(*[]*model.PodYaml)
	if !ok {
		return fmt.Errorf("parse error")
	}
	pod := &v1.Pod{}
	for _, pYaml := range *podYamls {
		if er := json.Unmarshal([]byte(pYaml.SqlPod), pod); er == nil {
			pYaml.Pod = pod
		}
		klog.Infof("===%+v", *pYaml)
	}

	return nil
}
func (s *StorageSqlImpl) QueryPodYamlsWithPodIp(data interface{}, podIp string) error {

	if podIp == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(100).Where("pod_ip = ?", podIp).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}
	podYamls, ok := data.(*[]*model.PodYaml)
	if !ok {
		return fmt.Errorf("parse error")
	}
	pod := &v1.Pod{}
	for _, pYaml := range *podYamls {
		if er := json.Unmarshal([]byte(pYaml.SqlPod), pod); er == nil {
			pYaml.Pod = pod
		}
		klog.Infof("===%+v", *pYaml)
	}

	return nil
}
func (s *StorageSqlImpl) QueryPodListWithNodeip(data interface{}, nodeIp string, isDeleted bool) error {

	if nodeIp == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(100).Where("host_ip =? AND is_deleted =?", nodeIp, isDeleted).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil
}
func (s *StorageSqlImpl) QueryPodUIDListByHostname(data interface{}, hostName string) error {

	if hostName == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(100).Where("hostname =?", hostName).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil
}
func (s *StorageSqlImpl) QueryPodUIDListByPodIP(data interface{}, podIp string) error {

	if podIp == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(100).Where("pod_ip =?", podIp).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil
}
func (s *StorageSqlImpl) QueryPodUIDListByPodName(data interface{}, podName string) error {

	if podName == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(100).Where("pod_name =?", podName).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil
}
func (s *StorageSqlImpl) QueryNodeYamlsWithNodeUid(data interface{}, nodeUid string) error {

	if nodeUid == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(1).Where("uid =?", nodeUid).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}
	nodYamls, ok := data.(*[]*model.NodeYaml)
	if !ok {
		return fmt.Errorf("parse error")
	}
	node := &v1.Node{}
	for _, pYaml := range *nodYamls {
		if er := json.Unmarshal([]byte(pYaml.SqlNode), node); er == nil {
			pYaml.Node = node
		}
	}

	return nil
}
func (s *StorageSqlImpl) QueryNodeYamlsWithNodeName(data interface{}, nodeName string) error {

	if nodeName == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(1).Where("node_name =?", nodeName).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}
	nodYamls, ok := data.(*[]*model.NodeYaml)
	if !ok {
		return fmt.Errorf("parse error")
	}
	node := &v1.Node{}
	for _, pYaml := range *nodYamls {
		if er := json.Unmarshal([]byte(pYaml.SqlNode), node); er == nil {
			pYaml.Node = node
		}
	}

	return nil
}
func (s *StorageSqlImpl) QueryNodeYamlsWithNodeIP(data interface{}, nodeIp string) error {

	if nodeIp == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(1).Where("node_ip =? ", nodeIp).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}
	nodYamls, ok := data.(*[]*model.NodeYaml)
	if !ok {
		return fmt.Errorf("parse error")
	}
	node := &v1.Node{}
	for _, pYaml := range *nodYamls {
		if er := json.Unmarshal([]byte(pYaml.SqlNode), node); er == nil {
			pYaml.Node = node
		}
		klog.Infof("===%+v", *pYaml)
	}

	return nil
}
func (s *StorageSqlImpl) QueryNodeUIDListWithNodeIp(data interface{}, nodeIp string) error {

	if nodeIp == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(100).Where("node_ip =? ", nodeIp).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil
}
func (s *StorageSqlImpl) QueryPodYamlsWithNodeIP(data interface{}, nodeIp string) error {

	// returnResult := make([]*model.PodYaml, 0)
	if nodeIp == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(100).Where("host_ip =? AND is_deleted =?", nodeIp, false).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
	}
	returnResult, ok := data.(*[]*model.PodYaml)
	if !ok {
		return fmt.Errorf("parse error")
	}
	pod := &v1.Pod{}
	sloMap := make(map[int][]*model.SloTraceData)
	var mutex1 sync.Mutex
	var wg sync.WaitGroup
	for _, pYaml := range *returnResult {
		if er := json.Unmarshal([]byte(pYaml.SqlPod), pod); er == nil {
			poYaml := &model.PodYaml{
				AuditID:           pYaml.AuditID,
				ClusterName:       pYaml.ClusterName,
				HostIP:            pYaml.HostIP,
				PodIP:             pYaml.PodIP,
				Namespace:         pYaml.Namespace,
				PodUid:            pYaml.PodUid,
				CreationTimestamp: pod.CreationTimestamp.Time,
				DebugUrl:          "http://lunettes.lunettes.svc:8080/api/v1/debugpod?name=" + pYaml.PodName,
				PodName:           pYaml.PodName,
				Status:            string(pod.Status.Phase),
			}
			key := len(*returnResult)
			wg.Add(1)
			go func() {
				defer wg.Done()
				slotraceResult := make([]*model.SloTraceData, 0)
				s.QuerySloTraceDataWithPodUID(&slotraceResult, pYaml.PodUid)
				mutex1.Lock()
				sloMap[key] = slotraceResult
				mutex1.Unlock()
			}()
			*returnResult = append(*returnResult, poYaml)
		}
	}
	wg.Wait()

	for k, res := range *returnResult {
		res.SLOType = "OutOfDate"
		res.SLOResult = "OutOfDate"
		res.SLO = "OutOfDate"
		if len(sloMap[k]) > 0 {
			res.SLO = time.Duration(sloMap[k][0].PodSLO).String()
			if sloMap[k][0].StartUpResultFromCreate == "success" || sloMap[k][0].DeleteResult == "success" || sloMap[k][0].UpgradeResult == "success" {
				res.SLOResult = "success"
			} else {
				res.SLOResult = "fail"
			}
			for i := range sloMap[k] {
				res.SLOType = sloMap[k][i].Type
			}
		}
	}

	return nil
}

func (s *StorageSqlImpl) QueryPodInfoWithPodUid(data interface{}, podUid string) error {

	if podUid == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("stage_timestamp desc").Limit(1).Where("pod_uid =?", podUid).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil
}

func (s *StorageSqlImpl) QueryNodephaseWithNodeName(data interface{}, nodeName string) error {

	if nodeName == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("start_time desc").Limit(100).Where("node_name = ?", nodeName).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil
}
func (s *StorageSqlImpl) QueryNodephaseWithNodeUID(data interface{}, nodeUid string) error {

	if nodeUid == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("start_time desc").Limit(100).Where("node_uid = ?", nodeUid).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil
}
func (s *StorageSqlImpl) QuerySloTraceDataWithPodUID(data interface{}, podUid string) error {

	if podUid == "" {
		return fmt.Errorf("the params is error")
	}

	tx := s.DB.Debug().Order("created_time desc").Limit(100).Where("pod_uid =?", podUid).Find(data)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil
}
func (s *StorageSqlImpl) QueryCreateSloWithResult(data interface{}, requestParams *model.SloOptions) error {

	if requestParams == nil || requestParams.Result == "" {
		return customerrors.Error(customerrors.ErrParams, customerrors.NoDeliveryResult)
	}

	env := DELIVERY_ENV_PROD
	if requestParams.Env == "infra" {
		env = DELIVERY_ENV_INFRA
	}

	tx := s.DB.Order("created desc").Where("slo_violation_reason=?", requestParams.Result)
	// query := elastic.NewBoolQuery().Must(stringQuery)

	if requestParams.Cluster != "" {
		tx = tx.Where("cluster =?", requestParams.Cluster)
	}

	// add range query
	if !requestParams.From.IsZero() || !requestParams.To.IsZero() {
		//rangeQuery := elastic.NewRangeQuery("Created").TimeZone("UTC")
		if !requestParams.From.IsZero() {
			tx = tx.Where("created >=?", requestParams.From)
		}
		if !requestParams.To.IsZero() {
			tx = tx.Where("created <?", requestParams.To)
		}
		// query = query.Must(rangeQuery)
	}

	if requestParams.BizName != "" {
		//query = query.Must(stringQuery3)
		tx = tx.Where("biz_name =?", requestParams.BizName)
	}

	if requestParams.DeliveryStatus != "" {

		tx = tx.Where("delivery_status_orig =?", requestParams.DeliveryStatus)
		if env == DELIVERY_ENV_INFRA {
			tx = tx.Where("delivery_status =?", requestParams.DeliveryStatus)
		}
		//query = query.Must(stringQuery4)
	}

	if requestParams.SloTime != "" {
		sloduration, err := time.ParseDuration(requestParams.SloTime)
		if err == nil {
			tx = tx.Where("pod_slo=?", int(sloduration))
			if env == DELIVERY_ENV_INFRA {
				tx = tx.Where("delivery_slo=?", int(sloduration))
			}
			//query = query.Must(stringQuery5)
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

	tx = tx.Limit(querySize).Find(data)
	//tx.Debug()
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}

	return nil
}
func (s *StorageSqlImpl) QueryUpgradeSloWithResult(data interface{}, requestParams *model.SloOptions) error {

	if requestParams == nil || requestParams.Result == "" {
		return customerrors.Error(customerrors.ErrParams, customerrors.NoDeliveryResult)
	}
	tx := s.DB.Where("upgrade_result =?", requestParams.Result)

	if requestParams.Cluster != "" {
		tx = tx.Where("cluster =?", requestParams.Cluster)
	}

	if requestParams.Type != "" {
		tx = tx.Where("type =?", requestParams.Type)
	}

	// add range query
	if !requestParams.From.IsZero() || !requestParams.To.IsZero() {
		if !requestParams.From.IsZero() {
			tx = tx.Where("created_time >=?", requestParams.From)
		}
		if !requestParams.To.IsZero() {
			tx = tx.Where("created_time <?", requestParams.To)
		}
	}

	querySize := 300
	if requestParams.Count != "" {
		count, err := strconv.Atoi(requestParams.Count)
		if err == nil {
			querySize = count
		}
	}
	if querySize > 5000 {
		querySize = 5000
	}

	tx = tx.Limit(querySize).Order("created_time desc").Find(data)

	if tx.Error != nil {
		klog.Error(tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}
	return nil
}
func (s *StorageSqlImpl) QueryDeleteSloWithResult(data interface{}, requestParams *model.SloOptions) error {
	if requestParams == nil || requestParams.Result == "" {
		return customerrors.Error(customerrors.ErrParams, customerrors.NoDeliveryResult)
	}
	tx := s.DB.Where("delete_result =?", requestParams.Result)
	if requestParams.Cluster != "" {
		tx = tx.Where("cluster =?", requestParams.Cluster)
	}

	if requestParams.Type != "" {
		tx = tx.Where("type= ?", requestParams.Type)
	}

	// add range query
	if !requestParams.From.IsZero() || !requestParams.To.IsZero() {
		if !requestParams.From.IsZero() {
			tx = tx.Where("created_time >=?", requestParams.From)
		}

		if !requestParams.To.IsZero() {
			tx = tx.Where("created_time <?", requestParams.To)
		}

	}

	querySize := 300
	if requestParams.Count != "" {
		count, err := strconv.Atoi(requestParams.Count)
		if err == nil {
			querySize = count
		}
	}
	if querySize > 5000 {
		querySize = 5000
	}

	tx = tx.Limit(querySize).Order("created_time desc").Find(data)
	if tx.Error != nil {
		klog.Error(tx.Error)
		return fmt.Errorf("error%v", tx.Error)
	}
	return nil
}
func (s *StorageSqlImpl) QueryNodeYamlWithParams(data interface{}, debugparams *model.NodeParams) error {
	var resultOB *gorm.DB

	if debugparams.NodeName != "" {
		resultOB = s.DB.Order("stage_timestamp desc").Limit(1).Where("node_name =?", debugparams.NodeName).Find(data)
	} else if debugparams.NodeUid != "" {
		resultOB = s.DB.Order("stage_timestamp desc").Limit(1).Where("uid =?", debugparams.NodeUid).Find(data)
	} else if debugparams.NodeIp != "" {
		resultOB = s.DB.Order("stage_timestamp desc").Limit(1).Where("node_ip =?", debugparams.NodeIp).Find(data)
	}
	if resultOB.Error != nil {
		return fmt.Errorf("error%v", resultOB.Error)
	}
	return nil
}
func (s *StorageSqlImpl) QueryAuditWithAuditId(data interface{}, auditid string) error {

	awsAudit := make([]*model.Audit, 0)

	tx := s.DB.Where("audit_id = ?", auditid).Find(&awsAudit)
	if tx.Error != nil {
		klog.Errorf("db.Order Error: %s", tx.Error)
	}

	if len(awsAudit) == 0 {
		return fmt.Errorf("the query is error")
	}
	result := &k8saudit.Event{}
	auditByte := json.RawMessage(awsAudit[0].Content)

	if err := json.Unmarshal(auditByte, &result); err != nil {
		klog.Errorf("json unmarshal error: %s", err)
	}
	return nil

}
func (s *StorageSqlImpl) QueryEventPodsWithPodUid(data interface{}, auditid string) error {

	return nil

}
func (s *StorageSqlImpl) QueryEventNodeWithPodUid(data interface{}, auditid string) error {

	return nil

}
func (s *StorageSqlImpl) QueryEventWithTimeRange(data interface{}, from, to time.Time) error {

	return nil

}
func (s *StorageSqlImpl) QueryPodYamlWithParams(data interface{}, debugparams *model.PodParams) error {
	var resultOB *gorm.DB

	if debugparams.Name != "" {
		resultOB = s.DB.Order("stage_timestamp desc").Limit(10).Where("pod_name =?", debugparams.Name).Find(data)
	} else if debugparams.Hostname != "" {
		resultOB = s.DB.Order("stage_timestamp desc").Limit(10).Where("hostname =?", debugparams.Hostname).Find(data)
	} else if debugparams.Podip != "" {
		resultOB = s.DB.Order("stage_timestamp desc").Limit(10).Where("pod_ip = ?", debugparams.Podip).Find(data)
	}
	if resultOB.Error != nil {
		return fmt.Errorf("error%v", resultOB.Error)
	}
	return nil

}
