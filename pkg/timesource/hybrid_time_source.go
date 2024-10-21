package timesource

import (
	"math"
	"sync"
	"time"
)

type HybridTimeSource struct {
	physicalTimeUpdateInterval int
	logicalTimeBits            int

	// 逻辑时间的最大值，2^logicalTimeBits - 1
	maxLogicalTime int64

	physicalTime int64 // 物理时间 (精确到毫秒)
	logicalTime  int64 // 逻辑时间
	mu           sync.Mutex
}

var _ TimeSourcer = (*HybridTimeSource)(nil)

func NewHybridTimeSource(physicalTimeUpdateInterval int, logicalTimeBits int) *HybridTimeSource {
	ts := &HybridTimeSource{
		physicalTimeUpdateInterval: physicalTimeUpdateInterval,
		logicalTimeBits:            logicalTimeBits,
	}
	ts.maxLogicalTime = (1 << ts.logicalTimeBits) - 1
	ts.physicalTime = time.Now().UnixMilli()
	go ts.updatePhysicalTime()
	return ts
}

func (ts *HybridTimeSource) updatePhysicalTime() {
	ticker := time.NewTicker(time.Duration(ts.physicalTimeUpdateInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		<-ticker.C
		ts.mu.Lock()
		ts.physicalTime = time.Now().UnixMilli()
		ts.logicalTime = 0
		ts.mu.Unlock()
	}
}

func (ts *HybridTimeSource) GetTime(mode string) (int64, error) {

	ts.mu.Lock()
	defer ts.mu.Unlock()

	// 如果逻辑时间即将超过上限，更新物理时间并重置逻辑时间
	if ts.logicalTime >= ts.maxLogicalTime-10 {
		// fmt.Printf("Triggered physical time update\n")
		ts.physicalTime = time.Now().UnixMilli() // 更新物理时间
		ts.logicalTime = 0                       // 重置逻辑时间
	}
	ts.logicalTime++

	// 将物理时间和逻辑时间打包成一个 int64
	timestamp := ts.physicalTime*int64(math.Pow10(ts.logicalTimeBits)) + ts.logicalTime
	return timestamp, nil
}
