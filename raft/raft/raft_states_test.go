package raft

import (
	"fmt"
	//"strings"
	"math/rand"
	"testing"
	"time"
)

// Checks a leader selection with 5 nodes.
// Also makes sure that NOOPs and few functionalities
// of log replication to work.
func TestLeaderElection(t *testing.T) {
	config := DefaultConfig()
	config.ClusterSize = 5
	config.LogPath = randSeq(10)

	nodes, err := CreateLocalCluster(config)
	if err != nil {
		t.Errorf("Could not create nodes")
		return
	}
	time.Sleep(time.Millisecond * 500)
	if !checkNodes(nodes, config.ClusterSize) {
		t.Errorf("CreateLocalCluster FAILED")
		return
	}

	fmt.Printf("Before loop\n")
	leader := getLeader(nodes)
	fmt.Printf("after loop\n")
	if leader == nil {
		t.Errorf("Not found the leader")
		fmt.Printf("# nodes: %v\n", len(nodes))
		printNodes(nodes)
		return
	}

	time.Sleep(time.Millisecond * 500)
	if !checkMajorityTerms(nodes) {
		t.Errorf("Nodes are not on the same term (%v)", leader.GetCurrentTerm())
	}
	if !checkMajorityCommitIndex(nodes) {
		t.Errorf("Nodes dont have the same commit index (%v)", leader.commitIndex)
	}
	if !checkLogOrder(nodes) {
		t.Errorf("Nodes logs are not in an ok order")
		printNodes(nodes)
	}

	fmt.Printf("The disabled node is: %v\n", leader.Id)
	leader.Testing.PauseWorld(true)
	disableLeader := leader
	time.Sleep(time.Millisecond * 100)
	leader = getLeader(nodes)
	if leader == nil {
		t.Errorf("Leader is not the same %v is not located in node", leader.Id)
		return
	}

	fmt.Printf("We now enable %v\n", disableLeader.Id)
	disableLeader.Testing.PauseWorld(false)
	time.Sleep(time.Millisecond * 100)
	leader = getLeader(nodes)
	if leader == nil {
		t.Errorf("Leader is not the same %v is not located in node", leader.Id)
		return
	}
	time.Sleep(time.Millisecond * 500)
	if !checkMajorityTerms(nodes) {
		t.Errorf("Nodes are not on the same term (%v)", leader.GetCurrentTerm())
	}
	if !checkMajorityCommitIndex(nodes) {
		t.Errorf("Nodes dont have the same commit index (%v)", leader.commitIndex)
	}
	if !checkLogOrder(nodes) {
		t.Errorf("Nodes logs are not in an ok order")
		printNodes(nodes)
	}
	// t.Errorf("NOOP")

	fmt.Println("TestLeaderElection pass")
	// removeLogs(nodes)
	shutdownNodes(nodes)
}

// Test a raft of 5 nodes with TONS of disables
func TestDisruptiveEnvironment(t *testing.T) {
	config := DefaultConfig()
	config.ClusterSize = 5
	config.LogPath = randSeq(10)

	nodes, err := CreateLocalCluster(config)
	if err != nil {
		t.Errorf("Could not create nodes")
		return
	}
	time.Sleep(time.Millisecond * 500)

	// Distrupts at most 2 nodes at the time.
	for i := 0; i < 15; i++ {
		disabledNode1 := nodes[rand.Int()%config.ClusterSize]
		disabledNode2 := nodes[rand.Int()%config.ClusterSize]
		disableRandomNode(disabledNode1)
		disableRandomNode(disabledNode2)
		time.Sleep(time.Millisecond * 100)
		disabledNode1.Testing.PauseWorld(false)
		disabledNode2.Testing.PauseWorld(false)
	}

	leader := getLeader(nodes)
	if leader == nil {
		t.Errorf("Leader not found")
		return
	}
	time.Sleep(time.Millisecond * 500)
	if !checkMajorityTerms(nodes) {
		t.Errorf("Nodes are not on the same term (%v)", leader.GetCurrentTerm())
	}
	if !checkMajorityCommitIndex(nodes) {
		t.Errorf("Nodes dont have the same commit index (%v)", leader.commitIndex)
	}
	if !checkLogOrder(nodes) {
		t.Errorf("Nodes logs are not in an ok order")
		printNodes(nodes)
	}

	// Distrupts at most 4 nodes at the time.
	for i := 0; i < 1; i++ {
		disabledNode1 := nodes[rand.Int()%config.ClusterSize]
		disabledNode2 := nodes[rand.Int()%config.ClusterSize]
		disabledNode3 := nodes[rand.Int()%config.ClusterSize]
		disabledNode4 := nodes[rand.Int()%config.ClusterSize]
		disableRandomNode(disabledNode1)
		disableRandomNode(disabledNode2)
		disableRandomNode(disabledNode3)
		disableRandomNode(disabledNode4)
		time.Sleep(time.Millisecond * 200)
		disabledNode1.Testing.PauseWorld(false)
		disabledNode2.Testing.PauseWorld(false)
		disabledNode3.Testing.PauseWorld(false)
		disabledNode4.Testing.PauseWorld(false)
	}

	leader = getLeader(nodes)
	if leader == nil {
		t.Errorf("Leader not found")
		return
	}
	time.Sleep(time.Millisecond * 500)
	if !checkMajorityTerms(nodes) {
		t.Errorf("Nodes are not on the same term (%v)", leader.GetCurrentTerm())
	}
	if !checkMajorityCommitIndex(nodes) {
		t.Errorf("Nodes dont have the same commit index (%v)", leader.commitIndex)
	}
	if !checkLogOrder(nodes) {
		t.Errorf("Nodes logs are not in an ok order")
		printNodes(nodes)
	}
	// removeLogs(nodes)
}

/*
func TestClient(t *testing.T) {
	config := DefaultConfig()
	config.ClusterSize = 5
	config.LogPath = randSeq(10)

	nodes, err := CreateLocalCluster(config)
	if err != nil {
		t.Errorf("Could not create nodes")
		return
	}

	time.Sleep(time.Millisecond * 500)
	leader := getLeader(nodes)

	client, err := CreateClient(*nodes[0].GetLocalAddr())
	if err != nil {
		t.Errorf("Could not create client")
	}

	if client.Id != leader.getLastLogIndex() {
		t.Errorf("Client not really registered")
	}

	err = client.SendRequest(HASH_CHAIN_INIT, []byte("tacos"))
	if err != nil {
		t.Errorf("Could not excecute command.")
	}

		err = client.SendRequest(HASH_CHAIN_ADD, []byte("Hola como estas"))
		if err != nil {
			t.Errorf("Could not excecute command.")
		}
		fmt.Println(string(leader.getLastLogEntry().Data))
		if string(leader.getLastLogEntry().Data) != "Hola como estas" {
			t.Errorf("Data wasn't saved in log.")
		}
}
*/

func disableRandomNode(n *RaftNode) {
	n.Testing.PauseWorld(true)
}

func checkLogOrder(nodes []*RaftNode) bool {
	for _, n := range nodes {
		prevIndex := int64(-1)
		prevTerm := int64(-1)
		seen := make(map[uint64]bool)
		for _, entry := range n.logCache {
			if seen[entry.Index] || int64(entry.Index)-1 != prevIndex || int64(entry.TermId) < prevTerm {
				return false
			}

			seen[entry.Index] = true
			prevIndex = int64(entry.Index)
			prevTerm = int64(entry.TermId)
		}
	}
	return true
}

// Loops until it finds a majority leader in nodes.
func getLeader(nodes []*RaftNode) *RaftNode {
	//Check all and make sure that leader matches
	var leader *RaftNode
	leader = nil
	it := 1
	for leader == nil && it < 50 {
		fmt.Printf("%v\n", it)
		time.Sleep(time.Millisecond * 200)
		sums := make(map[string]int, nodes[0].config.ClusterSize)
		for _, n := range nodes {
			if n.LeaderAddr != nil {
				sums[n.LeaderAddr.Id]++
			}
		}
		fmt.Printf("mapa %v\n\n\n", sums)
		var maxNode string
		max := -1
		for k, v := range sums {
			if v > max {
				maxNode = k
				max = v
			}
		}

		if max > len(nodes)/2 {
			for _, n := range nodes {
				if maxNode == n.Id {
					leader = n
				}
			}
		}
		it++
	}

	if it >= 50 {
		return nil
	}
	return leader
}

func checkMajorityTerms(nodes []*RaftNode) bool {
	sums := make(map[uint64]int, nodes[0].config.ClusterSize)
	for _, n := range nodes {
		sums[n.GetCurrentTerm()]++
	}
	max := -1
	for _, v := range sums {
		if v > max {
			max = v
		}
	}

	if max > len(nodes)/2 {
		return true
	}
	return false
}

func checkMajorityCommitIndex(nodes []*RaftNode) bool {
	sums := make(map[uint64]int, nodes[0].config.ClusterSize)
	for _, n := range nodes {
		sums[n.commitIndex]++
	}
	max := -1
	for _, v := range sums {
		if v > max {
			max = v
		}
	}

	if max > len(nodes)/2 {
		return true
	}
	return false
}

func checkNodes(nodes []*RaftNode, clusterSize int) bool {
	for _, n := range nodes {
		if len(n.GetOtherNodes()) != clusterSize {
			return false
		}
	}
	return true
}

func printNodes(nodes []*RaftNode) {
	for _, n := range nodes {
		n.PrintLogCache()
		n.ShowState()
	}
}

func removeLogs(nodes []*RaftNode) {
	for _, n := range nodes {
		n.RemoveLogs()
	}
}

func shutdownNodes(nodes []*RaftNode) {
	for _, n := range nodes {
		n.IsShutdown = true
		n.gracefulExit <- true
	}
	time.Sleep(time.Millisecond * 200)
}

func randSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

/*
func checkMajority(leader *RaftNode, nodes []*RaftNode) bool {
	if leader == nil {
		return false
	}

	sum := 0
	for _, n := range nodes {
		if n.LeaderAddr != nil && leader.Id == n.LeaderAddr.Id {
			sum++
		}
	}

	fmt.Println(sum)
	if sum > len(nodes)/2 {
		return true
	} else {
		return false
	}
}*/
