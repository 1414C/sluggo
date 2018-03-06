package wscom

// Article is the central artifact (Set/Get)
type Article struct {
	Key   string // uuid.UUID
	Op    string // {AU, D}
	Valid bool
	Value []byte
	Type  string
}

// how does server registration work?

// CacheServerGroup is a read-only map of servers in the group
var CacheServerGroup map[string]bool

// Init initiallizes the server group for now
func Init() error {
	CacheServerGroup = make(map[string]bool)
	CacheServerGroup["192.168.1.79"] = true
	CacheServerGroup["192.168.1.76"] = true
	return nil
}
