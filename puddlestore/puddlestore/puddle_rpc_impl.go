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

func (server *PuddleRPCServer) ConnectImpl(req *ConnectRequest, rep *ConnectReply) error {
	rvreply, err := server.node.connect(req)
	*rep = rvreply
	return err
}

func (server *PuddleRPCServer) PwdImpl(req *PwdRequest, rep *PwdReply) error {
	rvreply, err := server.node.pwd(req)
	*rep = rvreply
	return err
}

func (server *PuddleRPCServer) LsImpl(req *LsRequest, rep *LsReply) error {
	rvreply, err := server.node.ls(req)
	*rep = rvreply
	return err
}

func (server *PuddleRPCServer) CdImpl(req *CdRequest, rep *CdReply) error {
	rvreply, err := server.node.cd(req)
	*rep = rvreply
	return err
}

func (server *PuddleRPCServer) MvImpl(req *MvRequest, rep *MvReply) error {
	rvreply, err := server.node.mv(req)
	*rep = rvreply
	return err
}

func (server *PuddleRPCServer) CpImpl(req *MvRequest, rep *MvReply) error {
	rvreply, err := server.node.cp(req)
	*rep = rvreply
	return err
}

func (server *PuddleRPCServer) MkdirImpl(req *MkdirRequest, rep *MkdirReply) error {
	rvreply, err := server.node.mkdir(req)
	*rep = rvreply
	return err
}

func (server *PuddleRPCServer) RmdirImpl(req *RmdirRequest, rep *RmdirReply) error {
	rvreply, err := server.node.rmdir(req)
	*rep = rvreply
	return err
}

func (server *PuddleRPCServer) MkfileImpl(req *MkfileRequest, rep *MkfileReply) error {
	rvreply, err := server.node.mkfile(req)
	*rep = rvreply
	return err
}

func (server *PuddleRPCServer) RmfileImpl(req *RmfileRequest, rep *RmfileReply) error {
	rvreply, err := server.node.rmfile(req)
	*rep = rvreply
	return err
}

func (server *PuddleRPCServer) WritefileImpl(req *WritefileRequest, rep *WritefileReply) error {
	rvreply, err := server.node.writefile(req)
	*rep = rvreply
	return err
}

func (server *PuddleRPCServer) CatImpl(req *CatRequest, rep *CatReply) error {
	rvreply, err := server.node.cat(req)
	*rep = rvreply
	return err
}
