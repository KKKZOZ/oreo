package timeoracle

import "time"

// TimeOracle is an interface that defines a method for retrieving the current time.
type TimeOracle interface {
	GetTime() time.Time
}
