/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: Local Chord node functions to interact with the Chord ring.     */
/*                                                                           */

package chord

import (
	"log"
	"time"
	"errors"
)

// This node is trying to join an existing ring that a remote node is a part of (i.e., other)
func (node *Node) join(other *RemoteNode) error {

	// Handle case of "other" being nil (first node on ring).
	if other == nil {
		return nil
	}

	node.Predecessor = nil
	succ, err := FindSuccessor_RPC(other, node.Id)
	if (EqualIds(succ.Id, node.Id)){
		return errors.New("node already exists")
	}
	node.ftLock.Lock()
	node.Successor = succ
	node.FingerTable[0].Node = succ
	node.ftLock.Unlock()
	return err
}

// Thread 2: Psuedocode from figure 7 of chord paper
func (node *Node) stabilize(ticker *time.Ticker) {
	for _ = range ticker.C {
		if node.IsShutdown {
			ticker.Stop()
			return
		}
		pred, err := GetPredecessorId_RPC(node.Successor)

		if err != nil {
			log.Fatal("GetPredecessorId_RPC error: " + err.Error())
		}

		if pred != nil && BetweenRightIncl(pred.Id, node.Id, node.Successor.Id) {
			node.ftLock.Lock()
			node.Successor = pred
			node.FingerTable[0].Node = pred
			node.ftLock.Unlock()
		}

		// If you are your own successor, do not notify yourself.
		if !EqualIds(node.Successor.Id, node.Id) {
			//fmt.Printf("calling notify on %v, from %v, %p\n", node.Successor.Id, node.Id, node)
			//fmt.Println("we are executing notify")
			err = Notify_RPC(node.Successor, node.RemoteSelf)
			if err != nil {
				log.Fatal("Notify_RPC error: " + err.Error())
			}
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
		//fmt.Println("inb4")
		if oldPred != nil {
			err := TransferKeys_RPC(node.RemoteSelf, remoteNode,
				oldPred.Id)
			if err != nil {
				log.Fatal("TransferKeys_RPC error: " + err.Error())
			}
		} else {
			err := TransferKeys_RPC(node.RemoteSelf, remoteNode,
				nil)
			if err != nil {
				log.Fatal("TransferKeys_RPC error: " + err.Error())
			}
		}
	}
	//This part is to handle the very initial case when there
	//are only two nodes (one existing and one newly joined)
	// and we get notified about the newly joined node
	//SO if we have ourselves as our successor we set the succesor to be
	//the new node we just found about.
	if EqualIds(node.Successor.Id, node.Id) {
		//we set the remote node to be our predecessor
		node.Successor = remoteNode
	}
}

// Psuedocode from figure 4 of chord paper
func (node *Node) findSuccessor(id []byte) (*RemoteNode, error) {
	//TODO students should implement this method

	// Check if id is between me and my immediate successor.
	// Check if I'm my own successor.
	// If so, return it.
	if BetweenRightIncl(id, node.Id, node.Successor.Id) ||
		EqualIds(node.Successor.Id, node.Id) {

		return node.Successor, nil
	}

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

	// Loop while id is not beteen the current node and the
	// calculated successor.
	for !Between(id, curr.Id, succ.Id) && !EqualIds(curr.Id, succ.Id) {
		curr, err = ClosestPrecedingFinger_RPC(curr, id)
		if err != nil {
			log.Fatal("ClosestPrecedingFinger_RPC error: " + err.Error())
		}

		succ, err = GetSuccessorId_RPC(curr)
		if err != nil {
			log.Fatal("GetSuccessorId_RPC error: " + err.Error())
		}
	}
	return curr, err
}
