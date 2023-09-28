package trace

import (
	"encoding/json"
	"testing"

	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFiledSelector(t *testing.T) {
	p := v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Labels: map[string]string{
				"alabels": "sync",
			},
			Annotations: map[string]string{
				"annot.test.com": "test",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name: "c1",
				},
				v1.Container{
					Name: "c2",
				},
			},
			Volumes: []v1.Volume{
				v1.Volume{Name: "vol1"},
			},
		},
	}

	fieldRef := NewFieldRef("spec.[name]containers.name", ".")
	valueMap := fieldRef.GetFieldValue(p)
	t.Logf("field len: %d\n", len(valueMap))
	for k, v := range valueMap {
		t.Logf("k:%s, v: %v\n", k, v)
	}

	fieldRef1 := NewFieldRef("metadata/annotations/annot.test.co", "/")
	valueMap1 := fieldRef1.GetFieldValue(p)
	t.Logf("field len: %d\n", len(valueMap1))
	for k, v := range valueMap1 {
		t.Logf("k:%s, v: %v\n", k, v)
	}

	fieldRef2 := NewFieldRef("spec.[name]volumes.name", ".")
	valueMap2 := fieldRef2.GetFieldValue(p)
	t.Logf("field len: %d\n", len(valueMap2))
	for k, v := range valueMap2 {
		t.Logf("k:%s, v: %v\n", k, v)
	}
}

func TestMai(t *testing.T) {
	spanStr := `[{"ObjectRef":{"Resource":"pods","Namespace":"","Name":"PodSpans","UID":"","APIGroup":"","APIVersion":"v1","ResourceVersion":"","Subresource":""},"ActionType":"PodCreate","LifeFlag":{"Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"pod:create:success"}],"FinishEvent":[{"Type":"operation","Operation":"condition:Ready:true"},{"Type":"operation","Operation":"pod:delete:success"}]},"Spans":[{"Name":"pod_ready_span","Type":"pod_ready_span","Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"pod:create:success"}],"EndEvent":[{"Type":"operation","Operation":"condition:Ready:true"}]},{"Name":"default_schedule_span","Type":"default_schedule_span","SpanOwner":"k8s","Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"schedule:default-scheduler:entry"}],"EndEvent":[{"Type":"operation","Operation":"schedule:binding:success"}]},{"Name":"ip_allocate_span","Type":"ip_allocate_span","SpanOwner":"k8s","Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"schedule:binding:success"}],"EndEvent":[{"Type":"operation","Operation":"condition:IPAllocated:true"}]},{"Name":"kubelet_delay_span","Type":"kubelet_delay_span","SpanOwner":"k8s","Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"schedule:binding:success"}],"EndEvent":[{"Type":"operation","Operation":"condition:ContainerDiskPressure:false"}]},{"Name":"total_volume_mount_span","Type":"total_volume_mount_span","Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"schedule:binding:success"}],"EndEvent":[{"Type":"event","Reason":"SuccessfulAttachOrMountVolume"}]},{"Name":"volume_attach_span","Type":"volume_attach_span","SpanOwner":"k8s","Mode":"direct-info","DirectEvent":[{"Type":"event","NameRex":"Successfully attached volume .*\\\\^(.*),","DurationRex":"Successfully attached volume .* elapsed time (.*)","Reason":"AttachVolume"},{"Type":"event","NameRex":"Multi-Attach error for volume \\\"(.*)\\\"","Reason":"FailedAttachVolume"}]},{"NameRef":"spec.[name]volumes.name","Type":"volume_mount_span","SpanOwner":"k8s","Mode":"direct-info","DirectEvent":[{"Type":"event","NameRex":"Successfully mounted volume .*\\\\[(.*)\\\\],","DurationRex":"Successfully mounted volume .* elapsed time (.*)","Reason":"MountVolume"},{"Type":"event","NameRex":"Failed mounted volume .*\\\\[(.*)\\\\],","DurationRex":"Failed mounted volume .* elapsed time (.*)","Reason":"MountVolume"},{"Type":"event","NameRex":"Unable to attach or mount volumes: unmounted volumes=\\\\[(.*)\\\\],","Reason":"FailedMount"}]},{"Name":"sandbox_create_span","Type":"sandbox_create_span","SpanOwner":"k8s","Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"schedule:binding:success"},{"Type":"operation","Operation":"condition:IPAllocated:true"},{"Type":"event","Reason":"MountVolume"}],"EndEvent":[{"Type":"event","Reason":"SuccessfulCreatePodSandBox"},{"Type":"event","Reason":"Pulled"},{"Type":"event","Reason":"Pulling"}]},{"Type":"image_pull_span","NameRef":"spec.[name]containers.image","SpanOwner":"k8s","Mode":"start-finish","NeedClose":true,"StartEvent":[{"Type":"event","NameRex":"Pulling image \\\"(.*)\\\"","Reason":"Pulling"}],"EndEvent":[{"Type":"event","NameRex":"Successfully pulled image \\\"(.*)\\\"","Reason":"Pulled"}]},{"Name":"pod_init_span","Type":"pod_init_span","SpanOwner":"custom","Mode":"start-finish","StartEvent":[{"Type":"event","Reason":"SuccessfulCreatePodSandBox"}],"EndEvent":[{"Type":"operation","Operation":"condition:Initialized:true"}]},{"Type":"container_create_span","NameRef":"spec.[name]containers.name","SpanOwner":"k8s","Mode":"direct-info","DirectEvent":[{"Type":"event","NameRex":"Created container (.*), elapsedTime .*","DurationRex":"Created container .*, elapsedTime (.*)","Reason":"Created"},{"Type":"event","NameRex":"Failed create container (.*)","DurationRex":"Failed create container .*, elapsedTime (.*) ,","Reason":"Failed"}]},{"Type":"container_start_span","NameRef":"spec.[name]containers.name","SpanOwner":"k8s","Mode":"start-finish","StartEvent":[{"Type":"event","NameRex":"Created container (.*)","Reason":"Created"}],"EndEvent":[{"Type":"event","NameRex":"Started container (.*)","Reason":"Started"}]},{"Name":"container_readiness_span","Type":"container_readiness_span","SpanOwner":"custom","Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"phase:Running:set"}],"EndEvent":[{"Type":"operation","Operation":"condition:ContainersReady:true"}]},{"Type":"container_poststart_span","NameRef":"spec.[name]containers.name","SpanOwner":"custom","Mode":"start-finish","StartEvent":[{"Type":"event","NameRex":"Starting to execute poststart hook for container (.*) with .*","Reason":"StartingPostStartHook"}],"EndEvent":[{"Type":"event","NameRex":"Successfully execute poststart hook for container (.*), elapsedTime .*","Reason":"SucceedPostStartHook"}]},{"Name":"pod_running_span","Type":"pod_running_span","Mode":"start-finish","StartEvent":[{"Type":"event","Reason":"Started"},{"Type":"event","Reason":"SucceedPostStartHook"}],"EndEvent":[{"Type":"operation","Operation":"phase:Running:set"}]},{"Name":"pod_readiness_span","Type":"pod_readiness_span","SpanOwner":"custom","Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"condition:ContainersReady:true"}],"EndEvent":[{"Type":"operation","Operation":"condition:Ready:true"}]}]},{"ObjectRef":{"Resource":"pods","Namespace":"","Name":"PodSpans","UID":"","APIGroup":"","APIVersion":"v1","ResourceVersion":"","Subresource":""},"ActionType":"PodDelete","LifeFlag":{"Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"pod:delete:success"}],"FinishEvent":[{"Type":"operation","Operation":"pod:deleted:success"}]},"Spans":[{"Name":"pod_delete_span","Type":"pod_delete_span","Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"pod:delete:success"}],"EndEvent":[{"Type":"operation","Operation":"pod:deleted:success"}]},{"Type":"finalizer_delete_span","NameRef":"metadata.[$]finalizers.$","SpanOwner":"k8s","Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"pod:delete:success"}],"EndEvent":[{"Type":"operation","NameRex":"finalizer:(.*):delete","Operation":"finalizer:delete"}]},{"Type":"container_kill_span","NameRef":"spec.[name]containers.name","SpanOwner":"k8s","Mode":"start-finish","StartEvent":[{"Type":"operation","Operation":"pod:delete:success"},{"Type":"event","NameRex":"Stopping container (.*)","Reason":"Killing"}],"EndEvent":[{"Type":"operation","NameRex":"containerState:(.*):terminated:.*","Operation":"containerState:set"},{"Type":"event","NameRex":"Stopping container (.*), elapsedTime .*","Reason":"SucceedKillingContainer"}]}]}]`
	var tmpConfig = NewResourceSpanConfigList()
	err := json.Unmarshal([]byte(spanStr), &tmpConfig)
	if err != nil {
		t.Errorf("failed to unmarshal span configmap: %v", err)
		return
	}
}
