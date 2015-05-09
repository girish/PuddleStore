package puddlestore

import (
	"../../raft/raft"
	"../../tapestry/tapestry"
	"math/rand"
	// "net/rpc"
	//"bufio"
	"fmt"
	//"os"
	//"strings"
)

const TAPESTRY_NODES = 20
const RAFT_NODES = 3

type Vguid string
type Aguid string
type Guid string

type PuddleNode struct {
	tnodes      []*tapestry.Tapestry
	rnodes      []*raft.RaftNode
	rootV       uint32
	clientPaths map[uint64]string       // client id -> curpath
	clients     map[uint64]*raft.Client // client id -> client

	Local      PuddleAddr
	raftClient *raft.Client
	server     *PuddleRPCServer
}

type PuddleAddr struct {
	Addr string
}

func Start() (p *PuddleNode, err error) {
	var puddle PuddleNode
	p = &puddle
	puddle.tnodes = make([]*tapestry.Tapestry, TAPESTRY_NODES)
	puddle.rnodes = make([]*raft.RaftNode, RAFT_NODES)
	puddle.clientPaths = make(map[uint64]string)
	puddle.clients = make(map[uint64]*raft.Client)

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

	// RPC server --------------------------------------
	puddle.server = newPuddlestoreRPCServer(p)
	puddle.Local = PuddleAddr{puddle.server.listener.Addr().String()}
	// -------------------------------------------------

	// Create puddle raft client. Persist until raft is settled
	client, err := CreateClient(puddle.Local)
	for err != nil {
		client, err = CreateClient(puddle.Local)
	}
	// if err != nil {
	//	panic("Could not create puddle raft client.")
	//}
	puddle.raftClient = puddle.clients[client.Id]
	if puddle.raftClient == nil {
		panic("Could not retrieve puddle raft client.")
	}
	// -------------------------------------------------

	// Create the root node ----------------------------
	_, err = puddle.mkdir(&MkdirRequest{puddle.raftClient.Id, "/"})
	if err != nil {
		panic("Could not create root node")
	}
	// -------------------------------------------------

	fmt.Printf("Started puddlestore, listening at %v\n", puddle.server.listener.Addr().String())

	return
}

func (puddle *PuddleNode) getRandomTapestryNode() tapestry.Node {
	index := rand.Int() % TAPESTRY_NODES
	return puddle.tnodes[index].GetLocalNode()
}

func (puddle *PuddleNode) getRandomRaftNode() *raft.RaftNode {
	index := rand.Int() % RAFT_NODES
	return puddle.rnodes[index]
}

func (puddle *PuddleNode) getCurrentDir(id uint64) string {
	curdir, ok := puddle.clientPaths[id]
	if !ok {
		panic("Did not found the current path of a client that is supposed to be registered")
	}
	return curdir
}

func (puddle *PuddleNode) getRootInode(id uint64) *Inode {
	inode, err := puddle.getInode("/", id)
	if err != nil {
		panic("Root inode not found!")
	}
	return inode
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
