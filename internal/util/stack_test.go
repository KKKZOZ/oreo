package util

import "testing"

func TestStack(t *testing.T) {
	tests := []struct {
		name           string
		input          []int // Inputs to push to the stack
		expectedPop    []int // Expected outcome when the items are popped
		expectedTop    int   // Expected outcome of top operation
		expectedTopErr bool  // Does the top operation supposed to return an error
	}{
		{"no inputs", []int{}, []int{}, 0, true},
		{"single input", []int{1}, []int{1}, 1, false},
		{"multiple inputs", []int{1, 2, 3}, []int{3, 2, 1}, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStack[int]()
			for _, i := range tt.input {
				s.Push(i)
			}

			// Test Top operation
			value, err := s.Top()
			if tt.expectedTopErr && err == nil {
				t.Errorf("Expected an error from Top() method but got no error")
			} else if !tt.expectedTopErr && err != nil {
				t.Errorf("Expected no error from Top() method but got an error = %v", err)
			} else if !tt.expectedTopErr && value != tt.expectedTop {
				t.Errorf("Expected value from Top() method = %d, but got = %d", tt.expectedTop, value)
			}

			// Test Pop operation
			var popped []int
			for s.Len() > 0 {
				val, err := s.Pop()
				if err != nil {
					t.Errorf("Expected no error from Pop() method but got an error = %v", err)
				}
				popped = append(popped, val)
			}

			if !equal(popped, tt.expectedPop) {
				t.Errorf(
					"When popping elements, expected = %v, but got = %v",
					tt.expectedPop,
					popped,
				)
			}
		})
	}
}

// Helper function to compare two slices
func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
