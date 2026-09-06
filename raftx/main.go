package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type LogEntry struct {
	command string
	term    int
}

type RoleType int

const (
	RoleTypeFollower  	RoleType = 0
	RoleTypeCandidate 	RoleType = 1
	RoleTypeLeader    	RoleType = 2
)

type Node struct {
	// identities of a server.
	id   		int
	role 		RoleType

	// persistent states on all servers.
	currentTerm 	int
	votedFor    	int	 	// could be null(-1)
	logs        	[]LogEntry

	// volatile states on all servers.
	commitIndex 	int
	lastApplied 	int

	// volatile state on leaders.
	nextIndex  		[]int
	matchIndex 		[]int

	lastHeartBeat 	time.Time

	peers 			[]*Node

	mu 				sync.Mutex
}

// RequestVoteArgs represents the message that goes to
// other nodes from a candidate as a RPC request.
type RequestVoteArgs struct {
	term			int
	candidateId		int
	lastLogIndex	int
	lastLogTerm		int
}

// RequestVoteReply represents the reply to the RequestVote RPC
// that tells the candidate about the requested vote's status.
type RequestVoteReply struct {
	term			int
	voteGranted		bool
}

func newNode(id int) *Node {
	return &Node{
		id: 			id,
		role:			RoleTypeFollower,
		votedFor: 		-1,
		lastHeartBeat: 	time.Now(),
	}
} 

func (n *Node) runElectionTimer() {
	for {
		electionTimeout := time.Duration(150+rand.Intn(150)) * time.Millisecond
		time.Sleep(electionTimeout)

		// wake up and lock the state.
		n.mu.Lock()
		if time.Since(n.lastHeartBeat) > electionTimeout {
			log.Printf("[%d]: No hearbeat touched. election timeout.\n", n.id)
			n.mu.Unlock()
			n.startElection()
		} else {
			// yes the heartbeat did reached.
			n.mu.Unlock()
		}
	}
}

// When a follower converts to a candidate. It does these in an order -
// Increment its term -> Vote for itself -> Reset Election Timer -> Send RequestVote RPC.
func (n *Node) startElection() {
	n.mu.Lock()

	n.currentTerm++
	n.role = RoleTypeCandidate
	n.votedFor = n.id
	votes := 1		// self vote
	n.lastHeartBeat = time.Now()

	term := n.currentTerm
	lastIndex, lastTerm := n.lastLogInfo()
	n.mu.Unlock()

	// send RequestVoteRPC.
	args := &RequestVoteArgs{
		term : 			term,
		candidateId: 	n.id,
		lastLogIndex:	lastIndex,
		lastLogTerm:	lastTerm,
	}
	
	for _, peer := range n.peers {
		if peer.id != n.id {
			reply := peer.RequestVote(args)

			n.mu.Lock()
			if reply.term > n.currentTerm {
				n.lastHeartBeat = time.Now()
				n.role = RoleTypeFollower
				n.votedFor = -1
				n.currentTerm = reply.term
				n.mu.Unlock()
				return
			}
			
			if n.role != RoleTypeCandidate || n.currentTerm != term {
				n.mu.Unlock()
				return
			}
			
			if reply.voteGranted { votes++ }
			
			if votes > len(n.peers) / 2 {
				n.role = RoleTypeLeader
				log.Printf("[LEADER][%d] is voted as the leader now.", n.id)
				n.mu.Unlock()
				return
			}
			n.mu.Unlock()
		}
	}
}

func (n *Node) lastLogInfo() (index, term int) {
	if(len(n.logs) == 0) {
		return 0, 0
	} else {
		return len(n.logs), n.logs[len(n.logs) - 1].term
	}
}

func (n *Node) RequestVote(args *RequestVoteArgs) *RequestVoteReply {
	n.mu.Lock()
	defer n.mu.Unlock()

	// if term of candidate is lesser, then reject the vote request
	// and send the term of node so that candidate could step down.
	if args.term < n.currentTerm {
		return &RequestVoteReply{
			term:			n.currentTerm,
			voteGranted:	false,
		}
	}

	// when would the vote be granted ?
	// when the term of the candidate is higher then the follower node.
	if args.term > n.currentTerm {
		n.currentTerm = args.term
		n.role = RoleTypeFollower
		n.votedFor = -1
	}

	index, term := n.lastLogInfo()

	// vote granting conditions.
	if ( n.votedFor == -1 || n.votedFor == args.candidateId ) && ( args.lastLogTerm > term || ( args.lastLogTerm == term && args.lastLogIndex >= index ) ) {
		n.votedFor = args.candidateId
		n.lastHeartBeat = time.Now()

		return &RequestVoteReply{
			term:			n.currentTerm,
			voteGranted: 	true,
		}
	}
	return &RequestVoteReply{
		term:				n.currentTerm,
		voteGranted:		false,
	}
}

func main() {
	// initiate 3 nodes and then run the goroutine.
	node1 := newNode(0)
	node2 := newNode(1)
	node3 := newNode(2)

	nodes := []*Node{node1, node2, node3}
	for _, n := range nodes {
		n.peers = nodes
	}	

	go node1.runElectionTimer()
	go node2.runElectionTimer()
	go node3.runElectionTimer()

	time.Sleep(time.Second * 2)
}
