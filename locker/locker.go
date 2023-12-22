package locker

import "time"

// Locker is an interface that defines the methods for locking and unlocking a resource.
type Locker interface {
	// Lock locks the specified resource with the given key and ID for the specified duration.
	// It returns an error if the resource cannot be locked.
	Lock(key string, id string, holdDuration time.Duration) error

	// Unlock unlocks the specified resource with the given key and ID.
	// It returns an error if the resource cannot be unlocked.
	Unlock(key string, id string) error
}
