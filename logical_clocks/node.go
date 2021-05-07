package clocks

import (
	"bytes"
	"errors"
	"fmt"
	"time"
)

type Node struct {
	Clock
	id     string
	status bool          // false means system is down
	buf    *bytes.Buffer // underlying buffer
	c      chan int
	log    []eventLog
}

func New(id string, cl Clock) *Node {
	return &Node{
		Clock:  cl,
		id:     id,
		status: true,
		c:      make(chan int),
		buf:    bytes.NewBuffer(nil),
	}
}

type Cluster struct {
	nodes []*Node
}

func NewCluster(clock func() Clock, id ...string) *Cluster {
	cl := &Cluster{}
	for _, id := range ids {
		cl.nodes = append(cl.nodes, NewNode(id, clock()))
	}
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

// simulates a system fault
func (no *Node) changeStatus(b bool) { no.status = b }

// increment the clock to simulate some kind of event generation and record in log.
func (no *Node) genInternalEvent() {
	no.Increment()
	no.addEventLog(internalEventMsg, "internal")
}

func (no *Node) genEvent(msg, stat string) { no.addEventLog(msg, stat) }

// add new event log to the nodes log of events.
func (no *Node) addEventLog(msg, status string) {
	no.log = append(no.log, eventLog{
		nodeId:    no.id,
		msg:       msg,
		status:    status,
		timestamp: no.Get(),
	})
}

// sorts the logs of each nodes in a cluster
func appendEventLogs(logs ...[]eventLog) []eventLog {
	dlog := make([]eventLog, 0, len(logs))
	for _, log := range logs {
		dlog = append(dlog, log...)
	}
	return dlog
}

// bubble sort to bubble things up
func sortLamportLog(dlog []eventLog) {
	for i := 0; i < len(dlog); i++ {
		swapped := false
		for j := 0; j < len(dlog)-i-1; j++ {
			if dlog[j].timestamp.(int) > dlog[j+1].timestamp.(int) {
				dlog[j], dlog[j+1] = dlog[j+1], dlog[j]
				swapped = true
			}
		}
		if !swapped {
			break
		}
	}
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
	no.genEvent(fmt.Sprintf("(nodeId, %s) (msg, %s)", no.id, msg), "send")

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
		b := make([]byte, 100)
		_, err := no.Read(b)
		if err != nil {
			fmt.Errorf("recv error: (nodeId, %s) %v\n", no.id, err)
			no.buf.Reset()
			return
		}
		no.Merge(r.Clock)
		no.genEvent(fmt.Sprintf("(nodeId, %s) (msg, %s)", no.id, string(b)), "recv")
		r.c <- 1
		return
	}
}
