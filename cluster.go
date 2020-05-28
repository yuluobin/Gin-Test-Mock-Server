package puddlestore

import (
	"github.com/yuluobin/Gin-Test-Mocker-Server/tapestry"
	"math/rand"
	"time"
)

// Cluster is a struct for all nodes in a puddlestore cluster. One should be able to shutdown
// this cluster and create a client for this cluster
type Cluster struct {
	config Config
	nodes  []*Tapestry
}

// Shutdown causes all the tapestry nodes to gracefully exit
func (c *Cluster) Shutdown() {
	for _, node := range c.nodes {
		node.GracefulExit()
	}

	time.Sleep(time.Second)
}

// NewClient creates a new Puddlestore client
// - After associating a tapestry client with puddlestore client, use a go function https://piazza.com/class/k5illljohg02m8?cid=1097
// to listen for node address leaving zookeeper
func (c *Cluster) NewClient() (Client, error) {
	// TODO: Return a new PuddleStore Client that implements the Client interface
	puddleClient, err := NewPuddleClient(c)
	if err != nil {
		return nil, err
	}

	return puddleClient, nil
}

// CreateCluster starts all nodes necessary for puddlestore
func CreateCluster(config Config) (*Cluster, error) {
	// TODO: Start your tapestry cluster with size config.NumTapestry. You should
	// also use the zkAddr (zookeeper address) found in the config and pass it to
	// your Tapestry constructor method
	var cluster Cluster
	cluster.config = config
	tapestries, err := tapestry.MakeRandomTapestries(int64(rand.Intn(1000)) /*rand*/, config.NumTapestry)
	if err != nil {
		return nil, err
	}
	for i := 0; i < config.NumTapestry; i++ {
		tap, err := newTapestry(tapestries[i], config.ZkAddr)
		if err != nil {
			return nil, err
		}
		cluster.nodes = append(cluster.nodes, tap)
	}

	return &cluster, nil
}
