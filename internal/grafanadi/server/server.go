package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/alipay/container-observability-service/internal/grafanadi/handler"
	interutils "github.com/alipay/container-observability-service/internal/grafanadi/utils"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/gocarina/gocsv"
	"github.com/gorilla/mux"
	"k8s.io/klog"
)

type ServerConfig struct {
	MetricsAddr    string
	ListenAddr     string
	ListenAuthAddr string
	Storage        data_access.StorageInterface
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
		r.Path("/").HandlerFunc(handlerWrapper(handler.RootFactory, s.Storage))
		r.Path("/containerlifecycle").HandlerFunc(handlerWrapper(handler.ContainerlifecycleFactory, s.Storage))
		r.Path("/containerevents").HandlerFunc(handlerWrapper(handler.ContainerEventsFactory, s.Storage))
		r.Path("/podinfotable").HandlerFunc(handlerWrapper(handler.PodInfoFactory, s.Storage))
		r.Path("/bizinfotable").HandlerFunc(handlerWrapper(handler.BizInfoFactory, s.Storage))
		r.Path("/podstatustable").HandlerFunc(handlerWrapper(handler.PodStatusFactory, s.Storage))
		r.Path("/deliverypodcreatetable").HandlerFunc(handlerWrapper(handler.PodDeliveryFactory, s.Storage))
		r.Path("/deliverypoddeletetable").HandlerFunc(handlerWrapper(handler.PodDeliveryFactory, s.Storage))
		r.Path("/deliverypodupgradetable").HandlerFunc(handlerWrapper(handler.PodDeliveryUpgradeFactory, s.Storage))
		r.Path("/containerstatus").HandlerFunc(handlerWrapper(handler.ContainerStatusFactory, s.Storage))
		r.Path("/keylifecycleevents").HandlerFunc(handlerWrapper(handler.PodPhaseFactory, s.Storage))
		r.Path("/deliverytrace").HandlerFunc(handlerWrapper(handler.TraceFactory, s.Storage))
		r.Path("/clusterdistribute").HandlerFunc(handlerWrapper(handler.DebuggingPodsFactory, s.Storage))
		r.Path("/namespacedistribute").HandlerFunc(handlerWrapper(handler.DebuggingPodsFactory, s.Storage))
		r.Path("/nodedistribute").HandlerFunc(handlerWrapper(handler.DebuggingPodsFactory, s.Storage))
		r.Path("/podtypedistribute").HandlerFunc(handlerWrapper(handler.DebuggingPodsFactory, s.Storage))
		r.Path("/podlist").HandlerFunc(handlerWrapper(handler.PodlistFactory, s.Storage))
		r.Path("/podnumber").HandlerFunc(handlerWrapper(handler.PodlistFactory, s.Storage))
		r.Path("/showyamls").HandlerFunc(handlerWrapper(handler.YamlFactory, s.Storage))
		r.Path("/podyamlgraphnodes").HandlerFunc(handlerWrapper(handler.NodeGraphParamsFactory, s.Storage))
		r.Path("/podyamlgraphedges").HandlerFunc(handlerWrapper(handler.NodeGraphParamsFactory, s.Storage))
		r.Path("/elasticaggregations").HandlerFunc(corsWrapper(interutils.ServeSLOGrafanaDI, s.Storage))
		r.Path("/rawdata").HandlerFunc(handlerWrapper(handler.RawdataFactory, s.Storage))

		err := http.ListenAndServe(s.Config.ListenAddr, r)
		if err != nil {
			klog.Errorf("failed to ListenAndServe at ListenAddr %s, err:%s", s.Config.ListenAddr, err.Error())
			panic(err.Error())
		}
	}()

	go func() {
		// router
		r := mux.NewRouter()

		// federation api
		r.Path("/apis/v1/debugpodlist").HandlerFunc(basicAuthWrapper(handlerWrapper(handler.DebugPodListFactory, s.Storage)))
		r.Path("/apis/v1/fed-debugpodlist").HandlerFunc(basicAuthWrapper(handlerWrapper(handler.FedDebugPodListFactory, s.Storage)))
		// federation api

		err := http.ListenAndServe(s.Config.ListenAuthAddr, r)
		if err != nil {
			klog.Errorf("failed to ListenAndServe at ListenAuthAddr %s, err:%s", s.Config.ListenAuthAddr, err.Error())
			panic(err.Error())
		}
	}()

	<-stopCh
	klog.Error("apiserver exiting")
}

func basicAuthWrapper(innerHandler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {

			// use basic auth and admin/admin for now, will be updated later.
			isUsernameOK := username == "admin"
			isPasswordOK := password == "admin"

			if isUsernameOK && isPasswordOK {
				innerHandler.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func corsWrapper(f func(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface), storage data_access.StorageInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		corsHeader(r, w)
		f(w, r, storage)
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
				// set cors header
				corsHeader(r, w)

				// set code
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

		// set header
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		// set cors header
		corsHeader(r, w)
		if err := json.NewEncoder(w).Encode(respObj); err != nil {
			log.Printf("json enc: %+v", err)
		}

	}
}

func csvHandlerWrapper(h handler.HandlerFunc, storage data_access.StorageInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		p := h(w, r, storage)
		var (
			err      error
			respObj  interface{}
			respBody []byte
		)
		code := http.StatusOK
		msg := "query success"
		defer func() {
			if err != nil {
				w.Header().Set("Content-Type", "text/csv")
				w.Header().Set("Content-Disposition", "attachment;filename=content.csv")
				// set cors header
				corsHeader(r, w)
				// set code
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
		fmt.Println("httpstatus", httpStatus)
		if err != nil {
			klog.Errorf("Process failed: %s", err.Error())
			code = httpStatus
			msg = err.Error()
			return
		}
		// if err := json.NewEncoder(w).Encode(respObj); err != nil {
		// 	log.Printf("json enc: %+v", err)
		// }
		if err := gocsv.Marshal(respObj, w); err != nil {
			log.Printf("json enc: %+v", err)
		}
		// set header
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment;filename=content.csv")
		// set cors header
		corsHeader(r, w)
		// write result to response
		w.WriteHeader(code)
		// no errors, write response.
		bodyLen := len(respBody)
		if bodyLen > 0 {
			w.Header().Set("Content-Length", strconv.Itoa(bodyLen))
			w.Write(respBody)
		}
	}
}
