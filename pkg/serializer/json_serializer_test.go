package serializer

import (
	"reflect"
	"testing"
)

type TestStruct struct {
	Number int
	String string
}

func TestNewJSONSerializer(t *testing.T) {
	if reflect.ValueOf(NewJSONSerializer()).IsNil() {
		t.Error("NewJSONSerializer() should not return nil")
	}
}

func TestJSONSerializer_Serialize(t *testing.T) {
	s := NewJSONSerializer()

	testStruct := TestStruct{Number: 123, String: "abc"}
	bs, err := s.Serialize(testStruct)
	if err != nil {
		t.Errorf("Serialize error = %v, wantErr nil", err)
	}

	if len(bs) == 0 {
		t.Errorf("Serialize() byte slice should not be empty")
	}

	// Serialize non-serializable types.
	_, err = s.Serialize(make(chan int))
	if err == nil {
		t.Error("Serialize() should return an error for non-json-marshalable types")
	}
}

func TestJSONSerializer_Deserialize(t *testing.T) {
	s := NewJSONSerializer()

	// Test data
	testStruct := TestStruct{Number: 123, String: "abc"}
	bs, _ := s.Serialize(testStruct)

	// Where to store loaded data
	var loadedStruct TestStruct
	if err := s.Deserialize(bs, &loadedStruct); err != nil {
		t.Errorf("Deserialize() error = %v", err)
	}

	if !reflect.DeepEqual(testStruct, loadedStruct) {
		t.Errorf("Original and deserialized data do not match. Original = %v, Deserialized = %v", testStruct, loadedStruct)
	}
}
