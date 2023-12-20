package main

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDebugCounter(t *testing.T) {
	t.Run("test less than limit", func(t *testing.T) {
		memoryDatabase := NewMemoryDatabase("localhost", 8321)
		go memoryDatabase.start()
		defer func() { <-memoryDatabase.msgChan }()
		defer func() { go memoryDatabase.stop() }()
		time.Sleep(100 * time.Millisecond)

		conn := NewMockMemoryConnection("localhost", 8321, 10, true,
			func() error { return errors.New("error") })

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
		memoryDatabase := NewMemoryDatabase("localhost", 8321)
		go memoryDatabase.start()
		defer func() { <-memoryDatabase.msgChan }()
		defer func() { go memoryDatabase.stop() }()
		time.Sleep(100 * time.Millisecond)

		conn := NewMockMemoryConnection("localhost", 8321, 3, true,
			func() error { return errors.New("error") })

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

func TestDebugFunc(t *testing.T) {

	t.Run("trigger debugFunc", func(t *testing.T) {
		memoryDatabase := NewMemoryDatabase("localhost", 8321)
		go memoryDatabase.start()
		defer func() { <-memoryDatabase.msgChan }()
		defer func() { go memoryDatabase.stop() }()
		time.Sleep(100 * time.Millisecond)

		conn := NewMockMemoryConnection("localhost", 8321, 3, true,
			func() error { return errors.New("my error") })

		err := conn.Put("key1", "value1")
		assert.Nil(t, err)

		err = conn.Put("key2", "value2")
		assert.Nil(t, err)

		err = conn.Put("key3", "value3")
		assert.Nil(t, err)

		err = conn.Put("key4", "value4")
		assert.EqualError(t, err, "my error")
	})

	t.Run("after triggerring debugFunc", func(t *testing.T) {
		memoryDatabase := NewMemoryDatabase("localhost", 8321)
		go memoryDatabase.start()
		defer func() { <-memoryDatabase.msgChan }()
		defer func() { go memoryDatabase.stop() }()
		time.Sleep(100 * time.Millisecond)

		conn := NewMockMemoryConnection("localhost", 8321, 3, true,
			func() error { return errors.New("my error") })

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

func TestCallTimes(t *testing.T) {
	testCases := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, testCase := range testCases {
		memoryDatabase := NewMemoryDatabase("localhost", 8321)
		go memoryDatabase.start()
		defer func() { <-memoryDatabase.msgChan }()
		defer func() { go memoryDatabase.stop() }()
		time.Sleep(100 * time.Millisecond)

		conn := NewMockMemoryConnection("localhost", 8321, 20, true,
			func() error { return errors.New("my error") })

		for i := 0; i < testCase; i++ {
			err := conn.Put("key1", "value1")
			assert.Nil(t, err)
		}
		assert.Equal(t, testCase, conn.callTimes)
	}
}
