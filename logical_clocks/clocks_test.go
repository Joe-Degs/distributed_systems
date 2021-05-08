package clocks

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestLamportMax(t *testing.T) {
	l := LamportClock{val: 1}
	if l.max(2) != 2 {
		t.Errorf("expected %d, got %d", 2, l.val)
	}
}

func TestLamportMerge(t *testing.T) {
	A, B := &LamportClock{}, &LamportClock{}
	A.val, B.val = 4, 5
	A.Merge(B) // A's clock is now max(4, 5) + 1
	if !B.HappensBefore(A.val) {
		t.Errorf("expected tsB(%d) <= tsA(%d)\n", B.val, A.val)
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
func lamportNode(id string) *Node {
	return NewNode(id, &LamportClock{})
}

// case: A -> B if A and B are in the same process and A genuinely occurs
// before A
func TestLamportClockSameProcess(t *testing.T) {
	A := lamportNode("A")
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
	A, B := lamportNode("A"), lamportNode("B")
	A.genInternalEvent()
	A.send(fmt.Sprintf("message from node %s, timestamp at send %d", A.id, A.Get().(int)), B)
	tsA := A.Get().(int)

	if B.HappensBefore(tsA) {
		t.Errorf("expected tsA(%d) <= tsB(%d)", tsA, B.Get().(int))
	}
}

// case: we can say A -> C if A -> B and B -> C, this is called the transitive closure
func TestLamportClockTransitiveClosure(t *testing.T) {
	cl := NewCluster(NewLamportClock, "a", "b", "c")
	for key, value := range cl.nodes {
		t.Log(key, value)
	}
	cl.nodes["a"].genInternalEvent()
	cl.nodes["a"].send("send event from 'a'", cl.nodes["b"])
	cl.nodes["b"].genInternalEvent()
	cl.nodes["b"].send("send event from 'b'", cl.nodes["c"])
	cl.appendSortLogs() // consolidates and sorts logs

	//t.Log(cl.dlog)

	for i, log := range cl.dlog {
		if i == len(cl.dlog)-1 {
			return
		}
		if !(log.timestamp.(int) <= cl.dlog[i+1].timestamp.(int)) {
			t.Errorf("lamport clock's transitivity closure error")
		}
		if log.status == "send" {
			if cl.dlog[i+1].status != "recv" {
				t.Errorf("expected a corresponding recieve after every send")
			}
		}
	}

}

// Lamport clock is a very cool concept, but it has been proven not
// to work in real world systems.
func TestProveLamportWrong(t *testing.T) {
	var wg sync.WaitGroup
	cl := NewCluster(NewLamportClock, "a", "b", "c", "d", "e", "f", "g", "h", "i", "j")

	keys := make([]string, 0, len(cl.nodes))
	for key, _ := range cl.nodes {
		keys = append(keys, key)
	}

	randomInt := func(min, max int) int {
		rand.Seed(time.Now().UnixNano())
		return rand.Intn(max-min) + min
	}

	randomKey := func() string {
		return keys[randomInt(0, len(keys))]
	}

	// random time.Sleep sends
	// this will run in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			k1 := randomKey()
			var k2 string
			for {
				k2 = randomKey()
				if k1 == k2 {
					k2 = randomKey()
				} else {
					break
				}
			}
			time.Sleep(time.Millisecond * time.Duration(randomInt(10, 50)))
			cl.nodes[k1].send(fmt.Sprintf("send event from %s to %s",
				cl.nodes[k1].id, cl.nodes[k2].id), cl.nodes[k2])
		}
		return
	}()

	// random time.Sleep internal events.
	// this will also run in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			time.Sleep(time.Millisecond * time.Duration(randomInt(10, 50)))
			cl.nodes[randomKey()].genInternalEvent()
		}
		return
	}()
	// wait for goroutines to finish
	wg.Wait()

	cl.appendLogs()
	for _, log := range cl.dlog {
		t.Log(log.msg)
		fmt.Println("")
	}
}
