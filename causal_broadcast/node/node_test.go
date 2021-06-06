package node

import (
	"reflect"
	"testing"
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
	e := &EventLog{
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

func TestEventLogJson(t *testing.T) {
	n := New("joe")
	n.Clock.AddMember("kelvin", 2)
	n.Clock.AddMember("messi", 4)
	event := &EventLog{
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
		logs []*EventLog
	}{
		{
			name: "queue of length 2",
			q:    &EventQueue{q: make([]*EventLog, 2)},
			logs: []*EventLog{&EventLog{Id: "joe"}, &EventLog{Id: "jude"}},
		},
		{
			name: "test reading till end",
			q:    &EventQueue{q: make([]*EventLog, 4)},
			logs: []*EventLog{&EventLog{Id: "joe"}, &EventLog{Id: "jude"}},
		},
	}

	for tn, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			for i, log := range tc.logs {
				tc.q.Append(log)
				if i == len(tc.q.q)-1 && tc.q.stop != true {
					t.Error("Queue must stop accepting elements if its full and complete reads")
				}
			}
			if tn == 0 {
				r := tc.q.Append(tc.logs[0])
				if !r {
					t.Error("Can't add to a full queue, read first")
				}
			}
		})
	}
}
