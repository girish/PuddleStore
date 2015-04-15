package raft

import (
	"math/rand"
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

			if req.Term < currTerm {
				// Vote request is from an old term, return false.
				rep <- RequestVoteReply{currTerm, false}
			} else {
				if votedFor == "" || votedFor == candidate.Id {
					// TODO: Check log to see si el vato la arma.
					r.setCurrentTerm(req.Term)

					// Set voted for field (already voted)
					r.setVotedFor(candidate.Id)

					// Election in progess. There is no leader
					// Set to nil
					r.LeaderAddr = nil

					// Respond true, reset eleciton timeout.
					rep <- RequestVoteReply{currTerm, true}
					electionTimeout = makeElectionTimeout()
				} else {
					// Already voted.
					rep <- RequestVoteReply{currTerm, false}
				}
			}

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
			if r.handleCompetingRequestVote(vote) {
				return r.doFollower
			} else {
			}

		case appendEnt := <-r.appendEntries:
			// req := appendEnt.request
			rep := appendEnt.reply
			currTerm := r.GetCurrentTerm()
			//leader := req.LeaderId

			/*if req.Term < currTerm {
				rep <- AppendEntriesReply{currTerm, false}
			} else {*/
			//r.LeaderAddr = &leader
			//r.setCurrentTerm(req.Term)
			//r.setVotedFor("")
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
	beats := r.makeBeats()

	// Send NOOP as first entry to all nodes.
	entries := make([]LogEntry, 1)
	entries[0] = LogEntry{r.getLastLogIndex() + 1, r.GetCurrentTerm(), NOOP, make([]byte, 0), ""}
	appendLogEntry(entries[0])
	f, maj := r.sendAppendEntries(entries)

	// Leader should fall back, but just got elected. How?
	if f || !maj {
		panic("Leader got in but had to fall back. Por que?")
	}

	if maj {
	}

	for {
		select {
		case <-beats:
			r.Debug("sendHeartBeats: entered\n")
			f, _ := r.sendHeartBeats()
			if f {
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
				return r.doFollower
			}

		case regClient := <-r.registerClient:
			// req := regClient.request
			rep := regClient.reply
			rep <- RegisterClientReply{ELECTION_IN_PROGRESS, 0, NodeAddr{"", ""}}
			// electionTimeout = makeElectionTimeout()
		}
	}

	return nil
}

/**
 * This function is called when the node is a candidate or leader, and a
 * competing RequestVote is called. It will return true if the caller should
 * fall back to the follower state.
 */
func (r *RaftNode) handleCompetingRequestVote(msg RequestVoteMsg) bool {
	req := msg.request
	rep := msg.reply
	currTerm := r.GetCurrentTerm()

	if req.Term < currTerm {
		rep <- RequestVoteReply{currTerm, false}
		return false
	} else if req.Term == currTerm {
		rep <- RequestVoteReply{currTerm, false}
		return false
	}
	rep <- RequestVoteReply{currTerm, true}
	return true
}

/**
 * This function is called to request votes from all other nodes. It takes
 * a channel which the result of the vote should be sent over: true for
 * successful election, false otherwise.
 */
func (r *RaftNode) requestVotes(electionResults chan bool) {
	// TODO: Students should implement this method
	// RequestVoteRequest
	go func() {
		nodes := r.GetOtherNodes()
		num_nodes := len(nodes)
		votes := 1
		r.setVotedFor(r.Id)
		for _, n := range nodes {
			reply, _ := r.RequestVoteRPC(&n,
				RequestVoteRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
					r.getLastLogIndex(), r.getLastLogTerm()})

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
	// TODO: Students should implement this method

	nodes := r.GetOtherNodes()
	num_nodes := len(nodes)
	succ_nodes := 1
	for _, n := range nodes {
		if n.Id == r.Id {
			continue
		}
		reply, err := r.AppendEntriesRPC(&n,
			AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
				r.getLastLogIndex(), r.getLastLogTerm(), make([]LogEntry, 0),
				r.commitIndex})

		if err != nil {
			succ_nodes++
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

func (r *RaftNode) sendAppendEntries(entries []LogEntry) (fallBack, sentToMajority bool) {
	// TODO: Students should implement this method

	nodes := r.GetOtherNodes()
	num_nodes := len(nodes)
	succ_nodes := 1

	for _, n := range nodes {
		if n.Id == r.Id {
			continue
		}
		reply, err := r.AppendEntriesRPC(&n,
			AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
				r.getLastLogIndex(), r.getLastLogTerm(), entries,
				r.commitIndex})

		if err != nil {
			succ_nodes++
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
