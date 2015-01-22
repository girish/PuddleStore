/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: Finger table related functions for a given Chord node.          */
/*                                                                           */

package chord

import (
	"fmt"
	"time"
)

/* A single finger table entry */
type FingerEntry struct {
	Start []byte       /* ID hash of (n + 2^i) mod (2^m)  */
	Node  *RemoteNode  /* RemoteNode that Start points to */
}

/* Create initial finger table that only points to itself, will be fixed later */
func (node *Node) initFingerTable() {
	//TODO students should implement this method
}

/* Called periodically (in a seperate go routine) to fix entries in our finger table. */
func (node *Node) fixNextFinger(ticker *time.Ticker) {
	for _ = range ticker.C {
		//TODO students should implement this method
	}
}

/* (n + 2^i) mod (2^m) */
func fingerMath(n []byte, i int, m int) []byte {
	//TODO students should implement this method
	return nil
}

/* Print contents of a node's finger table */
func PrintFingerTable(node *Node) {
	fmt.Printf("[%v] FingerTable:\n", HashStr(node.Id))
	for _, val := range node.FingerTable {
		fmt.Printf("\t{start:%v\tnodeLoc:%v %v}\n",
			HashStr(val.Start), HashStr(val.Node.Id), val.Node.Addr)
	}
}
