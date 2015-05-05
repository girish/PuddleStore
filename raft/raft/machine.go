package raft

import (
	"crypto/md5"
	"fmt"
	"strings"
)

func (r *RaftNode) processLog(entry LogEntry) ClientReply {
	Out.Printf("%v\n", entry)
	status := OK
	response := ""
	switch entry.Command {
	case HASH_CHAIN_INIT:
		if r.hash == nil {
			r.hash = entry.Data
			response = fmt.Sprintf("%v", r.hash)
		} else {
			status = REQ_FAILED
			response = "The hash chain should only be initialized once!"
		}
	case HASH_CHAIN_ADD:
		if r.hash == nil {
			status = REQ_FAILED
			response = "The hash chain hasn't been initialized yet"
		} else {
			sum := md5.Sum(r.hash)
			fmt.Printf("hash is changing from %v to %v\n", r.hash, sum)
			r.hash = sum[:]
			response = fmt.Sprintf("%v", r.hash)
		}
	// For each of the following idk what to do with the hash chain
	//TODO: Do the byte[] and string casting for entry.Data
	case REMOVE:
		//So by now we have received consensus, we need to delete
		r.requestMutex.Lock()
		key := string(entry.Data)
		delete(r.fileMap, key)
		r.requestMutex.Unlock()
		response = "The key " + key + " has been deleted."
	case SET:
		r.requestMutex.Lock()
		keyVal := string(entry.Data)
		keyValAr := strings.Split(keyVal, ":")
		r.fileMap[keyValAr[0]] = keyValAr[1]
		r.requestMutex.Unlock()
		response = "The key: " + keyValAr[0] + " was set with the value: " + keyValAr[1]

	default:
		response = "Success!"
	}

	reply := ClientReply{
		Status:     status,
		Response:   response,
		LeaderHint: *r.GetLocalAddr(),
	}

	if entry.CacheId != "" {
		r.AddRequest(entry.CacheId, reply)
	}

	r.requestMutex.Lock()
	msg, exists := r.requestMap[entry.Index]
	if exists {
		msg.reply <- reply
		delete(r.requestMap, entry.Index)
	}
	r.requestMutex.Unlock()

	return reply
}
