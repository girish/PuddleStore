package raft

import (
	"time"
)

type state func() state

/**
 * This method contains the logic of a Raft node in the follower state.
 */
func (r *RaftNode) doFollower() state {
	r.Out("Transitioning to FOLLOWER_STATE")
	r.State = FOLLOWER_STATE
	// TODO: Students should implement this method
	for {
		select {
		case shutdown := <-r.gracefulExit:
			if shutdown {
				return nil
			}
		}
	}
}

/**
 * This method contains the logic of a Raft node in the candidate state.
 */
func (r *RaftNode) doCandidate() state {
	r.Out("Transitioning to CANDIDATE_STATE")
	r.State = CANDIDATE_STATE
	// TODO: Students should implement this method
	return nil
}

/**
 * This method contains the logic of a Raft node in the leader state.
 */
func (r *RaftNode) doLeader() state {
	r.Out("Transitioning to LEADER_STATE")
	r.State = LEADER_STATE
	// TODO: Students should implement this method
	return nil
}

/**
 * This function is called when the node is a candidate or leader, and a
 * competing RequestVote is called. It will return true if the caller should
 * fall back to the follower state.
 */
func (r *RaftNode) handleCompetingRequestVote(msg RequestVoteMsg) bool {
	// TODO: Students should implement this method
	return true
}

/**
 * This function is called to request votes from all other nodes. It takes
 * a channel which the result of the vote should be sent over: true for
 * successful election, false otherwise.
 */
func (r *RaftNode) requestVotes(electionResults chan bool) {
	// TODO: Students should implement this method
	return
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
	return true, true
}

/**
 * This function will use time.After to create a random timeout.
 */
func makeElectionTimeout() <-chan time.Time {
	// TODO: Students should implement this method
	return nil
}
