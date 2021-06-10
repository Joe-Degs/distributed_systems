// Package clock implements an increment on send events only clock.
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
func (v *Vector) AddMember(id string, val int) {
	v.val[id] = val
}

func (v *Vector) GetId() string {
	return v.id
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

// Decrement
func (v *Vector) Decrement() {
	v.val[v.id]--
}

// Merge compares two clocks, keeping the highest values
// at each point in the clock.
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
