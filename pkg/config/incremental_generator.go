package config

import (
	"strconv"
	"sync"
)

type IncrementalGenerator struct {
	mu   sync.Mutex
	next int
}

func NewIncrementalGenerator() *IncrementalGenerator {
	return &IncrementalGenerator{next: 0}
}

func (i *IncrementalGenerator) GenerateId() string {
	i.mu.Lock()
	defer i.mu.Unlock()
	str := "T" + strconv.Itoa(i.next)
	i.next++
	return str
}
