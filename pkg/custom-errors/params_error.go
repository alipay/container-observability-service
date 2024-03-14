package customerrors

import "errors"

const (
	NoDeliveryResult ErrMsg = iota
)

var paramsErrors = []error{
	errors.New("the params is error, deliveryresult needed"),
}
