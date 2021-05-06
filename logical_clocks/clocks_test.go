package clocks

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"
)

type Node struct {
	Clock
	id     string
	status bool          // false means system is down
	buf    *bytes.Buffer // underlying buffer
	c      chan int
	log    []*eventLog
}

// TODO: every node having a log of events will actually be a cool thing.
// So we can just go through the logs of each node and see if there's actual
// eventual consistency going on instead of just comparing the final clocks because
// the goal is to actually arrive at a solution that guarantee eventual consistency
// in the system.

// eventLog is a struct to keep info of generated events in the system and their times.
type eventLog struct {
	nodeId    string
	msg       string
	status    string // status is -> internal, send, recieve
	timestamp interface{}
}

var (
	errSystemDown    = errors.New("system is down")
	internalEventMsg = "internal server event"
)

func New(name string, cl Clock) *Node {
	return &Node{Clock: cl, id: name, status: true, c: make(chan int), buf: bytes.NewBuffer(nil)}
}

// simulates a system fault
func (no *Node) changeStatus(b bool) { no.status = b }

// increment the clock to simulate some kind of event generation and record in log.
func (no *Node) genInternalEvent() {
	no.Increment()
	no.addEventLog(addEventLog(internalEventMsg, "internal"))
}

func (no *Node) genEvent(msg, stat string) { no.addEventLog(addEventLog(msg, stat)) }

// add new event log to the nodes log of events.
func (no *Node) addEventLog(msg, status string) {
	no.log = append(no.log, &eventLog{
		nodeId:    no.id,
		msg:       msg,
		status:    status,
		timestamp: no.Get(),
	})
}

// sorts the logs of each nodes in a cluster
func sortLamportEventLog(logs ...[]*eventLog) *[]eventLog {

	for i, l := range no.log {

	}
}

func (no *Node) serializeLog() {

}

// read data from node
func (no *Node) Read(p []byte) (n int, err error) {
	n, err = no.buf.Read(p)
	if err != nil {
		return n, err
	}
	no.buf.Reset()
	return
}

// write data to node
func (no *Node) Write(p []byte) (n int, err error) {
	if !no.status {
		return 0, errSystemDown
	}
	n, err = no.buf.Write(p)
	if err != nil {
		return n, err
	}
	return
}

func (no *Node) send(msg string, r *Node) {
	go r.recv(no) // launch go routine to listen for connections

	// write the message to the remote connection
	if _, err := r.Write([]byte(msg)); errors.Is(err, errSystemDown) {
		time.Sleep(time.Millisecond * 70) // waiting for response
		return
	}

	// increment clock before sends.
	no.Increment()
	no.genEvent(msg, "send")

	r.c <- 1 // new message alert to remote host
	<-no.c   // wait for the remote host to finish reading new msg
}

func (no *Node) recv(r *Node) {
	fmt.Println()
	select {
	case <-no.c:
		// abort if system is down
		if !no.status {
			no.buf.Reset()
			return
		}
		time.Sleep(time.Millisecond * 70) // simulate network latency
		fmt.Println("recv debug: done sleeping")
		b := make([]byte, 100)
		_, err := no.Read(b)
		if err != nil {
			fmt.Errorf("recv error: (nodeId, %s) %v\n", no.id, err)
			no.buf.Reset()
			return
		}
		no.Merge(r.Clock)
		no.genEvent(string(b), "recv")
		fmt.Println("recv debug: done merging")
		fmt.Printf("recv succesful: (nodeId, %s) (msg, %s) (timestamp, %v)\n", no.id, string(b), no.Get())
		r.c <- 1
		return
	}
}

//---------------------------// Test for helpers and other methods.

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

// ----------------------------------------------------------------------------

// wheew this was really crazy wow.

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
	A.genEvent()
	ts1 := A.Get()
	A.genEvent()

	if A.HappensBefore(ts1) {
		t.Errorf("lamport clocks error: expected %q <= %q\n", ts1, A.Get())
	}
}

// case: if A -> B if A is a send operation and B is the corresponding
// recieve.
func TestLamportClockWithSendAndRecv(t *testing.T) {
	A, B := NodeWithLamport("A"), NodeWithLamport("B")
	A.genEvent()
	A.send(fmt.Sprintf("message from node %s, timestamp at send %d", A.id, A.Get().(int)), B)
	tsA := A.Get().(int)

	if B.HappensBefore(tsA) {
		t.Errorf("expected tsA(%d) <= tsB(%d)", tsA, B.Get().(int))
	}
}

// case: we can say A -> C if A -> B and B -> C, this is called the transitive closure
func TestLamportClockTransitiveClosure(t *testing.T) {
	A, B, C := NodeWithLamport("A"), NodeWithLamport("B"), NodeWithLamport("C")

}
