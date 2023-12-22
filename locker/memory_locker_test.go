package locker

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSimpleLockAndUnlock tests the functionality of locking and unlocking a key in the memory locker.
func TestSimpleLockAndUnlock(t *testing.T) {
	locker := NewMemoryLocker()
	err := locker.Lock("key1", "txnId-1", 100*time.Millisecond)
	assert.Nil(t, err)
	err = locker.Unlock("key1", "txnId-1")
	assert.Nil(t, err)
}

// TestLockTwice tests the behavior of locking the same key twice using the MemoryLocker.
// It verifies that the second lock request is successful and that the duration between the two locks is greater than 90 milliseconds.
func TestLockTwice(t *testing.T) {
	locker := NewMemoryLocker()
	err := locker.Lock("key1", "txnId-1", 100*time.Millisecond)
	assert.Nil(t, err)
	startTime := time.Now()
	err = locker.Lock("key1", "txnId-2", 100*time.Millisecond)
	assert.Nil(t, err)
	endTime := time.Now()
	assert.True(t, endTime.Sub(startTime) > 90*time.Millisecond)
}

// TestConcurrentLock tests the concurrent locking and unlocking of a counter using the MemoryLocker.
func TestConcurrentLock(t *testing.T) {
	locker := NewMemoryLocker()

	counter := 0

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(100)

	for i := 1; i <= 100; i++ {
		go func(id string) {
			err := locker.Lock("counter", id, 50*time.Millisecond)
			assert.Nil(t, err)
			counter++
			err = locker.Unlock("counter", id)
			assert.Nil(t, err)
			waitGroup.Done()
		}(fmt.Sprint(i))
	}

	waitGroup.Wait()
	assert.Equal(t, 100, counter)
}

// checks that a lock can be acquired and then released without error.
func TestMemoryLocker_Lock_Then_Unlock(t *testing.T) {
	locker := NewMemoryLocker()

	err := locker.Lock("key1", "id1", time.Second*5)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	err = locker.Unlock("key1", "id1")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// ensures that trying to unlock with a different id should result in an error.
func TestMemoryLocker_Unlock_Fail(t *testing.T) {
	locker := NewMemoryLocker()

	_ = locker.Lock("key2", "id1", time.Second*5)

	err := locker.Unlock("key2", "id2")
	if err == nil {
		t.Fatalf("Expected an error, but got none")
	}
}

// verifies that a lock is automatically released after its hold duration.
func TestMemoryLocker_AutoUnlock(t *testing.T) {
	locker := NewMemoryLocker()

	_ = locker.Lock("key3", "id1", time.Second*1)

	time.Sleep(time.Second * 2) // wait for auto-unlock

	err := locker.Lock("key3", "id2", time.Second*5) // try locking again
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	err = locker.Unlock("key3", "id2")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// tests that the locking operation is blocking until the lock becomes available.
func TestMemoryLocker_Lock_Blocking(t *testing.T) {
	locker := NewMemoryLocker()

	_ = locker.Lock("key4", "id1", time.Second*5)

	go func() {
		time.Sleep(time.Second * 2)  // letting Lock to be held for a while
		locker.Unlock("key4", "id1") // unlock in another goroutine
	}()

	err := locker.Lock("key4", "id2", time.Second*5) // should block until lock is released
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	err = locker.Unlock("key4", "id2")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestMemoryLocker_LockSameID tests if the lock can be acquired by the same id before hold duration ends
func TestMemoryLocker_LockSameID(t *testing.T) {
	locker := NewMemoryLocker()

	_ = locker.Lock("key5", "id1", time.Second*5)

	// Try to reacquire the lock with the same id before hold duration ends
	err := locker.Lock("key5", "id1", time.Second*2)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	err = locker.Unlock("key5", "id1")
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
}

// TestMemoryLocker_ConcurrentLock_Unlock tests concurrent locking and unlocking
func TestMemoryLocker_ConcurrentLock_Unlock(t *testing.T) {
	locker := NewMemoryLocker()

	go func() {
		_ = locker.Lock("key6", "id1", time.Second*60)
		time.Sleep(time.Second * 5)
		_ = locker.Unlock("key6", "id1")
	}()

	time.Sleep(time.Second * 2)
	err := locker.Lock("key6", "id2", time.Second*60)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	err = locker.Unlock("key6", "id2")
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
}

// TestMemoryLocker_MultipleKeys tests locking and unlocking on multiple keys
func TestMemoryLocker_MultipleKeys(t *testing.T) {
	locker := NewMemoryLocker()
	keys := []string{"key7", "key8", "key9"}
	ids := []string{"id1", "id2", "id3"}

	for i, key := range keys {
		err := locker.Lock(key, ids[i], time.Second*60)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
	}

	for i, key := range keys {
		err := locker.Unlock(key, ids[i])
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
	}
}

// TestMemoryLocker_HoldDurationExpires tests if the lock is automatically released when the holdDuration expires.
func TestMemoryLocker_HoldDurationExpires(t *testing.T) {
	locker := NewMemoryLocker()

	_ = locker.Lock("key10", "id1", time.Second*1)

	time.Sleep(time.Second * 2)

	// Reacquire the lock immediately after the holdDuration expires
	err := locker.Lock("key10", "id2", time.Second*1)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	err = locker.Unlock("key10", "id2")
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
}

// TestMemoryLocker_ConcurrentLockSame tests concurrent locking on the same key
func TestMemoryLocker_ConcurrentLockSame(t *testing.T) {
	locker := NewMemoryLocker()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = locker.Lock("key11", "id1", time.Second*4)
		time.Sleep(time.Second * 2)
		_ = locker.Unlock("key11", "id1") // Unlock after 2 seconds
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 1)                    // Wait 1 second later to ensure the previous goroutine secures the lock first
		_ = locker.Lock("key11", "id2", time.Second*2) // Will block and wait for the lock
		_ = locker.Unlock("key11", "id2")
	}()

	wg.Wait()
	// If this test hangs, it means the locker doesn't handle concurrent locking properly
}

// TestMemoryLocker_AutoRelease tests if the lock is automatically released
func TestMemoryLocker_AutoRelease(t *testing.T) {
	locker := NewMemoryLocker()

	_ = locker.Lock("key12", "id1", time.Second*2)

	time.Sleep(time.Second * 4) // Wait for the lock to be auto-released

	// Attempt to lock the same key with a different id, should not block if the auto-release works properly
	err := locker.Lock("key12", "id2", time.Second*2)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	err = locker.Unlock("key12", "id2")
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
}

// TestMemoryLocker_MultipleConcurrentHolds tests multiple concurrent locks with different holdDurations
func TestMemoryLocker_MultipleConcurrentHolds(t *testing.T) {
	locker := NewMemoryLocker()

	count := int32(0)
	total := int32(100)

	for i := 0; i < int(total); i++ {
		go func(locker *MemoryLocker, id int) {
			err := locker.Lock("key16", fmt.Sprintf("id%d", id), time.Duration(id)*time.Millisecond)
			if err != nil {
				t.Fatalf("Failed to acquire the lock: %v", err)
			}

			atomic.AddInt32(&count, 1)
			time.Sleep(time.Duration(id) * time.Millisecond)

			err = locker.Unlock("key16", fmt.Sprintf("id%d", id))
			if err != nil {
				t.Fatalf("Failed to release the lock: %v", err)
			}

			atomic.AddInt32(&count, -1)
		}(locker, i)
	}

	time.Sleep(time.Second * 5)

	finalCount := atomic.LoadInt32(&count)
	if finalCount != 0 {
		t.Fatalf("Expected all locks to be released, but still held: %v", finalCount)
	}
}
