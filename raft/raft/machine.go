package raft

import (
	"crypto/md5"
	"fmt"
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
	default:
	}

	reply := ClientReply{
		Status:     status,
		Response:   response,
		LeaderHint: *r.GetLocalAddr(),
	}

	r.requestMutex.Lock()
	msg, exists := r.requestMap[entry.Index]
	if exists {
		msg.reply <- reply
		r.AddRequest(*msg.request, reply)
		delete(r.requestMap, entry.Index)
	}
	r.requestMutex.Unlock()

	return reply
}
