package ycsb

import (
	"io"
	"time"
)

// Measurer is used to capture measurements.
type Measurer interface {
	// Measure measures the latency of an operation.
	Measure(op string, start time.Time, latency time.Duration)

	// Summary writes a summary of the current measurement results to stdout.
	Summary()

	// GenerateExtendedOutputs is called at the end of the benchmark
	GenerateExtendedOutputs()

	// Output writes the measurement results to the specified writer.
	Output(w io.Writer) error
}
