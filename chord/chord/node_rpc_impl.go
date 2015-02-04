/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: RPC API implementation, these are the functions that actually   */
/*           get executed on a destination Chord node when a *_RPC()         */
/*           function is called.                                             */
/*                                                                           */

package chord

import (
	"bytes"
	"errors"
	"fmt"
)

/* Validate that we're executing this RPC on the intended node */
func validateRpc(node *Node, reqId []byte) error {
	if !bytes.Equal(node.Id, reqId) {
		errStr := fmt.Sprintf("Node ids do not match %v, %v", node.Id, reqId)
		return errors.New(errStr)
	}
	return nil
}

/* RPC */
func (node *Node) GetPredecessorId(req *RemoteId, reply *IdReply) error {
	if err := validateRpc(node, req.Id); err != nil {
		return err
	}
	// Predecessor may be nil, which is okay.
	if node.Predecessor == nil {
		reply.Id = nil
		reply.Addr = ""
		reply.Valid = false
	} else {
		reply.Id = node.Predecessor.Id
		reply.Addr = node.Predecessor.Addr
		reply.Valid = true
	}
	return nil
}

/* RPC */
func (node *Node) GetSuccessorId(req *RemoteId, reply *IdReply) error {
	if err := validateRpc(node, req.Id); err != nil {
		return err
	}
	//TODO students should implement this method

	reply.Id = node.Successor.Id
	reply.Addr = node.Successor.Addr
	reply.Valid = true
	return nil

	/* REMOVE WHEN SURE

	remNode, err := node.findSuccessor(req.Id)
	if err != nil {
		reply.Valid = false
		return err
	}
	reply.Id = remNode.Id
	reply.Addr = remNode.Addr
	reply.Valid = true
	return nil

	*/
}

/* RPC */
func (node *Node) SetPredecessorId(req *UpdateReq, reply *RpcOkay) error {
	if err := validateRpc(node, req.FromId); err != nil {
		return err
	}
	//TODO students should implement this method
	return nil
}

/* RPC */
func (node *Node) SetSuccessorId(req *UpdateReq, reply *RpcOkay) error {
	if err := validateRpc(node, req.FromId); err != nil {
		return err
	}
	//TODO students should implement this method
	return nil
}

/* RPC */
func (node *Node) Notify(req *NotifyReq, reply *RpcOkay) error {
	//TODO fix this
	if err := validateRpc(node, req.NodeId); err != nil {
		reply.Ok = false
		return err
	}
	remote_node := new(RemoteNode)
	remote_node.Id = req.UpdateId
	remote_node.Addr = req.UpdateAddr
	node.notify(remote_node)
	reply.Ok = true
	return nil
}

/* RPC */
func (node *Node) FindSuccessor(query *RemoteQuery, reply *IdReply) error {
	if err := validateRpc(node, query.FromId); err != nil {
		return err
	}
	//TODO students should implement this method
	remNode, err := node.findSuccessor(query.Id)
	if err != nil {
		reply.Valid = false
		return err
	}
	reply.Id = remNode.Id
	reply.Addr = remNode.Addr
	reply.Valid = true
	return nil
}

/* RPC */
func (node *Node) ClosestPrecedingFinger(query *RemoteQuery, reply *IdReply) error {
	if err := validateRpc(node, query.FromId); err != nil {
		return err
	}
	//TODO students should implement this method
	//remoteId and fromId
	for i := KEY_LENGTH - 1; i >= 0; i-- {
		if BetweenRightIncl(node.FingerTable[i].Node.Id, node.Id, query.Id) {
			reply.Id = node.FingerTable[i].Node.Id
			reply.Addr = node.FingerTable[i].Node.Addr
			reply.Valid = true
			return nil
		}
	}

	reply.Id = node.Successor.Id
	reply.Addr = node.Successor.Addr
	reply.Valid = true
	return nil

	reply.Valid = false

	//TODO: return some error
	return errors.New("There is no closest preceding finger")
}
