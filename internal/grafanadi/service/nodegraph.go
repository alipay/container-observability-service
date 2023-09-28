package service

import (
	"fmt"
	"strings"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
)

func ConvertPodGraph2Frame(podYamls []*storagemodel.PodYaml) ([]model.PodYamlGraphNodes, []model.PodYamlGraphEdges) {
	var nodes []model.PodYamlGraphNodes
	var edges []model.PodYamlGraphEdges

	for _, py := range podYamls {
		podId := py.PodUid
		podReady := py.Pod.Status.Phase
		if py.Pod.DeletionTimestamp != nil {
			podReady = "Terminating/Terminated"
		}
		readyCondition := 0
		for _, c := range py.Pod.Status.Conditions {
			if strings.Contains(string(c.Type), "Pressure") {
				if c.Status == v1.ConditionFalse {
					readyCondition = readyCondition + 1
				}
			} else {
				if c.Status == v1.ConditionTrue {
					readyCondition = readyCondition + 1
				}
			}
		}

		finalizerNum := len(py.Pod.Finalizers)

		secondStat := fmt.Sprintf("%d/%d condition", readyCondition, len(py.Pod.Status.Conditions))
		if py.Pod.DeletionTimestamp != nil {
			secondStat = fmt.Sprintf("%d finalizers", finalizerNum)
		}

		detail := fmt.Sprintf("%s, %d/%d condition, %d finalizers", string(podReady), readyCondition, len(py.Pod.Status.Conditions), len(py.Pod.Finalizers))
		b := float64(0)
		g := float64(1.0)
		y := float64(0)
		r := float64(0)

		if py.Pod.DeletionTimestamp == nil {
			if readyCondition != len(py.Pod.Status.Conditions) {
				r = float64(readyCondition) / float64(len(py.Pod.Status.Conditions))
				g = 0
			}
		} else {
			if len(py.Pod.Finalizers) != 0 {
				r = 1
				g = 0
			} else {
				g = 0
				y = 1
			}
		}
		podNode := model.PodYamlGraphNodes{
			Id: podId, Title: "Pod", SubTitle: py.PodName, MainStat: string(podReady), SecondaryStat: secondStat, Detail_one: detail, Blue: b, Green: g, Yellow: y, Red: r,
		}
		nodes = append(nodes, podNode)
		if len(py.Pod.OwnerReferences) != 0 {
			for _, own := range py.Pod.OwnerReferences {
				ownerNode := model.PodYamlGraphNodes{
					Id: string(own.UID), Title: own.Kind, SubTitle: own.Name, MainStat: "Unknown", SecondaryStat: "Unknown", Detail_one: string(own.UID), Blue: 1.0, Green: 0.0, Yellow: 0.0, Red: 0.0,
				}
				nodes = append(nodes, ownerNode)
				edges = append(edges, model.PodYamlGraphEdges{
					Id: uuid.New().String(), Target: ownerNode.Id, Source: podId, MainStat: "OwnnerReference",
				})

			}

		}
		// containers
		for _, c := range py.Pod.Spec.Containers {
			cready := "NotReady"
			restartCnt := 0
			uid := ""
			for _, cs := range py.Pod.Status.ContainerStatuses {
				if cs.Name == c.Name {
					if cs.Ready {
						cready = "Ready"
					}
					restartCnt = int(cs.RestartCount)
					uid = cs.ContainerID
				}
			}
			restarts := fmt.Sprintf("%d restarts", restartCnt)

			detail := fmt.Sprintf("Image: %s", c.Image)
			b := 0.0
			g := 1.0
			y := 0.0
			r := 0.0

			if cready != "Ready" {
				g = 0.0
				r = 1.0
			}

			if restartCnt > 0 {
				y = 1.0
				g = 0.0
			}

			cNode := model.PodYamlGraphNodes{
				Id: uid, Title: "Container", SubTitle: c.Name, MainStat: cready, SecondaryStat: restarts, Detail_one: detail, Blue: b, Green: g, Yellow: y, Red: r,
			}

			nodes = append(nodes, cNode)
			edges = append(edges, model.PodYamlGraphEdges{
				Id: uuid.New().String(), Target: podId, Source: cNode.Id, MainStat: "Container",
			})

		}
	}

	return nodes, edges

}
func ConvertPodYamlGraphNodes2Frame(nodes []model.PodYamlGraphNodes) model.DataFrame {
	var idAry, titleAry, subtitleAry, mainStatAry, secStatAry, detailAry []string
	var blueAry, greenAry, yellowAry, redAry []float64
	if len(nodes) == 0 {
		return model.DataFrame{}
	}
	for _, node := range nodes {
		idAry = append(idAry, node.Id)
		titleAry = append(titleAry, node.Title)
		subtitleAry = append(subtitleAry, node.SubTitle)
		mainStatAry = append(mainStatAry, node.MainStat)
		secStatAry = append(secStatAry, node.SecondaryStat)
		detailAry = append(detailAry, node.Detail_one)
		blueAry = append(blueAry, node.Blue)
		greenAry = append(greenAry, node.Green)
		yellowAry = append(yellowAry, node.Yellow)
		redAry = append(redAry, node.Red)
	}

	return model.DataFrame{
		Schema: model.SchemaType{
			Fields: []model.FieldType{
				{Name: "id", Type: "string"},
				{Name: "title", Type: "string"},
				{Name: "subtitle", Type: "string"},
				{Name: "mainStat", Type: "string"},
				{Name: "secondaryStat", Type: "string"},
				{Name: "detail__one", Type: "string"},
				{Name: "arc__blue", Type: "float64"},
				{Name: "arc__green", Type: "float64"},
				{Name: "arc__yellow", Type: "float64"},
				{Name: "arc__red", Type: "float64"},
			},
		},
		Data: model.DataType{
			Values: []interface{}{
				idAry, titleAry, subtitleAry, mainStatAry, secStatAry, detailAry, blueAry, greenAry, yellowAry, redAry,
			},
		},
	}

}
func ConvertPodYamlGraphEdges2Frame(edges []model.PodYamlGraphEdges) model.DataFrame {
	var idAry, sourceAry, targetAry, mainStatAry []string
	if len(edges) == 0 {
		return model.DataFrame{}
	}
	for _, edge := range edges {
		idAry = append(idAry, edge.Id)
		sourceAry = append(sourceAry, edge.Source)
		targetAry = append(targetAry, edge.Target)
		mainStatAry = append(mainStatAry, edge.MainStat)
	}

	return model.DataFrame{
		Schema: model.SchemaType{
			Fields: []model.FieldType{
				{Name: "id", Type: "string"},
				{Name: "source", Type: "string"},
				{Name: "target", Type: "string"},
				{Name: "mainStat", Type: "string"},
			},
		},
		Data: model.DataType{
			Values: []interface{}{
				idAry, sourceAry, targetAry, mainStatAry,
			},
		},
	}

}
