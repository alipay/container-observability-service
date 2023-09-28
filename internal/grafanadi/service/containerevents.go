package service

import (
	"strings"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

var (
	RestartFlag        string = "restartCount:"
	NotRestartFlag     string = "restartCount:0"
	RestartStatus      string = "Restart"
	ContainerEventTags string = "Container Events"
)

func AddStatusFromPodLifePhase(podPhases []*storagemodel.LifePhase) []model.AnnotationResponse {
	annotationSlice := make([]model.AnnotationResponse, 0)
	for _, v := range podPhases {
		if strings.Contains(v.OperationName, RestartFlag) && !strings.Contains(v.OperationName, NotRestartFlag) && !v.StartTime.IsZero() {
			annotationSlice = append(annotationSlice, model.AnnotationResponse{
				// Grafana expects unix milliseconds:
				// https://github.com/grafana/simple-json-datasource#annotation-api
				Time:  v.StartTime.Unix() * 1000,
				Title: RestartFlag,
				Text:  RestartStatus,
				Tags:  ContainerEventTags,
			})
		}
	}
	return annotationSlice
}
