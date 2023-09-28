package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/alipay/container-observability-service/pkg/spans"
	"github.com/alipay/container-observability-service/pkg/trace"

	"github.com/alipay/container-observability-service/pkg/aggregator"
	apiserver "github.com/alipay/container-observability-service/pkg/api"
	"github.com/alipay/container-observability-service/pkg/featuregates"
	"github.com/alipay/container-observability-service/pkg/kube"
	"github.com/alipay/container-observability-service/pkg/xsearch"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	_ "go.uber.org/automaxprocs"
	"k8s.io/klog/v2"
)

var (
	stopCh              = make(chan struct{})
	gracefulStop        = make(chan os.Signal, 1)
	serviceRestartCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "service_restart_count",
			Help: "service restart count",
		},
	)
)

func newRootCmd() *cobra.Command {
	options := &aggregator.AggregatorOptions{}

	cmd := &cobra.Command{
		Use:   "aggregator",
		Short: "This is aggregator command",
		Long:  `This is aggregator comand for lenettes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// flag.Parse()
			go func() {
				// Expose the registered metrics via HTTP.
				http.Handle("/metrics", promhttp.Handler())
				klog.Fatal(http.ListenAndServe(options.MetricsAddr, nil))
			}()

			// parse feature gates
			featuregates.Parse(options.FeatureGates)

			if options.JaegerCollector == "" && featuregates.IsEnabled(spans.JaegerFeature) {
				klog.Error("need --jaeger-collector commandline arguments")
				os.Exit(-1)
			}

			if options.OTLPCollector == "" && featuregates.IsEnabled(trace.TraceFeature) {
				klog.Error("need --otlp-collector commandline arguments")
				os.Exit(-1)
			}

			if options.Cluster == "" {
				klog.Error("--clusters commandline arguments")
				os.Exit(-1)
			}

			if options.APIServerEnabled {
				// create and start api server
				config := &apiserver.ServerConfig{
					ListenAddr: options.APIServerListenAddr,
				}
				esOptions := &xsearch.ElasticSearchConf{
					Endpoint: options.ElasticSearchEndpoint,
					User:     options.ElasticSearchUser,
					Password: options.ElasticSearchPassword,
				}
				config.ESConfig = esOptions
				server, err := apiserver.NewAPIServer(config)
				if err != nil {
					panic(err.Error())
				}

				go func() {
					server.StartServer(stopCh)
				}()
			}

			//init kube
			kube.InitKube(options.KubeConfigFile)
			//init api module
			apiserver.InitApi(options.ElasticSearchEndpoint, options.ElasticSearchUser, options.ElasticSearchPassword)
			//init esClient
			xsearch.InitZsearch(
				options.ElasticSearchEndpoint, options.ElasticSearchUser, options.ElasticSearchPassword, options.Cluster)

			agg, err := aggregator.NewAggregator(options)
			if err != nil {
				return err
			}

			agg.Run(stopCh)
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(
		&options.MetricsAddr, "metrics-addr", "",
		":9091",
		"metrics listen address (default :9091)")
	cmd.PersistentFlags().StringVarP(
		&options.JaegerCollector, "jaeger-collector", "",
		"",
		"jaeger collector grpc-server listen address")
	cmd.PersistentFlags().StringVarP(
		&options.OTLPCollector, "otlp-collector", "",
		"",
		"otlp collector grpc-server listen address")
	cmd.PersistentFlags().DurationVarP(
		&options.TraceTimeout, "trace-timeout", "",
		10*time.Minute,
		"timeout for a trace to be closed (default 10 minutes)")
	cmd.PersistentFlags().StringVarP(
		&options.KubeConfigFile, "kubeconfig", "",
		"",
		"Path to kubeconfig file with authorization and apiserver information.")
	cmd.PersistentFlags().StringVarP(
		&options.Cluster, "cluster", "",
		"",
		"Default cluster name")

	cmd.PersistentFlags().IntVarP(
		&options.Burst, "burst", "",
		1024,
		"The maximum burst for throttle (default 1024)")
	cmd.PersistentFlags().Float32VarP(
		&options.QPS, "qps", "",
		1024,
		"The maximum QPS to the master from this client (default 1024)")

	// elasticsearch
	cmd.PersistentFlags().StringVarP(
		&options.ElasticSearchEndpoint, "es-endpoint", "",
		"http://127.0.0.1:9200",
		"ElasticSearch endpoint")
	cmd.PersistentFlags().StringVarP(
		&options.ElasticSearchUser, "es-user", "",
		"",
		"ElasticSearch username")
	cmd.PersistentFlags().StringVarP(
		&options.ElasticSearchPassword, "es-password", "",
		"",
		"ElasticSearch password")
	cmd.PersistentFlags().StringVarP(
		&options.ElasticSearchIndexName, "es-index", "",
		"audit",
		"ElasticSearch index where audit logs are stored")
	cmd.PersistentFlags().DurationVarP(
		&options.ElasticSearchBufferDuration, "es-buffer-duration", "",
		10*time.Second,
		"query data only before this lag (default -10 seconds)")
	cmd.PersistentFlags().DurationVarP(
		&options.ElasticSearchFetchInterval, "es-fetch-interval", "",
		5*time.Second,
		"ElasticSearch query interval (default 5 seconds)")

	cmd.PersistentFlags().StringVarP(
		&options.FeatureGates, "feature-gates", "",
		"",
		"Enabled features split by comma")

	// for apiserver
	cmd.PersistentFlags().BoolVarP(
		&options.APIServerEnabled, "apiserver-enabled", "",
		false,
		"Enabled Lunettes API Server")
	cmd.PersistentFlags().StringVarP(
		&options.APIServerListenAddr, "apiserver-addr", "",
		":8080",
		"Lunettes API Server listen address")
	cmd.PersistentFlags().BoolVarP(
		&options.EnableTrace, "trace-enable", "",
		true,
		"If close the trace feature, default false")

	return cmd
}

func main() {

	go func() {
		listen := fmt.Sprintf("0.0.0.0:%s", "6060")
		klog.Infof("start pprof http server, listen on: %s", listen)
		http.ListenAndServe(listen, nil)
	}()

	klog.Infof("GOMAXPROCS: %d", runtime.GOMAXPROCS(0))

	var rootCmd = newRootCmd()
	// Hack for depress the warning "ERROR: logging before flag.Parse: I0614xxx" which throw by klog.
	klog.InitFlags(flag.CommandLine)
	flag.CommandLine.Parse([]string{})
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	prometheus.MustRegister(serviceRestartCount)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)

	go func() {
		<-gracefulStop
		klog.Warning("[lunettes.main]Existing...")
		serviceRestartCount.Inc()
		xsearch.XSearchClear.DoClear()
		close(stopCh)
	}()

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
