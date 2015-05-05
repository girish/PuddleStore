package puddlestore

import (
	"../../tapestry/tapestry"
	"fmt"
	"strconv"
)

func (puddle *PuddleNode) Connect(req *ConnectRequest) error {
	fmt.Println("me caga este pedo")

	return nil
}

func (puddle *PuddleNode) ls(req *lsRequest) (*lsReply, error) {
	fmt.Println("me caga este pedo")

	reply := lsReply{}
	elements := make([]string, FILES_PER_INODE)
	numElements := 0
	curdir := req.curdir

	// TODO: Support relative paths.
	if curdir[0] != '/' {
		panic("not valid path")
	}

	// First, get the current directory inode
	inode, err := puddle.getInode(curdir)
	if err != nil {
		return &lsReply{make([]string, 0), false}, err
	}

	// Second, get the data block from this inode.
	dataBlock, err := puddle.getInodeBlock(inode)
	if err != nil {
		return &lsReply{make([]string, 0), false}, err
	}

	// Then we get the name of all the block inodes
	dirInodes, err := puddle.getBlockInodes(curdir, dataBlock)
	if err != nil {
		return &lsReply{make([]string, 0), false}, err
	}

	for _, n := range dirInodes {
		elements[numElements] = n.name
		numElements++
	}

	reply.elements = elements
	reply.Ok = true

	return &reply, nil
}

func (puddle *PuddleNode) cd(req *cdRequest) error {
	fmt.Println("me caga este pedo")

	return nil
}

// Gets an inode from a given path
func (puddle *PuddleNode) getInode(path string) (*Inode, error) {
	tapestryNode := puddle.getRandomTapestryNode()
	data, err := tapestry.TapestryGet(tapestryNode, path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	inode := new(Inode)
	err = inode.GobDecode(data)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println("ls: Inode decoded")
	fmt.Println(inode)

	return inode, nil
}

func (puddle *PuddleNode) getInodeBlock(inode *Inode) ([]byte, error) {
	tapestryNode := puddle.getRandomTapestryNode()
	dataBlock, err := tapestry.TapestryGet(tapestryNode, string(inode.indirect))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return dataBlock, nil
}

func (puddle *PuddleNode) getBlockInodes(path string, data []byte) ([]*Inode, error) {
	tapestryNode := puddle.getRandomTapestryNode()
	files := make([]*Inode, FILES_PER_INODE)
	numFiles := 0

	var fileguid uint64
	for i := 0; i < FILES_PER_INODE; i++ {
		fileguid = uint64(data[i])

		// We finished the data block. There is no more.
		if fileguid == 0 {
			return files, nil
		}

		filevguid := fmt.Sprintf("%v:%v", path, strconv.FormatUint(fileguid, 10))
		fileData, err := tapestry.TapestryGet(tapestryNode, filevguid)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		inode := new(Inode)
		inode.GobDecode(fileData)

		files[numFiles] = inode
		numFiles++

	}
	return files, nil
}
