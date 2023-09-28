package podyaml

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/shares"
	"k8s.io/klog/v2"

	"github.com/alipay/container-observability-service/pkg/queue"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	corev1 "k8s.io/api/core/v1"
)

const (
	deleteMark           = "deleted"
	workerQueueCap       = 2000
	numberOfWorkerQueues = 1000
)

var (
	//Queue audit log queue
	Queue            *queue.BoundedQueue
	savePodInfoQueue *queue.BoundedQueue
)

type podOpStruct struct {
	clusterName string
	pod         *corev1.Pod
	t           time.Time
	auditID     string
	isCreate    bool
	isDeleting  bool
	isDeleted   bool
}

func init() {
	Queue = queue.NewBoundedQueue("podyaml", 50000, nil)
	Queue.StartLengthReporting(10 * time.Second)
	Queue.IsDropEventOnFull = false
	Queue.IsLockOSThread = true
	Queue.SetFilterItemFunc(func(item interface{}) bool {
		if item == nil {
			return true
		}

		auditEvent, ok := item.(*shares.AuditEvent)
		if !ok || auditEvent == nil {
			return true
		}

		if auditEvent.ObjectRef == nil || auditEvent.ObjectRef.Resource != "pods" {
			return true
		}
		if auditEvent.ResponseStatus.Code >= 300 {
			return true
		}
		return false
	})

	Queue.StartConsumers(1, func(v interface{}) {
		defer utils.IgnorePanic("StartConsumers")

		auditEvent, ok := v.(*shares.AuditEvent)
		if !ok || auditEvent == nil {
			return
		}
		auditEvent.Wait()

		consumePodOpStruct(auditEvent)
	})

	savePodInfoQueue = queue.NewBoundedQueue("podyaml-info", 40000, nil)
	savePodInfoQueue.StartLengthReporting(10 * time.Second)
	savePodInfoQueue.IsDropEventOnFull = false
	savePodInfoQueue.StartConsumers(100, func(v interface{}) {
		defer utils.IgnorePanic("StartConsumers")

		podOp, ok := v.(*podOpStruct)
		if !ok || podOp == nil {
			return
		}

		_ = xsearch.SavePodInfoToZSearch(podOp.clusterName, podOp.pod, "未开始", podOp.t, podOp.auditID, "", false)
	})
}

func consumePodOpStruct(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("StartConsumers")

	if auditEvent.ResponseRuntimeObj == nil {
		return
	}
	pod, ok := auditEvent.ResponseRuntimeObj.(*corev1.Pod)
	if !ok || pod == nil {
		return
	}

	if string(pod.UID) == "" {
		return
	}

	podOp := &podOpStruct{}
	podOp.clusterName = auditEvent.Annotations["cluster"]
	podOp.pod = pod
	podOp.t = auditEvent.StageTimestamp.Time
	podOp.auditID = string(auditEvent.AuditID)

	if auditEvent.Verb == "create" && auditEvent.ObjectRef.Subresource == "" {
		podOp.isCreate = true
	} else {
		podOp.isCreate = false
	}

	if auditEvent.Verb == "delete" && auditEvent.ObjectRef.Subresource == "" || pod.DeletionTimestamp != nil {
		podOp.isDeleting = true
	} else {
		podOp.isDeleting = false
	}

	if (pod.Finalizers == nil || len(pod.Finalizers) == 0) && pod.DeletionTimestamp != nil && (pod.GetDeletionGracePeriodSeconds() != nil && *pod.GetDeletionGracePeriodSeconds() == 0) {
		podOp.isDeleted = true
	}

	err := xsearch.SavePodYaml(podOp.clusterName, podOp.pod, podOp.t, podOp.auditID, podOp.isDeleting, podOp.isDeleted)
	klog.V(6).Infof("save pod yaml for %s, deleted: %t, err==nil: %t, err: %v\n", podOp.pod.UID, podOp.isDeleted, err == nil, err)
	savePodInfoQueue.Produce(podOp)
}
