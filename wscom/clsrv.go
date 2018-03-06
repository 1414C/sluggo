package wscom

import "fmt"

// Article is the central artifact (Set/Get)
type Article struct {
	Key   string // uuid.UUID
	Op    string // {AU, D}
	Valid bool
	Value []byte
	Type  string
}

// SuccessorInfo is used to hold the topology info
type SuccessorInfo struct {
	PID  uint64
	Addr string
}

// Message is used for inter-cache comms
type Message struct {
	OriginID  uint64                   // Origin Pn of message
	Mneumonic string                   // message instruction   (HB,ELECTION,ELECTED,ADD,REMOVE,ACK,JOINREQ)
	SubjectID uint64                   // subjectid refers to the process being tested, proposed for election, being elected, or being added/removed)
	Topology  map[uint64]SuccessorInfo // common map?
}

// Leader holds the leader Pid and the leader's address
var Leader SuccessorInfo

// MyPID is the Pid of the cache instance running on localhost (whatever that is)
var MyPID uint64

// SuccessorPID is the successor Pid to the cache instance running localhost (needed?)
var SuccessorPID uint64

// CurrentTopology is a map containing a list of pid-pid-successors based on the last message received
var CurrentTopology map[uint64]SuccessorInfo

// CacheServerGroup is a read-only map of servers in the group
var CacheServerGroup map[uint64]string

// Init initializes the server group
func Init(leaderID uint64, leaderAddr string) error {

	// set the Leader.ID and MyPID if first server to start
	if leaderID == 0 {
		Leader.PID = 1
		Leader.Addr = leaderAddr // or somehow choose an address?  in multi-NIC systems this may be difficult?
		MyPID = 1
		fmt.Println("set LeaderID = 1, MyPID = 1")
	} else {
		Leader.PID = leaderID
		Leader.Addr = leaderAddr
		fmt.Println("set LeaderID =", leaderID)
		fmt.Println("sending JOINREQ to leader")
	}

	CurrentTopology = make(map[uint64]SuccessorInfo)

	// need to get the current leader
	// if no leader, set myself as the leader

	// CacheServerGroup = make(map[string]bool)
	// CacheServerGroup["192.168.1.79"] = true
	// CacheServerGroup["192.168.1.76"] = true
	return nil
}
