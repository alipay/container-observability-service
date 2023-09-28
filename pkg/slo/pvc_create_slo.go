package slo

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/xsearch"

	"k8s.io/klog/v2"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/alipay/container-observability-service/pkg/utils"

	v1 "k8s.io/api/core/v1"

	k8s_audit "k8s.io/apiserver/pkg/apis/audit"

	"github.com/alipay/container-observability-service/pkg/queue"
)

type PersistentVolumeClaimMileStone struct {
	Type            string
	Cluster         string
	Namespace       string
	PVCName         string
	PVCUID          string
	TriggerAuditLog string
	CreateResult    string
	CreatedTime     time.Time
	TimeoutTime     time.Time
	//内部变量
	key   string
	mutex sync.Mutex
}

const (
	PVC_CREATE_TIMEOUT = "timeout"
	PVC_CREATE_SUCCESS = "success"
)

var (
	pvcCreateResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pvc_create_result_count",
			Help: "",
		},
		[]string{"cluster", "namespace", "result"},
	)

	pvcCreateQueue       *queue.BoundedQueue
	keyToMileStone       *utils.SafeMap
	currentPVCCreateTime time.Time
)

func init() {
	prometheus.MustRegister(pvcCreateResult)

	pvcCreateQueue = queue.NewBoundedQueue("slo-pvc-create", 10000, nil)
	pvcCreateQueue.StartLengthReporting(10 * time.Second)
	pvcCreateQueue.StartConsumers(1, func(item interface{}) {
		defer utils.IgnorePanic("pvcCreateConsumer")

		event, ok := item.(*k8s_audit.Event)
		if !ok {
			return
		}
		currentPVCCreateTime = event.StageTimestamp.Time
		//pvc create
		doPVCCreate(event)
		//pvc update&patch
		doPVCUpdateAndPath(event)
	})

	keyToMileStone = utils.NewSafeMap()
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sloOngoingSize.WithLabelValues("pvcCreate").Set(float64(keyToMileStone.Size()))
				checkPVCCreateTimeout()
			}
		}
	}()
}

func genKey(cluster, namespace, name string) string {
	return fmt.Sprintf("%s/%s/%s", cluster, namespace, name)
}

func finishPVCCreateWithResult(key string, result string) {
	v, ok := keyToMileStone.Get(key)
	if ok && v != nil {
		milestone := v.(*PersistentVolumeClaimMileStone)
		milestone.CreateResult = result
		pvcCreateResult.WithLabelValues(milestone.Cluster, milestone.Namespace, milestone.CreateResult).Inc()

		data, err := json.Marshal(milestone)
		if err == nil {
			_ = xsearch.SaveSloTraceData(milestone.Cluster, milestone.Namespace, milestone.PVCName, milestone.PVCUID, "create", data)
		}
	}
	keyToMileStone.Delete(key)
}

func checkPVCCreateTimeout() {
	toDeleteKeys := make([]string, 0)
	keyToMileStone.IterateWithFunc(func(i interface{}) {
		ms, ok := i.(*PersistentVolumeClaimMileStone)
		if !ok {
			return
		}
		if currentPVCCreateTime.After(ms.TimeoutTime) {
			toDeleteKeys = append(toDeleteKeys, ms.key)
		}
	})
	for _, key := range toDeleteKeys {
		finishPVCCreateWithResult(key, PVC_CREATE_TIMEOUT)
	}
}

func doPVCUpdateAndPath(auditEvent *k8s_audit.Event) {
	if auditEvent.ResponseStatus.Code >= 300 || auditEvent.ObjectRef.Resource != "persistentvolumeclaims" {
		return
	}

	pvc := &v1.PersistentVolumeClaim{}
	err := json.Unmarshal(auditEvent.ResponseObject.Raw, pvc)
	if err != nil {
		return
	}

	clusterName := auditEvent.Annotations["cluster"]
	key := genKey(clusterName, pvc.Namespace, pvc.Name)
	v, ok := keyToMileStone.Get(key)
	if ok && v != nil {
		if pvc.Status.Phase == "Bound" {
			v.(*PersistentVolumeClaimMileStone).CreateResult = PVC_CREATE_SUCCESS
			finishPVCCreateWithResult(key, PVC_CREATE_SUCCESS)
		}
	}
}

func doPVCCreate(auditEvent *k8s_audit.Event) {
	if auditEvent.ResponseStatus.Code >= 300 || auditEvent.Verb != "create" {
		return
	}
	if auditEvent.ObjectRef.Resource != "persistentvolumeclaims" || auditEvent.ObjectRef.Subresource != "" {
		return
	}
	pvc := &v1.PersistentVolumeClaim{}
	err := json.Unmarshal(auditEvent.ResponseObject.Raw, pvc)
	if err != nil {
		klog.Info(err)
		return
	}

	klog.Info(auditEvent.AuditID)

	clusterName := auditEvent.Annotations["cluster"]
	key := genKey(clusterName, pvc.Namespace, pvc.Name)
	v, ok := keyToMileStone.Get(key)
	if !ok || v == nil {
		ms := &PersistentVolumeClaimMileStone{
			Type:            PVC_CREATE,
			Cluster:         clusterName,
			Namespace:       pvc.Namespace,
			PVCName:         pvc.Name,
			PVCUID:          string(pvc.UID),
			TriggerAuditLog: string(auditEvent.AuditID),
			CreatedTime:     auditEvent.StageTimestamp.Time,
			TimeoutTime:     auditEvent.StageTimestamp.Time.Add(10 * time.Minute),
			key:             key,
			mutex:           sync.Mutex{},
		}
		keyToMileStone.Set(key, ms)
	}
}
