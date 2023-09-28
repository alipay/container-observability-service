package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/olivere/elastic"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
)

type nodeOpStruct struct {
	ClusterName    string
	Node           *corev1.Node
	StageTimeStamp time.Time
	AuditID        string
}

func queryNodeYamlWithNodeName(nodeName string) *nodeOpStruct {
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryNodeYaml", begin)
	}()

	result := &nodeOpStruct{}
	if nodeName == "" {
		return nil
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("nodeName: \"%s\"", nodeName))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := esClient.Search().Index("node_yaml").Type("_doc").Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return nil
	}

	for _, hit := range searchResult.Hits.Hits {
		if er := json.Unmarshal(*hit.Source, result); er == nil {
			return result
		}

	}

	return nil
}
