package generator

import (
	"fmt"
	"math/rand"

	"benchmark/pkg/util"
)

// ScrambledZipfian produces a sequence of items, such that some items are more popular than
// others, according to a zipfian distribution
type ScrambledZipfian struct {
	Number
	gen       *Zipfian
	min       int64
	max       int64
	itemCount int64
}

// NewScrambledZipfian creates a ScrambledZipfian generator.
func NewScrambledZipfian(min int64, max int64, zipfianConstant float64) *ScrambledZipfian {
	const (
		zetan               = float64(26.46902820178302)
		usedZipfianConstant = float64(0.99)
		itemCount           = int64(10000000000)
	)

	s := new(ScrambledZipfian)
	s.min = min
	s.max = max
	s.itemCount = max - min + 1
	// if zipfianConstant == usedZipfianConstant {
	// 	s.gen = NewZipfian(0, itemCount, zipfianConstant, zetan)
	// } else {
	fmt.Println("Re-calulate zetan...")
	s.gen = NewZipfianWithRange(0, s.itemCount, zipfianConstant)
	// }
	return s
}

// Next implements the Generator Next interface.
func (s *ScrambledZipfian) Next(r *rand.Rand) int64 {
	n := s.gen.Next(r)

	n = s.min + util.Hash64(n)%s.itemCount
	s.SetLastValue(n)
	return n
}
