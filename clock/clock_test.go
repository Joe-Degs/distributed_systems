package clock

import (
	"fmt"
	"testing"
)

type clocks struct {
	id []string
	vc []*Vector
}

func (c *clocks) announce(idx int, id string) {
	for index, cl := range c.vc {
		if index == idx {
			continue
		}
		cl.AddMember(id)
	}
}

func (c *clocks) get(id string) *Vector {
	for idx, ids := range c.id {
		if ids == id {
			return c.vc[idx]
		}
	}
	return nil
}

func clauster(ids ...string) *clocks {
	clockGroup := &clocks{
		id: ids,
		vc: make([]*Vector, 0, len(ids)),
	}
	for _, id := range ids {
		//fmt.Println(id)
		newClock := New(id)
		clockGroup.vc = append(clockGroup.vc, newClock)
	}

	// how do we broadcast the clocks to each possible node
	// every node broadcasts their existence to every other node
	// in the clock group.
	for idx := range clockGroup.vc {
		clockGroup.announce(idx, clockGroup.id[idx])
	}
	return clockGroup
}

func TestAddMember(t *testing.T) {
	newClauster := clauster("a", "b", "c")

	// add new member to the clock group.
	a := newClauster.get("a")
	a.Increment()
	a.AddMember("d")
	a.AddMember("e")

	// merge two events and see what happens
	newClauster.get("c").Merge(a)

	for _, cl := range newClauster.vc {
		fmt.Println(cl)
	}
}
