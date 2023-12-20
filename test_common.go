package main

type Person struct {
	Name string
	Age  int
}

type TestItem struct {
	Value string
}

func NewDefaultPerson() Person {
	return Person{
		Name: "John",
		Age:  30,
	}
}

func NewTestItem(value string) TestItem {
	return TestItem{
		Value: value,
	}
}

var inputItemList = []TestItem{
	NewTestItem("item1"),
	NewTestItem("item2"),
	NewTestItem("item3"),
	NewTestItem("item4"),
	NewTestItem("item5"),
}
