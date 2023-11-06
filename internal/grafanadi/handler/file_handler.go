package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	"k8s.io/klog/v2"
)

func FileDownload(w http.ResponseWriter, r *http.Request) {
	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("FileDownload").Observe(cost)
	}()
	fileName := getFilenameFromRequest(r)
	fmt.Println("filename", fileName)

	file, err := os.Open("statics/" + fileName)
	if err != nil {
		klog.Infof("can not find the file: %s ", fileName)
		return
	}
	defer file.Close()

	fileHeader := make([]byte, 2048)
	file.Read(fileHeader)

	// fileStat, _ := file.Stat()
	if strings.HasSuffix(fileName, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else if strings.HasSuffix(fileName, ".js") {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	} else if strings.HasSuffix(fileName, ".svg") || strings.HasSuffix(fileName, ".png") {
		w.Header().Set("Content-Type", "image/svg+xml")
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	// w.Header().Set("Content-Length", strconv.FormatInt(fileStat.Size(), 10))
	file.Seek(0, 0)
	io.Copy(w, file)

	return
}

func getFilenameFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	paths := strings.Split(r.URL.Path, "/")
	if len(paths) > 0 {
		return paths[len(paths)-1]
	}
	return ""
}
