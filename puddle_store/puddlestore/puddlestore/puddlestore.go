package puddlestore

import (
	"../../../raft/raft"
	"../../../tapestry/tapestry"
	//"bufio"
	//"fmt"
	//"os"
	//"strings"
)

const TAPESTRY_NODES = 20
const RAFT_NODES = 3

type Puddlestore struct {
	tnodes []*tapestry.Tapestry
	rnodes []*raft.RaftNode
	rootV  uint32
	paths  map[string]uint32
}

func Start() {
	puddlestore := new(Puddlestore)

	// Start runnning the tapestry nodes. --------------
	t, err := tapestry.Start(0, "")
	if err != nil {

	}
	puddlestore.tnodes[0] = t
	for i := 1; i < TAPESTRY_NODES; i++ {
		t, err = tapestry.Start(0, puddlestore.tnodes[0].GetLocalAddr())
		if err != nil {

		}
		puddlestore.tnodes[i] = t
	}
	// -------------------------------------------------

	// Run the Raft cluster ----------------------------
	raft.CreateLocalCluster(raft.DefaultConfig())
	// -------------------------------------------------

	// Create the root node ----------------------------
	// -------------------------------------------------
}
