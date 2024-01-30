package generator

import "math/rand"

type discretePair struct {
	Weight float64
	Value  int64
}

// Discrete generates a distribution by choosing from a discrete set of values.
type Discrete struct {
	Number
	values []discretePair
}

// NewDiscrete creates the generator.
func NewDiscrete() *Discrete {
	return &Discrete{}
}

// Next implements the Generator Next interface.
func (d *Discrete) Next(r *rand.Rand) int64 {
	sum := float64(0)

	for _, p := range d.values {
		sum += p.Weight
	}

	val := r.Float64()

	for _, p := range d.values {
		pw := p.Weight / sum
		if val < pw {
			d.SetLastValue(p.Value)
			return p.Value
		}

		val -= pw
	}

	panic("oops, should not get here.")
}

// Add adds a value with weight.
func (d *Discrete) Add(weight float64, value int64) {
	d.values = append(d.values, discretePair{Weight: weight, Value: value})
}
