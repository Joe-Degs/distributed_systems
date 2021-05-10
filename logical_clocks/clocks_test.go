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
	if !B.HappensBefore(A) {
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

	if A.Get().(int) <= ts1.(int) {
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

	if B.Get().(int) <= tsA {
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
// to work in real world systems because even though lamport clocks are
// consistent with causality, they do not characterize causality.
// In view of this if A -> B then the LC(A) <= LC(B) but we cannot say
// for sure that if LC(A) <= LC(B) then A -> B
func TestRandomBehaviourInClusterWithLamportClock(t *testing.T) {
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

func TestVectorClockIncrement(t *testing.T) {
	v := &VectorClock{val: make(map[string]int)}
	v.id = "a"
	v.Increment()
	t.Log(v)
}

func TestVectorClockMerge(t *testing.T) {
	cl := NewVectorClock()
	cl.(*VectorClock).registerId("e")
	vc := &VectorClock{val: make(map[string]int)}
	vc.val["a"] = 2
	vc.val["b"] = 3
	vc.val["d"] = 1
	cl.Merge(vc)
	t.Log(cl)
}

func TestVectorClockHappensBefore(t *testing.T) {
	// a = [2,2,0] and b = [3,2,0]
	a, b := &VectorClock{val: make(map[string]int)}, &VectorClock{val: make(map[string]int)}
	a.registerId("a")
	b.registerId("b")

	a.val["a"] = 0
	a.val["b"] = 1
	a.val["c"] = 1

	b.val["a"] = 0
	b.val["b"] = 2
	b.val["c"] = 1

	if !a.HappensBefore(b) {
		t.Errorf("expected a to happen before b")
	}
}

func TestVectorClockCluster(t *testing.T) {
	cluster := NewCluster(NewVectorClock, "alice", "bob", "carol")

	for _, node := range cluster.nodes {
		for id, _ := range node.Clock.Get().(map[string]int) {
			if id != "alice" && id != "bob" && id != "carol" {
				t.Errorf("expected all of this ids to be present: alice, bob and carol")
			}
		}
	}
}

func TestVectorClockCausalRelationships(t *testing.T) {
	vA, vB := &VectorClock{val: make(map[string]int)}, &VectorClock{val: make(map[string]int)}

	vA.val["a"] = 2
	vA.val["b"] = 2
	vA.val["c"] = 0

	vB.val["a"] = 1
	vB.val["b"] = 2
	vB.val["c"] = 3

	if vA.getCausalRelation(vB) != "concurrent" {
		t.Errorf("event should be concurrent")
	}
}
