package puddlestore

import (
	"bytes"
	"encoding/gob"
)

type Filetype int

const (
	DIR Filetype = iota
	FILE
)

const FILES_PER_INODE = 4

type Inode struct {
	name     string
	filetype Filetype
	size     uint32
	indirect guid
}

func CreateRootInode() *Inode {
	inode := new(Inode)
	inode.name = "root"
	inode.filetype = DIR
	inode.size = 0
	inode.indirect = ""
	return inode
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
