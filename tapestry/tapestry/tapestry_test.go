package tapestry

import (
	"fmt"
	"testing"
	"time"
)

func CheckGet(err error, result []byte, expected string, t *testing.T) {
	if err != nil {
		t.Errorf("Get errored out. returned: %v", err)
		return
	}

	if string(result) != expected {
		t.Errorf("Get(\"%v\") did not return expected result '%v'",
			string(result), expected)
	}
}

func FindRootOfHash(nodes []*Tapestry, hash ID) *Tapestry {
	if len(nodes) == 0 {
		return nil
	}
	root, _ := nodes[0].local.findRoot(nodes[0].local.node, hash)

	for _, node := range nodes {
		if equal_ids(node.local.node.Id, root.Id) {
			return node
		}
	}

	return nil
}

func TestPublishAndRegister(t *testing.T) {
	if DIGITS != 4 {
		t.Errorf("Test wont work unless DIGITS is set to 4.")
		return
	}
	if TIMEOUT > 3*time.Second && REPUBLISH > 2*time.Second {
		t.Errorf("Test will take too long unless TIMEOUT is set to 3 and REPUBLISH is set to 2.")
		return
	}

	port = 58000
	id := ID{5, 8, 3, 15}
	node0 := makeTapestry(id, "", t)
	id = ID{7, 0, 0xd, 1}
	node1 := makeTapestry(id, node0.local.node.Address, t)
	id = ID{9, 0, 0xf, 5}
	node2 := makeTapestry(id, node0.local.node.Address, t)
	id = ID{0xb, 0, 0xf, 0xa}
	node3 := makeTapestry(id, node0.local.node.Address, t)

	node0.Store("spoon", []byte("cuchara"))
	node1.Store("table", []byte("mesa"))
	node2.Store("chair", []byte("silla"))
	node3.Store("fork", []byte("tenedor"))

	time.Sleep(time.Second * 5)

	// Objects should persist after TIMEOUT seconds because
	// publish is called every two seconds.
	result, err := node1.Get("spoon")
	CheckGet(err, result, "cuchara", t)
	result, err = node2.Get("table")
	CheckGet(err, result, "mesa", t)
	result, err = node3.Get("chair")
	CheckGet(err, result, "silla", t)
	result, err = node0.Get("fork")
	CheckGet(err, result, "tenedor", t)

	// Root node of Hash(spoon) should no longer have a record
	// of this object after node0 leaves after TIMEOUT seconds.
	root := FindRootOfHash([]*Tapestry{node1, node2, node3}, Hash("chair"))
	node0.Leave()
	fmt.Printf("The root is: %v and the node0 id is: %v", root.local.node.Id, node0.local.node.Id)
	if root == nil {
		t.Errorf("Could not find Root of Hash")
	} else {
		replicas := root.local.store.Get("spoon")
		if len(replicas) == 0 && len(replicas) > 1 {
			t.Errorf("Replica of 'spoon' not in root node. What?")
		} else {
			time.Sleep(time.Second * 5)
			replicas = root.local.store.Get("spoon")
			if len(replicas) != 0 {
				t.Errorf("Replica of 'spoon' is in root node after node containing it left.")
			}
		}
	}
	//We add a new node that contains spoon and we should find it.
	id = ID{0x5, 2, 0xa, 0xa}
	node4 := makeTapestry(id, node2.local.node.Address, t)
	node4.Store("spoon", []byte("cuchara"))
	time.Sleep(time.Second * 5)
	replicas, _ := node1.local.tapestry.Get("spoon")
	fmt.Printf("id of root is: %v\n", root.local.node.Id)
	// printTable(node4.local.table)
	// printBackpointers(node4.local.backpointers)
	// printTable(node1.local.table)
	// printBackpointers(node1.local.backpointers)
	// printTable(node2.local.table)
	// printBackpointers(node2.local.backpointers)
	// printTable(node3.local.table)
	// printBackpointers(node3.local.backpointers)
	if len(replicas) == 0 {
		t.Errorf("'spoon' is not there even after a new node containing it joined")
	}

	node1.Leave()
	node2.Leave()
	node3.Leave()
}

func TestPutAndGet(t *testing.T) {
	if DIGITS != 4 {
		t.Errorf("Test wont work unless DIGITS is set to 4.")
	}

	port = 58001
	id := ID{5, 8, 3, 15}
	node0 := makeTapestry(id, "", t)
	id = ID{7, 0, 0xd, 1}
	node1 := makeTapestry(id, node0.local.node.Address, t)
	id = ID{7, 0, 0xf, 5}
	node2 := makeTapestry(id, node0.local.node.Address, t)
	id = ID{7, 0, 0xf, 0xa}
	node3 := makeTapestry(id, node0.local.node.Address, t)

	node0.Store("spoon", []byte("cuchara"))
	node1.Store("table", []byte("mesa"))
	node2.Store("chair", []byte("silla"))
	node3.Store("fork", []byte("tenedor"))

	result, err := node0.Get("spoon")
	CheckGet(err, result, "cuchara", t)
	result, err = node1.Get("spoon")
	CheckGet(err, result, "cuchara", t)
	result, err = node2.Get("spoon")
	CheckGet(err, result, "cuchara", t)
	result, err = node3.Get("spoon")
	CheckGet(err, result, "cuchara", t)

	result, err = node0.Get("table")
	CheckGet(err, result, "mesa", t)
	result, err = node1.Get("table")
	CheckGet(err, result, "mesa", t)
	result, err = node2.Get("table")
	CheckGet(err, result, "mesa", t)
	result, err = node3.Get("table")
	CheckGet(err, result, "mesa", t)

	result, err = node0.Get("chair")
	CheckGet(err, result, "silla", t)
	result, err = node1.Get("chair")
	CheckGet(err, result, "silla", t)
	result, err = node2.Get("chair")
	CheckGet(err, result, "silla", t)
	result, err = node3.Get("chair")
	CheckGet(err, result, "silla", t)

	result, err = node0.Get("fork")
	CheckGet(err, result, "tenedor", t)
	result, err = node1.Get("fork")
	CheckGet(err, result, "tenedor", t)
	result, err = node2.Get("fork")
	CheckGet(err, result, "tenedor", t)
	result, err = node3.Get("fork")
	CheckGet(err, result, "tenedor", t)

	node0.Leave()
	node1.Leave()
	node2.Leave()
	node3.Leave()
}
