package clocks

import (
	"fmt"
	"testing"
)

func TestLamportMax(t *testing.T) {
	n := lamportMax(1, 2)
	if n != 2 {
		t.Errorf("expected %d, got %d", 2, n)
	}
}

func TestLamportMerge(t *testing.T) {
	A, B := &LamportClock{}, &LamportClock{}
	A.N, B.N = 4, 5
	A.Merge(B) // A's clock is now max(4, 5) + 1
	if !B.HappensBefore(A.N) {
		t.Errorf("expected tsB(%d) <= tsA(%d)\n", B.N, A.N)
	}
}

func TestEventLogs(t *testing.T) {
	log1 := []eventLog{
		eventLog{timestamp: 5},
		eventLog{timestamp: 2},
		eventLog{timestamp: 3},
		eventLog{timestamp: 0},
	}
	log2 := []eventLog{
		eventLog{timestamp: 7},
		eventLog{timestamp: 1},
	}
	log := []eventLog{
		eventLog{timestamp: 4},
	}

	dlog := appendEventLogs(log1, log2, log)
	sortLamportLog(dlog)

	for i, l := range dlog {
		if i == len(dlog)-1 {
			break
		}
		a, b := l.timestamp.(int), dlog[i+1].timestamp.(int)
		if a > b {
			fmt.Errorf("expected %d < %d, got %d > %d", a, b, a, b)
		}
	}
	t.Log(dlog)
}

// list of things nodes can do
// 1. simulate event generation so that nodes increase counters
// 2. send and recieve messages to other nodes in the system
// 3. system can crash and not respond to reads or writes

// Now lets start writing some tests.
// for Lamport Clocks
func NodeWithLamport(id string) *Node {
	return New(id, &LamportClock{})
}

// case: A -> B if A and B are in the same process and A genuinely occurs
// before A
func TestLamportClockSameProcess(t *testing.T) {
	A := NodeWithLamport("A")
	A.genInternalEvent()
	ts1 := A.Get()
	A.genInternalEvent()

	if A.HappensBefore(ts1) {
		t.Errorf("lamport clocks error: expected %q <= %q\n", ts1, A.Get())
	}
}

// case: if A -> B if A is a send operation and B is the corresponding
// recieve.
func TestLamportClockWithSendAndRecv(t *testing.T) {
	A, B := NodeWithLamport("A"), NodeWithLamport("B")
	A.genInternalEvent()
	A.send(fmt.Sprintf("message from node %s, timestamp at send %d", A.id, A.Get().(int)), B)
	tsA := A.Get().(int)

	if B.HappensBefore(tsA) {
		t.Errorf("expected tsA(%d) <= tsB(%d)", tsA, B.Get().(int))
	}
}

// case: we can say A -> C if A -> B and B -> C, this is called the transitive closure
//func TestLamportClockTransitiveClosure(t *testing.T) {
//	A, B, C := NodeWithLamport("A"), NodeWithLamport("B"), NodeWithLamport("C")
//}
