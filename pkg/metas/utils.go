package metas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	jsoniter "github.com/json-iterator/go"
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/apis/audit"
	k8s_audit "k8s.io/apiserver/pkg/apis/audit"

	"k8s.io/apimachinery/pkg/util/strategicpatch"

	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

// GeneratePodFromEvent generate pod info from event.
// first pars from responseObject and then requestObject and last from event.objectRef
func GeneratePodFromEvent(event *audit.Event) (*corev1.Pod, error) {
	if event.ObjectRef.Resource != "pods" && event.ObjectRef.Resource != "networkmetadatas" {
		return nil, fmt.Errorf("invalid object kind %s", event.ObjectRef.Resource)
	}
	var pod *corev1.Pod
	//gvk := &schema.GroupVersionKind{Version: "v1", Kind: "pods"}
	// first, get pod  from responseObject
	if obj := GetObjectFromRuntimeUnknown(event.ResponseObject, nil); obj != nil {
		if p, ok := obj.(*corev1.Pod); ok {
			pod = p
		}
	}
	if pod != nil {
		klog.Info("pod from response is validate")
		return pod, nil
	}

	if obj := GetObjectFromRuntimeUnknown(event.RequestObject, nil); obj != nil {
		if p, ok := obj.(*corev1.Pod); ok {
			pod = p
		}
	}

	if pod != nil {
		klog.Info("pod from request is validate")
		// pod.Namespace or/and pod.Name may be empty, fill it.
		if pod.Namespace == "" && event.ObjectRef.Namespace != "" {
			pod.Namespace = event.ObjectRef.Namespace
		}
		if pod.Name == "" && event.ObjectRef.Name != "" {
			pod.Name = event.ObjectRef.Name
		}
		if pod.Name == "" && pod.GenerateName != "" {
			pod.Name = pod.GenerateName
		}
		return pod, nil
	}

	// old process
	// FIXME should delete
	pod = &corev1.Pod{}

	if event.Level == audit.LevelMetadata {
		pod.ObjectMeta.Name = event.ObjectRef.Name
		pod.ObjectMeta.Namespace = event.ObjectRef.Namespace
		return pod, nil
	}

	if event.Verb == "delete" || event.Verb == "patch" {
		pod.ObjectMeta.Name = event.ObjectRef.Name
		pod.ObjectMeta.Namespace = event.ObjectRef.Namespace
	} else if event.Verb == "create" || event.Verb == "update" {
		if event.RequestObject == nil || event.RequestObject.Raw == nil {
			return nil, fmt.Errorf("no request body found: %s", event.RequestURI)
		} else if err := json.Unmarshal(event.RequestObject.Raw, pod); err != nil {
			return nil, err
		}

		// if no namespace in Pod spec, get it from ObjectRef
		if pod.Namespace == "" {
			pod.ObjectMeta.Namespace = event.ObjectRef.Namespace
		}
	}
	return pod, nil
}

// WaitForAPIServer waits for the API Server's /healthz endpoint to report "ok" with timeout.
func WaitForAPIServer(client clientset.Interface, timeout time.Duration) error {
	var lastErr error

	err := wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		healthStatus := 0
		result := client.Discovery().RESTClient().Get().AbsPath("/healthz").Do(context.TODO()).StatusCode(&healthStatus)
		if result.Error() != nil {
			lastErr = fmt.Errorf("failed to get apiserver /healthz status: %v", result.Error())
			return false, nil
		}
		if healthStatus != http.StatusOK {
			content, _ := result.Raw()
			lastErr = fmt.Errorf("APIServer isn't healthy: %v", string(content))
			klog.Warningf("APIServer isn't healthy yet: %v. Waiting a little while.", string(content))
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		return fmt.Errorf("%v: %v", err, lastErr)
	}

	return nil
}

type PodPatch struct {
	Status corev1.PodStatus `json:"status,omitempty"`
}

func GetPodPatch(oldPod, newPod *corev1.Pod) ([]byte, *PodPatch, error) {
	oldData, err := json.Marshal(oldPod)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to Marshal oldData for pod %q/%q: %v", oldPod.Namespace, oldPod.Name, err)
	}

	newData, err := json.Marshal(newPod)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to Marshal newData for pod %q/%q: %v", newPod.Namespace, newPod.Name, err)
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.Pod{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to CreateTwoWayMergePatch for pod %q/%q: %v", newPod.Namespace, newPod.Name, err)
	}
	patchObj := PodPatch{}
	err = json.Unmarshal(patchBytes, &patchObj)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to Unmarshalto podPatch for pod %q/%q: %v", newPod.Namespace, newPod.Name, err)
	}

	return patchBytes, &patchObj, nil
}

func IsPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// IsPodFinishedNormally judge if a pod is in a `finished` state.
// For a job pod, the pod should be Failed/Succeed/Running
// For a non-job pod, the pod should be Ready
func IsPodFinishedNormally(pod *corev1.Pod) bool {
	if IsJobPod(pod) {
		return IsJobPodFinished(pod)
	}
	return IsPodReady(pod)
}

func IsJobPodFinished(pod *corev1.Pod) bool {
	return IsJobPod(pod) && (IsJobPodSucceeded(pod) || IsJobPodFailed(pod) || IsJobPodRunning(pod))
}

func IsJobPodSucceeded(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodSucceeded
}

func IsJobPodRunning(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodRunning
}

func IsJobPodFailed(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodFailed
}

func IsPodSchedulingFailed(pod *corev1.Pod) bool {
	for i := range pod.Status.Conditions {
		c := pod.Status.Conditions[i]
		if c.Type == corev1.PodScheduled &&
			c.Status == corev1.ConditionFalse &&
			c.Message != "" && c.Reason != "" {
			return true
		}
	}
	return false
}

func IsPodScheduled(pod *corev1.Pod) bool {
	// ignore phase
	// if pod.Status.Phase != corev1.PodPending {
	// 	return true
	// }

	for i := range pod.Status.Conditions {
		c := pod.Status.Conditions[i]
		if c.Type == corev1.PodScheduled {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

// PodKey return a pod name with namespace
func PodKey(pod *corev1.Pod) string {
	if pod == nil {
		return "not_available"
	}
	return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
}

// PodKeyForEvent return a pod name with namespace
func PodKeyForEvent(event *k8s_audit.Event) string {
	return fmt.Sprintf("%s/%s", event.ObjectRef.Namespace, event.ObjectRef.Name)
}

func pod2ContainersMap(pod *corev1.Pod) map[string]corev1.Container {
	containers := make(map[string]corev1.Container)
	for i := range pod.Spec.Containers {
		c := pod.Spec.Containers[i]
		containers[c.Name] = c
	}
	return containers
}

func PodDiff(oldPod, newPod *corev1.Pod, diffFunc func(oldContainer, newContainer corev1.Container) bool) bool {
	// FIXME assume containers name and count not change
	oldContainers := pod2ContainersMap(oldPod)
	newContainers := pod2ContainersMap(newPod)
	for ko, vo := range oldContainers {
		if vn, found := newContainers[ko]; found {
			if diffFunc(vo, vn) {
				return true
			}
		} else {
			klog.Errorf("unreachable code")
		}
	}
	return false
}

// IsImageChanged identify an upgrade event
func IsImageChanged(oldPod, newPod *corev1.Pod) bool {
	return PodDiff(oldPod, newPod, func(oldContainer, newContainer corev1.Container) bool {
		if oldContainer.Image != newContainer.Image {
			return true
		}
		return false
	})
}

// IsSpecChanged identify an update event
func IsSpecChanged(oldPod, newPod *corev1.Pod) bool {
	return PodDiff(oldPod, newPod, func(oldContainer, newContainer corev1.Container) bool {
		if oldContainer.Resources.Requests.Cpu().Cmp(*newContainer.Resources.Requests.Cpu()) != 0 ||
			oldContainer.Resources.Limits.Cpu().Cmp(*newContainer.Resources.Limits.Cpu()) != 0 ||
			oldContainer.Resources.Requests.Memory().Cmp(*newContainer.Resources.Requests.Memory()) != 0 ||
			oldContainer.Resources.Limits.Memory().Cmp(*newContainer.Resources.Limits.Memory()) != 0 ||
			oldContainer.Resources.Requests.StorageEphemeral().Cmp(*newContainer.Resources.Requests.StorageEphemeral()) != 0 ||
			oldContainer.Resources.Limits.StorageEphemeral().Cmp(*newContainer.Resources.Limits.StorageEphemeral()) != 0 {
			return true
		}
		return false
	})
}

type Unknown struct {
	runtime.TypeMeta `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
}

func GetObjectFromRuntimeUnknown(un *runtime.Unknown, gvk *schema.GroupVersionKind) runtime.Object {
	if un == nil || un.Raw == nil {
		return nil
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	if gvk != nil {
		un.SetGroupVersionKind(*gvk)
	} else if un.GroupVersionKind().Empty() {
		var unknown Unknown
		err := json.Unmarshal(un.Raw, &unknown)
		if err != nil {
			klog.Infof("json un marshal error: %s \n", err.Error())
			return nil
		}
		un.SetGroupVersionKind(unknown.GroupVersionKind())
	}

	if un.GroupVersionKind().Version == "" || un.GroupVersionKind().Kind == "" {
		return nil
	}

	kindGuessed := UnsafeGuessResourceToKind(un.Kind)
	for _, kind := range kindGuessed {
		if unicode.IsLower(rune(un.GroupVersionKind().Kind[0])) {
			un.SetGroupVersionKind(un.GroupVersionKind().GroupVersion().WithKind(kind))
		}
		klog.V(8).Infof("un.apiversion: %s, un.kind:%s \n", un.APIVersion, un.Kind)
		obj, err := scheme.Scheme.New(un.GroupVersionKind())
		if err != nil {
			klog.Errorf("error: %s\n", err.Error())
			continue
		}

		if err := json.Unmarshal(un.Raw, obj); err != nil {
			continue
		}
		return obj
	}
	return nil
}

func GetJsonFromRuntimeUnknown(un *runtime.Unknown) map[string]interface{} {
	if un == nil || un.Raw == nil {
		return nil
	}

	obj := make(map[string]interface{})
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	if err := json.Unmarshal(un.Raw, &obj); err != nil {
		return nil
	}
	if metaData, ok := obj["metadata"]; ok && metaData != nil {
		v, ok := obj["metadata"].(map[string]interface{})
		if ok {
			return v
		}
	}
	return nil
}

func MicroTime2Time(t metav1.MicroTime) time.Time {
	return t.Time
}

func IsDeleteImmediately(pod *corev1.Pod) bool {
	if pod.DeletionGracePeriodSeconds != nil &&
		*pod.DeletionGracePeriodSeconds == 0 &&
		pod.DeletionTimestamp != nil {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodScheduled {
				if condition.Status == corev1.ConditionFalse {
					return true
				}
				break
			}
		}
		return true
	}
	return false
}

func GetPodReadyTime(pod *corev1.Pod) time.Time {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			return c.LastTransitionTime.Time
		}
	}
	return time.Time{}
}

// GetJobPodFinishedTime use the max container terminated time as job finish time
func GetJobPodFinishedTime(pod *corev1.Pod) time.Time {
	t := time.Time{}
	for _, c := range pod.Status.ContainerStatuses {
		if c.State.Terminated != nil {
			if t.Before(c.State.Terminated.FinishedAt.Time) {
				t = c.State.Terminated.FinishedAt.Time
			}
		}
	}
	return t
}

func IsSidecarContainer(container *corev1.Container) bool {
	for _, env := range container.Env {
		if env.Name == "IS_SIDECAR" {
			return env.Value == "true"
		}
	}
	return strings.Contains(container.Name, "-sidecar-container")
}

func GetOwnerType(pod *corev1.Pod) string {
	if pod == nil {
		return ""
	}
	if len(pod.OwnerReferences) == 0 {
		return ""
	}

	return strings.ToLower(pod.OwnerReferences[0].Kind)
}

func IsJobPod(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}

	return pod.Spec.RestartPolicy != corev1.RestartPolicyAlways
}

func UnsafeGuessResourceToKind(resource string) []string {
	resourceName := resource
	if len(resourceName) == 0 {
		return []string{resource}
	}
	/*singularName := strings.ToLower(kindName)
	singular := kind.GroupVersion().WithResource(singularName)*/
	resourceName = strFirstToUpper(resourceName)

	if strings.HasSuffix(resourceName, "ies") {
		return []string{strings.TrimSuffix(resourceName, "ies") + "y"}
	}

	if strings.HasSuffix(resourceName, "es") {
		rs := strings.TrimSuffix(resourceName, "es")
		if !strings.HasSuffix(rs, "s") && !strings.HasSuffix(rs, "x") && !strings.HasSuffix(rs, "ch") &&
			!strings.HasSuffix(rs, "sh") && !strings.HasSuffix(rs, "z") {
			rs += "e"
		}
		return []string{rs}
	}

	var rs []string
	if strings.HasSuffix(resourceName, "s") {
		rs = append(rs, strings.TrimSuffix(resourceName, "s"))
	}

	return append(rs, resourceName)
}

func strFirstToUpper(str string) string {
	if len(str) < 1 {
		return ""
	}
	if len(str) == 1 {
		return strings.ToUpper(str)
	}

	return strings.ToUpper(str[0:1]) + str[1:]
}

type SloSpecItem struct {
	SloTime string
}

type SloSpec map[string]*SloSpecItem

func FetchSloSpec(obj runtime.Object) SloSpec {
	sloSpec := make(map[string]*SloSpecItem, 0)

	return sloSpec
}
