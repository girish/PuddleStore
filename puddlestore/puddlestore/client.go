package puddlestore

import (
	"fmt"
)

const MAX_RETIRES = 5

type Client struct {
	LocalAddr  string
	Id         uint64
	PuddleServ PuddleAddr
}

func CreateClient(remoteAddr PuddleAddr) (cp *Client, err error) {
	fmt.Println("Puddlestore Create client")
	cp = new(Client)

	request := ConnectRequest{}

	reply, err := ConnectRPC(&remoteAddr, request)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !reply.Ok {
		fmt.Errorf("Could not register Client.")
	}

	fmt.Println("Create client reply:", reply, err)
	cp.Id = reply.Id
	cp.PuddleServ = remoteAddr

	return
}

func (c *Client) SendRequest(command int, data []byte) (err error) {

	return nil
}

func (c *Client) Ls() (elements string, err error) {

	request := LsRequest{c.Id}

	remoteAddr := c.PuddleServ

	reply, err := lsRPC(&remoteAddr, request)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !reply.Ok {
		fmt.Errorf("Could not list directory contents.")
	}

	return reply.Elements, nil
}
