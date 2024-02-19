package client

import (
	"benchmark/pkg/generator"
	"benchmark/pkg/util"
	"math/rand"
	"slices"
	"testing"
	"time"
)

func TestZipfian(t *testing.T) {

	recordCount := 1000

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	insertStart := int64(0)
	insertCount := int64(recordCount) - insertStart

	keySequence := generator.NewCounter(insertStart)

	var keyrangeLowerBound int64 = insertStart
	var keyrangeUpperBound int64 = insertStart + insertCount - 1
	keyChooser := generator.NewScrambledZipfian(keyrangeLowerBound, keyrangeUpperBound, generator.ZipfianConstant)

	items := make([]int64, recordCount)

	for i := 0; i < recordCount; i++ {
		key := keySequence.Next(r)
		items[i] = util.Hash64(key)
	}

	for i := 0; i < recordCount; i++ {
		key := keyChooser.Next(r)
		item := util.Hash64(key)

		if !slices.Contains(items, item) {
			t.Errorf("item %d not found in items", item)
		}

	}
}
