package utils

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	grafanamodel "github.com/alipay/container-observability-service/internal/grafanadi/model"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"

	"github.com/olivere/elastic/v7"
)

type Util struct {
	Storage data_access.StorageInterface
}

func (u *Util) GetUid(podYamls []*model.PodYaml, key string, value *string) {

	switch key {
	case "name":
		if err := u.Storage.QueryPodUIDListByPodName(&podYamls, *value); err == nil {
			if len(podYamls) > 0 {
				*value = podYamls[0].PodUid
			}
		}
	case "hostname":
		if err := u.Storage.QueryPodUIDListByHostname(&podYamls, *value); err == nil {
			if len(podYamls) > 0 {
				*value = podYamls[0].PodUid
			}
		}
	case "podip":
		if err := u.Storage.QueryPodUIDListByPodIP(&podYamls, *value); err == nil {
			if len(podYamls) > 0 {
				*value = podYamls[0].PodUid
			}
		}
	}
}

func (u *Util) GetPodYaml(podYamls []*model.PodYaml, key string, value string) ([]*model.PodYaml, error) {

	var err error
	switch key {
	case "name":
		err = u.Storage.QueryPodUIDListByPodName(&podYamls, value)
	case "hostname":
		err = u.Storage.QueryPodUIDListByHostname(&podYamls, value)
	case "podip":
		err = u.Storage.QueryPodUIDListByPodIP(&podYamls, value)
	case "uid":
		err = u.Storage.QueryPodYamlsWithPodUID(&podYamls, value)
	}

	return podYamls, err
}

func ServeSLOGrafanaDI(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) {
	switch r.Method {
	case http.MethodOptions:
	case http.MethodGet:
		index := r.URL.Query().Get("index")
		docType := r.URL.Query().Get("doctype")
		timeAddr := r.URL.Query().Get("timeattr")
		filters := r.URL.Query().Get("filters")
		lte := r.URL.Query().Get("lte")
		gte := r.URL.Query().Get("gte")
		terms := r.URL.Query().Get("terms")
		percentiles := r.URL.Query().Get("percentiles")
		overtime := r.URL.Query().Get("overtime")

		if len(terms) == 0 && len(percentiles) == 0 {
			http.Error(w, "terms or percentiles empty", http.StatusBadRequest)
			return
		}

		evs := AggregationFromParams(index, docType, timeAddr, filters, lte, gte, terms, percentiles, overtime, storage)
		if err := json.NewEncoder(w).Encode(evs); err != nil {
			log.Printf("json enc: %+v", err)
		}

	default:
		http.Error(w, "bad method; supported OPTIONS, GET", http.StatusBadRequest)
		return
	}

}
func AggregationFromParams(index, doctype, timeAttr, filters, lte, gte, terms, percentile, overtime string, storage data_access.StorageInterface) interface{} {
	resultAry := make(map[string]interface{})
	esClient, ok := storage.(*data_access.StorageEsImpl)
	if !ok {
		return errors.New("parse errror")
	}

	boolSearch := elastic.NewBoolQuery()
	if len(filters) == 0 {
		return resultAry
	}
	filtersAry := strings.Split(filters, "*")

	for _, f := range filtersAry {
		fAry := strings.Split(f, ":")
		if len(fAry[1]) == 0 {
			continue
		}
		// add operation for query
		if strings.HasPrefix(fAry[1], "!") {
			fVal := strings.TrimPrefix(fAry[1], "!")
			if len(fVal) == 0 {
				continue
			}
			boolSearch = boolSearch.MustNot(elastic.NewTermsQuery(fAry[0], fVal))
		} else {
			boolSearch = boolSearch.Filter(elastic.NewTermsQuery(fAry[0], fAry[1]))
		}
	}

	if len(lte) > 0 && len(gte) > 0 {
		boolSearch = boolSearch.Filter(elastic.NewRangeQuery(timeAttr).Gte(gte).Lte(lte))

	}

	query := esClient.DB.Search().Index(index).Type(doctype).Query(boolSearch).Size(0)
	if len(overtime) == 0 {
		if len(terms) > 0 {
			termAggr := elastic.NewTermsAggregation().Field(terms).Size(100)
			query = query.Aggregation("aggvalue", termAggr)

		} else if len(percentile) > 0 {
			percentileAggr := elastic.NewPercentilesAggregation().Field(percentile).Percentiles(50, 90)
			query = query.Aggregation("aggvalue", percentileAggr)
		}
	} else {
		dataHisAggs := elastic.NewDateHistogramAggregation().Interval(overtime).Field(timeAttr)
		//elastic.NewHistogramAggregation().Field(timeAttr).Interval(float64(dur))
		if len(terms) > 0 {
			termAggr := elastic.NewTermsAggregation().Field(terms).Size(100)
			query = query.Aggregation("overtime", dataHisAggs.SubAggregation("aggvalue", termAggr))
		} else if len(percentile) > 0 {
			percentileAggr := elastic.NewPercentilesAggregation().Field(percentile).Percentiles(50, 90)
			query = query.Aggregation("overtime", dataHisAggs.SubAggregation("aggvalue", percentileAggr))
		}
	}

	result, err := query.Do(context.Background())
	if err != nil {
		return resultAry
	}

	if len(overtime) > 0 {
		otResult := make([]grafanamodel.DataFrame, 0)
		tmpResult := make(map[string][][]interface{})

		overtime, ok := result.Aggregations.DateHistogram("overtime")
		if !ok {
			return nil
		}
		for _, otb := range overtime.Buckets {
			timestamp := int64(otb.Key)

			// add new array
			if len(terms) > 0 {
				termAgg, ok := otb.Aggregations.Terms("aggvalue")
				if !ok {
					return otResult
				}
				for _, term := range termAgg.Buckets {
					key, _ := term.Key.(string)
					value := term.DocCount

					// get key
					itemAry, ok := tmpResult[key]
					if ok {
						timeAry := append(itemAry[0], timestamp)
						valueAry := append(itemAry[1], value)
						pairAry := [][]interface{}{timeAry, valueAry}
						tmpResult[key] = pairAry
					} else {
						timeAry := []interface{}{timestamp}
						valueAry := []interface{}{value}
						pairAry := [][]interface{}{timeAry, valueAry}
						tmpResult[key] = pairAry
					}
				}
			}
			if len(percentile) > 0 {
				perAgg, ok := otb.Percentiles("aggvalue")
				if !ok {
					return otResult
				}

				for key, value := range perAgg.Values {
					// get key
					perAry, ok := tmpResult[key]
					if ok {
						timeAry := append(perAry[0], timestamp)
						valueAry := append(perAry[1], value)
						pairAry := [][]interface{}{timeAry, valueAry}
						tmpResult[key] = pairAry
					} else {
						timeAry := []interface{}{timestamp}
						valueAry := []interface{}{value}
						pairAry := [][]interface{}{timeAry, valueAry}
						tmpResult[key] = pairAry
					}
				}
			}

		}
		for k, v := range tmpResult {
			schema := grafanamodel.SchemaType{Name: k, Fields: []grafanamodel.FieldType{{Name: "Time", Type: "time"}, {Name: "Value", Type: "number"}}}
			dataI := []interface{}{}
			for _, v := range v {
				dataI = append(dataI, v)
			}
			data := grafanamodel.DataType{Values: dataI}
			otResult = append(otResult, grafanamodel.DataFrame{Schema: schema, Data: data})
		}
		return otResult
	}

	if len(terms) > 0 && len(overtime) == 0 {
		items, ok := result.Aggregations.Terms("aggvalue")
		if !ok {
			return resultAry
		}
		for _, ib := range items.Buckets {
			key, ok := ib.Key.(string)
			if ok {
				resultAry[key] = ib.DocCount
			}
		}
	} else if len(percentile) > 0 && len(overtime) == 0 {
		ptile, ok := result.Aggregations.Percentiles("aggvalue")
		if !ok {
			return resultAry
		}

		return ptile.Values
	}
	return resultAry
}
