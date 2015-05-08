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

func (c *Client) Ls(path string) (elements string, err error) {

	request := LsRequest{c.Id, path}

	remoteAddr := c.PuddleServ

	reply, err := lsRPC(&remoteAddr, request)
	if err != nil {
		return
	}
	if !reply.Ok {
		fmt.Errorf("Could not list directory contents.")
	}

	return reply.Elements, nil
}

func (c *Client) Cd(path string) (err error) {
	request := CdRequest{c.Id, path}

	remoteAddr := c.PuddleServ

	reply, err := cdRPC(&remoteAddr, request)

	if err != nil {
		return
	}
	if !reply.Ok {
		fmt.Errorf("Could not change directory")
	}

	return nil
}

func (c *Client) Mkdir(path string) (err error) {
	request := MkdirRequest{c.Id, path}

	remoteAddr := c.PuddleServ

	reply, err := mkdirRPC(&remoteAddr, request)

	if err != nil {
		return
	}
	if !reply.Ok {
		fmt.Errorf("Could not create directory")
	}

	return nil
}

func (c *Client) Rmdir(path string) (err error) {
	request := RmdirRequest{c.Id, path}

	remoteAddr := c.PuddleServ

	reply, err := rmdirRPC(&remoteAddr, request)

	if err != nil {
		return
	}
	if !reply.Ok {
		fmt.Errorf("Could not create directory")
	}

	return nil
}

func (c *Client) Mkfile(path string) (err error) {
	request := MkfileRequest{c.Id, path}

	remoteAddr := c.PuddleServ

	reply, err := mkfileRPC(&remoteAddr, request)

	if err != nil {
		fmt.Println(err)
		return
	}
	if !reply.Ok {
		fmt.Errorf("Could not create file")
	}

	return nil
}
