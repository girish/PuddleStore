package tapestry

import (
	"fmt"
	"strings"
	"testing"
)

var port int

func equal_ids(id1, id2 ID) bool {
	if SharedPrefixLength(id1, id2) == DIGITS {
		return true
	}
	return false
}

func printTable(table *RoutingTable) {
	fmt.Printf("RoutingTable for node %v\n", table.local)
	id := table.local.Id.String()
	for i, row := range table.rows {
		for j, slot := range row {
			for _, node := range *slot {
				fmt.Printf(" %v%v  %v: %v %v\n", id[:i], strings.Repeat(" ", DIGITS-i+1), Digit(j), node.Address, node.Id.String())
			}
		}
	}
	fmt.Printf("\n\n")
}

func makeTapestryNode(id ID, addr string, t *testing.T) *TapestryNode {

	tapestry, err := start(id, port, addr)

	if err != nil {
		t.Errorf("Error while making a tapestry %v", err)
	}

	port++
	return tapestry.local
}

// Prints the backpointers
func printBackpointers(b *Backpointers) {
	bp := b
	fmt.Printf("Backpointers for node %v\n", bp.local)
	for i, set := range bp.sets {
		for _, node := range set.Nodes() {
			fmt.Printf(" %v %v: %v\n", i, node.Address, node.Id.String())
		}
	}
	fmt.Printf("\n\n")
}
