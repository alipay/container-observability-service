package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alipay/container-observability-service/internal/restapi/handler"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"

	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/gorilla/mux"
	"k8s.io/klog/v2"
)

type ServerConfig struct {
	MetricsAddr string
	ListenAddr  string
	Storage     data_access.StorageInterface
}

// Server is server to query trace stats
type Server struct {
	Config  *ServerConfig
	Storage data_access.StorageInterface
}

// NewAPIServer create new API server
func NewAPIServer(config *ServerConfig) (*Server, error) {
	return &Server{
		Config:  config,
		Storage: config.Storage,
	}, nil
}

func (s *Server) StartServer(stopCh chan struct{}) {
	klog.Info(utils.Dumps(s.Config))
	go func() {
		// router
		r := mux.NewRouter()
		// containerlifecycle

		r.Path("/apis/v1/debugging/pods").HandlerFunc(handlerWrapper(handler.PodResetResultFactory, s.Storage))

		log.Println("ListenAndServe ...")
		err := http.ListenAndServe(s.Config.ListenAddr, r)
		if err != nil {
			klog.Errorf("failed to ListenAndServe, err:%s", err.Error())
			panic(err.Error())
		}
		log.Println("ListenAndServed ...")
	}()
	<-stopCh
	klog.Error("apiserver exiting")
}

func handlerWrapper(h handler.HandlerFunc, storage data_access.StorageInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		p := h(w, r, storage)
		var (
			err     error
			respObj interface{}
		)
		code := http.StatusOK
		msg := "query success"

		defer func() {
			if err != nil {
				w.Header().Set("Content-Type", "application/json;charset=UTF-8")
				corsHeader(r, w)
				w.WriteHeader(code)
				errorResponse := model.Response{
					Code:    code,
					Status:  http.StatusText(code),
					Message: msg,
				}
				res, err := json.Marshal(errorResponse)
				if err != nil {
					klog.Errorf("Marshal response failed: %s", err.Error())
					return
				}
				w.Write(res)
			}
		}()

		klog.V(6).Infof("uri: %s", r.RequestURI)

		// parse request
		err = p.ParseRequest()
		if err != nil {
			klog.Errorf("ParseRequest failed: %s", err.Error())
			return
		}
		// parse valid request
		err = p.ValidRequest()
		if err != nil {
			klog.Errorf("ValidRequest failed: %s", err.Error())
			code = http.StatusBadRequest
			msg = err.Error()
			return
		}
		httpStatus, respObj, err := p.Process()
		if err != nil {
			klog.Errorf("Process failed: %s", err.Error())
			code = httpStatus
			msg = err.Error()
			return
		}
		corsHeader(r, w)

		if err := json.NewEncoder(w).Encode(respObj); err != nil {
			log.Printf("json enc: %+v", err)
		}
		// set header
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")

		// write result to response
	}
}
func corsHeader(r *http.Request, w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
}
