package generator

import (
	"math/rand"
	"sync/atomic"
)

// Counter generates a sequence of integers. [0, 1, ...]
type Counter struct {
	counter int64
}

// NewCounter creates the Counter generator.
func NewCounter(start int64) *Counter {
	return &Counter{
		counter: start,
	}
}

// Next implements Generator Next interface.
func (c *Counter) Next(_ *rand.Rand) int64 {
	return atomic.AddInt64(&c.counter, 1) - 1
}

// Last implements Generator Last interface.
func (c *Counter) Last() int64 {
	return atomic.LoadInt64(&c.counter) - 1
}
