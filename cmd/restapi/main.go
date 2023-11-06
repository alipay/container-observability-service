/*
Copyright 2019 Alipay.Inc.
*/

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/alipay/container-observability-service/internal/restapi/server"
	"github.com/alipay/container-observability-service/pkg/common"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/klog"
)

const (
	MetricNamePrefix = "lunettesdi_"
)

var (
	stopCh              = make(chan struct{})
	gracefulStop        = make(chan os.Signal, 1)
	serviceRestartCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "service_restart_count",
			Help: "service restart count",
		},
	)
)

func newRootCmd() *cobra.Command {
	config := &server.ServerConfig{}
	var cfgFile, kubeConfigFile string

	cmd := &cobra.Command{
		Use:   "restapi",
		Short: "This is restapi command",
		Long:  `This is restapi comand for eavesdropping`,
		RunE: func(cmd *cobra.Command, args []string) error {

			options, err := common.InitConfig(cfgFile)
			if err != nil {
				klog.Errorf("failed to get init config [%s], err:%s", cfgFile, err.Error())
				panic(err.Error())
			}
			storage, err := data_access.NewDBClient(options)
			if err != nil {
				klog.Errorf("failed to new DBClient [%s] err:%s", cfgFile, err.Error())
				panic(err.Error())
			}

			err = utils.InitKube(kubeConfigFile)
			if err != nil {
				klog.Errorf("failed to init kube client [%s], err:%s", kubeConfigFile, err.Error())
				panic(err.Error())
			}

			serverConfig := &server.ServerConfig{
				ListenAddr: config.ListenAddr,
				Storage:    storage,
			}
			hcsServer, err := server.NewAPIServer(serverConfig)
			if err != nil {
				panic(err.Error())
			}

			hcsServer.StartServer(stopCh)
			return nil
		},
	}

	// for server listen port
	cmd.PersistentFlags().StringVarP(&config.ListenAddr, "listen-addr", "", ":8080", "api server listen address (default :8080)")
	// for storage
	cmd.PersistentFlags().StringVarP(&cfgFile, "config-file", "", "/app/storage-config.yaml", "storage config file")
	// kubeconfig for k8s client
	cmd.PersistentFlags().StringVarP(&kubeConfigFile, "kubeconfig", "", "/etc/kubernetes/kubeconfig/admin.kubeconfig", "Path to kubeconfig file with authorization and apiserver information.")

	return cmd
}

func main() {
	var rootCmd = newRootCmd()
	klog.InitFlags(flag.CommandLine)
	flag.CommandLine.Parse([]string{})
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	prometheus.MustRegister(serviceRestartCount)

	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-gracefulStop
		klog.Warning("[lunettesdi.main]Existing...")
		serviceRestartCount.Inc()
		close(stopCh)
	}()

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}
