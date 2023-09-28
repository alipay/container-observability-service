package slo

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/shares"

	"github.com/alipay/container-observability-service/pkg/queue"
)

var (
	Queue       *queue.BoundedQueue
	ClusterName string
)

const (
	POD_UPGRADE = "pod_upgrade"
	POD_DELETE  = "pod_delete"
	POD_CREATE  = "pod_create"
	PVC_CREATE  = "pvc_create"
)

func init() {
	Queue = queue.NewBoundedQueue("slo-watcher", 200000, nil)
	Queue.StartLengthReporting(10 * time.Second)
	Queue.IsDropEventOnFull = false
	Queue.StartConsumers(1, func(v interface{}) {
		if v == nil {
			return
		}
		event, ok := v.(*shares.AuditEvent)
		if !ok || event == nil {
			return
		}

		event.CanProcess(shares.SLOProcessNode)
		//删除Pod SLO
		deleteQueue.Produce(event)
		//Upgrade SLO
		upgradeQueue.Produce(event)
		//Create SLO
		createQueue.Produce(event)
		//pvc create
		pvcCreateQueue.Produce(event)
		//notify children
		event.FinishProcess(shares.SLOProcessNode)
	})
}
