package main

import (
	"fmt"
	"time"
)

type TimeTest struct {
	Time time.Time
	Name string
}

func main() {
	defer fmt.Println(1)
	defer fmt.Println(2)
	defer fmt.Println(3)
}
