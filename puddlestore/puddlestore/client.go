package puddlestore

import (
	"fmt"
)

const MAX_RETIRES = 5

type Client struct {
	LocalAddr  string
	Id         int
	PuddleServ *PuddleNode
	SeqNum     uint64
}

func CreateClient(remoteAddr PuddleAddr) (cp *Client, err error) {
	fmt.Println("Puddlestore Create client")
	cp = new(Client)

	request := ConnectRequest{}

	ConnectRPC(&remoteAddr, request)

	return
}

func (c *Client) SendRequest(command int, data []byte) (err error) {

	return nil
}
