package service

import (
	"encoding/json"
	"fmt"
	eavesmodel "github.com/alipay/container-observability-service/internal/grafanadi/model"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"time"
)

func InsertTimeStamp(ts time.Time, timeAry []time.Time) []time.Time {
	if len(timeAry) == 0 {
		return []time.Time{ts}
	}

	tmpAry := []time.Time{}
	idx := 0
	insert := false
	for ; idx < len(timeAry); idx++ {
		if !timeAry[idx].Before(ts) && !insert {
			tmpAry = append(tmpAry, ts)
			insert = true
		}
		tmpAry = append(tmpAry, timeAry[idx])
	}

	if !insert {
		tmpAry = append(tmpAry, ts)
	}
	return tmpAry
}
func ConvertSloDataTrace2Graph(sdts []model.SloTraceData) eavesmodel.DataFrame {
	if len(sdts) == 0 {
		return eavesmodel.DataFrame{}
	}
	var (
		ts             []time.Time
		needChangeFlag = true
	)
	for _, sdt := range sdts {
		if sdt.DeletedTime.IsZero() && needChangeFlag {
			ts = InsertTimeStamp(time.Now(), ts)
			needChangeFlag = false
		}
		ts = InsertTimeStamp(sdt.CreatedTime, ts)
		if sdt.ReadyAt.After(sdt.CreatedTime) {
			ts = InsertTimeStamp(sdt.ReadyAt, ts)
		}
		if sdt.DeletedTime.After(sdt.CreatedTime) {
			ts = InsertTimeStamp(sdt.DeletedTime, ts)
		}
		if sdt.DeleteEndTime.After(sdt.CreatedTime) {
			ts = InsertTimeStamp(sdt.DeleteEndTime, ts)
		}
	}

	s, _ := json.Marshal(sdts)
	fmt.Println(string(s))
	b, _ := json.Marshal(ts)
	fmt.Println(string(b))

	fields := []eavesmodel.FieldType{
		{Name: "timestamp", Type: "time"},
	}
	var values []interface{}
	values = append(values, ts)
	for _, sdt := range sdts {
		spanAry := make([]string, len(ts))

		idx := 0

		// skip time spot before create
		for ; idx < len(ts) && ts[idx].Before(sdt.CreatedTime); idx++ {
		}
		// set the create span
		curStat := "Creating"
		set := false
		if sdt.ReadyAt.After(sdt.CreatedTime) {
			for ; idx < len(ts) && ts[idx].Before(sdt.ReadyAt); idx++ {
				spanAry[idx] = curStat
			}
			set = true
			curStat = "Running"
		}
		// set the running span

		if sdt.DeletedTime.IsZero() {
			for ; idx < len(ts); idx++ {
				spanAry[idx] = curStat
			}
		}

		if sdt.DeletedTime.After(sdt.CreatedTime) {
			for ; idx < len(ts) && ts[idx].Before(sdt.DeletedTime); idx++ {
				spanAry[idx] = curStat
			}
			set = true
			curStat = "Deleting"
		}
		// set the deleting span
		if sdt.DeleteEndTime.IsZero() {
			for ; idx < len(ts); idx++ {
				spanAry[idx] = curStat
			}
		}
		if sdt.DeleteEndTime.After(sdt.CreatedTime) {
			for ; idx < len(ts) && ts[idx].Before(sdt.DeleteEndTime); idx++ {
				spanAry[idx] = curStat
			}
			set = true
		}
		if !set {
			for ; idx < len(ts); idx++ {
				spanAry[idx] = curStat
			}
		}

		for ; idx < len(ts); idx++ {
			spanAry[idx] = ""
		}
		spanAry[idx-1] = ""

		values = append(values, spanAry)
		fields = append(fields, eavesmodel.FieldType{Name: sdt.PodName, Type: "string"})
	}

	sortMap := make(map[string]int, len(values))

	for i := 1; i < len(values); i++ {
		if strSlice, ok := values[i].([]string); ok {
			var runIndex int
			for ri, status := range strSlice {
				if status == "Creating" {
					runIndex = ri
					break
				}
			}
			sortMap[fields[i].Name] = runIndex
		}
	}

	for i := 2; i < len(fields); i++ {
		for j := i; j > 0 && sortMap[fields[j].Name] < sortMap[fields[j-1].Name]; j-- {
			fields[j], fields[j-1] = fields[j-1], fields[j]
			values[j], values[j-1] = values[j-1], values[j]
		}
	}

	return eavesmodel.DataFrame{Schema: eavesmodel.SchemaType{Fields: fields}, Data: eavesmodel.DataType{Values: values}}
}
