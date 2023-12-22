package timeoracle

import "time"

type TimeOracle interface {
	GetTime() time.Time
}
