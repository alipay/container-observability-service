package main

import (
	"flag"
	tkpReqProvider "github.com/alipay/container-observability-service/pkg/tkp_provider"
	"os"
	"os/signal"
	"syscall"

	"github.com/alipay/container-observability-service/pkg/common"
	"github.com/alipay/container-observability-service/pkg/utils"

	"github.com/alipay/container-observability-service/internal/grafanadi/server"
	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/spf13/cobra"
	pflag "github.com/spf13/pflag"

	"k8s.io/klog/v2"
)

var (
	stopCh       = make(chan struct{})
	gracefulStop = make(chan os.Signal, 1)
)

func newRootCmd() *cobra.Command {
	config := &server.ServerConfig{}
	var cfgFile, kubeConfigFile, tkpRefCfgFile string

	cmd := &cobra.Command{
		Use:   "grafanadi",
		Short: "This is grafanadi command",
		Long:  `This is grafanadi comand for lunettes`,
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

			err = tkpReqProvider.InitTkpReqConfig(tkpRefCfgFile)
			if err != nil {
				klog.Errorf("failed to init tkp config [%s] err:%s", tkpRefCfgFile, err.Error())
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
	cmd.PersistentFlags().StringVarP(&config.MetricsAddr, "metrics-addr", "", ":9091", "metrics listen address (default :9091)")
	cmd.PersistentFlags().StringVarP(&config.ListenAddr, "listen-addr", "", ":8080", "api server listen address (default :8080)")

	// for storage
	cmd.PersistentFlags().StringVarP(&cfgFile, "config-file", "", "/app/storage-config.yaml", "storage config file")
	cmd.PersistentFlags().StringVarP(&service.GrafanaUrl, "grafana-url", "", "", "grafana url")
	cmd.PersistentFlags().StringVarP(&tkpRefCfgFile, "tkp-req-config-file", "", "/app/tkp-req-config-file.json", "tkp req config file")

	// kubeconfig for k8s client
	cmd.PersistentFlags().StringVarP(&kubeConfigFile, "kubeconfig", "", "/etc/kubernetes/kubeconfig/admin.kubeconfig", "Path to kubeconfig file with authorization and apiserver information.")

	return cmd
}

func main() {
	var rootCmd = newRootCmd()
	klog.InitFlags(flag.CommandLine)
	flag.CommandLine.Parse([]string{})
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-gracefulStop
		klog.Warning("[grafanadi.main]Existing...")
		close(stopCh)
	}()

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}
