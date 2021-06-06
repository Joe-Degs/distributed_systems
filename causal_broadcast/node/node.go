// Package node contains implementation of nodes for a causal broadcast
// cluster.
package node

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Joe-Degs/distributed_systems/causal_broadcast/clock"
)

// Node represents a single actor in the system.
type Node struct {
	Clock *clock.Vector
	Id    string
	//status  bool
	//buf     *bytes.Buffer
	//c       chan int
	History []string // log of all events it has delivered.

	// events nodes has recieved but has not delivered due to causal
	// anomalies.
	Queue *EventQueue
}

/*
* EventLogs are what are to be exchanged between nodes in a cluster.
* All nodes will keep a log of all the events they have seen so far.
* Event logs can be marshalled to bytes, basically json format before
* being sent over the virtual wire.
*
* EventLog will have a method to retrieve the clock of the timestamp.
* */

// EventLog is log of a single recieved event.
type EventLog struct {
	Id        string `json:"id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Msg       string `json:"msg,omitempty"`
}

// Clock returns the clock of the eventlog at the time of event generation.
func (e *EventLog) Clock() *clock.Vector {
	c := clock.New(e.Id)
	for _, item := range strings.Split(strings.Trim(e.Timestamp, "[]"), " ") {
		split := strings.Split(item, ":")
		val, err := strconv.Atoi(split[1])
		if err != nil {
			panic(err)
		}
		c.AddMember(split[0], val)
	}
	return c
}

// Json returns eventlog in json format for sending over the wire.
func (e *EventLog) Json() ([]byte, error) {
	return json.Marshal(e)
}

// Unmarshal turns the json of eventlog back to EventLog type.
func Unmarshal(b []byte) (*EventLog, error) {
	e := new(EventLog)
	err := json.Unmarshal(b, e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// EventQueue contains all the events that can't be
// delivered to node due to causal anomalies.
type EventQueue struct {
	next, idx int
	q         []*EventLog
	buf       *EventLog
}

// Append adds element to back of queue
func (eq *EventQueue) Append(e *EventLog) bool {
	if eq.q == nil {
		eq.q = make([]*EventLog, 10)
	}

	// if at the end of the queue
	// check if theres nobody there and accept somebody.
	// if theres somebody check and its different from the person already there
	qlen := len(eq.q) - 1
	if eq.idx == qlen && eq.next != qlen {
		if eq.q[eq.idx] == nil {
			eq.q[eq.idx] = e
			return false
		} else if eq.q[eq.idx] != e {
			return false
		}
	}
	eq.q[eq.idx] = e
	eq.idx = (eq.idx + 1) % len(eq.q)
	return true
}

func (eq *EventQueue) reset() {
	eq.idx = (eq.idx + 1) % len(eq.q)
}

func (eq *EventQueue) Next() (*EventLog, bool) {
	eq.buf = eq.q[eq.next]

	// reset queue if all items have been read.
	qlen := len(eq.q) - 1
	if eq.next == qlen && eq.idx == qlen {
		eq.reset()
	}
	if eq.buf == nil {
		return nil, false
	}

	eq.next = (eq.next + 1) % len(eq.q)
	return eq.buf, true
}

func (eq *EventQueue) Backup() *EventLog {
	return eq.buf
}

func New(id string) *Node {
	return &Node{
		Id:      id,
		Clock:   clock.New(id),
		History: make([]string, 0, 5),
		Queue:   &EventQueue{},
	}
}

/*
* Finding an intuitive way to hook up writes an to corresponding reads on nodes
* is the big problem now. I want it to be different and better than the
* implementation for logical_clocks.
* */

func (n *Node) Read([]byte) (int, error) {
	return 0, nil
}

func (n *Node) Write([]byte) (int, error) {
	return 0, nil
}
