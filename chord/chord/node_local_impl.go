/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: Local Chord node functions to interact with the Chord ring.     */
/*                                                                           */

package chord

import (
	"fmt"
	"log"
	"time"
)

// This node is trying to join an existing ring that a remote node is a part of (i.e., other)
func (node *Node) join(other *RemoteNode) error {

	// Handle case of "other" being nil (first node on ring).
	if other == nil {
		return nil
	}

	node.Predecessor = nil
	succ, err := FindSuccessor_RPC(other, node.Id)
	node.Successor = succ
	return err
}

// Thread 2: Psuedocode from figure 7 of chord paper
func (node *Node) stabilize(ticker *time.Ticker) {
	for _ = range ticker.C {
		if node.IsShutdown {
			fmt.Printf("[%v-stabilize] Shutting down stabilize timer\n", HashStr(node.Id))
			ticker.Stop()
			return
		}

		//TODO students should implement this method
		pred, err := GetPredecessorId_RPC(node.Successor)
		if err != nil {
			log.Fatal("GetPredecessorId_RPC error: " + err.Error())
		}

		if Between(pred.Id, node.Id, node.Successor.Id) {
			node.Successor = pred
		}

		err = Notify_RPC(node.Successor, node.RemoteSelf)
		if err != nil {
			log.Fatal("Notify_RPC error: " + err.Error())
		}
	}
}

// Psuedocode from figure 7 of chord paper
func (node *Node) notify(remoteNode *RemoteNode) {

	//TODO students should implement this method
	if node.Predecessor == nil ||
		Between(remoteNode.Id, node.Predecessor.Id, node.Id) {

		oldPred := node.Predecessor
		node.Predecessor = remoteNode

		// TODO: transfer keys
		err := TransferKeys_RPC(node.RemoteSelf, remoteNode,
			oldPred.Id)
		if err != nil {
			log.Fatal("TransferKeys_RPC error: " + err.Error())
		}
	}
}

// Psuedocode from figure 4 of chord paper
func (node *Node) findSuccessor(id []byte) (*RemoteNode, error) {
	//TODO students should implement this method
	n, err := node.findPredecessor(id)
	if err != nil {
		log.Fatal("findPredecessor error: " + err.Error())
	}

	return FindSuccessor_RPC(n, id)

}

// Psuedocode from figure 4 of chord paper
func (node *Node) findPredecessor(id []byte) (*RemoteNode, error) {
	//TODO students should implement this method
	curr := node.RemoteSelf
	succ, err := GetSuccessorId_RPC(curr)
	for !Between(id, curr.Id, succ.Id) {
		curr, err = ClosestPrecedingFinger_RPC(curr, id)
		// TODO: arreglar mamadas
		succ, err = GetSuccessorId_RPC(curr)
	}
	return curr, err
}
