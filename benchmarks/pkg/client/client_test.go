package client

import (
	"math/rand"
	"sort"
	"testing"
	"time"

	"benchmark/pkg/generator"
)

func TestZipfian(t *testing.T) {
	recordCount := 10000

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	insertStart := int64(0)
	insertCount := int64(recordCount) - insertStart

	// keySequence := generator.NewCounter(insertStart)

	var keyrangeLowerBound int64 = insertStart
	var keyrangeUpperBound int64 = insertStart + insertCount - 1
	keyChooser := generator.NewScrambledZipfian(
		keyrangeLowerBound,
		keyrangeUpperBound,
		generator.ZipfianConstant,
	)

	// items := make([]int64, recordCount)
	itemMap := make(map[int64]int)

	// for i := 0; i < recordCount; i++ {
	// 	key := keySequence.Next(r)
	// }

	for i := 0; i < recordCount; i++ {
		key := keyChooser.Next(r)
		itemMap[key]++
	}

	type Pair struct {
		Key   int64
		Value int
	}

	pairs := make([]Pair, 0)
	for k, v := range itemMap {
		pairs = append(pairs, Pair{k, v})
	}

	// sort pairs by value in descending order
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value
	})

	for i := 0; i < 10; i++ {
		t.Logf("%d: %d", pairs[i].Key, pairs[i].Value)
	}

	t.Fail()
}
