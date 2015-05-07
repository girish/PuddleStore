package puddlestore

import (
	"../../tapestry/tapestry"
	"bytes"
	"encoding/gob"
	"fmt"
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
	indirect guid
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

func (puddle *PuddleNode) StoreInode(key string, inode *Inode) error {
	bytes, err := inode.GobEncode()
	if err != nil {
		return err
	}
	err = tapestry.TapestryStore(puddle.getRandomTapestryNode(), key, bytes)
	if err != nil {
		return err
	}

	return nil
}

// Gets an inode from a given path
func (puddle *PuddleNode) getInode(key string) (*Inode, error) {
	bytes, err := tapestry.TapestryGet(puddle.getRandomTapestryNode(), key)
	if err != nil {
		fmt.Println("ACA")
		return nil, err
	}

	inode := new(Inode)
	err = inode.GobDecode(bytes)
	if err != nil {
		fmt.Println(bytes)
		return nil, err
	}

	return inode, nil
}

func (puddle *PuddleNode) removeKey(key string) error {
	tapestryNodes, err := tapestry.TapestryLookup(puddle.getRandomTapestryNode(), key)
	if err != nil {
		return err
	}
	if len(tapestryNodes) == 0 {
		return fmt.Errorf("Could not find node.")
	}

	success, err := tapestry.TapestryRemove(tapestryNodes[0], key)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("Could not remove node.")
	}
	return nil
}

func (puddle *PuddleNode) getInodeBlock(key string) ([]byte, error) {
	tapestryNode := puddle.getRandomTapestryNode()
	blockPath := fmt.Sprintf("%v:%v", key, "indirect")
	dataBlock, err := tapestry.TapestryGet(tapestryNode, blockPath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return dataBlock, nil
}

func (puddle *PuddleNode) StoreIndirectBlock(key string, block []byte) error {
	err := tapestry.TapestryStore(puddle.getRandomTapestryNode(), key, block)
	if err != nil {
		return err
	}

	return nil
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
