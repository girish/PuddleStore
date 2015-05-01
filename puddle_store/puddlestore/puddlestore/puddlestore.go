package puddlestore

import (
	"../../../raft/raft"
	"../../../tapestry/tapestry"
	"math/rand"
	"net"
	"net/rpc"
	//"bufio"
	// "fmt"
	//"os"
	//"strings"
)

const TAPESTRY_NODES = 20
const RAFT_NODES = 3

type vguid string
type aguid string
type guid string

type Puddlestore struct {
	tnodes []*tapestry.Tapestry
	rnodes []*raft.RaftNode
	rootV  uint32
	paths  map[string]string

	Listener   net.Listener
	listenPort int
	RPCServer  *PuddleRPCServer
	IsShutdown bool
}

func CreatePuddleStore() {
}

func Start() (p *Puddlestore) {
	var puddlestore Puddlestore
	p = &puddlestore

	// Start runnning the tapestry nodes. --------------
	t, err := tapestry.Start(0, "")
	if err != nil {
		panic(err)
	}
	puddlestore.tnodes[0] = t
	for i := 1; i < TAPESTRY_NODES; i++ {
		t, err = tapestry.Start(0, puddlestore.tnodes[0].GetLocalAddr())
		if err != nil {
			panic(err)
		}
		puddlestore.tnodes[i] = t
	}
	// -------------------------------------------------

	// Run the Raft cluster ----------------------------
	raft.CreateLocalCluster(raft.DefaultConfig())
	// -------------------------------------------------

	// Create the root node ----------------------------
	vguid := randSeq(5)
	root := CreateRootInode()
	puddlestore.paths["/"] = vguid
	encodedRoot, err := root.GobEncode()
	if err != nil {
		panic(err)
	}
	tapestry.TapestryStore(puddlestore.tnodes[0].GetLocalNode(),
		vguid, encodedRoot)
	// -------------------------------------------------

	conn, localPort, err := OpenListener()
	if err != nil {
		panic(err)
	}
	puddlestore.Listener = conn
	puddlestore.listenPort = localPort

	// Start RPC server
	puddlestore.RPCServer = &PuddleRPCServer{p}
	rpc.RegisterName(conn.Addr().String(), puddlestore.RPCServer)
	go puddlestore.RPCServer.startRpcServer()

	return
}

func randSeq(n int) string {
	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
