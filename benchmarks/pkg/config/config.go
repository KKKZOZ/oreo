package config

import "time"

var (
	RemoteAddressList = []string{"localhost:8001"}
	TimeOracleUrl     = "http://localhost:8010"
	ZipfianConstant   = 0.5
	Latency           = 10 * time.Millisecond
)
