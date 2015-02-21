package tapestry

import (
	"testing"
)

// NOTE: This needs to set digits to 5 to work!
func TestFindRoot(t *testing.T) {
	if DIGITS != 4 {
		t.Errorf("Test wont work unless DIGITS is set to 4.")
	}

	port = 8080
	id := ID{5, 8, 3, 15}
	mainNode := makeTapestryNode(id, "", t)
	t.Errorf("Address is: %v", mainNode.node.Address)

	id = ID{7, 0, 13, 1}
	node1 := makeTapestryNode(id, mainNode.node.Address, t)
	id = ID{7, 0, 15, 5}
	node2 := makeTapestryNode(id, mainNode.node.Address, t)
	id = ID{7, 0, 15, 10}
	node3 := makeTapestryNode(id, mainNode.node.Address, t)

	printTable(mainNode.table)
	printTable(node1.table)
	printTable(node2.table)
	printTable(node3.table)

	t.Errorf("Test wont work unless DIGITS is set to 4.")
	mainNode.Leave()
}
