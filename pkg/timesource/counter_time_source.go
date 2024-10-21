package timesource

import "sync"

type CounterTimeSource struct {
	counter int64
	mu      sync.Mutex
}

func NewCounterTimeSource() *CounterTimeSource {
	return &CounterTimeSource{}
}

func (ts *CounterTimeSource) GetTime(mode string) (int64, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.counter++
	return ts.counter, nil
}
