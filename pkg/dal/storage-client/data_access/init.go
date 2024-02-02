package data_access

import (
	"fmt"

	"github.com/alipay/container-observability-service/pkg/common"
	"github.com/olivere/elastic/v7"
	"k8s.io/klog"
)

var (
	EsClient *elastic.Client
	XSearch  StorageInterface
)

func NewDBClient(opts *common.DBOptions) (StorageInterface, error) {
	switch opts.Driver {
	case "mysql":
		sqlStorage, err := ProvideSqlStorate(opts.MysqlOptions)
		if err != nil {
			return nil, err
		}
		XSearch = sqlStorage
		return XSearch, nil
	case "elasticsearch":
		esStorage, err := ProvideEsStorage(opts.ESOptions)

		if err != nil {
			klog.Infof("read esStorage err: %s", err.Error())
			return nil, err
		}
		XSearch = esStorage
		return XSearch, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", opts.Driver)
	}
}

// func InitApi(endPoint string, username string, password string, jaegerQueryAddress string) {
// 	// TODO: multi storage implement

// 	// ZSearch
// 	EsClient, _ = impl.InitZSearch(endPoint, username, password)
// 	storage = impl.NewZSearchStorage()
// 	// OceanBase

// 	// Init jaeger query service
// 	InitJaegerQuery(jaegerQueryAddress)
// }

// func GetStorage() Storage {
// 	return storage
// }
