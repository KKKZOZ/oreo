package serializer

import (
	"reflect"
	"testing"
)

type User struct {
	ID   int
	Name string
}

type NonExportedStruct struct {
	secret string
}

func TestNewGobSerializer(t *testing.T) {
	if reflect.ValueOf(NewGobSerializer()).IsNil() {
		t.Error("NewGobSerializer() should not return nil")
	}
}

func TestGobSerializer_Serialize(t *testing.T) {
	s := NewGobSerializer()

	// Test data.
	user1 := User{ID: 123, Name: "John Doe"}
	bs, err := s.Serialize(user1)
	if err != nil {
		t.Errorf("Serialize() error = %v, wantErr nil", err)
	}

	if len(bs) == 0 {
		t.Errorf("Serialize() byte slice should not be empty")
	}

	// Test non-public struct.
	type nonExported struct {
		secret string
	}
	ne := nonExported{secret: "shh"}

	_, err2 := s.Serialize(ne)
	if err2 == nil {
		t.Errorf("Serialize() should return an error for non-public structs")
	}
}

func TestGobSerializer_Deserialize(t *testing.T) {
	s := NewGobSerializer()

	// Test data.
	user1 := User{ID: 123, Name: "John Doe"}
	bs, _ := s.Serialize(user1)

	// Where to store the loaded data.
	var user2 User
	err := s.Deserialize(bs, &user2)
	if err != nil {
		t.Errorf("Deserialize() error = %v, wantErr nil", err)
	}

	if user1.ID != user2.ID || user1.Name != user2.Name {
		t.Errorf("Deserialize() = %v, want %v", user2, user1)
	}
}
