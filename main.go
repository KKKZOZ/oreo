package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type TimeTest struct {
	Time time.Time
	Name string
}

func main() {
	timeTest := TimeTest{
		Time: time.Now(),
		Name: "John",
	}
	fmt.Println(timeTest)
	timeStr := toJSONString(timeTest)
	timeStr2 := toJSONString(timeStr)
	fmt.Println("1: " + timeStr)
	fmt.Println("2: " + timeStr2)
	var timeObj TimeTest
	err := json.Unmarshal([]byte(timeStr2), &timeObj)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(timeObj)
}
