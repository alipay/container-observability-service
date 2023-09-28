package share

const (
	owner  = "Owner Name to be filled in"
	action = "Action to be filled in"
)

// 错误模块
var (
	SCHEDULER            = "scheduler"
	VOLUME               = "volume"
	NETWORK              = "network"
	ADMISSION            = "admission"
	SANDBOX              = "sandbox"
	KUBELET_DELAY        = "kubelet_delay"
	RUNTIME              = "runtime"
	IMAGE                = "image"
	CONTAINER_CREATE     = "container_create"
	CONTAINER_START      = "container_start"
	CONTAINER_POST_START = "container_post_start"
	CONTAINER_READINESS  = "container_readiness"
	POD_INITIALIZE       = "pod_initialize"
	POD_READINESS        = "pod_readiness"
	NODE_HEALTH          = "node_health"
	CONTAINER_KILL       = "container_kill"
)

// 扩展分析模块
var (
	SCHEDULER_MODULE     = "scheduler_module"
	VOLUME_MODULE        = "volume_module"
	NETWORK_MODULE       = "network_module"
	RUNTIME_MODULE       = "runtime_module"
	IMAGE_MODULE         = "image_module"
	NODE_HEALTH_MODULE   = "node_health_module"
	POD_READINESS_MODULE = "pod_readiness_module"
)

var (
	ExtentsAnalysisMap = map[string][]string{
		SCHEDULER:            []string{SCHEDULER_MODULE},
		VOLUME:               []string{VOLUME_MODULE},
		NETWORK:              []string{NETWORK_MODULE},
		RUNTIME:              []string{RUNTIME_MODULE},
		IMAGE:                []string{IMAGE_MODULE},
		CONTAINER_CREATE:     []string{RUNTIME_MODULE, NODE_HEALTH_MODULE},
		CONTAINER_START:      []string{RUNTIME_MODULE, NODE_HEALTH_MODULE},
		CONTAINER_READINESS:  []string{NODE_HEALTH_MODULE},
		CONTAINER_POST_START: []string{NODE_HEALTH_MODULE},
		POD_INITIALIZE:       []string{NODE_HEALTH_MODULE},
		POD_READINESS:        []string{NODE_HEALTH_MODULE},
		NODE_HEALTH:          []string{NODE_HEALTH_MODULE},
	}

	OwnerMap = map[string]string{
		SCHEDULER:            owner,
		VOLUME:               owner,
		NETWORK:              owner,
		RUNTIME:              owner,
		IMAGE:                owner,
		CONTAINER_CREATE:     owner,
		CONTAINER_START:      owner,
		CONTAINER_READINESS:  owner,
		CONTAINER_POST_START: owner,
		POD_INITIALIZE:       owner,
		POD_READINESS:        owner,
		NODE_HEALTH:          owner,
	}
	ActionMap = map[string]string{
		SCHEDULER:            "1.检查应用资源是否充足; 2.检查应用逻辑池等亲和性配置是否正确; 3.调度组协助分析",
		VOLUME:               "1.确定挂载的volume正常; 2.检查volume配置是否正确; 3. volume owner协助分析",
		NETWORK:              "1. IP资源不足",
		RUNTIME:              "1. 联系L2解决",
		IMAGE:                "1. 检查镜像配置是否正确；2. 镜像拉取超时请走镜像加速；",
		CONTAINER_CREATE:     "1.联系 runtime Owner 解决",
		CONTAINER_START:      "1.退出码不为0的应用Owner自行检查; 2.联系 runtime Owner 解决",
		CONTAINER_READINESS:  "1. 请应用Owner自行判断container readiness probe探测失败原因",
		CONTAINER_POST_START: "1. 请应用Owner自行判断container poststarthook执行失败原因",
		POD_INITIALIZE:       "1.业务自己行检查Initial容器是否有问题；2.联系 runtime Owner",
		POD_READINESS:        "1.业务自己行检查应用五元组信息是否配置正确；",
		NODE_HEALTH:          "1.查看物理节点是否正常 2.查看节点监控",
	}
)

type ReasonResult struct {
	PodName   string                 `json:"pod_name"`
	PodUid    string                 `json:"pod_uid"`
	Result    string                 `json:"result"`
	Module    string                 `json:"module"`
	HasError  bool                   `json:"has_error"`
	Diagnosis map[string]interface{} `json:"diagnosis"`
	Owner     string                 `json:"owner"`
	Action    string                 `json:"action"`
}

func (r *ReasonResult) SetOwner() {
	if len(r.Module) == 0 {
		return
	}
	if o, ok := OwnerMap[r.Module]; ok {
		r.Owner = o
	}

}
func (r *ReasonResult) SetAction() {
	if len(r.Action) == 0 && len(r.Module) > 0 {
		if a, ok := ActionMap[r.Module]; ok {
			r.Action = a
		}
	}
	//TODO 根据具体的错误给出具体action
}
