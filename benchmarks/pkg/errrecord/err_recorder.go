package errrecord

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
)

type ErrRecorder struct {
	opMap map[string]ErrCntMap
}

func NewErrRecorder() *ErrRecorder {
	return &ErrRecorder{
		opMap: make(map[string]ErrCntMap),
	}
}

func (er *ErrRecorder) Record(op string, err error) {
	if err == nil {
		return
	}
	if _, ok := er.opMap[op]; !ok {
		er.opMap[op] = make(ErrCntMap)
	}
	er.opMap[op][err.Error()]++
}

func (er *ErrRecorder) Summary() {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)

	fmt.Println("Error Summary:")
	for op, errMap := range er.opMap {
		errList := make([]ErrCntItem, 0)
		for err, cnt := range errMap {
			errList = append(errList, ErrCntItem{err, cnt})
		}
		// Sort the error list by error count in descending order.
		sort.Slice(errList, func(i, j int) bool {
			return errList[i].Count > errList[j].Count
		})

		// Print the operation header.
		fmt.Fprintf(writer, "\nOperation:\t%s\t\n", op)
		fmt.Fprintf(writer, "Error\tCount\t\n")
		fmt.Fprintf(writer, "-----\t-----\t\n")

		// Print the errors and counts in tabular format.
		for _, item := range errList {
			fmt.Fprintf(writer, "%s\t%d\t\n", item.Err, item.Count)
		}
	}

	// Make sure to flush the writer to output the content to the console.
	writer.Flush()
}
