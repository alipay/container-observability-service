package config

import "flag"

var (
	EnableStressMode = false
)

func init() {
	flag.BoolVar(&EnableStressMode, "enable_stress_mode", true, "if enable stress mod")
}
