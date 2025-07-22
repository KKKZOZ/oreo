package generator

import (
	"math/rand"
	"sync/atomic"

	"benchmark/pkg/util"
)

const (
	// WindowSize is the size of window of pending acks.
	WindowSize int64 = 1 << 20

	// WindowMask is used to turn an ID into a slot in the window.
	WindowMask int64 = WindowSize - 1
)

// AcknowledgedCounter reports generated integers via Last only
// after they have been acknoledged.
type AcknowledgedCounter struct {
	c Counter

	lock util.SpinLock

	window []bool
	limit  int64
}

// NewAcknowledgedCounter creates the counter which starts at start.
func NewAcknowledgedCounter(start int64) *AcknowledgedCounter {
	return &AcknowledgedCounter{
		c:      Counter{counter: start},
		lock:   util.SpinLock{},
		window: make([]bool, WindowSize),
		limit:  start - 1,
	}
}

// Next implements the Generator Next interface.
func (a *AcknowledgedCounter) Next(r *rand.Rand) int64 {
	return a.c.Next(r)
}

// Last implements the Generator Last interface.
func (a *AcknowledgedCounter) Last() int64 {
	return atomic.LoadInt64(&a.limit)
}

// Acknowledge makes a generated counter vaailable via Last.
func (a *AcknowledgedCounter) Acknowledge(value int64) {
	currentSlot := value & WindowMask
	if a.window[currentSlot] {
		panic("Too many unacknowledged insertion keys.")
	}

	a.window[currentSlot] = true

	if !a.lock.TryLock() {
		return
	}

	defer a.lock.Unlock()

	// move a contiguous sequence from the window
	// over to the "limit" variable

	limit := atomic.LoadInt64(&a.limit)
	beforeFirstSlot := limit & WindowMask
	index := limit + 1
	for ; index != beforeFirstSlot; index++ {
		slot := index & WindowMask
		if !a.window[slot] {
			break
		}

		a.window[slot] = false
	}

	atomic.StoreInt64(&a.limit, index-1)
}
