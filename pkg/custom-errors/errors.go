package customerrors

type ErrType int
type ErrMsg int

const (
	ErrParams ErrType = iota
)

var customErrors = [][]error{
	paramsErrors,
}

func Error(errType ErrType, errMsg ErrMsg) error {
	return customErrors[errType][errMsg]
}
