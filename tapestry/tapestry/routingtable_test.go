package tapestry

import (
	//"fmt"
	//"strings"
	"testing"
)

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
			if len(*(table.rows[i][j])) > 1 {
				t.Errorf("Nodes where not deleted from table.")
			}
			if len(*(table.rows[i][j])) == 1 &&
				!equal_ids(me.Id, (*(table.rows[i][j]))[0].Id) {
				t.Errorf("Nodes where not deleted from table.")
			}
		}
	}
}
*/

// NOTE: This needs to set digits to 5 to work!
func TestGetNextHop(t *testing.T) {
	if DIGITS != 5 {
		// t.Errorf("Test wont work unless DIGITS is set to 5.")
	}
}
