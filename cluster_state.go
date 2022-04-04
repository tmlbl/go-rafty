package blobert

import (
	"fmt"
	"io"

	"github.com/hashicorp/raft"
)

type ClusterState struct {
}

func (cs *ClusterState) Apply(log *raft.Log) interface{} {
	fmt.Println("Got raft log", log)
	return nil
}

func (cs *ClusterState) Restore(snapshot io.ReadCloser) error {
	fmt.Println("Restoring snapshot", snapshot)
	return nil
}

func (cs *ClusterState) Snapshot() (raft.FSMSnapshot, error) {
	fmt.Println("Snapshot requested")
	return nil, nil
}

var _ raft.FSM = &ClusterState{}
