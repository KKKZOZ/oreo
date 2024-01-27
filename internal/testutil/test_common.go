package testutil

import (
	"fmt"
	"math"
	"net/http"
	"time"
)

type Person struct {
	Name string
	Age  int
}

type TestItem struct {
	Value string
}

// NewDefaultPerson returns a new testutil.Person with default values.
// The default values are:
// - Name: "John"
// - Age: 30
func NewDefaultPerson() Person {
	return Person{
		Name: "John",
		Age:  30,
	}
}

// NewPerson creates a new testutil.Person with the given name and default age of 30.
func NewPerson(name string) Person {
	return Person{
		Name: name,
		Age:  30,
	}
}

// NewTestItem creates a new testutil.TestItem with the specified value.
func NewTestItem(value string) TestItem {
	return TestItem{
		Value: value,
	}
}

var InputItemList = []TestItem{
	NewTestItem("item1"),
	NewTestItem("item2"),
	NewTestItem("item3"),
	NewTestItem("item4"),
	NewTestItem("item5"),
}

// WaitForServer waits for the server at the specified address and port to respond with a heartbeat within the given timeout duration.
// It periodically sends HTTP GET requests to the server's heartbeat endpoint until a successful response is received or the timeout is reached.
// If a successful response is received within the timeout, it returns nil. Otherwise, it returns an error indicating that the server did not reply within the specified timeout.
func WaitForServer(address string, port int, timeout time.Duration) error {
	URL := fmt.Sprintf("http://%s:%d/heartbeat", address, port)
	ch := make(chan bool)
	go func() {
		for {
			_, err := http.Get(URL)
			if err == nil {
				ch <- true
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("server did not reply after %v", timeout)
	}
}

func RoughlyEqual(a, b time.Duration, threshold time.Duration) bool {
	return math.Abs(float64(a-b)) <= float64(threshold)
}

func RoughlyLessThan(result, expected time.Duration, threshold time.Duration) bool {
	return result <= expected+threshold
}
