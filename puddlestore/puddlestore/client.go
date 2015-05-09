package puddlestore

import (
	"fmt"
)

/*
	Client wrapper to avoid clients to call RPCs directly
*/

const MAX_RETRIES = 10

type Client struct {
	LocalAddr  string
	Id         uint64
	PuddleServ PuddleAddr
}

func CreateClient(remoteAddr PuddleAddr) (cp *Client, err error) {
	fmt.Println("Puddlestore Create client")
	cp = new(Client)

	request := ConnectRequest{}
	var reply *ConnectReply

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = ConnectRPC(&remoteAddr, request)
		if err == nil || err.Error() != "EOF" {
			break
		}
		retries++
	}
	if err != nil {
		fmt.Println(err)
		if err.Error() == "EOF" {
			err = fmt.Errorf("Could not access the puddle server.")
		}
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

func (c *Client) Pwd() (path string, err error) {
	request := PwdRequest{c.Id}
	var reply *PwdReply

	remoteAddr := c.PuddleServ

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = pwdRPC(&remoteAddr, request)
		if err == nil || err.Error() != "EOF" {
			break
		}
		retries++
	}
	if err != nil {
		if err.Error() == "EOF" {
			err = fmt.Errorf("Could not access the puddle server.")
		}
		return
	}

	if !reply.Ok {
		fmt.Errorf("Could not get present working directory.")
	}

	return reply.Path, nil
}

func (c *Client) Ls(path string) (elements string, err error) {

	request := LsRequest{c.Id, path}
	var reply *LsReply

	remoteAddr := c.PuddleServ

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = lsRPC(&remoteAddr, request)
		if err == nil || err.Error() != "EOF" {
			break
		}
		retries++
	}
	if err != nil {
		if err.Error() == "EOF" {
			err = fmt.Errorf("Could not access the puddle server.")
		}
		return
	}

	if !reply.Ok {
		fmt.Errorf("Could not list directory contents.")
	}

	return reply.Elements, nil
}

func (c *Client) Cd(path string) (err error) {
	request := CdRequest{c.Id, path}
	var reply *CdReply

	remoteAddr := c.PuddleServ

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = cdRPC(&remoteAddr, request)
		if err == nil || err.Error() != "EOF" {
			break
		}
		retries++
	}
	if err != nil {
		if err.Error() == "EOF" {
			err = fmt.Errorf("Could not access the puddle server.")
		}
		return
	}

	if !reply.Ok {
		fmt.Errorf("Could not change directory")
	}

	return nil
}

func (c *Client) Mkdir(path string) (err error) {
	request := MkdirRequest{c.Id, path}
	var reply *MkdirReply

	remoteAddr := c.PuddleServ

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = mkdirRPC(&remoteAddr, request)
		if err == nil || err.Error() != "EOF" {
			break
		}
		retries++
	}
	if err != nil {
		if err.Error() == "EOF" {
			err = fmt.Errorf("Could not access the puddle server.")
		}
		return
	}

	if !reply.Ok {
		fmt.Errorf("Could not create directory")
	}

	return nil
}

func (c *Client) Rmdir(path string) (err error) {
	request := RmdirRequest{c.Id, path}
	var reply *RmdirReply

	remoteAddr := c.PuddleServ

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = rmdirRPC(&remoteAddr, request)
		if err == nil || err.Error() != "EOF" {
			break
		}
		retries++
	}
	if err != nil {
		if err.Error() == "EOF" {
			err = fmt.Errorf("Could not access the puddle server.")
		}
		return
	}

	if !reply.Ok {
		fmt.Errorf("Could not create directory")
	}

	return nil
}

func (c *Client) Cat(path string, location, count uint32) ([]byte, uint32, error) {
	request := CatRequest{c.Id, path, location, count}
	var reply *CatReply
	var err error

	remoteAddr := c.PuddleServ

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = catRPC(&remoteAddr, request)
		if err == nil || err.Error() != "EOF" {
			break
		}
		retries++
	}
	if err != nil {
		if err.Error() == "EOF" {
			err = fmt.Errorf("Could not access the puddle server.")
		}
		return nil, 0, err
	}

	if !reply.Ok {
		fmt.Errorf("Could not create file")
	}

	return reply.Buffer, reply.Read, nil
}

func (c *Client) Mkfile(path string) (err error) {
	request := MkfileRequest{c.Id, path}
	var reply *MkfileReply

	remoteAddr := c.PuddleServ

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = mkfileRPC(&remoteAddr, request)
		if err == nil || err.Error() != "EOF" {
			break
		}
		retries++
	}
	if err != nil {
		if err.Error() == "EOF" {
			err = fmt.Errorf("Could not access the puddle server.")
		}
		return
	}

	if !reply.Ok {
		fmt.Errorf("Could not create file")
	}

	return nil
}

func (c *Client) Rmfile(path string) (err error) {
	request := RmfileRequest{c.Id, path}
	var reply *RmfileReply

	remoteAddr := c.PuddleServ

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = rmfileRPC(&remoteAddr, request)
		if err == nil || err.Error() != "EOF" {
			break
		}
		retries++
	}
	if err != nil {
		if err.Error() == "EOF" {
			err = fmt.Errorf("Could not access the puddle server.")
		}
		return
	}

	if !reply.Ok {
		fmt.Errorf("Could not create file")
	}

	return nil
}

func (c *Client) Writefile(path string, location uint32, buf []byte) (uint32, error) {
	request := WritefileRequest{c.Id, path, location, buf}
	var reply *WritefileReply
	var err error

	remoteAddr := c.PuddleServ

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = writefileRPC(&remoteAddr, request)
		if err == nil || err.Error() != "EOF" {
			break
		}
		retries++
	}
	if err != nil {
		if err.Error() == "EOF" {
			err = fmt.Errorf("Could not access the puddle server.")
		}
		return 0, err
	}

	if !reply.Ok {
		fmt.Errorf("Could not create file")
	}

	return reply.Written, nil
}
