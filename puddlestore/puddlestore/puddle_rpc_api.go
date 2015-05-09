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
		fmt.Println(err)
		return nil, err
	}

	return &reply, nil
}

type LsRequest struct {
	ClientId uint64
}

type LsReply struct {
	Ok       bool
	Elements string
}

func lsRPC(remotenode *PuddleAddr, request LsRequest) (*LsReply, error) {
	var reply LsReply

	err := makeRemoteCall(remotenode, "LsImpl", request, &reply)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &reply, nil
}

type CdRequest struct {
	Path     string
	ClientId uint64
}

type CdReply struct {
	Ok bool
}

func cdRPC(remotenode *PuddleAddr, request CdRequest) (*CdReply, error) {
	var reply CdReply

	err := makeRemoteCall(remotenode, "CdImpl", request, &reply)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &reply, nil
}

type MkdirRequest struct {
	Path     string
	ClientId uint64
}

type MkdirReply struct {
	Ok bool
}

func mkdirRPC(remotenode *PuddleAddr, request MkdirRequest) (*MkdirReply, error) {
	var reply MkdirReply

	err := makeRemoteCall(remotenode, "MkdirImpl", request, &reply)
	if err != nil {
		fmt.Println(err)
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
