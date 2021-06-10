package node

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestNewNode(t *testing.T) {
	n := New("joe")

	if n.Id != "joe" && n.Clock.Get() != 0 {
		t.Error("new node does not work!")
	}
}

func TestTimestampConversions(t *testing.T) {
	n := New("jude")
	n.Clock.AddMember("joe", 0)
	e := &Event{
		Id:        "joe",
		Timestamp: n.Clock.String(),
		Msg:       "message",
	}

	// print timestamp
	t.Log(e.Timestamp)
	cl := e.Clock()
	if e.Timestamp != cl.String() {
		t.Errorf("expected %s, got %s", e.Timestamp, cl.String())
	}
	t.Log(cl.String())
}

func TestEventJson(t *testing.T) {
	n := New("joe")
	n.Clock.AddMember("kelvin", 2)
	n.Clock.AddMember("messi", 4)
	event := &Event{
		Id:        n.Id,
		Timestamp: n.Clock.String(),
		Msg:       "cryptic message",
	}

	// marshal eventlog struct to json.
	json, err := event.Json()
	if err != nil {
		t.Error(err)
	}
	t.Log(string(json))

	// unnarshal json back to eventlog
	uevent, err := Unmarshal(json)
	if err != nil {
		t.Error(err)
	}

	// check if the two structs are equal.
	elements := []string{"Id", "Timestamp", "Msg"}
	a := reflect.Indirect(reflect.ValueOf(event))
	b := reflect.Indirect(reflect.ValueOf(uevent))
	for _, el := range elements {
		c := a.FieldByName(el)
		d := b.FieldByName(el)
		if c.String() != d.String() {
			t.Errorf("expected %s and %s to be equal", c.String(), d.String())
		}
	}
}

func TestEventQueue(t *testing.T) {
	tt := []struct {
		name string
		q    *EventQueue
		logs []*Event
	}{
		{
			name: "append queue on full should fail",
			q:    &EventQueue{q: make([]*Event, 2)},
			logs: []*Event{&Event{Id: "joe"}, &Event{Id: "jude"}},
		},
		{
			name: "using vacant spaces at the beginning of queue",
			q:    &EventQueue{q: make([]*Event, 4)},
			logs: []*Event{&Event{Id: "joe"}, &Event{Id: "jog"}, &Event{Id: "jay"}},
		},
	}

	for tn, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// append logs to queue
			for i, log := range tc.logs {
				tc.q.Append(log)
				if i == len(tc.q.q)-1 && tc.q.stop != true {
					t.Error("Queue must stop accepting elements if its full and complete reads")
				}
			}

			switch tn {
			case 0:
				r := tc.q.Append(tc.logs[0])
				if r {
					t.Error("Queue is full and closed, only reads are allowed.")
				}
			case 1:
				// this queue has length=4, first 3 items are set, idx=3 next=0
				t.Log(tc.q)
				r := tc.q.Append(&Event{Id: "jey"}) // idx=3, append and close
				if !tc.q.stop && !r {
					t.Error("queue is full and should be closed")
				}

				_, r = tc.q.Next() // next=1, open
				t.Log(tc.q)
				if tc.q.stop && !r {
					t.Error("queue should be opened now. Vacancy at the begining")
				}

				t.Log("------------NEXT IN THE LEAD----------------------------")
				r = tc.q.Append(&Event{Id: "jie"}) // idx=1, close, succesful
				t.Log("append", tc.q)
				if !tc.q.stop && !r {
					t.Error("Vacancy at the begining, append should be succesful")
				}

				r = tc.q.Append(nil) // idx=1, append rejected
				t.Log("append", tc.q)
				if tc.q.stop && r {
					t.Error("idx == next, this shd have aborted a long time ago")
				}

				// read the next item in queue
				_, r = tc.q.Next() // next=2, open
				t.Log("next", tc.q)
				if tc.q.stop && r {
					t.Error("next in the lead reads shd work!")
				}

				// append should work after a next
				r = tc.q.Append(nil) // idx=2, still open
				t.Log("append", tc.q)
				if !tc.q.stop && !r {
					t.Error("append after a read, shd work like a charm!")
				}

				// read next item in queue
				_, r = tc.q.Next() // next=3, close
				t.Log("next", tc.q)
				if tc.q.stop && !r {
					t.Error("next in the lead, reads shd work!")
				}

				// appends should work after a read
				r = tc.q.Append(nil) // idx=3
				t.Log("append", tc.q)
				if !tc.q.stop && !r {
					t.Error("appending after a read, shd work!")
				}

				r = tc.q.Append(nil) // this shd not work
				t.Log("append", tc.q)
				if !tc.q.stop && !r {
					t.Error("appending at end, idx == next")
				}

				_, r = tc.q.Next()
				t.Log("next", tc.q)
				if tc.q.stop && !r {
					t.Error("next shd work and queue shd be open for reading")
				}

				tc.q.Append(nil)
				t.Log("append", tc.q)

				t.Log("---------------------------------")
				tc.q.Next()
				t.Log("next", tc.q)

				tc.q.Append(nil)
				t.Log("append", tc.q)
			}
		})
	}
}

func TestNodeIO(t *testing.T) {
	//str := "[a:1 b:3]"
	//t.Log(str[1 : len(str)-1])
	//

	node := New("joe")
	node.Clock.AddMember("con", 2)
	node.Clock.AddMember("jon", 3)
	p, err := node.GenEvent("eat this for breakfast")
	if err != nil {
		t.Error("event generation error", err)
	}

	anodaNode := New("jude")
	if _, err := anodaNode.Write(p); err != nil {
		t.Error("Write to Node not succesful, error occured", err)
	}
	spew.Dump(anodaNode.History)
}
