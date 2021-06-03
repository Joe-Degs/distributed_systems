// Package clock implements a send only vector clock.
// This clock does not increment on merges(recieves). It only
// increments when its increment method is called.
// will be useful for my distributed system series.
package clock

import (
	"fmt"
	"strings"
)

type Vector struct {
	id  string
	val map[string]int
}

// New returns a vector clock with an id.
func New(id string) *Vector {
	vec := &Vector{val: make(map[string]int)}
	vec.id = id
	vec.val[vec.id] = 0
	return vec
}

// AddMember add new member to clock group
func (v *Vector) AddMember(id string) {
	v.val[id] = 0
}

// String returns a nicely formatted string of clocks value
func (v *Vector) String() string {
	return strings.Split(fmt.Sprintf("%v", v.val), "p")[1]
}

// Get the latest timestamp of the Vector
func (v *Vector) Get() int {
	return v.val[v.id]
}

// Increment increments the timestamp of the Vector
func (v *Vector) Increment() {
	v.val[v.id]++
}

// Merge consolidates two clocks, keeping the highest
// values in the clocks
func (v *Vector) Merge(vc *Vector) {
	for k, val := range vc.val {
		// if reciever has not seen this (id, value)
		// pair, add it to its clock
		if _, ok := v.val[k]; !ok {
			v.val[k] = val
		}
	}

	// now the reciever vector is either bigger
	// or equal to the that of the sender vector
	// use the pointwise maximum to see which
	// is bigger and keep those
	for id, val := range v.val {
		if vc.val[id] > val {
			v.val[id] = vc.val[id]
		}
	}
}
