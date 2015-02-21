package tapestry

import (
	"fmt"
	"strings"
	//"testing"
)

func printTable(table *RoutingTable) {
	id := table.local.Id.String()
	for i, row := range table.rows {
		for j, slot := range row {
			for _, node := range *slot {
				fmt.Printf(" %v%v  %v: %v\n", id[:i], strings.Repeat(" ", DIGITS-i+1), Digit(j), node.Id.String())
			}
		}
	}
}

// Adds 100,000 nodes to the table and removes them, checking
// that all where deleted.
/*
func TestSimpleAddAndRemove(t *testing.T) {
	// We need more stuff done before doing this.
	NUM_NODES := 100000
	me := Node{RandomID(), ""}
	table := NewRoutingTable(me)
	nodes := make([]Node, NUM_NODES)
	for i := 0; i < NUM_NODES; i++ {
		nodes[i] = Node{RandomID(), ""}
		table.Add(nodes[i])
	}
	printTable(table)
	for i := 0; i < NUM_NODES; i++ {
		table.Remove(nodes[i])
	}

	for i := 0; i < DIGITS; i++ {
		for j := 0; j < BASE; j++ {
			if len(*(table.rows[i][j])) != 0 {
				t.Errorf("Nodes where not deleted from table.")
			}
		}
	}
}
*/
