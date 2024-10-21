package timesource

import (
	"time"
)

type SimpleTimeSource struct{}

var _ TimeSourcer = (*SimpleTimeSource)(nil)

func NewSimpleTimeSource() *SimpleTimeSource {
	return &SimpleTimeSource{}
}

func (l *SimpleTimeSource) GetTime(mode string) (int64, error) {
	return time.Now().UnixMicro(), nil
}
