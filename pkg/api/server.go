package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alipay/container-observability-service/pkg/api_metrics"

	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"

	"github.com/olivere/elastic"
	"k8s.io/klog/v2"
)

const (
	tpl = `
<!DOCTYPE html>
<html lang="en">
 <head> 
  <meta charset="UTF-8" /> 
  <meta name="viewport" content="width=device-width, initial-scale=1.0" /> 
  <meta http-equiv="X-UA-Compatible" content="ie=edge" /> 
  <title>Document</title> 

  <link href="/statics/reset.css" rel="stylesheet" /> 
  <link href="/statics/app.css" rel="stylesheet" /> 
  <link href="/statics/jsonTree.css" rel="stylesheet" /> 
  <link href="/statics/css_family.css" rel="stylesheet" /> 

 </head> 
 <body> 
  <header id="header"> 
   <nav id="nav" class="clearfix"> 
    <ul class="menu menu_level1"> 
     <li data-action="expand" class="menu__item" id="expand-all"> <span class="menu__item-name">Expand>>></span> </li>
     <li data-action="collapse" class="menu__item" id="collapse-all"> <span class="menu__item-name">Collapse>>></span> </li>
    </ul> 
   </nav> 
   <div id="coords"></div> 
   <div id="debug"></div> 
  </header> 
  <div id="wrapper"> 
   <div id="tree"></div> 
  </div> 
  
  <script src="/statics/jsonTree.js"></script> 

  <script>
	var wrapper = document.getElementById("wrapper");
	var data = %s
	var tree = jsonTree.create(data, wrapper);
	document.getElementById("expand-all").addEventListener("click", function() {
  		tree.expand();
	})
	document.getElementById("collapse-all").addEventListener("click", function() {
  		tree.collapse();
	})

	tree.expand(function(node) {
   		return node.childNodes.length < 11 || node.label === 'Pod基本信息';
	});
	
	window.onload = function() {
		let allnodes = document.getElementsByClassName("jsontree_node");
		for (let i=0; i<allnodes.length; i++) {
    		let node = allnodes[i];
    		if (node.className !== "jsontree_node") {
        		continue;
    		}
			let label = node.getElementsByClassName("jsontree_label")[0].textContent
			let value = node.getElementsByClassName("jsontree_value")[0].textContent
    		if (label === '"NodeIP"') {
        		if (value[0] === '"' && value[value.length-1] === '"' && value.length > 2) {
	                let newValue = '/api/v1/debugpodyaml?nodeip=' + value.substr(1, value.length-2);
					let htl = value + '  <a target="_blank" href="' + newValue + '">podlist</a>';
            		node.getElementsByClassName("jsontree_value")[0].innerHTML = htl;
        		}
    		} else if (label === '"NodeName"') {
				if (value[0] === '"' && value[value.length-1] === '"' && value.length > 2) {
					let newValue = '/api/v1/debugnodeyaml?name=' + value.substr(1, value.length-2);
					let htl = value + ' <a target="_blank" href="' + newValue + '">nodeyaml</a>'
					node.getElementsByClassName("jsontree_value")[0].innerHTML = htl
				}
			} else if (node.getElementsByClassName("jsontree_label")[0].textContent === '"PlfID"') {
				let value = node.getElementsByClassName("jsontree_value")[0].textContent
				if (value[0] === '"' && value[value.length-1] === '"' && value.length > 2) {
					let newValue = '/api/v1/rawdata?plfid=' + value.substr(1, value.length-2);
					node.getElementsByClassName("jsontree_value")[0].innerHTML = '<a target="_blank" href="' + newValue + '">rawdata</a>'
				}
			} else if (label === '"dataSourceId"' || label === '"auditID"') {
				let value = node.getElementsByClassName("jsontree_value")[0].textContent
				if (value[0] === '"' && value[value.length-1] === '"' && value.length > 2) {
					let newValue = '/api/v1/rawdata?auditid=' + value.substr(1, value.length-2);
					node.getElementsByClassName("jsontree_value")[0].innerHTML = '<a target="_blank" href="' + newValue + '">' + value + '</a>'
				}
			} else if (node.getElementsByClassName("jsontree_value")[0].textContent.startsWith('"http:')) {
				let value = node.getElementsByClassName("jsontree_value")[0].textContent
				node.getElementsByClassName("jsontree_value")[0].innerHTML = '<a target="_blank" href="' + value.substr(1, value.length-2) + '">' + value + '</a>'
			} else if (label === '"podUID"' || label === '"PodUID"') {
				if (value[0] === '"' && value[value.length-1] === '"' && value.length > 2) {
					let newValue = '/api/v1/debugpodyaml?uid=' + value.substr(1, value.length-2);
					let htl = value + ' <a target="_blank" href="' + newValue + '">podyaml</a>' 
                    newValue = window.location.origin.split(":")[0] + ":" + window.location.origin.split(":")[1] + ':32526/search?limit=20&service=pod_delivery&tags=%7B%22uid%22%3A%20%22' + value.substr(1, value.length-2) + '%22%7D';
					htl = htl + ' <a target="_blank" href="' + newValue + '">jaeger</a>'
					node.getElementsByClassName("jsontree_value")[0].innerHTML = htl
				}
			}
		}
	}
</script>  
 </body>
</html>
`
)

// ServerConfig is configuration for api server
type ServerConfig struct {
	MetricsAddr string
	ESConfig    *xsearch.ElasticSearchConf
	ListenAddr  string
}

// Server is server to query trace stats
type Server struct {
	Config   *ServerConfig
	ESClient *elastic.Client
}

// NewAPIServer create new API server
func NewAPIServer(config *ServerConfig) (*Server, error) {
	esClient, err := elastic.NewClient(elastic.SetURL(config.ESConfig.Endpoint),
		elastic.SetBasicAuth(config.ESConfig.User, config.ESConfig.Password),
		elastic.SetSniff(false))
	if err != nil {
		return nil, err
	}

	return &Server{
		Config:   config,
		ESClient: esClient,
	}, nil
}

// StartServer start new api server
func (s *Server) StartServer(stopCh chan struct{}) {
	klog.Info(utils.Dumps(s.Config))

	go func() {
		//debug pod yaml
		http.HandleFunc("/api/v1/debugpodyaml", handlerWrapper(s, debugPodYamlFactory))
		http.HandleFunc("/api/v1/debugnodeyaml", handlerWrapper(s, debugNodeYamlFactory))

		http.HandleFunc("/api/v1/debugpod", handlerWrapper(s, debugPodFactory))
		http.HandleFunc("/api/v1/debugslo", handlerWrapper(s, sloFactory))
		http.HandleFunc("/api/v1/rawdata", handlerWrapper(s, rawDataFactory))
		http.HandleFunc("/fake", handlerWrapper(s, fakeFactory))
		//watch delivery info
		http.HandleFunc("/api/v1/watch", watch)
		//static file
		http.HandleFunc("/statics/reset.css", fileDownload)
		http.HandleFunc("/statics/app.css", fileDownload)
		http.HandleFunc("/statics/jsonTree.css", fileDownload)
		http.HandleFunc("/statics/css_family.css", fileDownload)
		http.HandleFunc("/statics/jsonTree.js", fileDownload)
		http.HandleFunc("/statics/icons.svg", fileDownload)

		http.ListenAndServe(s.Config.ListenAddr, nil)
	}()
	<-stopCh
	klog.Error("apiserver exiting")
}

func handlerWrapper(s *Server, h handlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		p := h(s, w, r)
		var (
			err      error
			respObj  interface{}
			respBody []byte
		)
		start := time.Now()
		code := http.StatusOK

		defer func() {
			if err != nil {
				w.WriteHeader(code)
				w.Write([]byte(err.Error()))
			}
			// record metrics
			api_metrics.IncResponseStatusCode(code)
			api_metrics.RecordAPIDuration(r.URL.Path, code, start)
		}()

		klog.V(6).Infof("uri: %s", r.RequestURI)

		// parse request
		err = p.ParseRequest()
		if err != nil {
			code = http.StatusBadRequest
			klog.Errorf("ParseRequest failed: %s", err.Error())
			return
		}

		// parse valid request
		err = p.ValidRequest()
		if err != nil {
			klog.Errorf("ValidRequest failed: %s", err.Error())
			code = http.StatusBadRequest
			return
		}

		// do main request
		code, respObj, err = p.Process()
		if err != nil {
			klog.Errorf("Process failed: %s", err.Error())
			code = http.StatusInternalServerError
			return
		}

		// convert result to []byte
		if str, ok := respObj.(string); ok {
			respBody = []byte(str)
			w.Header().Set("Content-Type", "text/json;charset=UTF-8")
		} else {
			tmp, err := json.Marshal(respObj)
			if err != nil {
				klog.Errorf("Marshal response failed: %s", err.Error())
				code = http.StatusInternalServerError
				return
			}

			respBody = []byte(strings.Replace(tpl, "%s", string(tmp), -1))
			w.Header().Set("Content-Type", "text/html;charset=UTF-8")
		}
		// set cors header
		corsHeader(r, w)

		// write result to response
		w.WriteHeader(code)
		// no errors, write response.
		len := len(respBody)
		if len > 0 {
			w.Header().Set("Content-Length", strconv.Itoa(len))
			w.Write(respBody)
		}
	})
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
