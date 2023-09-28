package mocks

import (
	"encoding/json"
	"time"

	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FakeStorage struct{}

func (f *FakeStorage) QuerySpanWithPodUid(data interface{}, uid string) error {
	if uid == "123" {
		res := make([]*model.Span, 0)
		res = append(res, &model.Span{
			Name: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}
		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if uid == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryLifePhaseWithPodUid(data interface{}, uid string) error {
	if uid == "123" {
		res := make([]*model.Span, 0)
		res = append(res, &model.Span{
			Name: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if uid == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryPodYamlsWithPodName(data interface{}, podName string) error {
	if podName == "abcdef" {
		res := make([]*model.PodYaml, 0)
		res = append(res, &model.PodYaml{
			PodName: "abcdef",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if podName == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryPodYamlsWithHostName(data interface{}, hostName string) error {
	if hostName == "123" {
		res := make([]*model.PodYaml, 0)
		res = append(res, &model.PodYaml{
			Hostname: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if hostName == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryPodYamlsWithPodIp(data interface{}, podIp string) error {
	if podIp == "123" {
		res := make([]*model.PodYaml, 0)
		res = append(res, &model.PodYaml{
			PodIP: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if podIp == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryPodListWithNodeip(data interface{}, nodeIp string, isDeleted bool) error {
	if nodeIp == "123" {
		res := make([]*model.PodYaml, 0)
		res = append(res, &model.PodYaml{
			HostIP: "123",
			PodUid: "12345",
			Pod:    &v1.Pod{Status: v1.PodStatus{PodIP: "abcdef"}},
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if nodeIp == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryPodYamlsWithPodUID(data interface{}, podUid string) error {
	if podUid == "123" {
		res := make([]*model.PodYaml, 0)
		res = append(res, &model.PodYaml{
			PodUid: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if podUid == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryPodUIDListByHostname(data interface{}, hostName string) error {
	if hostName == "123" {
		res := make([]*model.PodYaml, 0)
		res = append(res, &model.PodYaml{
			PodUid: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if hostName == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryPodUIDListByPodIP(data interface{}, podIp string) error {
	if podIp == "123" {
		res := make([]*model.PodYaml, 0)
		res = append(res, &model.PodYaml{
			PodUid: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if podIp == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryPodUIDListByPodName(data interface{}, podName string) error {
	if podName == "123" {
		res := make([]*model.PodYaml, 0)
		res = append(res, &model.PodYaml{
			PodUid: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if podName == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryNodeYamlsWithNodeName(data interface{}, nodeName string) error {
	if nodeName == "123" {
		res := make([]*model.NodeYaml, 0)
		res = append(res, &model.NodeYaml{
			NodeName: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if nodeName == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryNodeYamlsWithNodeUid(data interface{}, nodeUid string) error {
	if nodeUid == "123" {
		res := make([]*model.NodeYaml, 0)
		res = append(res, &model.NodeYaml{
			UID: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if nodeUid == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryNodeUIDListWithNodeIp(data interface{}, nodeIp string) error {
	if nodeIp == "123" {
		res := make([]*model.NodeYaml, 0)
		res = append(res, &model.NodeYaml{
			NodeIp: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if nodeIp == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryNodeYamlsWithNodeIP(data interface{}, nodeIp string) error {
	if nodeIp == "123" {
		res := make([]*model.NodeYaml, 0)
		res = append(res, &model.NodeYaml{
			NodeIp: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if nodeIp == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryPodYamlsWithNodeIP(data interface{}, nodeIp string) error {
	if nodeIp == "123" {
		res := make([]*model.PodYaml, 0)
		res = append(res, &model.PodYaml{
			PodIP: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if nodeIp == "fake" {
		return errors.New("fake uid")
	}

	return nil

}

func (f *FakeStorage) QueryPodInfoWithPodUid(data interface{}, podUid string) error {
	if podUid == "123" {
		res := make([]*model.PodInfo, 0)
		res = append(res, &model.PodInfo{
			PodUID: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if podUid == "fake" {
		return errors.New("fake uid")
	}

	return nil
}
func (f *FakeStorage) QueryNodephaseWithNodeName(data interface{}, nodeName string) error {
	if nodeName == "123" {
		res := make([]*model.NodeLifePhase, 0)
		res = append(res, &model.NodeLifePhase{
			NodeName: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if nodeName == "fake" {
		return errors.New("fake uid")
	}

	return nil

}
func (f *FakeStorage) QueryNodephaseWithNodeUID(data interface{}, nodeUid string) error {
	if nodeUid == "123" {
		res := make([]*model.NodeLifePhase, 0)
		res = append(res, &model.NodeLifePhase{
			UID: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if nodeUid == "fake" {
		return errors.New("fake uid")
	}

	return nil

}

func (f *FakeStorage) QuerySloTraceDataWithPodUID(data interface{}, podUid string) error {
	if podUid == "12345" {
		res := make([]*model.SloTraceData, 0)
		res = append(res, &model.SloTraceData{
			NodeIP: "12345",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if podUid == "fake" {
		return errors.New("fake uid")
	}

	return nil
}
func (f *FakeStorage) QueryCreateSloWithResult(data interface{}, requestParams *model.SloOptions) error {
	if requestParams.BizName == "12345" || requestParams.Type == "create" {
		res := make([]*model.Slodata, 0)
		res = append(res, &model.Slodata{
			NodeIP:  "12345",
			Cluster: "cluster",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if requestParams.BizName == "fake" {
		return errors.New("fake uid")
	}

	return nil
}
func (s *FakeStorage) QueryUpgradeSloWithResult(data interface{}, requestParams *model.SloOptions) error {
	if requestParams.BizName == "12345" || requestParams.Type == "create" {
		res := make([]*model.Slodata, 0)
		res = append(res, &model.Slodata{
			NodeIP:  "12345",
			Cluster: "cluster",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if requestParams.BizName == "fake" {
		return errors.New("fake uid")
	}

	return nil
}
func (s *FakeStorage) QueryDeleteSloWithResult(data interface{}, requestParams *model.SloOptions) error {
	if requestParams.BizName == "12345" {
		res := make([]*model.Slodata, 0)
		res = append(res, &model.Slodata{
			NodeIP: "12345",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if requestParams.BizName == "fake" {
		return errors.New("fake uid")
	}

	return nil
}
func (s *FakeStorage) QueryNodeYamlWithParams(data interface{}, requestParams *model.NodeParams) error {
	if requestParams.NodeName == "12345" {
		res := make([]*model.NodeYaml, 0)
		res = append(res, &model.NodeYaml{
			NodeIp: "12345",
			Node: &v1.Node{Status: v1.NodeStatus{Conditions: []v1.NodeCondition{{Type: "Ready"}}}, ObjectMeta: metav1.ObjectMeta{Name: "abcdef",
				Labels: map[string]string{"lunettes/node-sn": "ad",
					"machine.lunettes.com/biz-name": "abdba"},

				Annotations: map[string]string{"remedy.k8s.lunettes.com/remedy-state": "hello"}}, Spec: v1.NodeSpec{Taints: []v1.Taint{
				{Key: "nodeStruct.Node.Spec.Taints"},
			}}},
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if requestParams.NodeName == "fake" {
		return errors.New("fake uid")
	}

	return nil
}
func (s *FakeStorage) QueryAuditWithAuditId(data interface{}, auditid string) error {

	if auditid == "123" {
		res := make([]*model.Audit, 0)
		res = append(res, &model.Audit{
			AuditId: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if auditid == "fake" {
		return errors.New("fake uid")
	}

	return nil
}
func (s *FakeStorage) QueryEventPodsWithPodUid(data interface{}, auditid string) error {

	if auditid == "123" {
		res := make([]*model.Audit, 0)
		res = append(res, &model.Audit{
			AuditId: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if auditid == "fake" {
		return errors.New("fake uid")
	}

	return nil
}
func (s *FakeStorage) QueryEventNodeWithPodUid(data interface{}, auditid string) error {

	if auditid == "123" {
		res := make([]*model.Audit, 0)
		res = append(res, &model.Audit{
			AuditId: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if auditid == "fake" {
		return errors.New("fake uid")
	}

	return nil
}
func (s *FakeStorage) QueryEventWithTimeRange(data interface{}, from, to time.Time) error {

	if !from.IsZero() {
		res := make([]*model.Audit, 0)
		res = append(res, &model.Audit{
			AuditId: "123",
		})
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if from.IsZero() {
		return errors.New("fake uid")
	}

	return nil
}
func (s *FakeStorage) QueryPodYamlWithParams(data interface{}, requestParams *model.PodParams) error {
	if requestParams.Name == "12345" {
		res := make([]*model.PodYaml, 0)
		pod1 := &model.PodYaml{
			PodUid:      "12345",
			ClusterName: "cluster1",
			Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "abcdef", CreationTimestamp: metav1.NewTime(time.Now().Add(time.Second * 10)),
				Annotations: map[string]string{"remedy.k8s.lunettes.com/remedy-state": "hello"}}},
		}
		pod2 := &model.PodYaml{
			PodUid:      "54321",
			ClusterName: "cluster2",

			CreationTimestamp: time.Now().Add(time.Second),
			Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "abcdef", CreationTimestamp: metav1.NewTime(time.Now().Add(time.Second)),
				Annotations: map[string]string{"remedy.k8s.lunettes.com/remedy-state": "hello"}}},
		}
		pod3 := &model.PodYaml{
			PodUid:            "12345",
			ClusterName:       "cluster3",
			CreationTimestamp: time.Now().Add(time.Second * 2),
			Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "abcdef", CreationTimestamp: metav1.NewTime(time.Now().Add(time.Second * 2)),
				Annotations: map[string]string{"remedy.k8s.lunettes.com/remedy-state": "hello"}}},
		}
		pod4 := &model.PodYaml{
			PodUid:            "15",
			ClusterName:       "cluster4",
			CreationTimestamp: time.Now().Add(time.Second * 13),
			Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "abcdef", CreationTimestamp: metav1.NewTime(time.Now().Add(time.Second * 14)),
				Annotations: map[string]string{"remedy.k8s.lunettes.com/remedy-state": "hello"}}},
		}
		res = append(res, pod1, pod2, pod3, pod4)
		resStr, err := json.Marshal(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resStr, data)
		if err != nil {
			return err
		}

	} else if requestParams.Name == "fake" {
		return errors.New("fake uid")
	}

	return nil
}
