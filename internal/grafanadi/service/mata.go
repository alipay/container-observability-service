package service

import (
	"strconv"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	lunettes_model "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

func ConvertMeta2Frame(eavesdroppingMeta []*lunettes_model.LunettesMeta) []model.RecordTimeTable {
	bit := make([]model.RecordTimeTable, 0)
	if len(eavesdroppingMeta) == 0 {
		return nil
	}

	for _, meta := range eavesdroppingMeta {
		var timeDuration int64
		var lastRecord int64
		readTime, ok := meta.LastReadTime.(float64)
		if ok {
			timeDuration = time.Now().UnixNano() - int64(readTime)
			lastRecord = int64(readTime)
		} else {
			readTime, ok := meta.LastReadTime.(string)
			if ok {
				rTime, err := strconv.ParseInt(readTime, 10, 64)
				if err == nil {
					timeDuration = time.Now().UnixNano() - rTime
					lastRecord = rTime
				}
			}
		}
		t1 := time.Unix(0, lastRecord)

		table := model.RecordTimeTable{
			Cluster:      meta.ClusterName,
			TimeDuration: timeDuration,
			LastRecord:   t1.Format("2006-01-02 15:04:05"),
		}
		bit = append(bit, table)
	}

	return bit
}
