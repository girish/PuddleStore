package puddlestore

import (
	"fmt"
	"net/rpc"
)

var connMap = make(map[string]*rpc.Client)

type ConnectRequest struct {
	FromNode PuddleAddr
}

type ConnectReply struct {
	Ok bool
	Id uint64
}

func ConnectRPC(remotenode *PuddleAddr, request ConnectRequest) (*ConnectReply, error) {
	fmt.Println("(Puddlestore) RPC Connect to", remotenode.Addr)
	var reply ConnectReply

	err := makeRemoteCall(remotenode, "ConnectImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type LsRequest struct {
	ClientId uint64
	Path     string
}

type LsReply struct {
	Ok       bool
	Elements string
}

func lsRPC(remotenode *PuddleAddr, request LsRequest) (*LsReply, error) {
	var reply LsReply

	err := makeRemoteCall(remotenode, "LsImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type CdRequest struct {
	ClientId uint64
	Path     string
}

type CdReply struct {
	Ok bool
}

func cdRPC(remotenode *PuddleAddr, request CdRequest) (*CdReply, error) {
	var reply CdReply

	err := makeRemoteCall(remotenode, "CdImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type MkdirRequest struct {
	ClientId uint64
	Path     string
}

type MkdirReply struct {
	Ok bool
}

func mkdirRPC(remotenode *PuddleAddr, request MkdirRequest) (*MkdirReply, error) {
	var reply MkdirReply

	err := makeRemoteCall(remotenode, "MkdirImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type RmdirRequest struct {
	ClientId uint64
	Path     string
}

type RmdirReply struct {
	Ok bool
}

func rmdirRPC(remotenode *PuddleAddr, request RmdirRequest) (*RmdirReply, error) {
	var reply RmdirReply

	err := makeRemoteCall(remotenode, "RmdirImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type MkfileRequest struct {
	ClientId uint64
	Path     string
}

type MkfileReply struct {
	Ok bool
}

func mkfileRPC(remotenode *PuddleAddr, request MkfileRequest) (*MkfileReply, error) {
	var reply MkfileReply

	err := makeRemoteCall(remotenode, "MkfileImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type RmfileRequest struct {
	ClientId uint64
	Path     string
}

type RmfileReply struct {
	Ok bool
}

func rmfileRPC(remotenode *PuddleAddr, request RmfileRequest) (*RmfileReply, error) {
	var reply RmfileReply

	err := makeRemoteCall(remotenode, "RmfileImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type WritefileRequest struct {
	ClientId uint64
	Path     string
	Location uint32
	Buffer   []byte
}

type WritefileReply struct {
	Ok      bool
	Written uint32
}

func writefileRPC(remotenode *PuddleAddr, request WritefileRequest) (*WritefileReply, error) {
	var reply WritefileReply

	err := makeRemoteCall(remotenode, "WritefileImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type CatRequest struct {
	ClientId uint64
	Path     string
	Location uint32
	Count    uint32
}

type CatReply struct {
	Ok     bool
	Read   uint32
	Buffer []byte
}

func catRPC(remotenode *PuddleAddr, request CatRequest) (*CatReply, error) {
	var reply CatReply

	err := makeRemoteCall(remotenode, "CatImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

/* Helper function to make a call to a remote node */
func makeRemoteCall(remoteNode *PuddleAddr, method string, req interface{}, rsp interface{}) error {
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
		client.Close()
		delete(connMap, remoteNodeAddrStr)
		return err
	}

	return nil
}
