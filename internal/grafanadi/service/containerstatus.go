package service

import (
	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

var (
	ContainerStatusTags string = "Container Status"
	PodScheduled        string = "Scheduled"
	Running             string = "Running"
	Succeed             string = "Succeed"
	Failed              string = "Failed"
	Ready               string = "Ready"
	NotReady            string = "NotReady"
	Deleting            string = "Deleting"
)

func AddStatusFromSloData(sloList []*storagemodel.SloTraceData, lifePhases []*storagemodel.LifePhase) []model.AnnotationResponse {
	annotationSlice := make([]model.AnnotationResponse, 0)
	for _, slo := range sloList {
		if !slo.Scheduled.IsZero() && slo.Type == "create" {
			annotationSlice = append(annotationSlice, model.AnnotationResponse{
				// Grafana expects unix milliseconds:
				// https://github.com/grafana/simple-json-datasource#annotation-api
				Time:  slo.Scheduled.Unix() * 1000,
				Title: PodScheduled,
				Text:  PodScheduled,
				Tags:  ContainerStatusTags,
			})
		}
		if !slo.ContainersReady.IsZero() && slo.Type == "create" {
			annotationSlice = append(annotationSlice, model.AnnotationResponse{
				// Grafana expects unix milliseconds:
				// https://github.com/grafana/simple-json-datasource#annotation-api
				Time:  slo.ContainersReady.Unix() * 1000,
				Title: Ready,
				Text:  Ready,
				Tags:  ContainerStatusTags,
			})
		}
		if !slo.RunningAt.IsZero() && slo.Type == "create" {
			annotationSlice = append(annotationSlice, model.AnnotationResponse{
				// Grafana expects unix milliseconds:
				// https://github.com/grafana/simple-json-datasource#annotation-api
				Time:  slo.RunningAt.Unix() * 1000,
				Title: Running,
				Text:  Running,
				Tags:  ContainerStatusTags,
			})
		}
		if !slo.SucceedAt.IsZero() && slo.Type == "create" {
			annotationSlice = append(annotationSlice, model.AnnotationResponse{
				// Grafana expects unix milliseconds:
				// https://github.com/grafana/simple-json-datasource#annotation-api
				Time:  slo.SucceedAt.Unix() * 1000,
				Title: Succeed,
				Text:  Succeed,
				Tags:  ContainerStatusTags,
			})
		}
		if !slo.FailedAt.IsZero() && slo.Type == "create" {
			annotationSlice = append(annotationSlice, model.AnnotationResponse{
				// Grafana expects unix milliseconds:
				// https://github.com/grafana/simple-json-datasource#annotation-api
				Time:  slo.FailedAt.Unix() * 1000,
				Title: Failed,
				Text:  Failed,
				Tags:  ContainerStatusTags,
			})
		}
		if !slo.ReadyAt.IsZero() && slo.Type == "create" {
			annotationSlice = append(annotationSlice, model.AnnotationResponse{
				// Grafana expects unix milliseconds:
				// https://github.com/grafana/simple-json-datasource#annotation-api
				Time:  slo.ReadyAt.Unix() * 1000,
				Title: Ready,
				Text:  Ready,
				Tags:  ContainerStatusTags,
			})
		}
		if !slo.CreatedTime.IsZero() && slo.Type == "delete" {
			annotationSlice = append(annotationSlice, model.AnnotationResponse{
				// Grafana expects unix milliseconds:
				// https://github.com/grafana/simple-json-datasource#annotation-api
				Time:  slo.CreatedTime.Unix() * 1000,
				Title: Deleting,
				Text:  Deleting,
				Tags:  ContainerStatusTags,
			})
		}
	}

	for _, v := range lifePhases {
		if v.OperationName == "condition:Ready:true" {
			annotationSlice = append(annotationSlice, model.AnnotationResponse{
				// Grafana expects unix milliseconds:
				// https://github.com/grafana/simple-json-datasource#annotation-api
				Time:  v.StartTime.Unix() * 1000,
				Title: Ready,
				Text:  Ready,
				Tags:  ContainerStatusTags,
			})
		}
		if v.OperationName == "condition:Ready:false" {
			annotationSlice = append(annotationSlice, model.AnnotationResponse{
				// Grafana expects unix milliseconds:
				// https://github.com/grafana/simple-json-datasource#annotation-api
				Time:  v.StartTime.Unix() * 1000,
				Title: NotReady,
				Text:  NotReady,
				Tags:  ContainerStatusTags,
			})
		}
	}
	return annotationSlice
}
