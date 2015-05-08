package puddlestore

import (
	"../../raft/raft"
	"../../tapestry/tapestry"
	"fmt"
	"strconv"
	"strings"
)

func (puddle *PuddleNode) connect(req *ConnectRequest) (ConnectReply, error) {
	reply := ConnectReply{}
	// addr := req.FromNode.Addr
	// raftNode := puddle.getRandomRaftNode()
	// fromAddr := raft.NodeAddr{raft.AddrToId(addr, raftNode.GetConfig().NodeIdSize), addr}

	raftAddr := puddle.getRandomRaftNode().GetLocalAddr()

	client, err := raft.CreateClient(*raftAddr)
	if err != nil {
		fmt.Println(err)
		return ConnectReply{false, 0}, err
	}

	// Clients that just started the connection should start in root node.
	puddle.clientPaths[client.Id] = "/"
	puddle.clients[client.Id] = client

	reply.Ok = true
	reply.Id = client.Id
	fmt.Println("connect reply:", reply)
	return reply, nil
}

func (puddle *PuddleNode) ls(req *LsRequest) (LsReply, error) {
	reply := LsReply{}
	var elements [FILES_PER_INODE + 2]string
	numElements := 2 // Leave 2 spots for '.' and '..'

	curdir, ok := puddle.clientPaths[req.ClientId]
	clientId := req.ClientId
	// fmt.Printf("Lookingg for %v in clientPaths. Found %v\n", req.ClientId, curdir)
	if !ok {
		panic("Did not found the current path of a client that is supposed to be registered")
	}

	// First, get the current directory inode
	inode, err := puddle.getInode(curdir, clientId)
	if err != nil {
		return LsReply{false, ""}, err
	}

	// Empty file dir (debugging)
	/*if inode.size == 0 {
		elements[0] = "No files, but hey, it got got in."
		// reply.elements = makeString(elements)
		reply.Elements = makeString(elements)
		reply.Ok = true
		return reply, nil
	}*/

	// Second, get the data block from this inode.
	dataBlock, err := puddle.getInodeBlock(curdir, clientId)
	if err != nil {
		return LsReply{false, ""}, err
	}

	// Then we get the name of all the block inodes
	dirInodes, err := puddle.getBlockInodes(curdir, inode, dataBlock, clientId)
	if err != nil {
		return LsReply{false, ""}, err
	}

	elements[0] = "."
	elements[1] = ".."
	size := inode.size / tapestry.DIGITS
	fmt.Println("Size in ls:", size)
	for i := uint32(2); i < size+2; i++ {
		elements[numElements] = dirInodes[i-2].name
		numElements++
	}

	reply.Elements = makeString(elements)
	reply.Ok = true

	return reply, nil
}

func (puddle *PuddleNode) cd(req *CdRequest) (CdReply, error) {
	reply := CdReply{}

	path := req.Path
	clientId := req.ClientId

	if len(path) == 0 {
		return CdReply{false}, fmt.Errorf("Empty path")
	}

	// TODO: Support relative paths.
	if path[0] != '/' {
		panic("not valid path")
	}

	_, err := puddle.getInode(path, clientId)

	if err != nil { // Path does not exist.
		return CdReply{false}, err
	}

	// Changes the current path of the client
	puddle.clientPaths[req.ClientId] = path

	reply.Ok = true
	return reply, nil
}

func (puddle *PuddleNode) mkdir(req *MkdirRequest) (MkdirReply, error) {
	fmt.Println("Entered mkdir")
	reply := MkdirReply{}

	path := req.Path
	clientId := req.ClientId

	if len(path) == 0 {
		return reply, fmt.Errorf("Empty path")
	}

	dirInode, name, fullPath, dirPath, err := puddle.dir_namev(path, clientId)
	if err != nil {
		fmt.Println(err)
		return reply, err
	}

	// This is the root node creation.
	if dirInode == nil {

		// Create the root Inode and its block
		newDirInode := CreateDirInode(name)
		newDirBlock := CreateBlock()

		// Set block paths for the indirect block and dot references
		blockPath := fmt.Sprintf("%v:%v", fullPath, "indirect") // this will be '/:indirect'
		// dotPath := fullPath                                     // . -> '/'
		// dotdotPath := dirPath                                   // .. -> '/'

		// Hash the dot references to put them on the indirect block.
		// newDirInodeHash := tapestry.Hash(blockPath)
		blockHash := tapestry.Hash(blockPath)
		// dotHash := tapestry.Hash(dotPath)
		// dotdotHash := tapestry.Hash(dotdotPath)

		// Insert the dot hashes into the new indirect block. TODO: This should be the AGUID, not VGUID, use raft
		// IdIntoByte(newDirBlock.bytes, &dotHash, 0)
		// IdIntoByte(newDirBlock.bytes, &dotdotHash, tapestry.DIGITS)

		// Save the root Inode indirect block in tapestry
		puddle.StoreIndirectBlock(fullPath, newDirBlock.bytes, clientId)
		// tapestry.TapestryStore(puddle.getRandomTapestryNode(), blockPath, newDirBlock.bytes)

		newDirInode.indirect = hashToGuid(blockHash)
		fmt.Println(blockHash, "->", newDirInode.indirect)

		// Save the root Inode
		puddle.StoreInode(fullPath, newDirInode, clientId)
		/*encodedNewDirNode, err := newDirInode.GobEncode()
		if err != nil {
			fmt.Println(err)
			return reply, err
		}
		tapestry.TapestryStore(puddle.getRandomTapestryNode(), fullpath, encodedNewDirNode)
		*/

	} else {
		// Get indirect block from the directory that is going to create
		// the node
		dirBlock, err := puddle.getInodeBlock(dirPath, clientId)
		if err != nil {
			fmt.Println(err)
			return reply, err
		}

		// Create new inode and block
		newDirInode := CreateDirInode(name)
		newDirBlock := CreateBlock()

		// Declare block paths
		// dirBlockPath := fmt.Sprintf("%v:%v", dirPath, "indirect")
		blockPath := fmt.Sprintf("%v:%v", fullPath, "indirect")

		// Get hashes
		// newDirBlockHash := tapestry.Hash(blockPath)
		newDirInodeHash := tapestry.Hash(fullPath)
		//dotHash := tapestry.Hash(dotPath)
		//dotdotHash := tapestry.Hash(dotdotPath)

		fmt.Println("Dirpath: %v", dirPath)
		fmt.Println("Fullpath: %v", fullPath)
		fmt.Println("blockPath: %v", blockPath)
		fmt.Println("newDirInodeHAsh: %v", newDirInodeHash)

		// Write '.' and '..' to new dir
		//IdIntoByte(newDirBlock.bytes, &dotHash, 0)
		//IdIntoByte(newDirBlock.bytes, &dotdotHash, tapestry.DIGITS)

		// Write the new dir to the old dir and increase its size
		IdIntoByte(dirBlock, &newDirInodeHash, int(dirInode.size))
		dirInode.size += tapestry.DIGITS

		bytes := make([]byte, tapestry.DIGITS)
		IdIntoByte(bytes, &newDirInodeHash, 0)
		newDirInode.indirect = Guid(ByteIntoAguid(bytes, 0))
		fmt.Println("\n\n\n\n\n\n", newDirInodeHash, "->", newDirInode.indirect)

		// Save both blocks in tapestry
		puddle.StoreIndirectBlock(fullPath, newDirBlock.bytes, clientId)
		puddle.StoreIndirectBlock(dirPath, dirBlock, clientId)

		// Encode both inodes
		puddle.StoreInode(dirPath, dirInode, clientId)
		puddle.StoreInode(fullPath, newDirInode, clientId)
		/*
			encodedDirNode, err := dirInode.GobEncode()
			if err != nil {
				fmt.Println(err)
				return reply, err
			}
			encodedNewDirNode, err := newDirInode.GobEncode()
			if err != nil {
				fmt.Println(err)
				return reply, err
			}

			// Save both inodes in tapestry
			tapestry.TapestryRemove(puddle.getRandomTapestryNode(), dirpath)
			tapestry.TapestryStore(puddle.getRandomTapestryNode(), dirpath, encodedDirNode)
			tapestry.TapestryStore(puddle.getRandomTapestryNode(), fullpath, encodedNewDirNode)
		*/
	}

	reply.Ok = true
	return reply, nil
}

func (puddle *PuddleNode) getBlockInodes(path string, inode *Inode,
	data []byte, id uint64) ([]*Inode, error) {

	files := make([]*Inode, FILES_PER_INODE)
	numFiles := 0

	size := inode.size / tapestry.DIGITS
	for i := uint32(0); i < size; i++ {

		curAguid := ByteIntoAguid(data, i*tapestry.DIGITS)
		fmt.Println("Found", curAguid, "In directory. Attempting to get inode...")
		curInode, err := puddle.getInodeFromAguid(curAguid, id)
		if err != nil {
			return nil, err
		}
		fmt.Println("Success.")

		files[numFiles] = curInode
		numFiles++

	}
	return files, nil
}

// Tribute to Weenix's version of dir_namev in kernel/fs/namev.c
func (puddle *PuddleNode) dir_namev(pathname string, id uint64) (*Inode, string, string, string, error) {

	path := removeExcessSlashes(pathname)
	lastSlash := strings.LastIndex(path, "/")
	var dirPath, name string

	fmt.Println("Last slash:", lastSlash)

	if lastSlash == 0 && len(path) != 1 {
		return puddle.getRootInode(id), pathname[1:], pathname, "/", nil
	} else if lastSlash == 0 {
		return nil, "/", "/", "", nil
	} else if lastSlash != -1 && len(path) != 1 { // K. all good
		dirPath = path[:lastSlash]
		name = path[lastSlash+1:]
	} else if lastSlash == -1 { // No slashes at all (relative path probably)
		dirPath = puddle.getCurrentDir(id)
		name = path
	} else {
		panic("What should go here?")
	}

	if dirPath[0] != '/' {
		dirPath = puddle.getCurrentDir(id) + "/" + dirPath
	}

	dirInode, err := puddle.getInode(dirPath, id)
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

func ByteIntoAguid(bytes []byte, start uint32) Aguid {
	aguid := ""
	for i := uint32(0); i < tapestry.DIGITS; i++ {
		aguid += strconv.FormatUint(uint64(bytes[start+i]), tapestry.BASE)
	}
	return Aguid(strings.ToUpper(aguid))
}

func makeString(elements [FILES_PER_INODE + 2]string) string {
	ret := ""
	for _, s := range elements {
		if s == "" {
			break
		}
		ret += "\t" + s
	}
	return ret
}

func hashToGuid(id tapestry.ID) Guid {
	s := ""
	for i := 0; i < tapestry.DIGITS; i++ {
		s += strconv.FormatUint(uint64(byte(id[i])), tapestry.BASE)
	}
	return Guid(strings.ToUpper(s))
}
