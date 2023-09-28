package api

import (
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func fileDownload(w http.ResponseWriter, r *http.Request) {
	fileName := getFilenameFromRequest(r)
	filePath := "statics/" + fileName

	file, err := os.Open(filePath)
	if err != nil {
		klog.Infof("can not find the file: %s ", fileName)
		return
	}
	defer file.Close()

	fileHeader := make([]byte, 2048)
	file.Read(fileHeader)

	fileStat, _ := os.Stat(filePath)
	if strings.HasSuffix(fileName, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else if strings.HasSuffix(fileName, ".js") {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	} else if strings.HasSuffix(fileName, ".svg") || strings.HasSuffix(fileName, ".png") {
		w.Header().Set("Content-Type", "image/svg+xml")
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Length", strconv.FormatInt(fileStat.Size(), 10))

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
