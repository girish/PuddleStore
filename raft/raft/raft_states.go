package raft

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

type state func() state

/**
 * This method contains the logic of a Raft node in the follower state.
 */
func (r *RaftNode) doFollower() state {
	r.Out("Transitioning to FOLLOWER_STATE")
	r.State = FOLLOWER_STATE
	electionTimeout := makeElectionTimeout()
	// TODO: Students should implement this method
	for {
		select {
		case shutdown := <-r.gracefulExit:
			if shutdown {
				return nil
			}

		case vote := <-r.requestVote:
			req := vote.request
			rep := vote.reply
			currTerm := r.GetCurrentTerm()
			candidate := req.CandidateId
			votedFor := r.GetVotedFor()

			if votedFor == "" || votedFor == candidate.Id {
				// TODO: Check log to see si el vato la arma.
				if r.handleCompetingRequestVote(vote) {
					// return r.doFollower
					r.setCurrentTerm(req.Term)
					// Set voted for field (already voted)
					r.setVotedFor(candidate.Id)

					// Election in progess. There is no leader
					// Set to nil
					r.LeaderAddr = nil

					// Respond true, reset eleciton timeout.
					// rep <- RequestVoteReply{currTerm, true}
					electionTimeout = makeElectionTimeout()
				} else {
					// Other candidate has a less up to date log.
				}
			} else {
				// Already voted.
				rep <- RequestVoteReply{currTerm, false}
			}

		case appendEnt := <-r.appendEntries:
			req := appendEnt.request
			rep := appendEnt.reply
			currTerm := r.GetCurrentTerm()

			if req.Term < currTerm {
				rep <- AppendEntriesReply{currTerm, false}
			} else {
				r.LeaderAddr = &req.LeaderId
				r.setCurrentTerm(req.Term)

				entry := r.getLogEntry(req.PrevLogIndex)
				if entry == nil || entry.TermId != req.PrevLogTerm {
					rep <- AppendEntriesReply{currTerm, false}
				} else {
					if req.PrevLogIndex == r.getLastLogIndex() {
						r.truncateLog(req.PrevLogIndex + 1)
					}

					if len(req.Entries) > 0 {
						for i := range req.Entries {
							r.appendLogEntry(req.Entries[i])
						}
					}

					if req.LeaderCommit > r.commitIndex {
						r.commitIndex = uint64(math.Min(float64(req.LeaderCommit), float64(r.getLastLogIndex())))
					}

					// Calls process log from lastApplied until commit index.
					if r.lastApplied < r.commitIndex {
						i := r.lastApplied
						if r.lastApplied != 0 {
							i++
						}
						for ; i <= r.commitIndex; i++ {
							reply := r.processLog(*r.getLogEntry(i))
							if reply.Status == REQ_FAILED {
								panic("This should not happen ever! PLZ FIX")
							}
						}
						r.lastApplied = r.commitIndex
					}

					rep <- AppendEntriesReply{currTerm, true}
				}
			}
			electionTimeout = makeElectionTimeout()
		case regClient := <-r.registerClient:
			// req := regClient.request
			rep := regClient.reply
			if r.LeaderAddr != nil {
				rep <- RegisterClientReply{NOT_LEADER, 0, *r.LeaderAddr}
			} else {
				// Return election in progress.
				rep <- RegisterClientReply{ELECTION_IN_PROGRESS, 0, NodeAddr{"", ""}}
			}
		case <-electionTimeout:
			return r.doCandidate
			/*
				case client := <-r.clientRequest:
					return nil
				case register := <-r.registerClient:
					return nil
			*/
		}
	}
}

/**
 * This method contains the logic of a Raft node in the candidate state.
 */
func (r *RaftNode) doCandidate() state {
	r.Out("Transitioning to CANDIDATE_STATE")

	r.setCurrentTerm(r.GetCurrentTerm() + 1)
	r.State = CANDIDATE_STATE

	electionResults := make(chan bool)
	electionTimeout := makeElectionTimeout()
	r.requestVotes(electionResults)

	// TODO: Students should implement this method
	for {
		select {
		case shutdown := <-r.gracefulExit:
			if shutdown {
				return nil
			}

		case election := <-electionResults:
			if election {
				return r.doLeader
			} else {
				return r.doFollower
			}

		case vote := <-r.requestVote:
			req := vote.request
			candidate := req.CandidateId
			if r.handleCompetingRequestVote(vote) {
				r.setCurrentTerm(req.Term)
				// Set voted for field (already voted)
				r.setVotedFor(candidate.Id)

				// Election in progess. There is no leader
				// Set to nil
				r.LeaderAddr = nil

				// Respond true, reset eleciton timeout.
				// rep <- RequestVoteReply{currTerm, true}
				electionTimeout = makeElectionTimeout()
				return r.doFollower
			} else {
				// Other candidate has a less up to date log.
			}

		case appendEnt := <-r.appendEntries:
			req := appendEnt.request
			rep := appendEnt.reply
			currTerm := r.GetCurrentTerm()
			leader := req.LeaderId

			/*if req.Term < currTerm {
				rep <- AppendEntriesReply{currTerm, false}
			} else {*/
			r.LeaderAddr = &leader
			r.setCurrentTerm(req.Term)
			r.setVotedFor("")
			rep <- AppendEntriesReply{currTerm, true}
			return r.doFollower
			//}
			electionTimeout = makeElectionTimeout()

		case regClient := <-r.registerClient:
			// req := regClient.request
			rep := regClient.reply
			rep <- RegisterClientReply{ELECTION_IN_PROGRESS, 0, NodeAddr{"", ""}}

		case <-electionTimeout:
			return r.doCandidate
		}
	}
	return nil
}

/**
 * This method contains the logic of a Raft node in the leader state.
 */
func (r *RaftNode) doLeader() state {
	r.Out("Transitioning to LEADER_STATE")
	r.State = LEADER_STATE
	r.LeaderAddr = r.GetLocalAddr()
	beats := r.makeBeats()

	// Set up all next index values
	for _, n := range r.GetOtherNodes() {
		r.nextIndex[n.Id] = r.getLastLogIndex() + 1
	}

	r.sendNoop()

	for {
		select {
		case <-beats:
			r.Debug("sendHeartBeats: entered\n")
			f, _ := r.sendHeartBeats()
			if f {
				r.sendRequestFail()
				return r.doFollower
			}
			beats = r.makeBeats()

		case appendEnt := <-r.appendEntries:
			req := appendEnt.request
			rep := appendEnt.reply
			currTerm := r.GetCurrentTerm()
			leader := req.LeaderId

			if req.Term < currTerm {
				rep <- AppendEntriesReply{currTerm, false}
			} else {
				r.LeaderAddr = &leader
				r.setCurrentTerm(req.Term)
				r.setVotedFor("")
				rep <- AppendEntriesReply{currTerm, true}
				r.sendRequestFail()
				return r.doFollower
			}

		case vote := <-r.requestVote:
			req := vote.request
			candidate := req.CandidateId

			if r.handleCompetingRequestVote(vote) {
				r.setCurrentTerm(req.Term)

				// Set voted for field (already voted)
				r.setVotedFor(candidate.Id)

				// Election in progess. There is no leader
				// Set to nil
				r.LeaderAddr = nil
				r.sendRequestFail()
				return r.doFollower
			} else {
				// Other candidate has a less up to date log.
			}

		case regClient := <-r.registerClient:
			rep := regClient.reply

			entries := make([]LogEntry, 1)
			entries[0] = LogEntry{r.getLastLogIndex() + 1, r.GetCurrentTerm(), CLIENT_REGISTRATION, make([]byte, 0), ""}
			//we add it to the log
			r.appendLogEntry(entries[0])

			fallback, maj := r.sendAppendEntries(entries)

			if fallback {
				if maj {
					rep <- RegisterClientReply{OK, r.getLastLogIndex(), *r.LeaderAddr}
				} else {
					rep <- RegisterClientReply{REQ_FAILED, 0, *r.LeaderAddr}
				}
				//not sure we need to do this
				r.sendRequestFail()
				return r.doFollower
			}

			if !maj {
				rep <- RegisterClientReply{REQ_FAILED, 0, *r.LeaderAddr}
				//Truncate log, didnt happen
				r.truncateLog(r.getLastLogIndex())
			} else {
				//r.processLog(entries[0])
				//r.commitIndex++
				//r.sendAppendEntries(make([]LogEntry, 0))
				rep <- RegisterClientReply{OK, r.getLastLogIndex(), *r.LeaderAddr}
			}

		case clientReq := <-r.clientRequest:
			req := clientReq.request
			// rep := clientReq.reply
			//TODO: Should we first check that it's registered?
			entries := make([]LogEntry, 1)
			//Fill in the LogEntry based on the request data
			entries[0] = LogEntry{r.getLastLogIndex() + 1, r.GetCurrentTerm(), req.Command, req.Data, strconv.FormatUint(req.SequenceNum, 10)}
			r.appendLogEntry(entries[0])

			//we now send it to everyone
			fallback, maj := r.sendAppendEntries(entries)
			r.requestMap[req.SequenceNum] = clientReq

			if fallback {
				//Ahora no estoy tan seguro de hacer esto
				r.sendRequestFail()
				return r.doFollower
			}

			if !maj {
				//rep <- ClientReply{REQ_FAILED, 0, *r.LeaderAddr}
				//Truncate log, didn't happen
				//r.truncateLog(r.getLastLogIndex())
			} else {
				//Nos esperamos a que el heartbeat haga commit.
				//We now cache the response in the response map to reply to it later
				//r.requestMap[r.SequenceNum] = clientReq
			}
			// electionTimeout = makeElectionTimeout()
		}
	}

	return nil
}

func (r *RaftNode) sendRequestFail() {
	for _, v := range r.requestMap {
		v.reply <- ClientReply{REQ_FAILED, "", *r.LeaderAddr}
	}
}

func (r *RaftNode) sendNoop() bool {
	// Send NOOP as first entry to all nodes.
	entries := make([]LogEntry, 1)
	entries[0] = LogEntry{r.getLastLogIndex() + 1, r.GetCurrentTerm(), NOOP, make([]byte, 0), ""}
	r.appendLogEntry(entries[0])
	f, maj := r.sendAppendEntries(entries)

	// Leader should fall back, but just got elected. How?
	if f {
		panic("Leader got in but had to fall back. Por que?")
	}
	if !maj {
		panic("major got in but had to fall back. Por que?")
	}

	//No processLog until we increase the commit index. Which happens in
	//heartbeat.
	// if maj {
	// 	r.processLog(entries[0])
	// 	r.commitIndex++
	// }
	// r.sendAppendEntries(make([]LogEntry, 0))

	return true
}

/**
 * This function is called when the node is a candidate or leader, and a
 * competing RequestVote is called. It will return true if the caller should
 * fall back to the follower state.
 */
// DICE QUE ES SOLO CANDIDATE O LEADER, PERO SEGUN YO FOLLOWER JALA TAMBIEN NO?
func (r *RaftNode) handleCompetingRequestVote(msg RequestVoteMsg) bool {
	req := msg.request
	rep := msg.reply
	prevIndex := r.commitIndex
	prevTerm := r.getLogTerm(prevIndex)
	currTerm := r.GetCurrentTerm()

	if prevTerm > req.LastLogTerm { // My commit term is greater, no vote.
		rep <- RequestVoteReply{currTerm, false}
		return false
	} else if prevTerm < req.LastLogTerm { // My commit term is smaller, vote.
		rep <- RequestVoteReply{currTerm, true}
		return true
	} else { // Same commit indexes
		if prevIndex > req.LastLogIndex { // My commit index is greater, no vote.
			rep <- RequestVoteReply{currTerm, false}
			return false
		} else if prevIndex < req.LastLogIndex { // My commit index is smaller, vote.
			rep <- RequestVoteReply{currTerm, true}
			return true
		} else { // Commit index and term are the same. Tiebreaker with currTerms
			// Previous election or election that I already voted for or
			// currently pursuing. No vote.
			if req.Term <= currTerm {
				rep <- RequestVoteReply{currTerm, false}
				return false
			} else { // Greater election term, give vote.
				rep <- RequestVoteReply{currTerm, true}
				return true
			}
		}
	}
}

/**
 * This function is called to request votes from all other nodes. It takes
 * a channel which the result of the vote should be sent over: true for
 * successful election, false otherwise.
 */
func (r *RaftNode) requestVotes(electionResults chan bool) {
	go func() {
		nodes := r.GetOtherNodes()
		num_nodes := len(nodes)
		votes := 1
		r.setVotedFor(r.Id)
		for _, n := range nodes {
			//reply, _ := r.RequestVoteRPC(&n,
			//	RequestVoteRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
			//		r.getLastLogIndex(), r.getLastLogTerm()})
			reply, _ := r.RequestVoteRPC(&n,
				RequestVoteRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
					r.commitIndex, r.getLogTerm(r.commitIndex)})

			if reply == nil {
				// Could not reach node for vote.
				continue
			}

			if r.GetCurrentTerm() < reply.Term {
				r.Debug("YA VALIO %d < %d\n", r.GetCurrentTerm(), reply.Term)
				electionResults <- false // YA VALIO MADRE
				return
			}

			// r.Debug("RequestVotes: votes %v term %d\n", reply.VoteGranted, r.GetCurrentTerm())
			if reply.VoteGranted {
				votes++
			}
			if votes > num_nodes/2 {
				electionResults <- true
				r.Debug("RequestVotes: won with %d votes\n", votes)
				return
			}
		}

		electionResults <- false
		r.Debug("RequestVotes: lost with %d votes\n", votes)
	}()
}

/*func (r *RaftNode) sendHeartBeats() (fallBack, sentToMajority bool) {
	// TODO: Students should implement this method

	nodes := r.GetOtherNodes()
	num_nodes := len(nodes)
	succ_nodes := 1
	fail_nodes := 0
	succ_chan := make(chan bool)
	should_fallback := make(chan bool)
	for _, n := range nodes {
		if n.Id == r.Id {
			continue
		}

		go func() {
			reply, _ := r.AppendEntriesRPC(&n,
				AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
					r.getLastLogIndex(), r.getLastLogTerm(), make([]LogEntry, 0),
					r.commitIndex})

			if reply != nil {
				if reply.Success {
					succ_chan <- true
				} else {
					if r.GetCurrentTerm() < reply.Term {
						should_fallback <- true
					} else {
						// TODO: fix log, CHECK IF ITS LEADER
					}
				}
			} else {
				succ_chan <- false
			}
		}()
	}

	for {
		r.Debug("aaa\n")

		select {
		case vote := <-succ_chan:
			if vote {
				succ_nodes++
			} else {
				fail_nodes++
			}
		case <-should_fallback:
			return true, true
		}
		if succ_nodes > num_nodes/2 {
			return false, true
		} else if succ_nodes+fail_nodes >= num_nodes {
			return false, false
		}
	}

	// Should never get here.
	return false, false
} */
func (r *RaftNode) sendEmptyHeartBeats() {
	// TODO: Students should implement this method

	nodes := r.GetOtherNodes()
	for _, n := range nodes {
		if n.Id == r.Id {
			continue
		}
		r.AppendEntriesRPC(&n,
			AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
				r.getLastLogIndex(), r.getLastLogTerm(), make([]LogEntry, 0),
				r.commitIndex})
	}
}

/**
 * This function is used by the leader to send out heartbeats to each of
 * the other nodes. It returns true if the leader should fall back to the
 * follower state. (This happens if we discover that we are in an old term.)
 *
 * If another node isn't up-to-date, then the leader should attempt to
 * update them, and, if an index has made it to a quorum of nodes, commit
 * up to that index. Once committed to that index, the replicated state
 * machine should be given the new log entries via processLog.
 */
func (r *RaftNode) sendHeartBeats() (fallBack, sentToMajority bool) {

	nodes := r.GetOtherNodes()
	num_nodes := len(nodes)
	succ_nodes := 1
	for _, n := range nodes {
		if n.Id == r.Id {
			continue
		}
		prevLogIndex := r.nextIndex[n.Id] - 1
		//TODO: Que pasa si el next index esta mas arriba que el log del leader?
		// Esta madre va a tronar.
		// r.Out("nextIndex vs prevLog vs lastLog %v, %v, %v, %v", r.nextIndex[n.Id], prevLogIndex, r.getLastLogIndex(), r.logCache)
		prevLogTerm := r.getLogEntry(prevLogIndex).TermId
		reply, _ := r.AppendEntriesRPC(&n,
			AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
				prevLogIndex, prevLogTerm, make([]LogEntry, 0),
				r.commitIndex})

		if reply == nil {
			continue
		}

		if r.GetCurrentTerm() < reply.Term {
			return true, true
		}

		if reply.Success {
			succ_nodes++
			nextIndex := r.nextIndex[n.Id]
			//Aqui segun you se vuelven iguales porque ya apendearon
			r.matchIndex[n.Id] = r.nextIndex[n.Id] - 1
			if (nextIndex - 1) != r.getLastLogIndex() {
				// TODO
				// r.Out("nextIndex vs lastLog %v, %v", nextIndex, r.getLastLogIndex())
				entries := r.getLogEntries(nextIndex, r.getLastLogIndex())
				reply, _ = r.AppendEntriesRPC(&n,
					AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
						prevLogIndex, prevLogTerm, entries,
						r.commitIndex})

				if reply != nil && reply.Success {
					r.nextIndex[n.Id] = r.getLastLogIndex() + 1
					r.matchIndex[n.Id] = r.nextIndex[n.Id] - 1
				}
			}

		} else {
			r.nextIndex[n.Id]--
		}
	}

	// Updates commit index according to what the followers have in their logs already.
	for N := r.getLastLogIndex(); N > r.commitIndex; N-- {
		if r.hasMajority(N) && r.getLogTerm(N) == r.GetCurrentTerm() {
			r.commitIndex = N
		}
	}

	// Calls process log from lastApplied until commit index.
	if r.lastApplied != r.commitIndex {
		i := r.lastApplied
		if r.lastApplied != 0 {
			i++
		}
		for ; i <= r.commitIndex; i++ {
			reply := r.processLog(*r.getLogEntry(i))
			if reply.Status == REQ_FAILED {
				panic("This should not happen ever! PLZ FIX")
			}
		}
		r.lastApplied = r.commitIndex
	}

	if succ_nodes > num_nodes/2 {
		return false, true
	}

	return false, false
}

func (r *RaftNode) sendAppendEntries(entries []LogEntry) (fallBack, sentToMajority bool) {
	// TODO: Students should implement this method

	nodes := r.GetOtherNodes()
	num_nodes := len(nodes)
	succ_nodes := 1

	for _, n := range nodes {
		if n.Id == r.Id {
			continue
		}
		//we have the - 1 because we just added it (before this call) to our own log
		prevLogIndex := r.getLastLogIndex() - 1
		prevLogTerm := r.getLogTerm(prevLogIndex)
		reply, _ := r.AppendEntriesRPC(&n,
			AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
				prevLogIndex, prevLogTerm, entries,
				r.commitIndex})

		if reply != nil {
			succ_nodes++
			//We increase the next index for this node
			//It's lastLogIndex + 1 porque en lastLogIndex es donde lo
			//acaba de agregar.
			r.nextIndex[n.Id] = r.getLastLogIndex() + 1
		} else {
			continue
		}

		if r.GetCurrentTerm() < reply.Term {
			return true, true
		} else if !reply.Success {
			// TODO
		}
	}

	if succ_nodes > num_nodes/2 {
		return false, true
	}

	return false, false
}

/**
 * This function will use time.After to create a random timeout.
 */
func makeElectionTimeout() <-chan time.Time {
	// TODO: Students should implement this method
	millis := rand.Int()%150 + 150
	return time.After(time.Millisecond * time.Duration(millis))
}

func (r *RaftNode) makeBeats() <-chan time.Time {
	return time.After(r.config.HeartbeatFrequency)
}
