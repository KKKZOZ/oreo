package testutil

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type TxnTopic string

const (
	DConn    TxnTopic = "CONN"
	DRead    TxnTopic = "READ"
	DWrite   TxnTopic = "WRITE"
	DDelete  TxnTopic = "DELETE"
	DCommit  TxnTopic = "COMMIT"
	DAbort   TxnTopic = "ABORT"
	DInfo    TxnTopic = "INFO"
	DTest    TxnTopic = "TEST"
	DPrepare TxnTopic = "PREPARE"
	DConUpdt TxnTopic = "CONUPDT"
	DTSR     TxnTopic = "TSR"
)

var (
	debugStart     time.Time
	debugVerbosity int
)

func init() {
	debugVerbosity = getVerbosity()
	debugStart = time.Now()
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

// Retrieve the verbosity level from an environment variable
func getVerbosity() int {
	v := os.Getenv("VERBOSE")
	level := 0
	if v != "" {
		var err error
		level, err = strconv.Atoi(v)
		if err != nil {
			log.Fatalf("Invalid verbosity %v", v)
		}
	}
	return level
}

func Debug(topic TxnTopic, format string, a ...interface{}) {
	if debugVerbosity >= 1 {
		time := time.Since(debugStart).Microseconds()
		time /= 10
		prefix := fmt.Sprintf("%07d %v ", time, string(topic))
		format = prefix + format
		log.Printf(format, a...)
	}
}
