package measurement

import (
	"fmt"
	"io"
	"time"
)

type fcsventry struct {
	// end time of the operation in us from unix epoch
	endUs int64
	// latency of the operation in us
	latencyUs int64
}

type fcsvs struct {
	opCsv map[string][]fcsventry
}

func (c *fcsvs) GenerateExtendedOutputs() {
}

func InitFCSV() *fcsvs {
	return &fcsvs{
		opCsv: make(map[string][]fcsventry),
	}
}

func (c *fcsvs) Measure(op string, start time.Time, lan time.Duration) {
	c.opCsv[op] = append(c.opCsv[op], fcsventry{
		endUs:     start.UnixMicro(),
		latencyUs: lan.Microseconds(),
	})
}

func (c *fcsvs) Output(w io.Writer) error {
	_, err := fmt.Fprintln(w, "operation,timestamp_us,latency_us")
	if err != nil {
		return err
	}
	for op, entries := range c.opCsv {
		for _, entry := range entries {
			_, err := fmt.Fprintf(w, "%s,%d,%d\n", op, entry.endUs, entry.latencyUs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *fcsvs) Summary() {
	// do nothing as csvs don't keep a summary
}
