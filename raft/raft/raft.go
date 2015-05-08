package raft

import (
	"../../tapestry/tapestry"
	"crypto/sha1"
	"fmt"
	"math/big"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

/* Node's can be in three possible states */
type NodeState int

// Tapestry's id
type ID tapestry.ID

const (
	FOLLOWER_STATE NodeState = iota
	CANDIDATE_STATE
	LEADER_STATE
	JOIN_STATE
)

type RaftNode struct {
	Id         string
	Listener   net.Listener // Node listener socket
	listenPort int
	State      NodeState // The state this node is currently in
	LeaderAddr *NodeAddr // Our current leader

	config     *Config
	IsShutdown bool
	RPCServer  *RaftRPCServer
	mutex      sync.Mutex
	Testing    *TestingPolicy

	/* Raft log cache (do not use directly) */
	logCache []LogEntry

	/* File descriptors and values for persistent state */
	raftLogFd   FileData
	raftMetaFd  FileData
	stableState NodeStableState
	ssMutex     sync.Mutex

	/* Leader specific volitile state */
	commitIndex uint64
	lastApplied uint64
	leaderMutex sync.Mutex
	nextIndex   map[string]uint64 //we update it only when we actually append and sentToMaj = true (sned noop, client req and reg)
	matchIndex  map[string]uint64

	/* Channels to send/recv various RPC messages */
	appendEntries  chan AppendEntriesMsg
	requestVote    chan RequestVoteMsg
	clientRequest  chan ClientRequestMsg
	registerClient chan RegisterClientMsg
	gracefulExit   chan bool

	/* The replicated state machine */
	hash         []byte
	requestMutex sync.Mutex
	requestMap   map[uint64]ClientRequestMsg

	/*The map that we need to keep the state of PuddleStore*/
	fileMap    map[string]string
	fileMapMtx sync.Mutex

	/*The map that we need to keep the locks of PuddleStore*/
	lockMap    map[string]bool
	lockMapMtx sync.Mutex
}

type NodeAddr struct {
	Id   string
	Addr string
}

func CreateNode(localPort int, remoteAddr *NodeAddr, config *Config) (rp *RaftNode, err error) {
	var r RaftNode
	rp = &r
	var conn net.Listener

	r.IsShutdown = false
	r.config = config

	// init rpc channels
	r.appendEntries = make(chan AppendEntriesMsg)
	r.requestVote = make(chan RequestVoteMsg)
	r.clientRequest = make(chan ClientRequestMsg)
	r.registerClient = make(chan RegisterClientMsg)
	r.gracefulExit = make(chan bool)

	r.hash = nil
	r.requestMap = make(map[uint64]ClientRequestMsg)

	r.commitIndex = 0
	r.lastApplied = 0
	r.nextIndex = make(map[string]uint64)
	r.matchIndex = make(map[string]uint64)

	r.fileMap = make(map[string]string)

	r.lockMap = make(map[string]bool)

	r.Testing = NewTesting()
	r.Testing.PauseWorld(false)

	switch {
	case localPort != 0 && remoteAddr != nil:
		conn, err = OpenPort(localPort)
		if err != nil {
			return
		}
	case localPort != 0:
		conn, err = OpenPort(localPort)
		if err != nil {
			return
		}
	case remoteAddr != nil:
		conn, localPort, err = OpenListener()
		if err != nil {
			return
		}
	default:
		conn, localPort, err = OpenListener()
		if err != nil {
			return
		}
	}

	// Create node ID based on listener address
	r.Id = AddrToId(conn.Addr().String(), config.NodeIdSize)

	r.Listener = conn
	r.listenPort = localPort
	Out.Printf("Started node with id:%v, listening at %v\n", r.Id, conn.Addr().String())

	freshNode, err := r.initStableStore()
	if err != nil {
		Error.Printf("Error intitializing the stable store: %v", err)
		return nil, err
	}

	r.setLocalAddr(&NodeAddr{Id: r.Id, Addr: conn.Addr().String()})

	// Start RPC server
	r.RPCServer = &RaftRPCServer{rp}
	rpc.RegisterName(r.GetLocalAddr().Addr, r.RPCServer)
	go r.RPCServer.startRpcServer()

	if freshNode {
		r.State = JOIN_STATE
		if remoteAddr != nil {
			err = JoinRPC(remoteAddr, r.GetLocalAddr())
		} else {
			Out.Printf("Waiting to start nodes until all have joined\n")
			go r.startNodes()
		}
	} else {
		r.State = FOLLOWER_STATE
		go r.run()
	}

	return
}

func (r *RaftNode) startNodes() {
	r.mutex.Lock()
	r.AppendOtherNodes(*r.GetLocalAddr())
	r.mutex.Unlock()

	for len(r.GetOtherNodes()) < r.config.ClusterSize {
		time.Sleep(time.Millisecond * 100)
	}

	for _, otherNode := range r.GetOtherNodes() {
		if r.Id != otherNode.Id {
			Out.Printf("(%v) Starting node-%v\n", r.Id, otherNode.Id)
			StartNodeRPC(otherNode, r.GetOtherNodes())
		}
	}

	// Start the Raft finite-state-machine, initially in follower state
	go r.run()
}

func CreateLocalCluster(config *Config) ([]*RaftNode, error) {
	if config == nil {
		config = DefaultConfig()
	}
	err := CheckConfig(config)
	if err != nil {
		return nil, err
	}

	nodes := make([]*RaftNode, config.ClusterSize)

	nodes[0], err = CreateNode(0, nil, config)
	for i := 1; i < config.ClusterSize; i++ {
		nodes[i], err = CreateNode(0, nodes[0].GetLocalAddr(), config)
		if err != nil {
			return nil, err
		}
	}

	return nodes, nil
}

func CreateDefinedLocalCluster(config *Config, ports []int) ([]*RaftNode, error) {
	if config == nil {
		config = DefaultConfig()
	}
	err := CheckConfig(config)
	if err != nil {
		return nil, err
	}

	nodes := make([]*RaftNode, config.ClusterSize)

	nodes[0], err = CreateNode(ports[0], nil, config)
	if err != nil {
		Error.Printf("Error creating first node: %v", err)
		return nodes, err
	}
	for i := 1; i < config.ClusterSize; i++ {
		nodes[i], err = CreateNode(ports[i], nodes[0].GetLocalAddr(), config)
		if err != nil {
			Error.Printf("Error creating %v-th node: %v", i, err)
			return nil, err
		}
	}

	return nodes, nil
}

func OpenPort(port int) (net.Listener, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	conn, err := net.Listen("tcp4", fmt.Sprintf("%v:%v", hostname, port))
	return conn, err
}

func AddrToId(addr string, length int) string {
	h := sha1.New()
	h.Write([]byte(addr))
	v := h.Sum(nil)
	keyInt := big.Int{}
	keyInt.SetBytes(v[:length])
	return keyInt.String()
}

func (r *RaftNode) Exit() {
	Out.Printf("Abruptly shutting down node!")
	os.Exit(0)
}

func (r *RaftNode) GracefulExit() {
	r.Testing.PauseWorld(true)
	Out.Printf("Gracefully shutting down node!")
	r.gracefulExit <- true
}

func (r *RaftNode) GetConfig() *Config {
	return r.config
}

func (r *RaftNode) run() {
	curr := r.doFollower
	for curr != nil {
		curr = curr()
	}
}
