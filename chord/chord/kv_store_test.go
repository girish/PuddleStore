package chord

import (
	// "fmt"
	"math/rand"
	"strconv"
	"testing"
	// "time"
)

// -------- Create Node / Create Defined Node ------------
/*
This test does a local put and get on a single node
that is part of a ring
*/
// func TestSingleNodeLocalPutAndGetSimple(t *testing.T) {
// 	node, err := CreateNode(nil)
// 	if err != nil {
// 		t.Errorf("Unable to create node, received error:%v\n", err)
// 	}

// 	node2, err := CreateNode(node.RemoteSelf)
// 	if err != nil {
// 		t.Errorf("Unable to create node, received error:%v\n", err)
// 	}

// 	value := "valueOne"
// 	key := "keyOne"

// 	fmt.Printf("%p", node2)

// 	Put(node, key, value)
// 	value2, _ := Get(node, key)

// 	if (value != value2) {
// 		t.Errorf("TestSingleNodeLocalPutAndGetSimple: Value 1 and 2 are not the same")
// 	}
// }

// func TestRemotePutAndGetSimple(t *testing.T) {
// 	node, err := CreateNode(nil)
// 	if err != nil {
// 		t.Errorf("Unable to create node, received error:%v\n", err)
// 	}

// 	node2, err := CreateNode(node.RemoteSelf)
// 	if err != nil {
// 		t.Errorf("Unable to create node, received error:%v\n", err)
// 	}

// 	value := "valueOne"
// 	key := "keyOne"

// 	Put(node, key, value)
// 	value2, _ := Get(node2, key)

// 	if (value != value2) {
// 		t.Errorf("TestRemotePutAndGetSimple: Value 1 and 2 are not the same")
// 	}
// }

// func TestRemotePutAndGetByBundles(t *testing.T) {
// 	nNodes := 25
// 	nodes, _ := CreateNNodes(nNodes)

// 	for i := 0; i < nNodes; i++ {
// 		Put(nodes[i], strconv.Itoa(i*i), strconv.Itoa(i*i*i))
// 	}

// 	for i := nNodes - 1; i < 0; i++ {
// 		val := nNodes - i
// 		valueOne := strconv.Itoa(val*val*val)
// 		valueTwo, _ := Get(nodes[i+1], strconv.Itoa(val*val))
// 		if (valueOne != valueTwo) {
// 			t.Errorf("TestRemotePutAndGetByBundles: Value 1 and 2 are not the same")
// 		}
// 	}
// }

// func TestRemotePutAndGetSequence(t *testing.T) {
// 	nNodes := 25
// 	nodes, _ := CreateNNodes(nNodes)

// 	for i := 0; i < nNodes-1; i++ {
// 		Put(nodes[i], strconv.Itoa(i*i), strconv.Itoa(i*i*i))
// 		valueOne := string(i*i*i)
// 		valueTwo, _ := Get(nodes[i+1], strconv.Itoa(i*i))
// 		if (valueOne != valueTwo) {
// 			t.Errorf("TestRemotePutAndGetSequence: Value 1 and 2 are not the same")
// 		}
// 	}
// }

func TestRemotePutAndGetBundleRandom(t *testing.T) {
	nNodes := 10
	numRange := 100
	base := make(map[int]int64, numRange)
	result := make(map[int]int64, numRange)
	nodes, _ := CreateNNodesRandom(nNodes)
	//time.Sleep(3*time.Second)
	for i := 0; i < numRange; i++ {
		base[i] = int64(i * i)
		//Now we randomly pick a node and put the value in it
		nodeIndex := rand.Intn(9)
		Put(nodes[nodeIndex], strconv.Itoa(i), strconv.Itoa(i*i))
	}

	for i := 0; i < numRange; i++ {
		nodeIndex := rand.Intn(9)
		val, _ := Get(nodes[nodeIndex], strconv.Itoa(i))
		result[i], _ = strconv.ParseInt(val, 10, 32)
	}

	equal := true
	for i := 0; i < numRange; i++ {
		if result[i] != base[i] {
			equal = false
		}
	}
	if !equal {
		t.Errorf("TestRemotePutAndGetBundleRandom: result")
	}
}
