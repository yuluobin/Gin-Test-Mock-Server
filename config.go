package puddlestore

// Config for the puddlestore distributed file system
type Config struct {
	// BlockSize is size of a block in bytes. Direct and indirect blocks
	// are the same size. inodes have no set size to make life easier
	BlockSize uint64

	// NumReplicas is the amount of tapestry nodes to replicate each (VGUID, data) pair to
	NumReplicas int

	// NumTapestry is the number of tapestry nodes to start during cluster creation
	NumTapestry int

	// ZkAddr is the address of a zookeeper node
	ZkAddr string
}

// DefaultConfig is the default config for puddlestore. It is `lightweight` on purpose
// for testing reasons
func DefaultConfig() Config {
	return Config{
		BlockSize:   64,
		NumReplicas: 2,
		NumTapestry: 2,
		ZkAddr:      "localhost:2181",
	}
}
