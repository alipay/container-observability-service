package base_processor

import "github.com/alipay/container-observability-service/pkg/shares"

func init() {
	shares.BaseObjectProcessor.Register("ObjectDecoder", &ObjectDecoder{})
	shares.BaseObjectProcessor.Register("SLOTimeComputer", &SLOTimeComputer{})
}
