/*
*
消费WatcherQueue中的审计日志
*/
package podphase

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/shares"

	"github.com/alipay/container-observability-service/pkg/queue"
)

var (
	WatcherQueue *queue.BoundedQueue
)

func init() {
	WatcherQueue = queue.NewBoundedQueue("podphase-watcher", 100000, nil)
	WatcherQueue.StartLengthReporting(10 * time.Second)
	WatcherQueue.IsDropEventOnFull = false
	//WatcherQueue.IsLockOSThread = true
	WatcherQueue.StartConsumers(100, func(item interface{}) {
		if item == nil {
			return
		}
		auditEvent, ok := item.(*shares.AuditEvent)
		if !ok || auditEvent == nil {
			return
		}
		auditEvent.Wait()

		// pods资源
		if auditEvent.ObjectRef.Resource == "pods" {
			if auditEvent.Verb == "create" {
				if auditEvent.ObjectRef.Subresource == "" {
					processPodCreation(auditEvent)
				} else if auditEvent.ObjectRef.Subresource == "binding" {
					processPodBinding(auditEvent)
				}
			} else if auditEvent.Verb == "delete" {
				if auditEvent.ObjectRef.Subresource == "" {
					processPodDeletion(auditEvent)
				}
			} else if auditEvent.Verb == "patch" {
				processPodPatch(auditEvent)
			} else if auditEvent.Verb == "update" {
				processPodUpdate(auditEvent)
			}

			return
		}

		// events资源
		if auditEvent.ObjectRef.Resource == "events" {
			if auditEvent.Verb == "create" {
				if auditEvent.ObjectRef.Subresource == "" {
					processPodEventCreation(auditEvent)
				}
			} else if auditEvent.Verb == "patch" {
				processPodEventPatch(auditEvent)
			}
		}
	})
}
