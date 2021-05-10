package clocks

import "fmt"

// Cluster represents groups of nodes communicating
type Cluster struct {
	nodes map[string]*Node
	dlog  []eventLog
}

func NewCluster(clock func() Clock, ids ...string) *Cluster {
	cl := &Cluster{nodes: make(map[string]*Node)}
	for _, id := range ids {
		cl.nodes[id] = NewNode(id, clock())
	}
	if _, ok := clock().(*VectorClock); ok {
		for _, node := range cl.nodes {
			node.Clock.(*VectorClock).registerId(node.id)
			cl.announcePresence(node.id)
		}
	}
	return cl
}

func (cl *Cluster) announcePresence(id string) {
	// if a node joins a cluster, it has to annouce its presence
	// especially in the case of vector clocks
	// for every node in the system, you get their id and give them your
	// id and you include them in your clocks.

	if len(cl.nodes) == 1 {
		return
	}
	for nid, node := range cl.nodes {
		if nid == id {
			continue
		}
		node.Clock.(*VectorClock).addMember(id)
	}
}

func (cl *Cluster) Get(id string) *Node {
	if node, ok := cl.nodes[id]; ok {
		return node
	}
	return nil
}

// consolidates all logs in the cluster and sort them based on the
// happens before lamport relationship
func (cl *Cluster) appendLogs() {
	for _, value := range cl.nodes {
		cl.dlog = append(cl.dlog, value.log...)
	}
	//sortLamportLog(cl.dlog)
}

func (cl *Cluster) appendSortLogs() {
	cl.appendLogs()
	sortLamportLog(cl.dlog)
}

func (cl *Cluster) Send(from, to, msg string) error {
	sender, recipient := cl.Get(from), cl.Get(to)
	if sender == nil || recipient == nil {
		return fmt.Errorf("node ids not in cluster: %s, %s", from, to)
	}
	sender.send(msg, recipient)
	return nil
}
