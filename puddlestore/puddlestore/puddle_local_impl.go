package puddlestore

import (
	"../../raft/raft"
	"../../tapestry/tapestry"
	"fmt"
	"strconv"
	"strings"
)

func (puddle *PuddleNode) connect(req *ConnectRequest) (*ConnectReply, error) {
	reply := ConnectReply{}
	addr := req.FromNode.Addr
	raftNode := puddle.getRandomRaftNode()
	fromAddr := raft.NodeAddr{raft.AddrToId(addr, raftNode.GetConfig().NodeIdSize), addr}

	client, err := raft.CreateClient(fromAddr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Clients that just started the connection should start in root node.
	puddle.clientPaths[addr] = "/"
	puddle.clients[addr] = client

	reply.Ok = true
	return &reply, nil
}

func (puddle *PuddleNode) ls(req *lsRequest) (*lsReply, error) {
	reply := lsReply{}
	elements := make([]string, FILES_PER_INODE)
	numElements := 0

	curdir, ok := puddle.clientPaths[req.FromNode.Addr]
	if !ok {
		panic("Did not found the current path of a client that is supposed to be registered")
	}

	// First, get the current directory inode
	inode, err := puddle.getInode(curdir)
	if err != nil {
		return &lsReply{make([]string, 0), false}, err
	}

	// Empty file dir (debugging)
	if inode.size == 0 {
		reply.elements[0] = "No files, but hey, it got got in."
		reply.Ok = true
		return &reply, nil
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

func (puddle *PuddleNode) cd(req *cdRequest) (*cdReply, error) {
	reply := cdReply{}

	path := req.path

	if len(path) == 0 {
		return nil, fmt.Errorf("Empty path")
	}

	// TODO: Support relative paths.
	if path[0] != '/' {
		panic("not valid path")
	}

	_, err := puddle.getInode(path)

	if err != nil { // Path does not exist.
		return &cdReply{false}, err
	}

	// Changes the current path of the client
	puddle.clientPaths[req.FromNode.Addr] = path

	reply.Ok = true
	return &reply, nil
}

func (puddle *PuddleNode) mkdir(req *mkdirRequest) (*mkdirReply, error) {
	reply := mkdirReply{}

	path := req.path
	addr := req.FromNode.Addr

	if len(path) == 0 {
		return nil, fmt.Errorf("Empty path")
	}

	dirInode, name, fullpath, dirpath, err := puddle.dir_namev(path, addr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	dirBlock, err := puddle.getInodeBlock(dirInode)
	dirBlockPath := fmt.Sprintf("%v:%v", dirpath, "indirect")

	newDirInode := CreateDirInode(name)

	newDirBlock := CreateBlock()

	blockPath := fmt.Sprintf("%v:%v", fullpath, "indirect")
	dotPath := fullpath
	dotdotPath := dirpath

	newDirInodeHash := tapestry.Hash(blockPath)
	dotHash := tapestry.Hash(dotPath)
	dotdotHash := tapestry.Hash(dotdotPath)

	IdIntoByte(newDirBlock.bytes, &dotHash, 0)
	IdIntoByte(newDirBlock.bytes, &dotdotHash, tapestry.DIGITS)

	IdIntoByte(dirBlock, &newDirInodeHash, int(dirInode.size))
	dirInode.size += tapestry.DIGITS

	tapestry.TapestryStore(puddle.getRandomTapestryNode(), blockPath, newDirBlock.bytes)
	tapestry.TapestryStore(puddle.getRandomTapestryNode(), dirBlockPath, newDirBlock.bytes)

	encodedDirNode, err := dirInode.GobEncode()
	encodedNewDirNode, err := newDirInode.GobEncode()

	tapestry.TapestryStore(puddle.getRandomTapestryNode(), dirpath, encodedDirNode)
	tapestry.TapestryStore(puddle.getRandomTapestryNode(), fullpath, encodedNewDirNode)

	reply.Ok = true
	return &reply, nil
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

// Tribute to Weenix's version of dir_namev in kernel/fs/namev.c
func (puddle *PuddleNode) dir_namev(pathname string, addr string) (*Inode, string, string, string, error) {

	path := removeExcessSlashes(pathname)
	lastSlash := strings.LastIndex(path, "/")
	var dirPath, name string

	if lastSlash != -1 && len(path) != 1 { // K. all good
		dirPath = path[:lastSlash]
		name = path[lastSlash+1:]
	} else if lastSlash == -1 { // No slashes at all (relative path probably)
		dirPath = puddle.getCurrentDir(addr)
		name = path
	} else { // This is root
		return puddle.rootInode, "", "/", "", nil
	}

	if dirPath[0] != '/' {
		dirPath = puddle.getCurrentDir(addr) + "/" + dirPath
	}

	dirInode, err := puddle.getInode(dirPath)
	if err != nil { // Dir path does not exist
		fmt.Println(err)
		return nil, "", "", "", err
	}

	return dirInode, name, dirPath + "/" + name, dirPath, nil
}

func removeExcessSlashes(path string) string {
	var firstNonSlash, lastNonSlash, start int

	onlySlashes := true
	str := path

	length := len(path)

	// Nothing to do
	if path[0] != '/' && path[length-1] != '/' {
		return str
	}

	// Get the first non slash
	for i := 0; i < length; i++ {
		if str[i] != '/' {
			onlySlashes = false
			firstNonSlash = i
			break
		}
	}

	// Get the last non slash
	for i := length - 1; i >= 0; i-- {
		if str[i] != '/' {
			lastNonSlash = i
			break
		}
	}

	// Guaranteed to be the root path
	if onlySlashes {
		str = "/"
		return str
	} else {
		length = lastNonSlash - firstNonSlash + 1
		if str[0] == '/' {
			start = firstNonSlash - 1
			length++
		} else {
			start = 0
		}

		str = path[start : start+length]
	}

	length = len(str)
	for i := 0; i < length; i++ {
		if i+1 == length {
			break
		}

		if str[i] == '/' && str[i+1] == '/' {
			str = str[:i] + str[i+1:]
			length -= 1
			i -= 1
		}
	}

	return str
}

func IdIntoByte(bytes []byte, id *tapestry.ID, start int) {
	for i := 0; i < tapestry.DIGITS; i++ {
		bytes[start+i] = byte(id[i])
	}
}
