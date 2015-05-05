package puddlestore

import (
	"../../raft/raft"
	"../../tapestry/tapestry"
	"math/rand"
	"net"
	// "net/rpc"
	//"bufio"
	"fmt"
	//"os"
	//"strings"
)

const TAPESTRY_NODES = 20
const RAFT_NODES = 3

type vguid string
type aguid string
type guid string

type PuddleNode struct {
	tnodes []*tapestry.Tapestry
	rnodes []*raft.RaftNode
	rootV  uint32
	paths  map[string]string

	server     *PuddleRPCServer
	Id         string
	Addr       string
	Listener   net.Listener
	listenPort int
	RPCServer  *PuddleRPCServer
	IsShutdown bool
}

type PuddleAddr struct {
	Addr string
}

func CreatePuddleStore() {
}

func Start() (p *PuddleNode, err error) {
	var puddle PuddleNode
	p = &puddle
	puddle.tnodes = make([]*tapestry.Tapestry, TAPESTRY_NODES)
	puddle.rnodes = make([]*raft.RaftNode, RAFT_NODES)

	// Start runnning the tapestry nodes. --------------
	t, err := tapestry.Start(0, "")
	if err != nil {
		panic(err)
	}

	puddle.tnodes[0] = t
	for i := 1; i < TAPESTRY_NODES; i++ {
		t, err = tapestry.Start(0, puddle.tnodes[0].GetLocalAddr())
		if err != nil {
			panic(err)
		}
		puddle.tnodes[i] = t
	}
	// -------------------------------------------------

	// Run the Raft cluster ----------------------------
	puddle.rnodes, err = raft.CreateLocalCluster(raft.DefaultConfig())
	if err != nil {
		panic(err)
	}
	// -------------------------------------------------

	// Create the root node ----------------------------
	// vguid := randSeq(5)
	root := CreateRootInode()
	// puddlestore.paths["/"] = vguid
	encodedRoot, err := root.GobEncode()
	if err != nil {
		panic(err)
	}
	err = tapestry.TapestryStore(puddle.tnodes[0].GetLocalNode(),
		"/", encodedRoot)
	if err != nil {
		panic(err)
	}
	// -------------------------------------------------

	// RPC server --------------------------------------
	puddle.server = newPuddlestoreRPCServer(p)

	fmt.Printf("Started puddlestore, listening at %v\n", puddle.server.listener.Addr().String())
	// -------------------------------------------------

	return
}

func (puddle *PuddleNode) getRandomTapestryNode() tapestry.Node {
	index := rand.Int() % TAPESTRY_NODES
	return puddle.tnodes[index].GetLocalNode()
}

func randSeq(n int) string {
	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (p *PuddleNode) run() {
	for {
	}
}
