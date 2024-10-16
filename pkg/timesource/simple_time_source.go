package timesource

import (
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/config"
)

type SimpleTimeSource struct{}

var _ TimeSourcer = (*SimpleTimeSource)(nil)

func NewSimpleTimeSource() *SimpleTimeSource {
	return &SimpleTimeSource{}
}

func (l *SimpleTimeSource) GetTime(mode string) (int64, error) {
	if config.Debug.DebugMode && mode == "start" {
		// simulate the latency of the HTTP request
		// used in benchmark
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}
	return time.Now().UnixMicro(), nil
}
