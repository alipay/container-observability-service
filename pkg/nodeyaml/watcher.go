package nodeyaml

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/shares"

	"github.com/alipay/container-observability-service/pkg/queue"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	corev1 "k8s.io/api/core/v1"
)

var (
	Queue *queue.BoundedQueue
)

type nodeOpStruct struct {
	clusterName string
	node        *corev1.Node
	t           time.Time
	auditID     string
}

func init() {
	Queue = queue.NewBoundedQueue("nodeyaml-watcher", 100000, nil)
	Queue.StartLengthReporting(10 * time.Second)
	Queue.StartConsumers(1, processAuditEvent)
	Queue.IsDropEventOnFull = false
	Queue.SetFilterItemFunc(func(item interface{}) bool {
		if item == nil {
			return true
		}

		auditEvent, ok := item.(*shares.AuditEvent)
		if !ok || auditEvent == nil {
			return true
		}

		if auditEvent.ObjectRef.Resource != "nodes" {
			return true
		}

		if auditEvent.ResponseStatus.Code >= 300 {
			return true
		}

		return false
	})
}

func processAuditEvent(v interface{}) {
	defer utils.IgnorePanic("StartConsumers")

	if v == nil {
		return
	}

	auditEvent, ok := v.(*shares.AuditEvent)
	if !ok || auditEvent == nil {
		return
	}
	auditEvent.Wait()

	if auditEvent.ObjectRef.Resource != "nodes" {
		return
	}

	if auditEvent.ResponseStatus.Code >= 300 {
		return
	}
	if auditEvent.ResponseRuntimeObj == nil {
		return
	}

	var node *corev1.Node
	ok = false
	if node, ok = auditEvent.ResponseRuntimeObj.(*corev1.Node); !ok {
		return
	}

	/*if err := json.Unmarshal(auditEvent.ResponseObject.Raw, node); err != nil {
		return
	}*/
	if string(node.UID) == "" {
		return
	}

	nodeOp := &nodeOpStruct{}
	nodeOp.clusterName = auditEvent.Annotations["cluster"]
	nodeOp.node = node
	nodeOp.t = auditEvent.StageTimestamp.Time
	nodeOp.auditID = string(auditEvent.AuditID)

	_ = xsearch.SaveNodeYaml(nodeOp.clusterName, nodeOp.node, nodeOp.t, nodeOp.auditID)
}
