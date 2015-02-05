package chord

import (
	// "fmt"
	// "testing"
	// "time"
)

// -------- Create Node / Create Defined Node ------------

// func TestCreateNodeSingleNode(t *testing.T) {
// 	node, err := CreateNode(nil)
// 	if err != nil {
// 		t.Errorf("Unable to create node, received error:%v\n", err)
// 	}

// 	node2, err := CreateNode(node.RemoteSelf)
// 	if err != nil {
// 		t.Errorf("Unable to create node, received error:%v\n", err)
// 	}
// 	// Ensure finger table of both nodes is updated
// 	// (this may not be needed)
// 	time.Sleep(1)
// 	if !EqualIds(node2.Successor.Id, node.Id) {
// 		t.Errorf("Nodes not linked correctly", err)
// 	}
// }

// func TestCreateNodeTwoNodes(t *testing.T) {
// 	node, err := CreateNode(nil)
// 	if err != nil {
// 		t.Errorf("Unable to create node, received error:%v\n", err)
// 	}

// 	node2, err := CreateNode(node.RemoteSelf)
// 	if err != nil {
// 		t.Errorf("Unable to create node, received error:%v\n", err)
// 	}
// 	// Ensure finger table of both nodes is updated
// 	// (this may not be needed)
// 	time.Sleep(time.Second)
// 	if !EqualIds(node2.Successor.Id, node.Id) {
// 		t.Errorf("Nodes not linked correctly")
// 	}
// 	if !EqualIds(node2.Predecessor.Id, node.Id) {
// 		t.Errorf("Nodes not linked correctly")
// 	}
// }

// func TestCreateNodeMultipleNodes(t *testing.T) {
// 	n := 10
// 	nodes, err := CreateNNodes(10)
// 	if err != nil {
// 		t.Errorf("CreateNNodes error: %v \n", err)
// 	}

// 	time.Sleep(time.Second)

// 	for _, node := range nodes {
// 		fmt.Println(NodeStr(node))
// 	}

// 	for i := 0; i < n; i += 1 {
// 		curr := nodes[i]
// 		predId := (((i - 1) % n) + n) % n * 10
// 		pred := make([]byte, KEY_LENGTH)
// 		pred[0] = byte(predId)

// 		succId := (((i + 1) % n) + n) % n * 10
// 		succ := make([]byte, KEY_LENGTH)
// 		succ[0] = byte(succId)

// 		fmt.Println(predId, succId)

// 		if curr == nil || curr.Predecessor == nil {
// 			t.Errorf("Node %v has no predecessor (and it should)\n",
// 				curr.Id)
// 			return
// 		}
// 		if !EqualIds(curr.Predecessor.Id, pred) {
// 			t.Errorf("%v Previous node mismatch: \n prev: %v, must be: %v \n",
// 				curr.Id, curr.Predecessor.Id, pred)
// 		}

// 		if curr == nil || curr.Successor == nil {
// 			t.Errorf("Node %v has no successor (and it should)\n",
// 				curr.Id)
// 			return
// 		}
// 		if !EqualIds(curr.Successor.Id, succ) {
// 			t.Errorf("%v Successor node mismatch: \n succ: %v, must be: %v\n",
// 				curr.Id, curr.Successor.Id, succ)
// 		}
// 	}
// }
