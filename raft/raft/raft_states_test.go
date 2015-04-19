package raft

import (
	"fmt"
	//"strings"
	"testing"
	"time"
)

// Loops until it finds a majority leader in nodes.
func getLeader(nodes []*RaftNode) *RaftNode {
	//Check all and make sure that leader matches
	var leader *RaftNode
	leader = nil
	it := 1
	for leader == nil {
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

	return leader
}

func checkSameTerms(nodes []*RaftNode) bool {
	term := nodes[0].GetCurrentTerm()
	for _, n := range nodes {
		if n.GetCurrentTerm() != term {
			return false
		}
	}
	return true
}

func checkSameCommitIndex(nodes []*RaftNode) bool {
	index := nodes[0].commitIndex
	for _, n := range nodes {
		n.PrintLogCache()
		n.ShowState()
		if n.commitIndex != index {
			return false
		}
	}
	return true
}

// Checks a leader selection with 5 nodes.
// Also makes sure that NOOPs and few functionalities
// of log replication to work.
func TestLeaderElection(t *testing.T) {
	config := DefaultConfig()
	config.ClusterSize = 5

	nodes := make([]*RaftNode, config.ClusterSize)
	var err error
	nodes[0], err = CreateNode(0, nil, config)
	if err != nil {
		t.Errorf("Could not create node")
		return
	}

	for i := 1; i < config.ClusterSize; i++ {
		nodes[i], err = CreateNode(0, nodes[0].GetLocalAddr(), config)
		if err != nil {
			t.Errorf("Could not create node")
			return
		}
	}

	time.Sleep(time.Millisecond * 500)
	fmt.Printf("Before loop")
	leader := getLeader(nodes)
	fmt.Printf("after loop")
	if leader == nil {
		t.Errorf("Leader is not the same %v is not located in node", leader.Id)
	} else {
		fmt.Println("Initial leader election worked.")
	}

	time.Sleep(time.Millisecond * 500)
	if !checkSameTerms(nodes) {
		t.Errorf("Nodes are not on the same term (%v)", leader.GetCurrentTerm())
	}
	if !checkSameCommitIndex(nodes) {
		t.Errorf("Nodes dont have the same commit index (%v)", leader.commitIndex)
	}

	leader.Testing.PauseWorld(true)
	disableLeader := leader
	fmt.Printf("The disabled node is: %v\n", leader.Id)
	leader = getLeader(nodes)
	if leader == nil {
		t.Errorf("Leader is not the same %v is not located in node", leader.Id)
	}

	disableLeader.Testing.PauseWorld(false)
	fmt.Printf("We now enable %v\n", disableLeader.Id)
	leader = getLeader(nodes)
	if leader == nil {
		t.Errorf("Leader is not the same %v is not located in node", leader.Id)
	}
	time.Sleep(time.Millisecond * 1000)
	if !checkSameTerms(nodes) {
		t.Errorf("Nodes are not on the same term (%v)", leader.GetCurrentTerm())
	}
	if !checkSameCommitIndex(nodes) {
		t.Errorf("Nodes dont have the same commit index (%v)", leader.commitIndex)
	}
	// t.Errorf("NOOP")
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
