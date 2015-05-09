package puddlestore

import (
	"fmt"
)

const MAX_RETIRES = 5

type Client struct {
	LocalAddr  string
	Id         int
	PuddleServ PuddleAddr
	SeqNum     uint64
}

func CreateClient(remoteAddr PuddleAddr) (cp *Client, err error) {
	fmt.Println("Puddlestore Create client")
	cp = new(Client)

	request := ConnectRequest{}

	ConnectRPC(&remoteAddr, request)

	cp.PuddleServ = remoteAddr

	return
}

func (c *Client) SendRequest(command int, data []byte) (err error) {

	return nil
}

func (c *Client) Ls() (reply *LsReply, err error) {

	request := LsRequest{}

	remoteAddr := c.PuddleServ
	fmt.Println("Puddlestore Ls to addr", remoteAddr)

	reply, err = lsRPC(&remoteAddr, request)

	return reply, nil
}
