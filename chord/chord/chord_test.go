package chord

import (
	"testing"
	"time"
)

// -------- Create Node / Create Defined Node ------------

func TestCreateNodeSingleNode(t *testing.T) {
	node, err := CreateNode(nil)
	if err != nil {
		t.Errorf("Unable to create node, received error:%v\n", err)
	}

	node2, err := CreateNode(node.RemoteSelf)
	if err != nil {
		t.Errorf("Unable to create node, received error:%v\n", err)
	}
	// Ensure finger table of both nodes is updated
	// (this may not be needed)
	time.Sleep(1)
	if !EqualIds(node2.Successor.Id, node.Id) {
		t.Errorf("Nodes not linked correctly", err)
	}
}

func TestCreateNodeTwoNodes(t *testing.T) {
	node, err := CreateNode(nil)
	if err != nil {
		t.Errorf("Unable to create node, received error:%v\n", err)
	}

	node2, err := CreateNode(node.RemoteSelf)
	if err != nil {
		t.Errorf("Unable to create node, received error:%v\n", err)
	}
	// Ensure finger table of both nodes is updated
	// (this may not be needed)
	time.Sleep(1)
	if !EqualIds(node2.Successor.Id, node.Id) {
		t.Errorf("Nodes not linked correctly", err)
	}
}

func TestCreateNodeMultipleNodes(t *testing.T) {
	nodes := make([]*Node, 10)

	id := make([]byte, KEY_LENGTH)
	id[0] = byte(0)
	curr, err := CreateDefinedNode(nil, id)
	nodes[0] = curr
	if err != nil {
		t.Errorf("Unable to create node, received error:%v\n", err)
	}

	for i := 1; i < 10; i += 1 {
		id := make([]byte, KEY_LENGTH)
		id[0] = byte(i * 10)
		curr, err := CreateDefinedNode(nodes[0].RemoteSelf, id)
		nodes[i] = curr
		if err != nil {
			t.Errorf("Unable to create node, received error:%v\n", err)
		}
	}

	time.Sleep(1)

	for i := 0; i < 10; i += 1 {
		curr := nodes[i]
		prevId := (i*10 - 10) % 100
		prev := make([]byte, KEY_LENGTH)
		prev[0] = byte(prevId)

		succId := (i*10 + 10) % 100
		succ := make([]byte, KEY_LENGTH)
		succ[0] = byte(succId)

		if curr == nil || curr.Predecessor == nil {
			t.Errorf("Node %v has no predecessor (and it should)\n",
				curr.Id)
			return
		}
		if !EqualIds(curr.Predecessor.Id, prev) {
			t.Errorf("Previous node mismatch: \n prev: %v",
				"must be: %v\n",
				curr.Predecessor.Id, prev)
		}

		if curr == nil || curr.Successor == nil {
			t.Errorf("Node %v has no successor (and it should)\n",
				curr.Id)
		}
		if !EqualIds(curr.Successor.Id, succ) {
			t.Errorf("Successor node mismatch: \n prev: %v",
				"must be: %v\n",
				curr.Predecessor.Id, prev)
		}
	}
}

func TestCreateDefinedNode(t *testing.T) {
}

/*

func TestPaPendejiar(t *testing.T) {
	h := HashKey("unstring")
	a := big.Int{}
	a.SetInt64(7)
	b := big.Int{}
	b.SetInt64(5)
	t.Errorf("%v\n", a)
	t.Errorf("%v\n", b)
	b.Add(&b, &a)
	t.Errorf("%v\n", b)
	fmt.Println(h[0])
}

*/
