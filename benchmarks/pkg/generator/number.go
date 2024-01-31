package generator

// Number is a common generator.
type Number struct {
	LastValue int64
}

// SetLastValue sets the last value generated.
func (n *Number) SetLastValue(value int64) {
	n.LastValue = value
}

// Last implements the Generator Last interface.
func (n *Number) Last() int64 {
	return n.LastValue
}
