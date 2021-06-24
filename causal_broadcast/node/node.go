// Package node contains implementation of nodes for a causal broadcast
// cluster.
package node

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Joe-Degs/distributed_systems/causal_broadcast/clock"
)

// Node represents a single actor in the system.
type Node struct {
	Clock *clock.Vector
	Id    string
	//status  bool
	buf *bytes.Buffer
	//c       chan int

	// holds events that have been delivered on client.
	// seen contains the id of nodes its aware of.
	History, seen []string // log of all events it has delivered.

	// events nodes has recieved but has not delivered due to causal
	// anomalies.
	Queue *EventQueue
}

/*
* Events are what are to be exchanged between nodes in a cluster.
* All nodes will keep a log of all the events they have seen so far.
* Event logs can be marshalled to bytes, basically json format before
* being sent over the virtual wire.
*
* Event will have a method to retrieve the clock of the timestamp.
* */

// Event is log of a single recieved event.
type Event struct {
	Id        string `json:"id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Msg       string `json:"msg,omitempty"`
}

// Clock returns the clock of the eventlog at the time of event generation.
func (e *Event) Clock() *clock.Vector {
	c := clock.New(e.Id)
	for _, item := range strings.Split(e.Timestamp[1:len(e.Timestamp)-1], " ") {
		split := strings.Split(item, ":")
		val, err := strconv.Atoi(split[1])
		if err != nil {
			panic(err)
		}
		c.AddMember(split[0], val)
	}
	return c
}

// Marshal returns the json of an event.
func (e *Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// Unmarshal turns the json of eventlog back to Event type.
func Unmarshal(b []byte) (*Event, error) {
	e := new(Event)
	err := json.Unmarshal(b, e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// EventQueue contains recent events that can't be
// delivered to node due to causal anomalies.
type EventQueue struct {
	next, idx int
	stop      bool
	q         []*Event
	buf       *Event
	mu        *sync.Mutex
}

func NewQueue(length int) *EventQueue {
	return &EventQueue{
		q:  make([]*Event, length),
		mu: &sync.Mutex{},
	}
}

// Append puts an item at the next vacant slot of the queue.
func (eq *EventQueue) Append(e *Event) bool {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	if eq.stop {
		// if queue is closed but idx and next are not
		// pointing to the same element, open queue
		if eq.next > 0 && eq.idx != eq.next {
			eq.stop = false
		} else {
			return false
		}
	}

	// queue is open
	if eq.idx > 0 {
		if eq.idx == eq.next {
			// if where we are reading the next item is same
			// as where we put the next item, close queue and return.
			eq.stop = true
			return false
		} else if eq.idx == len(eq.q)-1 {
			// when appending to the end of the queue,
			// append and check if there's vacancy at the
			// start of the queue.
			eq.q[eq.idx] = e
			if eq.next > 0 {
				// continue appending if there's vacancy.
				eq.reset()
			} else {
				// stop if there's no vacancy at beginning of queue.
				eq.stop = true
			}
			return true
		}
	}

	eq.q[eq.idx] = e
	eq.idx = (eq.idx + 1) % len(eq.q)

	// if after appending eq.idx becomes equal to eq.next
	// stop halt the next append.
	if eq.idx > 0 && eq.idx == eq.next {
		eq.stop = true
	}

	return true
}

func (eq *EventQueue) reset() {
	eq.stop = false
	eq.idx = (eq.idx + 1) % len(eq.q)

	// funny story:
	//	i was trying to lock this function and defer the unlock
	//	after execution, fucking naive right?. The methods calling this
	//	method are already locking the struct before accessing it so
	//	locking it only creates a deadlock or livelock?. One of them.
	//	pretty sure its a deadlock.
}

// Next returns the next item in the queue.
func (eq *EventQueue) Next() (*Event, bool) {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	move := func() {
		eq.buf = eq.q[eq.next]
		eq.next = (eq.next + 1) % len(eq.q)
	}

	if eq.stop {
		if eq.idx == eq.next {
			move()
			eq.stop = false
			return eq.buf, true
		} else if eq.idx == len(eq.q)-1 {
			// queue insertion at the end, vacancy at the start
			eq.reset()
		} else if eq.next != eq.idx {
			// if idx and next are not the same open queue for appending.
			eq.reset()
		}
	}

	move()
	return eq.buf, true
}

func (eq *EventQueue) Retry() *Event {
	return eq.buf
}

func New(id string) *Node {
	return &Node{
		Id:      id,
		Clock:   clock.New(id),
		History: make([]string, 0, 5),
		Queue:   NewQueue(10),
		buf:     new(bytes.Buffer),
	}
}

/*
* Finding an intuitive way to hook up writes an to corresponding reads on nodes
* is the big problem now. I want it to be different and better than the
* implementation for logical_clocks.
* */

var ErrEventQueued = errors.New("new event queued for later delivery")

// Read reads message from underlying buffer and adds it
// to the log of events it has seen
func (n *Node) ProcessEvent(event *Event) error {
	// if no event is passed in, read from node's buffer and
	// unmarshal it.
	var eventJson string
	var err error
	if event == nil {
		eventJson := n.buf.String()
		if eventJson == "" || eventJson == "<nil>" {
			return errors.New("message is probably empty")
		}
		event, err = Unmarshal([]byte(eventJson))
		if err != nil {
			return err
		}
	}

	eventClock := event.Clock()

	// an event has arrived. we have to know whether its safe to deliver
	// or we abort delivery and stick in a queue and try again sometime.
	if eventClock.IsCausallyConsistentWith(n.Clock) {
		// alls good deliver the message.
		n.Clock.Merge(eventClock)
		n.History = append(n.History, eventJson)
		return nil
	}

	stat := n.Queue.Append(event)
	if !stat {
		// if append fails try again until you can insert.
		// this feel extremely inefficient, i have to find
		// a better way to do this!.
		// TODO(joe)
		go func() {
			tts := 100
		LOOP:
			for {
				select {
				case <-time.After(time.Millisecond * time.Duration(tts)):
					stat = n.Queue.Append(event)
					if stat {
						break LOOP
					}
					tts = (tts % 1000) + 100
				}
			}
		}()
	}
	return ErrEventQueued
}

// GenEvent generates
func (n *Node) GenEvent(msg string) ([]byte, error) {
	// this is an event that will be sent out to other
	// nodes in the cluster.
	n.Clock.Increment()
	event := &Event{
		Id:        n.Id,
		Timestamp: n.Clock.String(),
		Msg:       msg,
	}

	p, err := event.Marshal()
	if err != nil {
		n.Clock.Decrement()
		return nil, err
	}
	return p, nil
}

// Write writes len(p) bytes to the underlying buffer of the node.
// and calls ProcessEvent to handle delivery on the node.
func (n *Node) Write(p []byte) (l int, err error) {
	// writes will write to the underlying buffer whatever is
	// in the byte slice that was supplied
	if l, err = n.buf.Write(p); err != nil || l != len(p) {
		return
	}
	if err := n.ProcessEvent(nil); err != nil {
		if errors.Is(err, ErrEventQueued) {
			// TODO(joe)
			// start the queue draining process
			return 0, nil
		} else {
			return l, err
		}
	}
	n.buf.Reset()
	return
}
