package mock

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMongo_DebugCounter(t *testing.T) {
	t.Run("test less than limit", func(t *testing.T) {

		conn := NewMockMongoConnection("localhost", 27017, 10, true,
			func() error { return errors.New("error") })
		conn.Connect()

		err := conn.Put("key1", "value1")
		assert.Nil(t, err)

		err = conn.Put("key2", "value2")
		assert.Nil(t, err)

		err = conn.Put("key3", "value3")
		assert.Nil(t, err)

		err = conn.Put("key4", "value4")
		assert.Nil(t, err, "error")
	})

	t.Run("test equal the limit", func(t *testing.T) {
		conn := NewMockMongoConnection("localhost", 27017, 3, true,
			func() error { return errors.New("error") })
		conn.Connect()

		err := conn.Put("key1", "value1")
		assert.Nil(t, err)

		err = conn.Put("key2", "value2")
		assert.Nil(t, err)

		err = conn.Put("key3", "value3")
		assert.Nil(t, err)

		err = conn.Put("key4", "value4")
		assert.EqualError(t, err, "error")
	})

}

func TestMongo_DebugFunc(t *testing.T) {

	t.Run("trigger debugFunc", func(t *testing.T) {
		conn := NewMockMongoConnection("localhost", 27017, 3, true,
			func() error { return errors.New("my error") })
		conn.Connect()

		err := conn.Put("key1", "value1")
		assert.Nil(t, err)

		err = conn.Put("key2", "value2")
		assert.Nil(t, err)

		err = conn.Put("key3", "value3")
		assert.Nil(t, err)

		err = conn.Put("key4", "value4")
		assert.EqualError(t, err, "my error")
	})

	t.Run("after triggering debugFunc", func(t *testing.T) {
		conn := NewMockMongoConnection("localhost", 27017, 3, true,
			func() error { return errors.New("my error") })
		conn.Connect()

		err := conn.Put("key1", "value1")
		assert.Nil(t, err)

		err = conn.Put("key2", "value2")
		assert.Nil(t, err)

		err = conn.Put("key3", "value3")
		assert.Nil(t, err)

		err = conn.Put("key4", "value4")
		assert.EqualError(t, err, "my error")

		err = conn.Put("key5", "value5")
		assert.Nil(t, err)

		err = conn.Put("key6", "value6")
		assert.Nil(t, err)
	})
}

func TestMongo_CallTimes(t *testing.T) {
	testCases := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, testCase := range testCases {
		conn := NewMockMongoConnection("localhost", 27017, 20, true,
			func() error { return errors.New("my error") })
		conn.Connect()

		for i := 0; i < testCase; i++ {
			err := conn.Put("key1", "value1")
			assert.Nil(t, err)
		}
		assert.Equal(t, testCase, conn.PutTimes)
	}
}
