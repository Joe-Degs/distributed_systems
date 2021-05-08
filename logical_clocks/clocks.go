package clocks

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
	HappensBefore(interface{}) bool
}

// Couple of assumptions we can make from A -> B is that
// 1. A could have been the cause of B
// 2. The two are related in a way
// 3. A and B are in the same process and A happens before B
// 4. A is a send operation and B is the corresponding recieve
// 5. if A -> B and B -> C then A -> B(transitive closure) is likely to be true

// Lamport Clocks help in the ordering of events in our systems with the following
// sequences. The assumption is you have your systems wired to communicate.
// And all the nodes in the systems are keeping some kind of integer counter as
// the their clock
// -> If you generate an event in your system, increase your clock by 1
// -> If you send a message, add your clock to the message
// -> If you recieve a message, take max of the sender's clock value and your
//    clock value and increase by 1 and thats your clock value
// Lamport clocks is a very cool concept all that, but here's the catch, in real
// world systems if two independent systems that have broken connection and
// are not communicating but generating events internally finally get back
// communicating, comparing their clocks will show that some events happened
// before some and vice versa but that will be erroneous because those events
// are not causally connected per our definition of what makes events connected
// based on their lamport clocks.
// also my english sucks bad.
type LamportClock struct {
	val int
}

func NewLamportClock() Clock { return &LamportClock{} }

func (l *LamportClock) max(j interface{}) int {
	if l.val >= j.(int) {
		return l.val
	}
	return j.(int)
}

// return clock value or time
func (l *LamportClock) Get() interface{} { return l.val }

// increase the value of clock
func (l *LamportClock) Increment() { l.val += 1 }

// update the clock based on some recieved event
func (l *LamportClock) Merge(cl Clock) {
	l.val = l.max(cl.Get()) + 1
}

// checks the clock values to see which event happened first
func (l *LamportClock) HappensBefore(time interface{}) bool {
	if l.val <= time.(int) {
		return true
	}
	return false
}

// VectorClocks are a shift from the lamport clocks that do not characterize
//
type VectorClock struct {
	val map[string]int
}

//func NewVectorClock() Clock {
//	return &VectorClock{Val: make(map[string]int)}
//}
