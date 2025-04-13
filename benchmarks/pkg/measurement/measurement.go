package measurement

import (
	"benchmark/ycsb"
	"bufio"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var header = []string{"Operation", "Takes(s)", "Count", "OPS", "Avg(us)", "Min(us)", "Max(us)", "50th(us)", "90th(us)", "95th(us)", "99th(us)", "99.9th(us)", "99.99th(us)"}

type measurement struct {
	sync.RWMutex

	measurer ycsb.Measurer
}

func (m *measurement) measure(op string, start time.Time, lan time.Duration) {
	m.Lock()
	m.measurer.Measure(op, start, lan)
	m.Unlock()
}

func (m *measurement) output(outFile string) {
	m.RLock()
	defer m.RUnlock()

	var w *bufio.Writer
	if outFile == "" {
		w = bufio.NewWriter(os.Stdout)
	} else {
		f, err := os.Create(outFile)
		if err != nil {
			panic("failed to create output file: " + err.Error())
		}
		defer f.Close()
		w = bufio.NewWriter(f)
	}

	err := m.measurer.Output(w)
	if err != nil {
		panic("failed to write output: " + err.Error())
	}

	err = w.Flush()
	if err != nil {
		panic("failed to flush output: " + err.Error())
	}
}

func (m *measurement) summary() {
	m.RLock()
	globalMeasure.measurer.Summary()
	m.RUnlock()
}

// InitMeasure initializes the global measurement.
func InitMeasure() {
	globalMeasure = new(measurement)
	measurementType := "histogram"
	switch measurementType {
	case "histogram":
		globalMeasure.measurer = InitHistograms()
	case "raw", "csv":
		globalMeasure.measurer = InitCSV()
	default:
		panic("unsupported measurement type: " + measurementType)
	}

	faultToleranceMeasure = new(measurement)
	faultToleranceMeasure.measurer = InitFCSV()
	// EnableWarmUp(p.GetInt64(prop.WarmUpTime, 0) > 0)
}

// Output prints the complete measurements.
func Output() {
	globalMeasure.measurer.GenerateExtendedOutputs()
	globalMeasure.output("")
	fmt.Println("##################################################")
	faultToleranceMeasure.measurer.GenerateExtendedOutputs()
	faultToleranceMeasure.output("timeline.csv")
}

// Summary prints the measurement summary.
func Summary() {
	globalMeasure.summary()
}

// EnableWarmUp sets whether to enable warm-up.
func EnableWarmUp(b bool) {
	if b {
		atomic.StoreInt32(&warmUp, 1)
	} else {
		atomic.StoreInt32(&warmUp, 0)
	}
}

// IsWarmUpFinished returns whether warm-up is finished or not.
func IsWarmUpFinished() bool {
	return atomic.LoadInt32(&warmUp) == 0
}

// Measure measures the operation.
func Measure(op string, start time.Time, lan time.Duration) {
	if IsWarmUpFinished() {
		globalMeasure.measure(op, start, lan)
	}
}

func FMeasure(op string, end time.Time, lan time.Duration) {
	faultToleranceMeasure.measure(op, end, lan)
}

var globalMeasure *measurement
var faultToleranceMeasure *measurement
var warmUp int32 // use as bool, 1 means in warmup progress, 0 means warmup finished.
