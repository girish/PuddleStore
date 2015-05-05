package puddlestore

import (
	"fmt"
	"net"
	"net/rpc"
)

type PuddleRPCServer struct {
	node     *PuddleNode
	listener net.Listener
	rpc      *rpc.Server
}

func newPuddlestoreRPCServer(puddle *PuddleNode) (server *PuddleRPCServer) {
	server = new(PuddleRPCServer)
	server.node = puddle
	server.rpc = rpc.NewServer()
	listener, _, err := OpenListener()
	server.rpc.RegisterName(listener.Addr().String(), server)
	server.listener = listener

	if err != nil {
		panic("AA")
	}

	go func() {
		for {
			conn, err := server.listener.Accept()
			if err != nil {
				fmt.Printf("(%v) Raft RPC server accept error: %v\n", err)
				continue
			}
			go server.rpc.ServeConn(conn)
		}
	}()

	return
}

func (server *PuddleRPCServer) startRpcServer() {
	for {
		if server.node.IsShutdown {
			fmt.Printf("(%v) Shutting down RPC server\n")
			return
		}
		conn, err := server.node.Listener.Accept()
		if err != nil {
			if !server.node.IsShutdown {
				fmt.Printf("(%v) Raft RPC server accept error: %v\n", err)
			}
			continue
		}
		if !server.node.IsShutdown {
			go rpc.ServeConn(conn)
		} else {
			conn.Close()
		}
	}
}

func (server *PuddleRPCServer) ConnectImpl(req *ConnectRequest, rep *ConnectReply) error {

	fmt.Println("(Puddlestore) RPC Connect impl WE DID IT")
	err := server.node.Connect(req)

	return err
}
