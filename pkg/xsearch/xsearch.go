package xsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	jsoniter "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// EnsureIndex to check if an index exists, create it if not exists.
func EnsureIndex(client *elastic.Client, index, mapping string) error {
	ctx := context.Background()
	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		ics := client.CreateIndex(index)
		if mapping != "" {
			ics = ics.BodyString(mapping)
		}
		createIndex, err := ics.Do(ctx)
		if err != nil {
			return err
		}

		if !createIndex.Acknowledged {
			return fmt.Errorf("create index without acknowledged")
		}
	}

	return nil
}
func EnsurePipeline(client *elastic.Client, index, mapping string) error {
	ctx := context.Background()
	_, err := client.IngestGetPipeline(index).Do(ctx)
	if err != nil {
		ics := client.IngestPutPipeline(index)
		if mapping != "" {
			ics = ics.BodyString(mapping)
		}
		createpipeline, err := ics.Do(ctx)
		if err != nil {
			return err
		}

		if !createpipeline.Acknowledged {
			return fmt.Errorf("create pipeline without acknowledged")
		}
	}

	return nil
}

const (
	defaultTypeName = "_doc"
	defaultMapping  = `
	{
		"mappings" : {
			"dynamic" : "false",
			"properties" : {
			}
    	}
  	}
	`

	//node yaml
	nodeYamlIndexName = "node_yaml"
	nodeYamlMapping   = `
	{
		"mappings": {
			"dynamic": "false",
			"properties": {
				"auditID": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"clusterName": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"namespace": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"nodeIp": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"nodeName": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"stageTimestamp": {
					"type": "date"
				},
				"uid": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				}
			}
		},
		"settings": {
			"index": {
				"number_of_shards": "1",
				"number_of_replicas": "1"
			}
		}
	}`
	ztimePipelineName    = "ztimestamp"
	ztimePipelineMapping = `
	{
		"processors" : [
		  {
			"set" : {
			  "field" : "ztimestamp",
			  "value" : "{{_ingest.timestamp}}"
			}
		  }
		]
	}`

	//pod 生命周期中各个事件
	podLifePhaseIndexName = "pod_life_phase"
	podLifePhaseTypeName  = "_doc"
	podLifePhaseMapping   = `{
		"mappings": {
			"dynamic": "false",
			"properties": {
				"clusterName": {
					"type": "keyword"
				},
				"dataSourceId": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"endTime": {
					"type": "date"
				},
				"endTimeStr": {
					"type": "date"
				},
				"endTimeUnix": {
					"type": "long"
				},
				"hasErr": {
					"type": "boolean"
				},
				"namespace": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"operationName": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"podName": {
					"type": "keyword"
				},
				"podUID": {
					"type": "keyword"
				},
				"startTime": {
					"type": "date"
				},
				"startTimeStr": {
					"type": "date"
				},
				"startTimeUnix": {
					"type": "long"
				}
			}
		}
	}`

	// pod yaml
	podYamlIndexName = "pod_yaml"
	podYamlTypeName  = "_doc"
	podYamlmap       = `
	{
		"mappings": {
			"dynamic": "false",
			"properties": {
				"stageTimestamp": {
					"type": "date"
				},
				"auditID": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"pod": {
					"enabled": false
				},
				"clusterName": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"isDeleted": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"namespace": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"podIP": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"podUID": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"hostIP": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"hostname": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"images": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"podName": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				}
			}
		}
	}`
	//slo data
	sloDataIndexName = "slo_data"
	sloDataTypeName  = "_doc"

	//slo trace data
	// change from slo_trace_data to slo_trace_data_daily
	sloTraceDataIndexName = "slo_trace_data_daily"
	sloTraceDataTypeName  = "_doc"
	sloTraceDataMapping   = `
	{
		"mappings": {
			"dynamic": "false",
			"properties": {
				"Appname": {
					"type": "text"
				},
				"BizId": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"BizName": {
					"type": "keyword"
				},
				"Cluster": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"Cores": {
					"type": "integer"
				},
				"CreatedTime": {
					"type": "date"
				},
				"DeleteResult": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"DeliveryDuration": {
					"type": "long"
				},
				"DeliverySLO": {
					"type": "keyword"
				},
				"DeliveryStatus": {
					"type": "keyword"
				},
				"DeliveryWorkload": {
					"type": "keyword"
				},
				"LifeDuration": {
					"type": "long",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				},
				"Namespace": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"NodeIP": {
					"type": "keyword",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				},
				"NodeName": {
					"type": "keyword"
				},
				"PodName": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"PodUID": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"ResourceHint": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"SLOViolationReason": {
					"type": "keyword",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"SloHint": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"StartUpResultFromCreate": {
					"type": "keyword",
					"fields": {
						"field": {
							"type": "text"
						},
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"Type": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"UpgradeResult": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				}
			}
		}
	}`

	auditStagingName    = "audit_staging"
	auditStagingMapping = `{
		"mappings": {
			"dynamic": "false",
			"properties": {
				"@timestamp": {
					"type": "date"
				},
				"annotations": {
					"properties": {
						"cluster": {
							"type": "text",
							"fields": {
								"keyword": {
									"type": "keyword",
									"ignore_above": 256
								}
							}
						}
					}
				},
				"auditID": {
					"type": "keyword"
				},
				"host": {
					"properties": {
						"name": {
							"type": "keyword"
						}
					}
				},
				"kind": {
					"type": "keyword"
				},
				"objectRef": {
					"properties": {
						"apiGroup": {
							"type": "keyword"
						},
						"apiVersion": {
							"type": "keyword"
						},
						"name": {
							"type": "text",
							"fields": {
								"keyword": {
									"type": "keyword",
									"ignore_above": 256
								}
							}
						},
						"namespace": {
							"type": "keyword"
						},
						"resource": {
							"type": "keyword"
						},
						"resourceVersion": {
							"type": "keyword"
						},
						"subresource": {
							"type": "keyword"
						},
						"uid": {
							"type": "keyword"
						}
					}
				},
				"requestReceivedTimestamp": {
					"type": "date"
				},
				"responseObject": {
					"properties": {
						"involvedObject": {
							"properties": {
								"name": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"namespace": {
									"type": "keyword"
								},
								"uid": {
									"type": "keyword"
								}
							}
						},
						"metadata": {
							"properties": {
								"managedFields": {
									"properties": {
										"fieldsV1": {
											"properties": {
												"f:metadata": {
													"type": "object",
													"enabled": false
												}
											}
										}
									}
								}
							}
						},
						"reason": {
							"type": "keyword"
						}
					}
				},
				"responseStatus": {
					"properties": {
						"code": {
							"type": "long"
						},
						"reason": {
							"type": "text",
							"fields": {
								"keyword": {
									"type": "keyword",
									"ignore_above": 256
								}
							}
						},
						"status": {
							"type": "keyword"
						}
					}
				},
				"stageTimestamp": {
					"type": "date"
				},
				"timestamp": {
					"type": "date"
				},
				"verb": {
					"type": "keyword"
				}
			}
		}
	}`

	podSummaryFeedbackIndexName = "pod_summary_feedback"
	podSummaryFeedbackTypeName  = "_doc"
	podSummaryFeedbackMapping   = `{
		"mappings": {
			"properties": {
			  	"ClusterName": {
					"type": "keyword"
				},
				"Namespace": {
					"type": "keyword"
				},
				"PodName": {
					"type": "keyword"
				},
				"PodUID": {
					"type": "keyword"
				},
				"PodIP": {
					"type": "keyword"
				},
				"NodeName": {
					"type": "keyword"
				},
				"Feedback": {
					"type": "keyword"
				},
				"Score": {
					"type": "keyword"
				},
				"Comment": {
					"type": "keyword"
				},
				"Summary": {
					"type": "keyword"
				},
				"CreateTime": {
					"type": "date"
				}
			}
		}
	}`
)

var esClient *elastic.Client
var EsConfig *ElasticSearchConf
var cluster string

func InitZsearch(zsearchEndPoint, username, password string, extraInfo interface{}) {
	if esClient == nil {
		client, err := elastic.NewClient(
			elastic.SetURL(zsearchEndPoint),
			elastic.SetBasicAuth(username, password),
			elastic.SetSniff(false),
			elastic.SetTraceLog(log.New(os.Stdout, "", log.LstdFlags)),
		)
		if err != nil {
			panic(err)
		}

		esClient = client
		cluster = extraInfo.(string)
	}

	//node yaml
	err := EnsureIndex(esClient, nodeYamlIndexName, nodeYamlMapping)
	if err != nil {
		log.Printf("index: %s, mammping: %s\n", nodeYamlIndexName, nodeYamlMapping)
		panic(err)
	}

	err = EnsureIndex(esClient, podLifePhaseIndexName, podLifePhaseMapping)
	if err != nil {
		log.Printf("index: %s, mammping: %s\n", podLifePhaseIndexName, podLifePhaseMapping)
		panic(err)
	}
	err = EnsureIndex(esClient, podYamlIndexName, podYamlmap)
	if err != nil {
		log.Printf("index: %s, mammping: %s\n", podYamlIndexName, podYamlmap)
		panic(err)
	}

	err = EnsureIndex(esClient, sloDataIndexName, "")
	if err != nil {
		log.Printf("index: %s, mammping: %s\n", sloDataIndexName, "")
		panic(err)
	}
	//slo trace data
	err = EnsureIndex(esClient, sloTraceDataIndexName, sloTraceDataMapping)
	if err != nil {
		log.Printf("index: %s, mammping: %s\n", sloTraceDataIndexName, sloTraceDataMapping)
		panic(err)
	}

	err = EnsureIndex(esClient, podInfoIndexName, sloPodInfoMapping)
	if err != nil {
		log.Printf("index: %s, mammping: %s\n", podInfoIndexName, sloPodInfoMapping)
		panic(err)
	}

	//audit_staging
	err = EnsureIndex(esClient, auditStagingName, auditStagingMapping)
	if err != nil {
		log.Printf("index: %s, mammping: %s\n", auditStagingName, auditStagingMapping)
		panic(err)
	}

	err = EnsureIndex(esClient, podSummaryFeedbackIndexName, podSummaryFeedbackMapping)
	if err != nil {
		log.Printf("index: %s, mammping: %s\n", podSummaryFeedbackIndexName, podSummaryFeedbackMapping)
		panic(err)
	}

	err = EnsurePipeline(esClient, ztimePipelineName, ztimePipelineMapping)
	if err != nil {
		log.Printf("index: %s, mammping: %s\n", ztimePipelineName, ztimePipelineMapping)
		panic(err)
	}
}

// SavePodLifePhase save pod life phase to zsearch
// 存储 Pod Phase 到 zsearch pod_life_phase 这个索引
// TODO tangbo 后续在 zsearch pod_life_phase 中增加大类信息，后续会消除这个 API
var podLifeBuffer *utils.BufferUtils = nil

func SavePodLifePhase(clusterName string, namespace string, podUID string, podName string, operationName string,
	hasErr bool, startTime, endTime time.Time, extraInfo map[string]interface{}, dataSourceID string) error {
	if podLifeBuffer == nil {
		podLifeBuffer = utils.NewBufferUtils(podLifePhaseIndexName, 1000, 10*time.Second, false, func(datas map[string]interface{}) error {
			if datas == nil {
				return nil
			}

			klog.V(6).Infof("do bulk, data size: %d", len(datas))
			err := utils.ReTry(func() error {
				bulkService := esClient.Bulk()
				for id, data := range datas {
					doc := elastic.NewBulkIndexRequest().Index(podLifePhaseIndexName).Type(podLifePhaseTypeName).Id(id).Doc(data).UseEasyJSON(true)
					bulkService = bulkService.Add(doc)
				}
				_, err := bulkService.Do(context.Background())
				if err != nil {
					return err
				}
				return nil
			}, 1*time.Second, 5)

			if err != nil {
				return err
			}
			return nil
		},
		)

		podLifeBuffer.DoClearData()
		//add graceful clear
		XSearchClear.AddCleanWork(func() {
			podLifeBuffer.Stop()
		})
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.DebugMethodDurationMilliSeconds.WithLabelValues("SavePodLifePhase").Observe(cost)
	}()

	docMap := make(map[string]interface{})
	docMap["clusterName"] = clusterName
	docMap["namespace"] = namespace
	docMap["podUID"] = podUID
	docMap["podName"] = podName
	docMap["operationName"] = operationName
	docMap["hasErr"] = hasErr
	docMap["startTime"] = startTime
	docMap["endTime"] = endTime
	docMap["extraInfo"] = extraInfo
	docMap["dataSourceId"] = dataSourceID

	hashCodeStr := fmt.Sprintf("%d", utils.StringHashcode(operationName))
	docID := clusterName + "_" + namespace + "_" + podUID + "_" + dataSourceID + "_" + hashCodeStr

	//insert to es with retry
	err := podLifeBuffer.SaveData(docID, docMap)
	if err != nil {
		klog.Error("Do save data error", err)
		return err
	}

	return nil
}

// SavePodYaml save pod yaml
type PodYamlDic struct {
	data []byte
	sync.WaitGroup
}

var jsonter = jsoniter.ConfigCompatibleWithStandardLibrary

func (p *PodYamlDic) conStruct(cluster string, pod *corev1.Pod, t time.Time, auditID string, isBeginDelete, isDeleted bool) {
	p.Add(1)
	defer p.Done()

	dic := make(map[string]interface{}, 15)
	dic["stageTimestamp"] = t
	dic["auditID"] = auditID
	dic["pod"] = pod
	if pod != nil {
		dic["podIP"] = pod.Status.PodIP
		dic["namespace"] = pod.Namespace
		dic["podUID"] = pod.UID
		dic["podName"] = pod.Name
		if pod.Status.HostIP != "" {
			dic["hostIP"] = pod.Status.HostIP
		}
		images := make([]string, 0)
		for _, c := range pod.Spec.Containers {
			images = append(images, c.Image)
		}
		dic["images"] = images
	}
	dic["clusterName"] = cluster
	dic["isBeginDelete"] = fmt.Sprintf("%t", isBeginDelete)
	dic["isDeleted"] = fmt.Sprintf("%t", isDeleted)

	bts, _ := jsonter.Marshal(dic)
	p.data = bts
	bts = nil
	//runtime.GC()
}

var podYamlBuffer *utils.BufferUtils = nil

func SavePodYaml(cluster string, pod *corev1.Pod, t time.Time, auditID string, isBeginDelete, isDeleted bool) error {
	defer utils.IgnorePanic("SavePodYaml")

	if podYamlBuffer == nil {
		podYamlBuffer = utils.NewBufferUtils(podYamlIndexName, 2000, 10*time.Second, true, func(datas map[string]interface{}) error {
			if datas == nil {
				return nil
			}

			klog.V(6).Infof("do bulk for %s, data size: %d ", podYamlIndexName, len(datas))
			err := utils.ReTry(func() error {
				bulkService := esClient.Bulk()
				for id, data := range datas {
					podYamlDic, ok := data.(*PodYamlDic)
					if !ok {
						continue
					}
					podYamlDic.Wait()

					doc := elastic.NewBulkIndexRequest().Index(podYamlIndexName).Type(podYamlTypeName).Id(id).Doc(json.RawMessage(podYamlDic.data)).UseEasyJSON(true)
					bulkService = bulkService.Add(doc)

					//释放内存
					podYamlDic.data = nil
				}

				_, err := bulkService.Do(context.Background())
				if err != nil {
					klog.Errorf("save pod yaml error: %s", err.Error())
					return err
				}
				return nil
			}, 1*time.Second, 10)

			if err != nil {
				return err
			}
			return nil
		},
		)

		podYamlBuffer.DoClearData()
		//add graceful clear
		XSearchClear.AddCleanWork(func() {
			podYamlBuffer.Stop()
		})
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.DebugMethodDurationMilliSeconds.WithLabelValues("SavePodYaml").Observe(cost)
	}()

	docID := fmt.Sprintf("%s_%s", cluster, pod.UID)
	podYamlDic := &PodYamlDic{}
	go podYamlDic.conStruct(cluster, pod, t, auditID, isBeginDelete, isDeleted)

	//insert to es with retry
	err := podYamlBuffer.SaveData(docID, podYamlDic)

	if err != nil {
		klog.Errorf("SavePodYaml Err: %s", err)
	}

	return nil
}

// SaveSloData save slo data
// 存储 SLO 数据到 ZSearch slo_data 索引
var sloBuffer *utils.BufferUtils = nil

func SaveSloData(cluster string, namespace string, podName string, podUID string, sloType string, sloData []byte) error {
	defer utils.IgnorePanic("SaveSloData")

	if sloBuffer == nil {
		sloBuffer = utils.NewBufferUtils(sloDataIndexName, 1000, 10*time.Second, true, func(datas map[string]interface{}) error {
			if datas == nil {
				return nil
			}

			klog.V(6).Infof("do bulk, data size: %d", len(datas))
			err := utils.ReTry(func() error {
				bulkService := esClient.Bulk()
				for id, data := range datas {
					doc := elastic.NewBulkIndexRequest().Index(sloDataIndexName).Type(sloDataTypeName).Id(id).Doc(json.RawMessage(data.([]byte))).UseEasyJSON(true)
					bulkService = bulkService.Add(doc)
					data = nil
				}
				_, err := bulkService.Do(context.Background())
				if err != nil {
					return err
				}
				return nil
			}, 1*time.Second, 10)

			if err != nil {
				return err
			}
			return nil
		},
		)

		sloBuffer.DoClearData()
		//add graceful clear
		XSearchClear.AddCleanWork(func() {
			sloBuffer.Stop()
		})
	}

	begin := time.Now()
	defer func() {
		metrics.DebugMethodDurationMilliSeconds.
			WithLabelValues("SaveSloData").Observe(utils.TimeSinceInMilliSeconds(begin))
	}()

	docID := fmt.Sprintf("%s_%s_%s_%s_%s", cluster, namespace, podName, podUID, sloType)

	err := sloBuffer.SaveData(docID, sloData)

	if err != nil {
		klog.Errorf("SaveSloData Err: %s", err)
	}

	return nil
}

// 存储 SLO Trace 数据到 ZSearch slo_trace_data 索引
var sloTraceBuffer *utils.BufferUtils = nil

func SaveSloTraceData(cluster string, namespace string, name string, uid string, sloType string, sloData []byte) error {
	defer utils.IgnorePanic("SaveSloTraceData")

	if sloTraceBuffer == nil {
		sloTraceBuffer = utils.NewBufferUtils(sloTraceDataIndexName, 1000, 10*time.Second, true, func(datas map[string]interface{}) error {
			if datas == nil {
				return nil
			}

			klog.V(6).Infof("do bulk, data size: %d", len(datas))
			err := utils.ReTry(func() error {
				bulkService := esClient.Bulk()
				for id, data := range datas {
					doc := elastic.NewBulkIndexRequest().Index(sloTraceDataIndexName).Type(sloTraceDataTypeName).Id(id).Doc(json.RawMessage(data.([]byte))).UseEasyJSON(true)
					bulkService = bulkService.Add(doc)

					data = nil
				}
				_, err := bulkService.Do(context.Background())
				if err != nil {
					return err
				}
				return nil
			}, 1*time.Second, 10)

			if err != nil {
				return err
			}
			return nil
		},
		)

		sloTraceBuffer.DoClearData()
		//add graceful clear
		XSearchClear.AddCleanWork(func() {
			sloTraceBuffer.Stop()
		})
	}

	begin := time.Now()
	defer func() {
		metrics.DebugMethodDurationMilliSeconds.
			WithLabelValues("SaveSloTraceData").Observe(utils.TimeSinceInMilliSeconds(begin))
	}()

	/*err := utils.ReTry(func() error {
		_, err := esClient.Index().Index(sloTraceDataIndexName).Type(sloTraceDataTypeName).
			BodyString(sloDataStr).Do(context.Background())
		if err != nil {
			klog.Info(err)
			return err
		}

		return nil
	}, 1*time.Second, 20)*/
	docID := fmt.Sprintf("%s_%s_%s_%s_%s", cluster, namespace, name, uid, sloType)
	err := sloTraceBuffer.SaveData(docID, sloData)

	if err != nil {
		klog.Errorf("SaveSloTraceData Err: %s", err)
	}

	return nil
}

// SaveNodeYaml save node yaml
var sloNodeYamlBuffer *utils.BufferUtils = nil

func SaveNodeYaml(cluster string, node *corev1.Node, t time.Time, auditID string) error {
	defer utils.IgnorePanic("SaveNodeYaml")

	if sloNodeYamlBuffer == nil {
		sloNodeYamlBuffer = utils.NewBufferUtils(nodeYamlIndexName, 1000, 10*time.Second, true, func(data map[string]interface{}) error {
			if data == nil {
				return nil
			}

			klog.V(6).Infof("do bulk, data size: %d", len(data))
			err := utils.ReTry(func() error {
				bulkService := esClient.Bulk()
				for id, doc := range data {
					doc := elastic.NewBulkIndexRequest().Index(nodeYamlIndexName).Type(defaultTypeName).Id(id).Doc(doc).UseEasyJSON(true)
					bulkService = bulkService.Add(doc)
				}
				_, err := bulkService.Do(context.Background())
				if err != nil {
					return err
				}
				return nil
			}, 1*time.Second, 10)

			if err != nil {
				return err
			}
			return nil
		},
		)

		sloNodeYamlBuffer.DoClearData()
		//add graceful clear
		XSearchClear.AddCleanWork(func() {
			sloNodeYamlBuffer.Stop()
		})
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.DebugMethodDurationMilliSeconds.WithLabelValues("SaveNodeYaml").Observe(cost)
	}()

	dic := make(map[string]interface{})
	dic["stageTimestamp"] = t
	dic["auditID"] = auditID
	dic["node"] = node
	dic["clusterName"] = cluster
	if node != nil {
		dic["nodeName"] = node.Name
		dic["namespace"] = node.Namespace
		dic["uid"] = string(node.UID)
	}

	docID := fmt.Sprintf("%s_%s", cluster, dic["nodeName"])
	/*err := utils.ReTry(func() error {
		_, err := esClient.Index().Index(nodeYamlIndexName).Type(defaultTypeName).
			Id(docID).BodyJson(dic).Do(context.Background())
		if err != nil {
			return err
		}

		return nil
	}, 1*time.Second, 10)*/

	err := sloNodeYamlBuffer.SaveData(docID, dic)

	if err != nil {
		klog.Errorf("SaveNodeYaml Err: %s", err)
	}

	return nil
}
