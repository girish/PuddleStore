package puddlestore

import (
	"../../raft/raft"
	"../../tapestry/tapestry"
	"fmt"
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

// Easiest piece of code of my life.
func (puddle *PuddleNode) pwd(req *PwdRequest) (PwdReply, error) {
	reply := PwdReply{true, puddle.getCurrentDir(req.ClientId)}
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

	if path[0] != '/' {
		path = puddle.getCurrentDir(clientId) + "/" + path
	}
	path = removeExcessSlashes(path)

	// fmt.Printf("Lookingg for %v in clientPaths. Found %v\n", req.ClientId, curdir)

	// First, get the current directory inode
	inode, err := puddle.getInode(path, clientId)
	if err != nil {
		return LsReply{false, ""}, err //fmt.Errorf("Path does not exist")
	}
	if inode.filetype != DIR {
		return LsReply{false, ""}, fmt.Errorf("Not a directory")
	}

	// Second, get the data block from this inode.
	dataBlock, err := puddle.getInodeBlock(path, clientId)
	if err != nil {
		return LsReply{false, ""}, err //fmt.Errorf("Directory is currupt")
	}

	// Then we get the name of all the block inodes
	dirInodes, err := puddle.getBlockInodes(path, inode, dataBlock, clientId)
	if err != nil {
		return LsReply{false, ""}, err //fmt.Errorf("Error getting directory contents")
	}

	elements[0] = "./"
	elements[1] = "../"
	size := inode.size / tapestry.DIGITS
	fmt.Println("Size in ls:", size)
	for i := uint32(2); i < size+2; i++ {
		elements[numElements] = dirInodes[i-2].name
		if dirInodes[i-2].filetype == DIR {
			elements[numElements] += "/"
		}
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
		cdInode, err := puddle.getInode(path, clientId)

		if err != nil { // Path does not exist.
			return CdReply{false}, fmt.Errorf("Path does not exist.")
		}
		if cdInode.filetype != DIR {
			return CdReply{false}, fmt.Errorf("Cannot cd into a file.")
		}

		// Changes the current path of the client
		puddle.clientPaths[req.ClientId] = path
	}

	reply.Ok = true
	return reply, nil
}

func (puddle *PuddleNode) mv(req *MvRequest) (MvReply, error) {
	reply := MvReply{}

	source := req.Source
	dest := req.Dest
	clientId := req.ClientId

	// Format paths before starting
	if source[0] != '/' {
		source = puddle.getCurrentDir(clientId) + "/" + source
	}
	source = removeExcessSlashes(source)

	if dest[0] != '/' {
		dest = puddle.getCurrentDir(clientId) + "/" + dest
	}
	dest = removeExcessSlashes(dest)

	// Same exact path. Just return.
	if dest == source {
		reply.Ok = true
		return reply, nil
	}

	// Get the name of the new inode, dir path, and dir inode of source and dest
	sourceDirInode, _, _, sourceDirPath, err := puddle.dir_namev(source, clientId)
	if err != nil {
		return reply, fmt.Errorf("Path does not exist")
	}
	destDirInode, newName, _, destDirPath, err := puddle.dir_namev(dest, clientId)
	if err != nil {
		return reply, fmt.Errorf("Path does not exist")
	}

	// Check that the source is not a directory (not supported)
	sourceInode, err := puddle.getInode(source, clientId)
	if err != nil {
		return reply, fmt.Errorf("Source file does not exist")
	}
	if sourceInode.filetype != FILE {
		return reply, fmt.Errorf("Cannot move something that is not a file.")
	}

	// Check if they are in the same directory (rename)
	if destDirPath == sourceDirPath {
		sourceInode.name = newName
		// store the modified source dir inode
		err = puddle.storeInode(source, sourceInode, clientId)
		if err != nil {
			return reply, err
		}
		return reply, fmt.Errorf("This is a rename")
	}

	if destDirInode.size == tapestry.DIGITS*FILES_PER_INODE {
		return reply, fmt.Errorf("Not enough space in destination directory")
	}

	_, err = puddle.getInode(dest, clientId)
	if err == nil {
		return reply, fmt.Errorf("File with that name exists.")
	}

	// Get sourceDir, source and dest vguids
	// sourceDirHash := tapestry.Hash(sourceDirPath)
	sourceHash := tapestry.Hash(source)
	// destDirHash := tapestry.Hash(destDirPath)
	destHash := tapestry.Hash(dest)

	// sourceDirAguid := Aguid(hashToGuid(sourceDirHash))
	sourceAguid := Aguid(hashToGuid(sourceHash))
	//destDirAguid := Aguid(hashToGuid(destDirHash))
	destAguid := Aguid(hashToGuid(destHash))

	// Get the vguid where all of the info is at
	res, err := puddle.getRaftVguid(sourceAguid, clientId)
	sourceVguid := Vguid(strings.Split(string(res), ":")[1])
	if err != nil {
		return reply, err
	}

	sourceInode.name = newName

	// Get source and dest dir blocks
	sourceDirBlock, err := puddle.getInodeBlock(sourceDirPath, clientId)
	if err != nil {
		return reply, err
	}
	destDirBlock, err := puddle.getInodeBlock(destDirPath, clientId)
	if err != nil {
		return reply, err
	}

	// Get that vguid from the sourceDirBlock and remove
	err = puddle.removeEntryFromBlock(sourceDirBlock, sourceVguid,
		sourceDirInode.size, clientId)
	if err != nil {
		return reply, err
	}
	sourceDirInode.size -= tapestry.DIGITS

	// Map the new path to the same vguid the source pointed to.
	err = puddle.setRaftVguid(destAguid, sourceVguid, clientId)
	if err != nil {
		return reply, err
	}
	// Remove the new old mapping of source
	err = puddle.removeRaftVguid(sourceAguid, clientId)
	if err != nil {
		return reply, err
	}

	// Add the new entry to dest dir
	IdIntoByte(destDirBlock, &destHash, int(destDirInode.size))
	destDirInode.size += tapestry.DIGITS

	// store the modified source dir block
	err = puddle.storeIndirectBlock(sourceDirPath, sourceDirBlock, clientId)
	if err != nil {
		return reply, err
	}
	// store the modified source dir block
	err = puddle.storeIndirectBlock(destDirPath, destDirBlock, clientId)
	if err != nil {
		return reply, err
	}

	// store the modified source dir inode
	err = puddle.storeInode(sourceDirPath, sourceDirInode, clientId)
	if err != nil {
		return reply, err
	}
	// store the modified dest dir inode
	err = puddle.storeInode(destDirPath, destDirInode, clientId)
	if err != nil {
		return reply, err
	}

	err = puddle.storeInode(dest, sourceInode, clientId)
	if err != nil {
		return reply, err
	}

	reply.Ok = true
	return reply, nil
}

func (puddle *PuddleNode) cp(req *MvRequest) (MvReply, error) {
	reply := MvReply{}

	source := req.Source
	dest := req.Dest
	clientId := req.ClientId

	// Format paths before starting
	if source[0] != '/' {
		source = puddle.getCurrentDir(clientId) + "/" + source
	}
	source = removeExcessSlashes(source)

	if dest[0] != '/' {
		dest = puddle.getCurrentDir(clientId) + "/" + dest
	}
	dest = removeExcessSlashes(dest)

	// Same exact path. Just return.
	if dest == source {
		reply.Ok = true
		return reply, nil
	}

	// Get the name of the new inode, dir path, and dir inode of source and dest
	destDirInode, newName, _, destDirPath, err := puddle.dir_namev(dest, clientId)
	if err != nil {
		return reply, fmt.Errorf("Path does not existt")
	}

	// Check that the source is not a directory (not supported)
	sourceInode, err := puddle.getInode(source, clientId)
	if err != nil {
		return reply, fmt.Errorf("Source file does not exist")
	}
	if sourceInode.filetype != FILE {
		return reply, fmt.Errorf("Cannot move something that is not a file.")
	}

	if destDirInode.size == tapestry.DIGITS*FILES_PER_INODE {
		return reply, fmt.Errorf("Not enough space in destination directory")
	}

	_, err = puddle.getInode(dest, clientId)
	if err == nil {
		return reply, fmt.Errorf("File with that name exists.")
	}

	// Get sourceDir, source and dest vguids
	// sourceDirHash := tapestry.Hash(sourceDirPath)
	//sourceHash := tapestry.Hash(source)
	// destDirHash := tapestry.Hash(destDirPath)
	destHash := tapestry.Hash(dest)

	// sourceDirAguid := Aguid(hashToGuid(sourceDirHash))
	//sourceAguid := Aguid(hashToGuid(sourceHash))
	//destDirAguid := Aguid(hashToGuid(destDirHash))
	//destAguid := Aguid(hashToGuid(destHash))

	// Get the vguid where all of the info is at
	// res, err := puddle.getRaftVguid(sourceAguid, clientId)
	// sourceVguid := Vguid(strings.Split(string(res), ":")[1])
	if err != nil {
		return reply, err
	}

	sourceInode.name = newName

	fmt.Println("\n\n\n1111111111\n\n\n\n")

	// Get source and dest dir blocks
	// sourceDirBlock, err := puddle.getInodeBlock(sourceDirPath, clientId)
	if err != nil {
		return reply, err
	}
	sourceBlock, err := puddle.getInodeBlock(source, clientId)
	if err != nil {
		return reply, err
	}
	destDirBlock, err := puddle.getInodeBlock(destDirPath, clientId)
	if err != nil {
		return reply, err
	}

	// Get that vguid from the sourceDirBlock and remove
	/*err = puddle.removeEntryFromBlock(sourceDirBlock, sourceVguid,
		sourceDirInode.size, clientId)
	if err != nil {
		return reply, err
	}
	sourceDirInode.size -= tapestry.DIGITS
	*/

	// Map the new path to the same vguid the source pointed to.
	// err = puddle.setRaftVguid(destAguid, sourceVguid, clientId)
	// if err != nil {
	//	return reply, err
	//}

	/*
		// Remove the new old mapping of source
		err = puddle.removeRaftVguid(sourceAguid, clientId)
		if err != nil {
			return reply, err
		}*/

	// Add the new entry to dest dir
	IdIntoByte(destDirBlock, &destHash, int(destDirInode.size))
	destDirInode.size += tapestry.DIGITS

	// store the modified source dir block
	//err = puddle.storeIndirectBlock(sourceDirPath, sourceDirBlock, clientId)
	//if err != nil {
	//	return reply, err
	//}

	fmt.Println("\n\n\n222222222222\n\n\n\n")
	err = puddle.storeIndirectBlock(dest, sourceBlock, clientId)
	if err != nil {
		return reply, err
	}
	err = puddle.storeIndirectBlock(destDirPath, destDirBlock, clientId)
	if err != nil {
		return reply, err
	}

	// store the modified source dir inode
	// err = puddle.storeInode(sourceDirPath, sourceDirInode, clientId)
	if err != nil {
		return reply, err
	}
	fmt.Println("\n\n\n33333333333333\n\n\n\n")
	// store the modified dest dir inode
	err = puddle.storeInode(destDirPath, destDirInode, clientId)
	if err != nil {
		return reply, err
	}

	err = puddle.storeInode(dest, sourceInode, clientId)
	if err != nil {
		return reply, err
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

	if path[0] != '/' {
		path = puddle.getCurrentDir(clientId) + "/" + path
	}
	path = removeExcessSlashes(path)

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

		// Hash the dot references to put them on the indirect block.
		blockHash := tapestry.Hash(blockPath)

		// Save the root Inode indirect block in tapestry
		puddle.storeIndirectBlock(fullPath, newDirBlock.bytes, clientId)

		newDirInode.indirect = hashToGuid(blockHash)
		fmt.Println(blockHash, "->", newDirInode.indirect)

		// Save the root Inode
		puddle.storeInode(fullPath, newDirInode, clientId)

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
		blockPath := fmt.Sprintf("%v:%v", fullPath, "indirect")

		// Get hashes
		newDirInodeHash := tapestry.Hash(fullPath)

		fmt.Println("Dirpath: %v", dirPath)
		fmt.Println("Fullpath: %v", fullPath)
		fmt.Println("blockPath: %v", blockPath)
		fmt.Println("newDirInodeHAsh: %v", newDirInodeHash)

		// Write the new dir to the old dir and increase its size
		IdIntoByte(dirBlock, &newDirInodeHash, int(dirInode.size))
		dirInode.size += tapestry.DIGITS

		bytes := make([]byte, tapestry.DIGITS)
		IdIntoByte(bytes, &newDirInodeHash, 0)
		newDirInode.indirect = Guid(ByteIntoAguid(bytes, 0))
		fmt.Println("\n\n\n\n\n\n", newDirInodeHash, "->", newDirInode.indirect)

		// Save both blocks in tapestry
		puddle.storeIndirectBlock(fullPath, newDirBlock.bytes, clientId)
		puddle.storeIndirectBlock(dirPath, dirBlock, clientId)

		// Encode both inodes
		puddle.storeInode(dirPath, dirInode, clientId)
		puddle.storeInode(fullPath, newDirInode, clientId)
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

	if path[0] != '/' {
		path = puddle.getCurrentDir(clientId) + "/" + path
	}
	path = removeExcessSlashes(path)

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
		panic("Root file does not exist")
	} else {
		// Get indirect block from the directory that is going to create
		// the node
		dirBlock, err := puddle.getInodeBlock(dirPath, clientId)
		if err != nil {
			fmt.Println(err)
			return reply, err
		}

		// Create new inode and block
		newFileInode := CreateFileInode(name)
		newFileBlock := CreateBlock()

		// Get hashes
		newFileInodeHash := tapestry.Hash(fullPath)

		// Write the new dir to the old dir and increase its size
		IdIntoByte(dirBlock, &newFileInodeHash, int(dirInode.size))
		dirInode.size += tapestry.DIGITS

		bytes := make([]byte, tapestry.DIGITS)
		IdIntoByte(bytes, &newFileInodeHash, 0)
		newFileInode.indirect = Guid(ByteIntoAguid(bytes, 0))
		fmt.Println("\n\n\n\n\n\n", newFileInodeHash, "->", newFileInode.indirect)

		// Save both blocks in tapestry
		puddle.storeIndirectBlock(fullPath, newFileBlock.bytes, clientId)
		puddle.storeIndirectBlock(dirPath, dirBlock, clientId)

		// Encode both inodes
		puddle.storeInode(dirPath, dirInode, clientId)
		puddle.storeInode(fullPath, newFileInode, clientId)

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

	if rmInode.filetype != DIR {
		return reply, fmt.Errorf("Cannot remove something that is not a dir.")
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
	res, err := puddle.getRaftVguid(aguid, clientId)
	vguid := Vguid(strings.Split(string(res), ":")[1])
	if err != nil {
		return reply, err
	}

	// Get that vguid from the block and zero out the contents
	err = puddle.removeEntryFromBlock(dirBlock, vguid, dirInode.size, clientId)
	if err != nil {
		return reply, err
	}
	dirInode.size -= tapestry.DIGITS

	// Remove anode -> vnode mapping from raft.
	err = puddle.removeRaftVguid(aguid, clientId)
	if err != nil {
		return reply, err
	}

	// store the modified dir block
	err = puddle.storeIndirectBlock(dirPath, dirBlock, clientId)
	if err != nil {
		return reply, err
	}

	// store the modified dir inode
	err = puddle.storeInode(dirPath, dirInode, clientId)
	if err != nil {
		return reply, err
	}

	reply.Ok = true
	return reply, nil
}

func (puddle *PuddleNode) rmfile(req *RmfileRequest) (RmfileReply, error) {
	reply := RmfileReply{}

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
		return reply, fmt.Errorf("File does not exist")
	}

	if rmInode.filetype != FILE {
		return reply, fmt.Errorf("Cannot remove something that is not a file.")
	}

	dirBlock, err := puddle.getInodeBlock(dirPath, clientId)
	if err != nil {
		fmt.Println(err)
		return reply, err
	}

	// Get rmInode's vguid
	hash := tapestry.Hash(fullPath)
	aguid := Aguid(hashToGuid(hash))
	res, err := puddle.getRaftVguid(aguid, clientId)
	vguid := Vguid(strings.Split(string(res), ":")[1])
	if err != nil {
		return reply, err
	}

	// Get that vguif from the block and zero out the contents
	err = puddle.removeEntryFromBlock(dirBlock, vguid, dirInode.size, clientId)
	if err != nil {
		return reply, err
	}
	dirInode.size -= tapestry.DIGITS

	// Remove anode -> vnode mapping from raft.
	err = puddle.removeRaftVguid(aguid, clientId)
	if err != nil {
		return reply, err
	}

	// store the modified dir block
	err = puddle.storeIndirectBlock(dirPath, dirBlock, clientId)
	if err != nil {
		return reply, err
	}

	// store the modified dir inode
	err = puddle.storeInode(dirPath, dirInode, clientId)
	if err != nil {
		return reply, err
	}

	reply.Ok = true
	return reply, nil
}

func (puddle *PuddleNode) writefile(req *WritefileRequest) (WritefileReply, error) {
	reply := WritefileReply{}

	path := req.Path
	// length := len(path)
	buf := req.Buffer
	location := req.Location
	count := uint32(len(buf))
	clientId := req.ClientId

	if path[0] != '/' {
		path = puddle.getCurrentDir(clientId) + "/" + path
	}
	path = removeExcessSlashes(path)

	inode, err := puddle.getInode(path, clientId)
	if err != nil {
		return reply, fmt.Errorf("File does not exist")
	}
	if inode.filetype != FILE {
		return reply, fmt.Errorf("File is a directory")
	}

	blockNo := location / BLOCK_SIZE
	blockOffset := location % BLOCK_SIZE
	block, err := puddle.getFileBlock(path, blockNo, clientId)
	if err != nil {
		if err.Error() == "Tapestry error" {
			return reply, err
		}
		block = make([]byte, BLOCK_SIZE)
	}

	var i uint32
	for i = 0; i < count; i++ {
		// Reached limit of a certain block
		if blockOffset == BLOCK_SIZE {
			// Save previous block first, then change to next one
			err = puddle.storeFileBlock(path, blockNo, block, clientId)
			if err != nil {
				return reply, fmt.Errorf("File does not exist")
			}
			blockNo++
			if blockNo == FILES_PER_INODE {
				break
			}
			block, err = puddle.getFileBlock(path, blockNo, clientId)
			if err != nil {
				if err.Error() == "Tapestry error" {
					return reply, err
				}
				block = make([]byte, BLOCK_SIZE)
			}
			blockOffset = 0
		}

		block[blockOffset] = buf[i]
		blockOffset++
	}

	if i == count {
		err = puddle.storeFileBlock(path, blockNo, block, clientId)
		if err != nil {
			return reply, fmt.Errorf("File does not exist")
		}
	}

	reply.Ok = true
	reply.Written = i
	return reply, nil
}

func (puddle *PuddleNode) cat(req *CatRequest) (CatReply, error) {
	reply := CatReply{}

	path := req.Path
	// length := len(path)
	buf := make([]byte, BLOCK_SIZE*FILES_PER_INODE) // Set buf to max len possible
	location := req.Location
	count := req.Count
	clientId := req.ClientId

	if path[0] != '/' {
		path = puddle.getCurrentDir(clientId) + "/" + path
	}
	path = removeExcessSlashes(path)

	inode, err := puddle.getInode(path, clientId)
	if err != nil {
		return reply, fmt.Errorf("File does not exist")
	}
	if inode.filetype != FILE {
		return reply, fmt.Errorf("File is a directory")
	}

	blockNo := location / BLOCK_SIZE
	blockOffset := location % BLOCK_SIZE
	fmt.Println(location, " : ", BLOCK_SIZE)
	fmt.Println(blockNo, " : ", blockOffset)
	block, err := puddle.getFileBlock(path, blockNo, clientId)
	if err != nil {
		if err.Error() == "Tapestry error" {
			return reply, err
		}
		block = make([]byte, BLOCK_SIZE)
	}

	var i uint32
	for i = 0; i < count; i++ {
		// Reached limit of a certain block
		if blockOffset == BLOCK_SIZE {
			// Save previous block first, then change to next one
			// err = puddle.storeFileBlock(path, blockNo, block, clientId)
			//if err != nil {
			//	return reply, fmt.Errorf("File does not exist")
			//}
			fmt.Printf("Changing blocks %v -> %v at count %v, %v\n", blockNo, blockNo+1, i,
				blockOffset)
			blockNo++
			if blockNo == FILES_PER_INODE {
				break
			}
			block, err = puddle.getFileBlock(path, blockNo, clientId)
			if err != nil {
				if err.Error() == "Tapestry error" {
					return reply, err
				}
				block = make([]byte, BLOCK_SIZE)
			}
			blockOffset = 0
		}

		buf[i] = block[blockOffset]
		blockOffset++
	}

	if i == count {
		err = puddle.storeFileBlock(path, blockNo, block, clientId)
		if err != nil {
			return reply, fmt.Errorf("File does not exist")
		}
	}

	reply.Ok = true
	reply.Read = i
	reply.Buffer = buf
	return reply, nil
}

// Helper method used in 'ls'
func (puddle *PuddleNode) getBlockInodes(path string, inode *Inode,
	data []byte, id uint64) ([]*Inode, error) {

	files := make([]*Inode, FILES_PER_INODE)
	numFiles := 0

	size := inode.size / tapestry.DIGITS
	for i := uint32(0); i < size; i++ {

		curAguid := ByteIntoAguid(data, i*tapestry.DIGITS)
		fmt.Println("Found", curAguid, "In directory. Attempting to get inode...")
		fmt.Println(curAguid)
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
