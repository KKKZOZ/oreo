package util

import (
	"encoding/binary"
	"hash/fnv"
)

// Hash64 returns a fnv Hash of the integer.
func Hash64(n int64) int64 {
	var b [8]byte
	binary.BigEndian.PutUint64(b[0:8], uint64(n))
	hash := fnv.New64a()
	hash.Write(b[0:8])
	result := int64(hash.Sum64())
	if result < 0 {
		return -result
	}
	return result
}

// BytesHash64 returns the fnv hash of a bytes
func BytesHash64(b []byte) int64 {
	hash := fnv.New64a()
	hash.Write(b)
	return int64(hash.Sum64())
}

// StringHash64 returns the fnv hash of a string
func StringHash64(s string) int64 {
	hash := fnv.New64a()
	hash.Write(Slice(s))
	return int64(hash.Sum64())
}
