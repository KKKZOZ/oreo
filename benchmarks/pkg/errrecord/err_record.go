package errrecord

import "sync"

type ErrCntMap map[string]int

type ErrCntItem struct {
	Err   string
	Count int
}

var mu sync.Mutex
var errRecorder *ErrRecorder

func init() {
	errRecorder = NewErrRecorder()
}

func Record(op string, err error) {
	mu.Lock()
	defer mu.Unlock()
	errRecorder.Record(op, err)
}

func Summary() {
	errRecorder.Summary()
}
