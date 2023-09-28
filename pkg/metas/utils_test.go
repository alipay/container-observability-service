package metas

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"

	k8s_audit "k8s.io/apiserver/pkg/apis/audit"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	normalEvent = `
	{
		"log": {
		  "file": {
			"path": "/audit/audit.log"
		  }
		},
		"annotations": {
		  "authorization.k8s.io/decision": "allow",
		  "cluster": "staging",
		  "authorization.k8s.io/reason": ""
		},
		"source": "/audit/audit.log",
		"responseObject": {
		  "metadata": {
			"deletionGracePeriodSeconds": 0,
			"uid": "33d7d262-ce52-43ef-bc68-8db6ef8a813d",
			"managedFields": [
			  {
				"apiVersion": "v1",
				"fieldsV1": {
				  "f:metadata": {
					"f:labels": {
					  "f:pod-template-hash": {},
					  "f:app": {},
					  ".": {}
					},
					"f:ownerReferences": {
					  "k:{\"uid\":\"f89a8bc6-a112-4ef4-864b-fb7345c3950c\"}": {
						"f:uid": {},
						"f:controller": {},
						"f:apiVersion": {},
						"f:kind": {},
						"f:name": {},
						"f:blockOwnerDeletion": {},
						".": {}
					  },
					  ".": {}
					},
					"f:generateName": {}
				  },
				  "f:spec": {
					"f:enableServiceLinks": {},
					"f:securityContext": {},
					"f:restartPolicy": {},
					"f:schedulerName": {},
					"f:containers": {
					  "k:{\"name\":\"busybox\"}": {
						"f:imagePullPolicy": {},
						"f:terminationMessagePath": {},
						"f:terminationMessagePolicy": {},
						"f:image": {},
						"f:command": {},
						"f:name": {},
						".": {},
						"f:resources": {
						  "f:limits": {
							"f:cpu": {},
							".": {}
						  },
						  "f:requests": {
							"f:cpu": {},
							".": {}
						  },
						  ".": {}
						}
					  }
					},
					"f:dnsPolicy": {},
					"f:terminationGracePeriodSeconds": {}
				  }
				},
				"manager": "kube-controller-manager",
				"time": "2023-06-19T03:07:38Z",
				"operation": "Update",
				"fieldsType": "FieldsV1"
			  },
			  {
				"apiVersion": "v1",
				"fieldsV1": {
				  "f:status": {
					"f:conditions": {
					  "k:{\"type\":\"ContainersReady\"}": {
						"f:message": {},
						"f:lastTransitionTime": {},
						"f:lastProbeTime": {},
						"f:type": {},
						"f:status": {},
						".": {},
						"f:reason": {}
					  },
					  "k:{\"type\":\"Initialized\"}": {
						"f:lastTransitionTime": {},
						"f:lastProbeTime": {},
						"f:type": {},
						"f:status": {},
						".": {}
					  },
					  "k:{\"type\":\"Ready\"}": {
						"f:message": {},
						"f:lastTransitionTime": {},
						"f:lastProbeTime": {},
						"f:type": {},
						"f:status": {},
						".": {},
						"f:reason": {}
					  }
					},
					"f:podIPs": {
					  "k:{\"ip\":\"192.168.220.30\"}": {
						"f:ip": {},
						".": {}
					  },
					  ".": {}
					},
					"f:startTime": {},
					"f:hostIP": {},
					"f:phase": {},
					"f:containerStatuses": {},
					"f:podIP": {}
				  }
				},
				"manager": "kubelet",
				"time": "2023-06-19T03:08:34Z",
				"operation": "Update",
				"fieldsType": "FieldsV1"
			  }
			],
			"resourceVersion": "2829097",
			"name": "busy-b6f7b6db6-c2szh",
			"namespace": "kube-system",
			"creationTimestamp": "2023-06-19T03:07:38Z",
			"generateName": "busy-b6f7b6db6-",
			"labels": {
			  "app": "busybox",
			  "pod-template-hash": "b6f7b6db6"
			},
			"selfLink": "/api/v1/namespaces/kube-system/pods/busy-b6f7b6db6-c2szh",
			"deletionTimestamp": "2023-06-19T03:08:04Z",
			"ownerReferences": [
			  {
				"uid": "f89a8bc6-a112-4ef4-864b-fb7345c3950c",
				"controller": true,
				"apiVersion": "apps/v1",
				"kind": "ReplicaSet",
				"name": "busy-b6f7b6db6",
				"blockOwnerDeletion": true
			  }
			]
		  },
		  "apiVersion": "v1",
		  "kind": "Pod",
		  "spec": {
			"nodeName": "node2",
			"dnsPolicy": "ClusterFirst",
			"terminationGracePeriodSeconds": 30,
			"enableServiceLinks": true,
			"serviceAccountName": "default",
			"volumes": [
			  {
				"name": "default-token-g85vf",
				"secret": {
				  "secretName": "default-token-g85vf",
				  "defaultMode": 420
				}
			  }
			],
			"serviceAccount": "default",
			"securityContext": {},
			"priority": 0,
			"restartPolicy": "Always",
			"tolerations": [
			  {
				"effect": "NoExecute",
				"tolerationSeconds": 300,
				"operator": "Exists",
				"key": "node.kubernetes.io/not-ready"
			  },
			  {
				"effect": "NoExecute",
				"tolerationSeconds": 300,
				"key": "node.kubernetes.io/unreachable",
				"operator": "Exists"
			  }
			],
			"containers": [
			  {
				"imagePullPolicy": "Always",
				"image": "busybox",
				"terminationMessagePolicy": "File",
				"terminationMessagePath": "/dev/termination-log",
				"name": "busybox",
				"resources": {
				  "requests": {
					"cpu": "100m"
				  },
				  "limits": {
					"cpu": "100m"
				  }
				},
				"command": [
				  "sh",
				  "-c",
				  "echo The app is running! && sleep 3600"
				],
				"volumeMounts": [
				  {
					"mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
					"name": "default-token-g85vf",
					"readOnly": true
				  }
				]
			  }
			],
			"schedulerName": "default-scheduler"
		  },
		  "status": {
			"phase": "Running",
			"podIP": "192.168.220.30",
			"containerStatuses": [
			  {
				"image": "busybox:latest",
				"imageID": "docker-pullable://busybox@sha256:5acba83a746c7608ed544dc1533b87c737a0b0fb730301639a0179f9344b1678",
				"restartCount": 0,
				"ready": false,
				"name": "busybox",
				"started": false,
				"state": {
				  "terminated": {
					"reason": "Error",
					"exitCode": 137,
					"startedAt": "2023-06-19T03:07:55Z",
					"containerID": "docker://e339f071cc46b3bf2464aa1c1d37f748750e7d016afc548213f829113ce7061f",
					"finishedAt": "2023-06-19T03:08:34Z"
				  }
				},
				"lastState": {},
				"containerID": "docker://e339f071cc46b3bf2464aa1c1d37f748750e7d016afc548213f829113ce7061f"
			  }
			],
			"hostIP": "192.168.220.30",
			"startTime": "2023-06-19T03:07:38Z",
			"qosClass": "Burstable",
			"conditions": [
			  {
				"type": "Initialized",
				"lastTransitionTime": "2023-06-19T03:07:38Z",
				"status": "True"
			  },
			  {
				"reason": "ContainersNotReady",
				"lastTransitionTime": "2023-06-19T03:08:34Z",
				"message": "containers with unready status: [busybox]",
				"type": "Ready",
				"status": "False"
			  },
			  {
				"reason": "ContainersNotReady",
				"message": "containers with unready status: [busybox]",
				"type": "ContainersReady",
				"lastTransitionTime": "2023-06-19T03:08:34Z",
				"status": "False"
			  },
			  {
				"type": "PodScheduled",
				"lastTransitionTime": "2023-06-19T03:07:38Z",
				"status": "True"
			  }
			],
			"podIPs": [
			  {
				"ip": "192.168.220.30"
			  }
			]
		  }
		},
		"apiVersion": "audit.k8s.io/v1",
		"ztimestamp": "2023-06-19T03:08:42.450489Z",
		"host": {
		  "os": {
			"codename": "Core",
			"name": "CentOS Linux",
			"family": "redhat",
			"version": "7 (Core)",
			"platform": "centos"
		  },
		  "containerized": true,
		  "name": "filebeat-9nc5d",
		  "architecture": "x86_64"
		},
		"beat": {
		  "hostname": "filebeat-9nc5d",
		  "name": "filebeat-9nc5d",
		  "version": "6.6.2"
		},
		"requestReceivedTimestamp": "2023-06-19T03:08:40.583032Z",
		"auditID": "7627d14d-9fe0-4d68-a1cb-901558cadf97",
		"objectRef": {
		  "apiVersion": "v1",
		  "resource": "pods",
		  "namespace": "kube-system",
		  "name": "busy-b6f7b6db6-c2szh"
		},
		"offset": 45132615,
		"level": "RequestResponse",
		"kind": "Event",
		"verb": "delete",
		"prospector": {
		  "type": "log"
		},
		"userAgent": "kubelet/v1.18.4 (linux/amd64) kubernetes/c96aede",
		"requestURI": "/api/v1/namespaces/kube-system/pods/busy-b6f7b6db6-c2szh",
		"responseStatus": {
		  "metadata": {},
		  "code": 200
		},
		"input": {
		  "type": "log"
		},
		"stageTimestamp": "2023-06-19T03:08:40.591203Z",
		"sourceIPs": [
		  "192.168.220.30"
		],
		"@timestamp": "2023-06-19T03:08:41.454Z",
		"stage": "ResponseComplete",
		"requestObject": {
		  "apiVersion": "v1",
		  "kind": "DeleteOptions",
		  "preconditions": {
			"uid": "33d7d262-ce52-43ef-bc68-8db6ef8a813d"
		  },
		  "gracePeriodSeconds": 0
		},
		"user": {
		  "groups": [
			"system:nodes",
			"system:authenticated"
		  ],
		  "username": "system:node:node2"
		}
	  }
	`
)

func newBasePod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-pod",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "container-1",
				Image: "image:v1",
				Resources: corev1.ResourceRequirements{
					Limits:   corev1.ResourceList{corev1.ResourceCPU: *resource.NewQuantity(100, resource.DecimalExponent), corev1.ResourceMemory: *resource.NewQuantity(1024, resource.DecimalExponent)},
					Requests: corev1.ResourceList{corev1.ResourceCPU: *resource.NewQuantity(100, resource.DecimalExponent), corev1.ResourceMemory: *resource.NewQuantity(1024, resource.DecimalExponent)},
				},
			}},
		},
	}
}

func TestIsImageChanged(t *testing.T) {

	pod1 := newBasePod()
	pod2 := newBasePod()
	b := IsImageChanged(pod1, pod2)
	if b != false {
		t.Error("pod image should not be changed")
	}
	pod2.Spec.Containers[0].Image = "image:v2"

	b = IsImageChanged(pod1, pod2)
	if b != true {
		t.Error("pod image should be changed")
	}
}

func TestSpecChanged(t *testing.T) {

	pod1 := newBasePod()
	pod2 := newBasePod()
	b := IsSpecChanged(pod1, pod2)
	if b != false {
		t.Error("pod spec should not be changed")
	}

	pod2.Spec.Containers[0].Resources.Limits = corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(200, resource.DecimalExponent),
		corev1.ResourceMemory: *resource.NewQuantity(1024, resource.DecimalExponent)}
	b = IsSpecChanged(pod1, pod2)
	if b != true {
		t.Error("pod spec should be changed")
	}
}

func TestGeneratePodFromEvent(t *testing.T) {
	testCase := []struct {
		caseName  string
		source    string
		namespace string
		name      string
		disabled  bool
	}{
		{
			caseName:  "normal-create-event",
			source:    normalEvent,
			namespace: "kube-system",
			name:      "busy-b6f7b6db6-c2szh",
		},
	}

	for _, tc := range testCase {
		if tc.disabled {
			continue
		}
		var event k8s_audit.Event
		err := json.Unmarshal([]byte(tc.source), &event)
		if err != nil {
			t.Fatal("error should be nil", err.Error())
		}

		pod, err := GeneratePodFromEvent(&event)
		if err != nil {
			t.Fatal("error should be nil", err.Error())
		}
		// t.Logf("pod is: %s", utils.Dumps(pod))
		assert.Equal(t, tc.namespace, pod.Namespace, "parse pod namespace failed")
		assert.Equal(t, tc.name, pod.Name, "parse pod name failed")
	}

}
