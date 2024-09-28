package raft

import (
	"sync"
	"time"

	"github.com/vshishkar/raft/log"
	"github.com/vshishkar/raft/net"
)

type rState string

const (
	follower  rState = "foll"
	candidate rState = "cand"
	leader    rState = "lead"
)

type Raft struct {
	mu    sync.Mutex
	id    int64
	peers map[int64]*net.PeerConnection

	currentTerm      int64
	votedFor         int64
	state            rState
	lastHbOrElection time.Time
	electionTimeout  time.Duration

	log log.Log
}
