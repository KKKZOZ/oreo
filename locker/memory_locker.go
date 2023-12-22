package locker

import (
	"errors"
	"sync"
	"time"
)

type MemoryLocker struct {
	mu     sync.Mutex
	locks  map[string]string
	cond   *sync.Cond
	timers map[string]*time.Timer
}

var AMemoryLocker = NewMemoryLocker()

// NewMemoryLocker creates a new instance of MemoryLocker.
// MemoryLocker is a type that provides a mechanism for locking and unlocking memory resources.
// It initializes the locks and timers maps and returns a pointer to the newly created MemoryLocker.
func NewMemoryLocker() *MemoryLocker {
	ml := &MemoryLocker{
		locks:  make(map[string]string),
		timers: make(map[string]*time.Timer),
	}
	ml.cond = sync.NewCond(&ml.mu)
	return ml
}

// Lock locks the memory locker for the given key and ID for a specified duration.
// If the memory locker is already locked by someone else, the function will block until it is unlocked.
// Once locked, the memory locker will automatically unlock after the specified hold duration.
// If the lock is released before the hold duration expires, the timer will be stopped.
// After the hold duration expires, the lock will be released and the corresponding timer will be removed.
// The function is thread-safe.
func (ml *MemoryLocker) Lock(key string, id string, holdDuration time.Duration) error {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	for ml.locks[key] != "" && ml.locks[key] != id {
		ml.cond.Wait()
	}

	if timer, ok := ml.timers[key]; ok {
		timer.Stop()
	}

	ml.locks[key] = id
	timer := time.AfterFunc(holdDuration, func() {
		ml.mu.Lock()
		defer ml.mu.Unlock()
		if ml.locks[key] == id {
			delete(ml.locks, key)
			delete(ml.timers, key)
			ml.cond.Broadcast()
		}
	})

	ml.timers[key] = timer
	return nil
}

// Unlock releases the lock for the given key and ID.
// If the ID does not match the one that holds the lock, an error is returned.
// It also stops the timer associated with the key, if any.
// After releasing the lock, it broadcasts a signal to wake up any goroutines waiting on the lock.
func (ml *MemoryLocker) Unlock(key string, id string) error {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if ml.locks[key] != id {
		return errors.New("the id does not match the one that holds the lock")
	}
	delete(ml.locks, key)
	if timer, ok := ml.timers[key]; ok {
		timer.Stop()
	}
	delete(ml.timers, key)
	ml.cond.Broadcast()
	return nil
}
