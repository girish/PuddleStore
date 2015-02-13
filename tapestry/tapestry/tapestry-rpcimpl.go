package tapestry

import (
	"fmt"
	"net"
	//	"net/http"
	"net/rpc"
)

/*
	This file contains the RPC server impementation of Tapestry functions.  The RPC just proxies
	remote method invocations to the local node.

	Some functions defined in this file are for the Tapestry struct, some are for the TapestryRPCServer struct
*/

/*
	Receives remote invocations of methods for the local tapestry node
*/
type TapestryRPCServer struct {
	tapestry *Tapestry
	listener net.Listener
	rpc      *rpc.Server
}

/*
	Creates the tapestry RPC server of a tapestry node.  The RPC server receives function invocations,
	and proxies them to the tapestrynode implementations
*/
func newTapestryRPCServer(port int, tapestry *Tapestry) (server *TapestryRPCServer, err error) {
	// Create the RPC server
	server = new(TapestryRPCServer)
	server.tapestry = tapestry
	server.rpc = rpc.NewServer()
	server.rpc.Register(server)
	server.rpc.Register(NewBlobStoreRPC(tapestry.blobstore))
	server.listener, err = net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, fmt.Errorf("Tapestry RPC server unable to listen on tcp port %v, reason: %v", port, err)
	}

	// Start the RPC server
	go func() {
		for {
			cxn, err := server.listener.Accept()
			if err != nil {
				Debug.Printf("Server %v closing: %s\n", port, err)
				return
			}
			go server.rpc.ServeConn(cxn)
		}
	}()

	return
}

/*
	Force kill the tapestry RPC server.  For testing involuntary node deletion
*/
func (server *TapestryRPCServer) kill() {
	server.listener.Close()
}

type NextHopRequest struct {
	To Node
	Id ID
}
type NextHopResponse struct {
	HasNext bool
	Next    Node
}

type RegisterRequest struct {
	To   Node
	From Node
	Key  string
}
type RegisterResponse struct {
	IsRoot bool
}

type FetchRequest struct {
	To  Node
	Key string
}
type FetchResponse struct {
	To     Node
	IsRoot bool
	Values []Node
}

type RemoveBadNodesRequest struct {
	To       Node
	BadNodes []Node
}

type NodeRequest struct {
	To   Node
	Node Node
}

type AddNodeMulticastRequest struct {
	To      Node
	NewNode Node
	Level   int
}
type TransferRequest struct {
	To   Node
	From Node
	Data map[string][]Node
}

type GetBackpointersRequest struct {
	To    Node
	From  Node
	Level int
}

type NotifyLeaveRequest struct {
	To          Node
	From        Node
	Replacement *Node
}

func (server *TapestryRPCServer) validate(expect Node) error {
	if server.tapestry.local.node != expect {
		return fmt.Errorf("Remote node expected us to be %v, but we are %v", expect, server.tapestry.local.node)
	}
	return nil
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) Hello(req Node, rsp *Node) (err error) {
	*rsp = server.tapestry.local.node
	return
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) GetNextHop(req NextHopRequest, rsp *NextHopResponse) (err error) {
	err = server.validate(req.To)
	if err == nil {
		rsp.HasNext, rsp.Next, err = server.tapestry.local.GetNextHop(req.Id)
	}
	return
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) Register(req RegisterRequest, rsp *RegisterResponse) (err error) {
	err = server.validate(req.To)
	if err == nil {
		rsp.IsRoot, err = server.tapestry.local.Register(req.Key, req.From)
	}
	return
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) Fetch(req FetchRequest, rsp *FetchResponse) (err error) {
	err = server.validate(req.To)
	if err == nil {
		rsp.IsRoot, rsp.Values, err = server.tapestry.local.Fetch(req.Key)
	}
	return
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) RemoveBadNodes(req RemoveBadNodesRequest, rsp *Node) error {
	err := server.validate(req.To)
	if err != nil {
		return err
	}
	return server.tapestry.local.RemoveBadNodes(req.BadNodes)
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) AddNode(req NodeRequest, rsp *[]Node) (err error) {
	err = server.validate(req.To)
	if err != nil {
		return
	}
	neighbours, err := server.tapestry.local.AddNode(req.Node)
	*rsp = append(*rsp, neighbours...)
	return
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) AddNodeMulticast(req AddNodeMulticastRequest, rsp *[]Node) (err error) {
	err = server.validate(req.To)
	if err != nil {
		return err
	}
	neighbours, err := server.tapestry.local.AddNodeMulticast(req.NewNode, req.Level)
	*rsp = append(*rsp, neighbours...)
	return err
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) Transfer(req TransferRequest, rsp *Node) error {
	err := server.validate(req.To)
	if err != nil {
		return err
	}
	return server.tapestry.local.Transfer(req.From, req.Data)
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) AddBackpointer(req NodeRequest, rsp *Node) error {
	err := server.validate(req.To)
	if err != nil {
		return err
	}
	return server.tapestry.local.AddBackpointer(req.Node)
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) RemoveBackpointer(req NodeRequest, rsp *Node) error {
	err := server.validate(req.To)
	if err != nil {
		return err
	}
	return server.tapestry.local.RemoveBackpointer(req.Node)
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) GetBackpointers(req GetBackpointersRequest, rsp *[]Node) (err error) {
	err = server.validate(req.To)
	if err != nil {
		return err
	}
	backpointers, err := server.tapestry.local.GetBackpointers(req.From, req.Level)
	*rsp = append(*rsp, backpointers...)
	return
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) NotifyLeave(req NotifyLeaveRequest, rsp *Node) error {
	err := server.validate(req.To)
	if err != nil {
		return err
	}
	return server.tapestry.local.NotifyLeave(req.From, req.Replacement)
}
