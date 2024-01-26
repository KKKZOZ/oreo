package mock

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedis_DebugCounter(t *testing.T) {
	t.Run("test less than limit", func(t *testing.T) {

		conn := NewMockRedisConnection("localhost", 6379, 10, true,
			0, func() error { return errors.New("error") })

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
		conn := NewMockRedisConnection("localhost", 6379, 3, true,
			0, func() error { return errors.New("error") })

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

func TestRedis_DebugFunc(t *testing.T) {

	t.Run("trigger debugFunc", func(t *testing.T) {
		conn := NewMockRedisConnection("localhost", 6379, 3, true,
			0, func() error { return errors.New("my error") })

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
		conn := NewMockRedisConnection("localhost", 6379, 3, true,
			0, func() error { return errors.New("my error") })

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

func TestRedis_CallTimes(t *testing.T) {
	testCases := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, testCase := range testCases {
		conn := NewMockRedisConnection("localhost", 6379, 20, true,
			0, func() error { return errors.New("my error") })

		for i := 0; i < testCase; i++ {
			err := conn.Put("key1", "value1")
			assert.Nil(t, err)
		}
		assert.Equal(t, testCase, conn.PutTimes)
	}
}
