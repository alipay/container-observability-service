package pods

import (
	"regexp"
	"strings"
	"time"

	"github.com/alipay/container-observability-service/pkg/reason/modules"
	"github.com/alipay/container-observability-service/pkg/reason/share"
	"github.com/alipay/container-observability-service/pkg/reason/utils"
	"github.com/alipay/container-observability-service/pkg/shares"
	utils2 "github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

var (
	InvalidMountConfig  = "InvalidMountConfig"  // 用户自身问题  hostPath卷配置错误
	AccessError         = "AccessError"         // 权限错误or延迟
	NFSError            = "NFSError"            // NFS版本或转换协议不支持
	NoSpaceLeftOnDevice = "NoSpaceLeftOnDevice" // 磁盘空间不足

	FailedMount = "FailedMount" // 其他原因
)

func init() {
	modules.ShareModuleFactory.Register(share.VOLUME, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.VOLUME, VolumeReason)
	})
}

func VolumeReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	podYaml := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if podYaml == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_volume_mount")

	SuccessfullyMountedVolumes := map[string]bool{}

	eventLen := len(auditEvents)
	for idx := eventLen - 1; idx >= 0; idx-- {
		hyEvent := auditEvents[idx]
		if endTime != nil && hyEvent.Event.StageTimestamp.After(*endTime) {
			continue
		}

		reason := ""
		if hyEvent.Reason != "" {
			reason = hyEvent.Reason
		}

		msg := ""
		ev, ok := hyEvent.ResponseRuntimeObj.(*v1.Event)
		if ok && ev != nil {
			msg = ev.Message
		}

		//v1.16版本
		if reason == "SuccessfulCreatePodSandBox" {
			return "", false
		}
		//1.14版本没有明确信息，可以依赖是否开始镜像操作
		if reason == "Pulling" || reason == "Pulled" {
			return "", false
		}

		// 1.16后若EnableEventEnhancement 不报此reason
		if reason == "SuccessfulAttachOrMountVolume" {
			return "", false
		}

		//regUnmounted := regexp.MustCompile("unmounted volumes=\\[(.*?)\\]")
		//unmountedVolumes := regUnmounted.FindStringSubmatch(msg)

		if reason != "MountVolume" {
			continue
		}

		var volume string
		regVolume := regexp.MustCompile("\\[(.*?)\\]")
		volumeMatches := regVolume.FindStringSubmatch(msg)
		if len(volumeMatches) >= 2 {
			volume = volumeMatches[1]
		}

		if reason == "MountVolume" && strings.Contains(msg, "Successfully mounted") {
			SuccessfullyMountedVolumes[volume] = true
			continue
		}

		if reason == "MountVolume" && strings.Contains(msg, "Failed mounted") {
			// 如果已经成功挂载，跳过本次分析
			if SuccessfullyMountedVolumes[volume] == true {
				continue
			}

			if strings.Contains(msg, "hostPath type check failed") {
				return InvalidMountConfig, true
			}
			if strings.Contains(msg, "no relationship found between node") {
				return AccessError, true
			}
			if strings.Contains(msg, "requested NFS version or transport protocol is not supported") {
				return NFSError, true
			}
			if strings.Contains(msg, "no space left on device") {
				return NoSpaceLeftOnDevice, true
			}
			return FailedMount, true
		}

	}
	return "", false
}
