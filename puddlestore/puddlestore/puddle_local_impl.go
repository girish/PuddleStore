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

	path := req.Path
	clientId := req.ClientId
	var ok bool

	if path == "" {
		path, ok = puddle.clientPaths[req.ClientId]
		if !ok {
			panic("Did not found the current path of a client that is supposed to be registered")
		}
	}

	// fmt.Printf("Lookingg for %v in clientPaths. Found %v\n", req.ClientId, curdir)

	// First, get the current directory inode
	inode, err := puddle.getInode(path, clientId)
	if err != nil {
		return LsReply{false, ""}, err
	}

	// Second, get the data block from this inode.
	dataBlock, err := puddle.getInodeBlock(path, clientId)
	if err != nil {
		return LsReply{false, ""}, err
	}

	// Then we get the name of all the block inodes
	dirInodes, err := puddle.getBlockInodes(path, inode, dataBlock, clientId)
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
		puddle.clientPaths[req.ClientId] = "/"
		reply.Ok = true
		return reply, nil
	}

	if path[0] != '/' {
		path = puddle.getCurrentDir(clientId) + "/" + path
	}
	path = removeExcessSlashes(path)
	length := len(path)
	fmt.Println("CD path:", path)

	lastSlash := strings.LastIndex(path, "/")

	if length > 1 && path[length-1] == '.' && path[length-2] == '.' {
		if lastSlash == 0 {
			puddle.clientPaths[req.ClientId] = "/"
		} else {
			splits := strings.Split(path, "/")
			if len(splits) <= 3 {
				puddle.clientPaths[req.ClientId] = "/"
			} else {
				path = strings.Join(splits[:len(splits)-2], "/")
				fmt.Println("path:", path)
				puddle.clientPaths[req.ClientId] = path
			}
		}
	} else if path[length-1] == '.' { // Just stay where you are
		if lastSlash == 0 {
			puddle.clientPaths[req.ClientId] = "/"
		} else {
			puddle.clientPaths[req.ClientId] = path[:lastSlash]
		}
	} else {

		fmt.Println(path)
		_, err := puddle.getInode(path, clientId)

		if err != nil { // Path does not exist.
			return CdReply{false}, fmt.Errorf("Path does not exist")
		}

		// Changes the current path of the client
		puddle.clientPaths[req.ClientId] = path
	}

	reply.Ok = true
	return reply, nil
}

func (puddle *PuddleNode) mkdir(req *MkdirRequest) (MkdirReply, error) {
	fmt.Println("Entered mkdir")
	reply := MkdirReply{}

	path := req.Path
	length := len(path)
	clientId := req.ClientId

	if length == 0 {
		return reply, fmt.Errorf("Empty path")
	}
	if (length > 2 && path[length-1] == '.' && path[length-2] == '.') ||
		path[length-1] == '.' {
		return reply, fmt.Errorf("There already exists a file/dir with that name.")
	}

	dirInode, name, fullPath, dirPath, err := puddle.dir_namev(path, clientId)
	if err != nil {
		fmt.Println(err)
		return reply, err
	}

	// File we are about to make should not exist.
	_, err = puddle.getInode(fullPath, clientId)
	if err == nil {
		return reply, fmt.Errorf("There already exists a file/dir with that name.")
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

//mkfile
func (puddle *PuddleNode) mkfile(req *MkfileRequest) (MkfileReply, error) {
	fmt.Println("Entered mkfile")
	reply := MkfileReply{}

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
		dotPath := fullPath
		dotdotPath := dirPath

		// Get hashes
		newDirInodeHash := tapestry.Hash(blockPath)
		//dotHash := tapestry.Hash(dotPath)
		//dotdotHash := tapestry.Hash(dotdotPath)

		fmt.Println("Dirpath: %v", dirPath)
		fmt.Println("Fullpath: %v", fullPath)
		fmt.Println("blockPath: %v", blockPath)
		fmt.Println("dotPath: %v", dotPath)
		fmt.Println("dotdotPath: %v", dotdotPath)

		// Write '.' and '..' to new dir
		//IdIntoByte(newDirBlock.bytes, &dotHash, 0)
		//IdIntoByte(newDirBlock.bytes, &dotdotHash, tapestry.DIGITS)

		// Write the new dir to the old dir and increase its size
		IdIntoByte(dirBlock, &newDirInodeHash, int(dirInode.size))
		dirInode.size += tapestry.DIGITS

		// Save both blocks in tapestry
		puddle.StoreIndirectBlock(fullPath, newDirBlock.bytes, clientId)
		puddle.StoreIndirectBlock(dirPath, newDirBlock.bytes, clientId)

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

func (puddle *PuddleNode) rmdir(req *RmdirRequest) (RmdirReply, error) {
	reply := RmdirReply{}

	path := req.Path
	length := len(path)
	clientId := req.ClientId

	if (length > 2 && path[length-1] == '.' && path[length-2] == '.') ||
		path[length-1] == '.' {
		return reply, fmt.Errorf("Invalid.")
	}

	dirInode, _, fullPath, dirPath, err := puddle.dir_namev(path, clientId)
	if err != nil {
		return reply, fmt.Errorf("Path does not exist")
	}

	rmInode, err := puddle.getInode(fullPath, clientId)
	if err != nil {
		return reply, fmt.Errorf("Directory does not exist")
	}

	if rmInode.size != 0 {
		return reply, fmt.Errorf("Directory is not empty")
	}

	dirBlock, err := puddle.getInodeBlock(dirPath, clientId)
	if err != nil {
		fmt.Println(err)
		return reply, err
	}

	// Get rmInode's vguid
	hash := tapestry.Hash(fullPath)
	aguid := Aguid(hashToGuid(hash))
	vguid, err := puddle.getRaftVguid(aguid, clientId)
	if err != nil {
		return reply, err
	}

	// Get that vguif from the block and zero out the contents
	pointer, err := puddle.lookupInode(dirBlock, vguid, dirInode.size, clientId)
	if err != nil {
		return reply, err
	}
	// MakeZeros(dirBlock, pointer)
	RemoveEntryFromBlock(dirBlock, pointer, dirInode.size)
	dirInode.size -= tapestry.DIGITS

	// Remove anode -> vnode mapping from raft.
	err = puddle.removeRaftVguid(aguid, clientId)
	if err != nil {
		return reply, err
	}

	// Store the modified dir block
	err = puddle.StoreIndirectBlock(dirPath, dirBlock, clientId)
	if err != nil {
		return reply, err
	}

	// Store the modified dir inode
	err = puddle.StoreInode(dirPath, dirInode, clientId)
	if err != nil {
		return reply, err
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

	path = removeExcessSlashes(path)

	if dirPath[0] != '/' {
		dirPath = puddle.getCurrentDir(id) + "/" + dirPath
	}

	dirInode, err := puddle.getInode(dirPath, id)
	if err != nil { // Dir path does not exist
		fmt.Println(err)
		return nil, "", "", "", err
	}

	dirPath = removeExcessSlashes(dirPath)
	fullPath := removeExcessSlashes(dirPath + "/" + name)

	return dirInode, name, fullPath, dirPath, nil
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

func MakeZeros(bytes []byte, start uint32) {
	for i := uint32(0); i < tapestry.DIGITS; i++ {
		bytes[start+i] = 0
	}
}

// Removes an entry from a directory block. If it not the last entry,
// It moves and replaces the last entry with the removing entry.
func RemoveEntryFromBlock(bytes []byte, start uint32, size uint32) {
	if start == size-tapestry.DIGITS { // Last one
		MakeZeros(bytes, start)
	} else {
		for i := uint32(0); i < tapestry.DIGITS; i++ {
			bytes[start+i] = bytes[size-tapestry.DIGITS+i]
		}
	}
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
