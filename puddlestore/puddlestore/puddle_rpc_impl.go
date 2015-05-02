package puddlestore

import (
	"fmt"
	"net/rpc"
)

type PuddleRPCServer struct {
	node *Puddlestore
}

func (server *PuddleRPCServer) startRpcServer() {
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
