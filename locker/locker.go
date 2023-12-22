package locker

import "time"

type Locker interface {
	Lock(key string, id string, holdDuration time.Duration) error
	Unlock(key string, id string) error
}
