package tapestry

import (
	"testing"
	// "time"
)

func CheckFindRoot(node *TapestryNode, target ID, expected ID,
	t *testing.T) {
	result, _ := node.findRoot(node.node, target)
	if !equal_ids(result.Id, expected) {
		t.Errorf("findRoot of %v is not %v (gives %v)", target, expected,
			result.Id)
	}
}

// NOTE: This needs to set digits to 5 to work!
func TestFindRoot(t *testing.T) {
	if DIGITS != 4 {
		t.Errorf("Test wont work unless DIGITS is set to 4.")
	}

	port = 58000
	id := ID{5, 8, 3, 15}
	mainNode := makeTapestryNode(id, "", t)
	// t.Errorf("Address is: %v", mainNode.node.Address)

	id = ID{7, 0, 13, 1}
	node1 := makeTapestryNode(id, mainNode.node.Address, t)
	id = ID{7, 0, 15, 5}
	node2 := makeTapestryNode(id, mainNode.node.Address, t)
	id = ID{7, 0, 15, 10}
	node3 := makeTapestryNode(id, mainNode.node.Address, t)

	printTable(mainNode.table)
	printBackpointers(mainNode.backpointers)
	printTable(node1.table)
	printBackpointers(node1.backpointers)
	printTable(node2.table)
	printBackpointers(node2.backpointers)
	printTable(node3.table)

	id = ID{3, 0xf, 8, 0xa}
	CheckFindRoot(mainNode, id, mainNode.node.Id, t)
	id = ID{5, 2, 0, 0xc}
	CheckFindRoot(mainNode, id, mainNode.node.Id, t)
	id = ID{5, 8, 0xf, 0xf}
	CheckFindRoot(mainNode, id, mainNode.node.Id, t)
	id = ID{7, 0, 0xc, 3}
	CheckFindRoot(mainNode, id, node1.node.Id, t)
	id = ID{6, 0, 0xf, 4}
	CheckFindRoot(mainNode, id, node2.node.Id, t)

	// t.Errorf("Test wont work unless DIGITS is set to 4.")
	mainNode.Leave()
}
