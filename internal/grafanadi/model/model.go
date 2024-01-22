package model

// Take this as Example: https://sriramajeyam.com/grafana-infinity-datasource/wiki/json
type BizInfoTable struct {
	ClusterName string `json:"ClusterName,omitempty"`
	Namespace   string `json:"Namespace,omitempty"`
}

type PodStatusTable struct {
	PodPhase      string `json:"PodPhase,omitempty"`
	UpTime        string `json:"UpTime,omitempty"`
	State         string `json:"State,omitempty"`
	CreateTime    string `json:"CreatedAt,omitempty"`
	LastTimeStamp string `json:"LastActiveAt,omitempty"`
}

type PodInfoTable struct {
	ClusterName string `json:"ClusterName,omitempty"`
	Namespace   string `json:"Namespace,omitempty"`
	PodName     string `json:"PodName,omitempty"`
	PodUID      string `json:"PodUID,omitempty"`
	PodIP       string `json:"PodIP,omitempty"`
	PodYaml     string `json:"PodYaml,omitempty"`
	NodeName    string `json:"NodeName,omitempty"`
	NodeIP      string `json:"NodeIP,omitempty"`
	NodeYaml    string `json:"NodeYaml,omitempty"`
}
type PodListTable struct {
	DeliveryTime       string `json:"交付时间,omitempty"`
	Namespace          string `json:"namespace,omitempty"`
	Cluster            string `json:"cluster,omitempty"`
	PodUID             string `json:"PodUid,omitempty"`
	PodName            string `json:"PodName,omitempty"`
	NodeIP             string `json:"node,omitempty"`
	SLO                string `json:"交付SLO,omitempty"`
	Result             string `json:"交付结果,omitempty"`
	DebugUrl           string `json:"诊断链接,omitempty"`
	SLOViolationReason string `json:"SLOViolationReason"`
}
type DeliveryPodCreateOrDeleteTable struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
type ClusterTable struct {
	Cluster string `json:"cluster,omitempty"`
	Number  string `json:"number,omitempty"`
}
type NamespaceTable struct {
	Namespace string `json:"namespace,omitempty"`
	Number    string `json:"number,omitempty"`
}
type NodeTable struct {
	Node   string `json:"node,omitempty"`
	Number string `json:"number,omitempty"`
}
type SloTable struct {
	Slo    string `json:"SLO,omitempty"`
	Number string `json:"number,omitempty"`
}

type DeliveryPodUpgradeTable struct {
	Index string `json:"index,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type KeyLifecycleEvents struct {
	StartTime     string `json:"startTime,omitempty"`
	UserAgent     string `json:"UserAgent,omitempty"`
	PlfID         string `json:"PlfID,omitempty"`
	TraceStage    string `json:"TraceStage,omitempty"`
	OperationName string `json:"operationName,omitempty"`
	Reason        string `json:"reason,omitempty"`
	Message       string `json:"message,omitempty"`
}

// Node Graph for Pod Yaml
type PodYamlGraphNodes struct {
	Id, Title, SubTitle, MainStat, SecondaryStat, Detail_one string
	Blue, Green, Yellow, Red                                 float64
}

type PodYamlGraphEdges struct {
	Id, Source, Target, MainStat string
}

func ConvertPodYamlGraphNodes2Frame(nodes []PodYamlGraphNodes) DataFrame {
	var idAry, titleAry, subtitleAry, mainStatAry, secStatAry, detailAry []string
	var blueAry, greenAry, yellowAry, redAry []float64

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

	return DataFrame{
		Schema: SchemaType{
			Fields: []FieldType{
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
		Data: DataType{
			Values: []interface{}{
				idAry, titleAry, subtitleAry, mainStatAry, secStatAry, detailAry, blueAry, greenAry, yellowAry, redAry,
			},
		},
	}

}

func ConvertPodYamlGraphEdges2Frame(edges []PodYamlGraphEdges) DataFrame {
	var idAry, sourceAry, targetAry, mainStatAry []string
	for _, edge := range edges {
		idAry = append(idAry, edge.Id)
		sourceAry = append(sourceAry, edge.Source)
		targetAry = append(targetAry, edge.Target)
		mainStatAry = append(mainStatAry, edge.MainStat)
	}

	return DataFrame{
		Schema: SchemaType{
			Fields: []FieldType{
				{Name: "id", Type: "string"},
				{Name: "source", Type: "string"},
				{Name: "target", Type: "string"},
				{Name: "mainStat", Type: "string"},
			},
		},
		Data: DataType{
			Values: []interface{}{
				idAry, sourceAry, targetAry, mainStatAry,
			},
		},
	}

}

// ExtraPropertyConfig config for property extractor
type ExtraPropertyConfig struct {
	Name       string `json:"name,omitempty"`
	ValueRex   string `json:"valueRex,omitempty"`   //json path to Value
	NeedChange bool   `json:"needChange,omitempty"` //is need change
	Resource   string `json:"resource,omitempty"`   //resource to extract
}

type PodSummary struct {
	DebugStage []string    `json:"DebugStage,omitempty"`
	ResultCode []string    `json:"ResultCode,omitempty"`
	Component  interface{} `json:"Component,omitempty"`
	Summary    interface{} `json:"Summary,omitempty"`
}
