package chord

import (
	"testing"
	"time"
)

func TestCreateNode(t *testing.T) {
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
