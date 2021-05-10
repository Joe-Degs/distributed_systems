package clocks

import (
	"fmt"
	"strings"
)

// This package implements bunch of logical clocks i'm learning about
// as part of a distributed systems course. Clocks help in ordering events
// in systems and keeping order is a very important topic in distributed systems.
// On single node machines, ordering is abstracted away by the lower level layers.
// Writing concurrent code has never been easier, but now we communicate and
// distribute work to multiple machines sitting in several datacenters at
// separate parts of the world. We have to coordinate actions in these systems,
// its our job to maintain order in the systems and how do we do that with
// all the nodes in the system as unreliable as you can imagine?.

// This is why logical clocks are very important to distributed computing because
// they help us to have some kind of consensus in the system.
// The whole concept boils down to figuring out what happens before what based on
// the clock of the actors in the system. If we say A -> B ('->' happens before)
// then the clock value of A < clock value of B.

// Clock interface is to make testing easy. And i think this is really cool
// stuff over here.
type Clock interface {
	Get() interface{} // returns the current value of the clock
	Increment()       // increases the value of the clock
	Merge(Clock)      // finds the max of clock values and increments
	HappensBefore(Clock) bool
	String() string
}

// Couple of assumptions we can make from A -> B is that
// 1. A could have been the cause of B
// 2. The two are related in a way
// 3. A and B are in the same process and A happens before B
// 4. A is a send operation and B is the corresponding recieve
// 5. if A -> B and B -> C then A -> B(transitive closure) is likely to be true

// Clocks help in the ordering of events in our systems with the following
// sequences. The assumption is you have your systems wired to communicate.
// And all the nodes in the systems are keeping some kind of integer counter as
// the their clock

// -> If you generate an event in your system, increase your clock by 1
// -> If you send a message, add your clock to the message
// -> If you recieve a message, take max of the sender's clock value and your
//    clock value and increase by 1 and thats your clock value

// Lamport clocks have the following relation A -> B => LC(A) <= LC(B)
// meaning if A happens before B then the lamport clock of A is less than or equal
// to that of B but it doesnt work the other way round.
// Lamport clocks is a very cool concept all that, but here's the catch, in real
// world systems if two independent systems that have broken connection and
// are not communicating but generating events internally finally get back
// communicating, comparing their clocks will show that some events happened
// before some and vice versa but that will be erroneous because those events
// are not causally connected per our definition of what makes events connected
// based on their clocks.
// also my english sucks bad.
type LamportClock struct {
	val int
}

func NewLamportClock() Clock { return &LamportClock{} }

func (l *LamportClock) String() string {
	return fmt.Sprintf("%d", l.val)
}

func (l *LamportClock) max(j interface{}) int {
	if l.val >= j.(int) {
		return l.val
	}
	return j.(int)
}

// return current timestamp in the system
func (l *LamportClock) Get() interface{} { return l.val }

// increase the value of clock
func (l *LamportClock) Increment() { l.val++ }

// update the clock based on some recieved event
func (l *LamportClock) Merge(cl Clock) {
	l.val = l.max(cl.Get()) + 1
}

// checks the clock values to see which event happened first
func (l *LamportClock) HappensBefore(cl Clock) bool {
	if l.val <= cl.Get().(int) {
		return true
	}
	return false
}

// VectorClocks are just a sequence of integers that represent the clock
// value of all other nodes in the systems. This is an upgrade from lamport
// clocks because nodes keep not only their clock but that of every other node
// in the system. Vector Clocks are better than lamport clocks in that they not
// only characterize causality but are also consistent with causality.
// This relation can be stated as A -> B <=> VC(A) <= VC(B).
// It implies that if A happens before B then the vector clock of A is less than
// or equal to the vector clock of B but also if the vector clock of A is less than
// or equal to that B we know that A happens before B
type VectorClock struct {
	val map[string]int
	id  string // keep track of nodes entry in the clock.
}

// in vector clocks, every node must have an entry in their clock
// that corresponds to the last known clock value of every other node
// in the system.
//
// so the question is how do you join the cluster. what if there are multiple
// clusters and they are merged together.
// If you join a cluster you should announce your presence so your clock gets merged
// in all the others clocks
//
// this is going to be harder than i thought it.
//
// First things first. when you join the cluster. you get the last know clock value
// of every single node in the cluster.
//
// On sends, you increment your own value in the list and sends it out with the message.
//
// On recieve, the reciever node compares its clock with that of the sender nodes
// finding the max of the two, it proceeds to increment the clock.
// recieves are peculiar because, how do you know which of the clocks are bigger when
// comparing?. you use the pointwise maximum i.e comparing every single index point with
// the corresponding one in the other clock.
// so for every index point 'i', VC(A) < VC(B) if VC(A)i <= VC(B) and VC(A) != VC(B)
//
// On internal events, a node increments just its clock value. and continues living life.

func NewVectorClock() Clock {
	return &VectorClock{val: make(map[string]int)}
}

// keeps track of clocks id in vector.
func (vc *VectorClock) registerId(id string) {
	vc.val[id] = 0
	vc.id = id
}

func (vc *VectorClock) addMember(id string) {
	vc.val[id] = 0
}

func (vc *VectorClock) String() string {
	return strings.Split(fmt.Sprintf("%v", vc.val), "p")[1]
}

func (vc *VectorClock) Get() interface{} {
	return vc.val
}

func (vc *VectorClock) Increment() {
	vc.val[vc.id]++
}

// Merge finds the max of two vectors and increments
// the owner of the clocks position in the vector.
func (vc *VectorClock) Merge(cl Clock) {

	// get a set of keys of the two clocks
	keys := make(map[string]struct{})
	for key, _ := range vc.val {
		keys[key] = struct{}{}
	}

	m := cl.Get().(map[string]int)

	for key, _ := range m {
		_, ok := keys[key]
		if !ok {
			keys[key] = struct{}{}
		}
	}

	for id, _ := range keys {
		if vc.val[id] < m[id] {
			vc.val[id] = m[id]
		}
	}
	vc.Increment()
}

// checks to see if the clocks timestamp happens before its own
// using the following rule
// VC(A) < VC(B) if VC(A)i <= VC(B)i for all i and VC(A) != VC(B)
func (vc *VectorClock) HappensBefore(cl Clock) bool {
	if fmt.Sprint(vc) == fmt.Sprint(cl) {
		return false
	}
	return lessThan(vc.val, cl.Get().(map[string]int))
}

// compares two vector clocks and returns true if a < b
func lessThan(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if v > b[k] {
			return false
		}
	}
	return true
}

// The point of most of the stuff implemented here it to find
// causal relationships between events happenning on distributed nodes
// communicating over unreliable networks. And thats one of the big problems
// in distributed systems design finding causal relationships between events
// independently executing nodes.
// With vector clocks, so long as nodes keep on communicating and exchanging
// clocks and timestamps of events, we are guaranteed to know
// if an event A -> B or if A || B.
func (vc *VectorClock) getCausalRelation(other Clock) string {
	// three of the following values must be returned.
	// happens-before, concurrent, unknown
	m := other.Get().(map[string]int)
	if lessThan(vc.val, m) || lessThan(vc.val, m) {
		return "happens-before"
	}

	if !lessThan(vc.val, m) && !lessThan(m, vc.val) {
		return "concurrent"
	}
	return "unknown"
}
