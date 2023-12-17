package main

import (
	"errors"
	"fmt"
	"time"
)

type TimeTest struct {
	Time time.Time
	Name string
}

func main() {
	err1 := errors.New("hello")
	err2 := errors.New("hello")
	if err1.Error() == err2.Error() {
		fmt.Println("equal")
	} else {
		fmt.Println("not equal")
	}
}
