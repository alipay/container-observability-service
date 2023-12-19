package common

import "flag"

var (
	LunettesNs = "lunettes"
)

func init() {
	flag.StringVar(&LunettesNs, "lunettes-namespace", "lunettes", "the namespace of lunettes")
}
