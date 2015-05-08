package puddlestore

import (
	"../../raft/raft"
	"../../tapestry/tapestry"
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"
)

type Filetype int

const (
	DIR Filetype = iota
	FILE
)

const BLOCK_SIZE = 4096
const FILES_PER_INODE = 4

type Inode struct {
	name     string
	filetype Filetype
	size     uint32
	indirect Guid
}

type Block struct {
	bytes []byte
}

func CreateRootInode() *Inode {
	inode := new(Inode)
	inode.name = "root"
	inode.filetype = DIR
	inode.size = 0
	inode.indirect = ""
	return inode
}

func CreateDirInode(name string) *Inode {
	inode := new(Inode)
	inode.name = name
	inode.filetype = DIR
	//	inode.size = tapestry.DIGITS * 2 // for '.' and '..'
	inode.size = 0
	inode.indirect = ""
	return inode
}

func CreateBlock() *Block {
	block := new(Block)
	block.bytes = make([]byte, BLOCK_SIZE)
	return block
}

func (d *Inode) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(d.name)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.filetype)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.size)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.indirect)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (d *Inode) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&d.name)
	if err != nil {
		return err
	}
	err = decoder.Decode(&d.filetype)
	if err != nil {
		return err
	}
	err = decoder.Decode(&d.size)
	if err != nil {
		return err
	}
	return decoder.Decode(&d.indirect)
}

// Stores inode as data
func (puddle *PuddleNode) StoreInode(path string, inode *Inode, id uint64) error {

	hash := tapestry.Hash(path)

	aguid := Aguid(hashToGuid(hash))
	vguid := Vguid(randSeq(tapestry.DIGITS))

	// Encode the inode
	bytes, err := inode.GobEncode()
	if err != nil {
		return err
	}

	// Set the new aguid -> vguid pair with raft
	err = puddle.setRaftVguid(aguid, vguid, id)
	if err != nil {
		return err
	}

	// Store data in tapestry with key: vguid
	err = tapestry.TapestryStore(puddle.getRandomTapestryNode(), string(vguid), bytes)
	if err != nil {
		return err
	}

	return nil
}

// Gets an inode from a given path
func (puddle *PuddleNode) getInode(path string, id uint64) (*Inode, error) {

	hash := tapestry.Hash(path)

	aguid := Aguid(hashToGuid(hash))

	// Get the vguid using raft
	bytes, err := puddle.getTapestryData(aguid, id)

	inode := new(Inode)
	err = inode.GobDecode(bytes)
	if err != nil {
		fmt.Println(bytes)
		return nil, err
	}

	return inode, nil
}

func (puddle *PuddleNode) getInodeFromAguid(aguid Aguid, id uint64) (*Inode, error) {
	// Get the vguid using raft
	bytes, err := puddle.getTapestryData(aguid, id)

	inode := new(Inode)
	err = inode.GobDecode(bytes)
	if err != nil {
		fmt.Println(bytes)
		return nil, err
	}

	return inode, nil
}

func (puddle *PuddleNode) getInodeBlock(key string, id uint64) ([]byte, error) {
	blockPath := fmt.Sprintf("%v:%v", key, "indirect")
	hash := tapestry.Hash(blockPath)
	aguid := Aguid(hashToGuid(hash))

	return puddle.getTapestryData(aguid, id)
}

func (puddle *PuddleNode) getTapestryData(aguid Aguid, id uint64) ([]byte, error) {
	tapestryNode := puddle.getRandomTapestryNode()
	response, err := puddle.getRaftVguid(aguid, id)
	if err != nil {
		return nil, err
	}

	ok := strings.Split(string(response), ":")[0]
	vguid := strings.Split(string(response), ":")[1]
	if ok != "SUCCESS" {
		return nil, fmt.Errorf("Could not get raft vguid: %v", response)
	}

	data, err := tapestry.TapestryGet(tapestryNode, string(vguid))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (puddle *PuddleNode) StoreIndirectBlock(inodePath string, block []byte,
	id uint64) error {

	blockPath := fmt.Sprintf("%v:%v", inodePath, "indirect")
	hash := tapestry.Hash(blockPath)

	aguid := Aguid(hashToGuid(hash))
	vguid := Vguid(randSeq(tapestry.DIGITS))

	// Set the new aguid -> vguid pair with raft
	err := puddle.setRaftVguid(aguid, vguid, id)
	if err != nil {
		return err
	}

	err = tapestry.TapestryStore(puddle.getRandomTapestryNode(), string(vguid), block)
	if err != nil {
		return err
	}

	return nil
}

func (puddle *PuddleNode) setRaftVguid(aguid Aguid, vguid Vguid, id uint64) error {
	// Get the raft client struct
	c, ok := puddle.clients[id]
	if !ok {
		panic("Attempted to get client from id, but not found.")
	}

	data := fmt.Sprintf("%v:%v", aguid, vguid)

	res, err := c.SendRequestWithResponse(raft.SET, []byte(data))
	if err != nil {
		return err
	}
	if res.Status != raft.OK {
		return fmt.Errorf("Could not get response from raft.")
	}

	return nil
}

func (puddle *PuddleNode) getRaftVguid(aguid Aguid, id uint64) (Vguid, error) {
	// Get the raft client struct
	c, ok := puddle.clients[id]
	if !ok {
		panic("Attempted to get client from id, but not found.")
	}

	res, err := c.SendRequestWithResponse(raft.GET, []byte(aguid))
	if err != nil {
		return "", err
	}
	if res.Status != raft.OK {
		return "", fmt.Errorf("Could not get response from raft.")
	}

	return Vguid(res.Response), nil
}

func (puddle *PuddleNode) removeRaftVguid(aguid Aguid, id uint64) error {
	// Get the raft client struct
	c, ok := puddle.clients[id]
	if !ok {
		panic("Attempted to get client from id, but not found.")
	}

	res, err := c.SendRequestWithResponse(raft.REMOVE, []byte(aguid))
	if err != nil {
		return err
	}
	if res.Status != raft.OK {
		return fmt.Errorf("Could not get response from raft.")
	}

	return nil
}
