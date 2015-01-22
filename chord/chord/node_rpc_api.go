/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: RPC API used to communicate with other nodes in the Chord ring. */
/*                                                                           */

package chord

import (
	"errors"
	"fmt"
	"net/rpc"
)

type RemoteId struct {
	Id []byte
}

type RemoteQuery struct {
	FromId []byte
	Id     []byte
}

type IdReply struct {
	Id    []byte
	Addr  string
	Valid bool
}

type RpcOkay struct {
	Ok bool
}

type KeyValueReq struct {
	NodeId []byte
	Key    string
	Value  string
}

type KeyValueReply struct {
	Key   string
	Value string
}

type TransferReq struct {
	NodeId   []byte
	FromId   []byte
	FromAddr string
	PredId   []byte
}

/* RPC connection map cache */
var connMap = make(map[string]*rpc.Client)

/*                    */
/* Chord Node RPC API */
/*                    */

/* Get the predecessor ID of a remote node */
func GetPredecessorId_RPC(remoteNode *RemoteNode) (*RemoteNode, error) {
	var reply IdReply
	err := makeRemoteCall(remoteNode, "GetPredecessorId", RemoteId{remoteNode.Id}, &reply)
	if err != nil {
		return nil, err
	}

	if !reply.Valid {
		return nil, err
	}

	rNode := new(RemoteNode)
	rNode.Id = reply.Id
	rNode.Addr = reply.Addr
	return rNode, err
}

/* Get the successor ID of a remote node */
func GetSuccessorId_RPC(remoteNode *RemoteNode) (*RemoteNode, error) {
	var reply IdReply
	err := makeRemoteCall(remoteNode, "GetSuccessorId", RemoteId{remoteNode.Id}, &reply)
	if err != nil {
		return nil, err
	}
	rNode := new(RemoteNode)
	rNode.Id = reply.Id
	rNode.Addr = reply.Addr
	return rNode, err
}

/* Notify a remote node that we believe we are its predecessor */
func Notify_RPC(remoteNode, us *RemoteNode) error {
	if remoteNode == nil {
		return errors.New("RemoteNode is empty!")
	}
	var reply RpcOkay
	err := makeRemoteCall(remoteNode, "Notify", us, &reply)
	if !reply.Ok {
		return errors.New(fmt.Sprintf("RPC replied not valid from %v", remoteNode.Id))
	}

	return err
}

/* Find the closest preceding finger from a remote node for an ID */
func ClosestPrecedingFinger_RPC(remoteNode *RemoteNode, id []byte) (*RemoteNode, error) {
	if remoteNode == nil {
		return nil, errors.New("RemoteNode is empty!")
	}
	var reply IdReply
	err := makeRemoteCall(remoteNode, "ClosestPrecedingFinger", RemoteQuery{remoteNode.Id, id}, &reply)

	rNode := new(RemoteNode)
	rNode.Id = reply.Id
	rNode.Addr = reply.Addr
	return rNode, err
}

/* Find the successor node of a given ID in the entire ring */
func FindSuccessor_RPC(remoteNode *RemoteNode, id []byte) (*RemoteNode, error) {
	if remoteNode == nil {
		return nil, errors.New("RemoteNode is empty!")
	}
	var reply IdReply
	err := makeRemoteCall(remoteNode, "FindSuccessor", RemoteQuery{remoteNode.Id, id}, &reply)

	rNode := new(RemoteNode)
	rNode.Id = reply.Id
	rNode.Addr = reply.Addr
	return rNode, err
}

/*                   */
/* Datastore RPC API */
/*                   */

/* Get a value from a remote node's datastore for a given key */
func Get_RPC(locNode *RemoteNode, key string) (string, error) {
	if locNode == nil {
		return "", errors.New("RemoteNode is empty!")
	}

	var reply KeyValueReply
	req := KeyValueReq{locNode.Id, key, ""}
	err := makeRemoteCall(locNode, "GetLocal", &req, &reply)

	return reply.Value, err
}

/* Put a key/value into a datastore on a remote node */
func Put_RPC(locNode *RemoteNode, key string, value string) error {
	if locNode == nil {
		return errors.New("RemoteNode is empty!")
	}

	var reply KeyValueReply
	req := KeyValueReq{locNode.Id, key, value}
	err := makeRemoteCall(locNode, "PutLocal", &req, &reply)

	return err
}

/* Inform a successor node that we should now take care of IDs between (node.Id : predId] */
/* This should trigger the successor node to transfer the relevant keys back to node      */
func TransferKeys_RPC(succ *RemoteNode, node *RemoteNode, predId []byte) error {
	if succ == nil {
		return errors.New("RemoteNode is empty!")
	}
	var reply RpcOkay

	var req TransferReq
	req.NodeId = succ.Id
	req.FromId = node.Id
	req.FromAddr = node.Addr
	req.PredId = predId

	err := makeRemoteCall(succ, "TransferKeys", &req, &reply)
	if !reply.Ok {
		fmt.Println(err)
	}

	return err
}

/* Helper function to make a call to a remote node */
func makeRemoteCall(remoteNode *RemoteNode, method string, req interface{}, rsp interface{}) error {
	// Dial the server if we don't already have a connection to it
	remoteNodeAddrStr := remoteNode.Addr
	var err error
	client, ok := connMap[remoteNodeAddrStr]
	if !ok {
		client, err = rpc.Dial("tcp", remoteNode.Addr)
		if err != nil {
			return err
		}
		connMap[remoteNodeAddrStr] = client
	}

	// Make the request
	uniqueMethodName := fmt.Sprintf("%v.%v", remoteNodeAddrStr, method)
	err = client.Call(uniqueMethodName, req, rsp)
	if err != nil {
		return err
	}

	return nil
}
