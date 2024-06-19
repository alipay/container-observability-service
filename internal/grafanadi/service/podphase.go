package service

import (
	"fmt"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/utils"
)

const (
	traceStageAdmission  = "准入阶段"
	traceStageDecision   = "决策阶段"
	traceStageScheduling = "调度阶段"
	traceStageRunning    = "运行阶段"
)

var (
	reasonToTraceStageMap = map[string]string{
		"Scheduled": traceStageScheduling,

		"Pulling": traceStageRunning,
		"Pulled":  traceStageRunning,
	}

	operationNameToTraceStageMap = map[string]string{
		"condition:Ready:true": traceStageRunning,

		"apicreate": traceStageAdmission,
		"scheduled": traceStageScheduling,

		"Enters default-scheduler": traceStageScheduling,

		"pod phase: Running": "完成交付",
	}
	GrafanaUrl string
)

func ConvertPodPhase2Frame(podPhases []*storagemodel.LifePhase) model.DataFrame {
	timeAry := []time.Time{}
	tagsAry := []string{}
	opsAry := []string{}
	typeAry := []string{}
	uaAry := []string{}
	stateAry := []string{}
	plfAry := []string{}
	traceStageAry := []string{}
	reasonAry := []string{}
	messageAry := []string{}

	for _, phase := range podPhases {
		if phase.StartTime.IsZero() {
			continue
		}
		timeAry = append(timeAry, phase.StartTime)
		op := phase.OperationName
		t := "Operation"

		state := "Info"
		if phase.HasErr {
			state = "ERROR!!"
		}
		dic := make(map[string]interface{})
		dic["operationName"] = phase.OperationName
		var userAgent string
		var phaseMessage string
		var phaseReason string
		var currentTraceStage string
		if phase.ExtraInfo != nil {
			agent := phase.ExtraInfo.(map[string]interface{})["UserAgent"]
			if agent != "" && agent != nil {
				userAgent = agent.(string)
			}
			message := phase.ExtraInfo.(map[string]interface{})["Message"]
			if message != nil && message != "" {
				phaseMessage = message.(string)
			}
			ext := phase.ExtraInfo.(map[string]interface{})["eventObject"]
			if ext != nil {
				reason := ext.(map[string]interface{})["reason"]
				if reason != "" && reason != nil {
					dic["reason"] = reason
					phaseReason = reason.(string)
				}
				msg := ext.(map[string]interface{})["message"]
				if msg != "" && msg != nil {
					phaseMessage = msg.(string)
				}
				op = msg.(string)
			}
			ext = phase.ExtraInfo.(map[string]interface{})["auditEvent.ResponseObject"]
			if ext != nil {
				auditMessage := ext.(map[string]interface{})["message"]
				if auditMessage != nil && auditMessage != "" {
					phaseMessage = auditMessage.(string)
				}
			}
			currentTraceStage = phase.TraceStage
			var traceStage string = "abc"
			var ok bool
			if dic["reason"] != nil {
				traceStage, ok = reasonToTraceStageMap[dic["reason"].(string)]
				if !ok {
					if dic["operationName"] != nil {
						traceStage, ok = operationNameToTraceStageMap[dic["operationName"].(string)]
						if !ok {
							traceStage = "ToBeFilled"
						}
					} else {
						traceStage = "ToBeFilled"
					}
				}
			} else if dic["operationName"] != nil {
				traceStage, ok = operationNameToTraceStageMap[dic["operationName"].(string)]
				if !ok {
					traceStage = "ToBeFilled"
				}
			}

			if traceStage != "ToBeFilled" {
				currentTraceStage = traceStage
			}

		}
		if len(op) == 0 {
			op = phaseReason
			t = "Event"
		}
		tags := fmt.Sprintf("\"%s\" UA=\"%s\" Level=\"%s\"", op, userAgent, state)
		stateAry = append(stateAry, convertNil(state))
		tagsAry = append(tagsAry, convertNil(tags))
		opsAry = append(opsAry, convertNil(op))
		typeAry = append(typeAry, t)

		uaAry = append(uaAry, convertNil(userAgent))
		hashCodeStr := fmt.Sprintf("%d", utils.StringHashcode(phase.OperationName))
		plfId := phase.ClusterName + "_" + phase.Namespace + "_" + phase.PodUID + "_" + phase.DataSourceId + "_" + hashCodeStr
		plfUrl := fmt.Sprintf("http://%s/d/rawdatalinks/rawdata?orgId=1&var-plfid=%s", GrafanaUrl, plfId)
		plfAry = append(plfAry, convertNil(plfUrl))
		reasonAry = append(reasonAry, convertNil(phaseReason))
		messageAry = append(messageAry, convertNil(phaseMessage))
		traceStageAry = append(traceStageAry, convertNil(currentTraceStage))

	}

	return model.DataFrame{
		Schema: model.SchemaType{
			Fields: []model.FieldType{
				{Name: "StartTime", Type: "time"},
				{Name: "Tags", Type: "string"},
				{Name: "Type", Type: "string"},
				{Name: "OperationName", Type: "string"},
				{Name: "UserAgent", Type: "string"},
				{Name: "TraceStage", Type: "string"},
				{Name: "State", Type: "string"},
				{Name: "Reason", Type: "string"},
				{Name: "Message", Type: "string"},
				{Name: "PlfID", Type: "string"},
			},
		},
		Data: model.DataType{
			Values: []interface{}{
				timeAry, tagsAry, typeAry, opsAry, uaAry, traceStageAry, stateAry, reasonAry, messageAry, plfAry,
			},
		},
	}
}
