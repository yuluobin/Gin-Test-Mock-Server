package puddlestore

import (
	"fmt"
	"github.com/brown-csci1380-s20/puddlestorenew-puddlestorenew-cwang147-byu18-mxu57/tapestry"
	"github.com/samuel/go-zookeeper/zk"
	"path/filepath"
)

// Tapestry is a wrapper for a single Tapestry node. It is responsible for
// maintaining a zookeeper connection and implementing methods we provide
type Tapestry struct {
	tap *tapestry.Node //# Uncomment this
	zk  *zk.Conn
}

// NewTapestry creates a new tapestry struct. Uncomment this function
func newTapestry(tap *tapestry.Node, zkAddr string) (*Tapestry, error) {
	//TODO: Setup a zookeeper connection and return a Tapestry struct
	conn, err := connectZk(zkAddr)
	if err != nil {
		return nil, err
	}
	exists, _, err := conn.Exists("/tapestry")
	if err != nil {
		return nil, fmt.Errorf("error: zookeeper fail to find target, reason is %v", err)
	}
	if !exists {
		_, err = conn.Create("/tapestry", nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			return nil, err
		}
	}
	// Tapestry register them in ZooKeeper
	// we will simply use file paths as unique IDs for files and directories.
	err = createEphSeq(conn, filepath.Join("/tapestry", tap.Addr()), []byte(tap.Addr()))
	if err != nil {
		return nil, err
	}

	return &Tapestry{
		tap: tap,
		zk:  conn,
	}, nil
}

// GracefulExit closes the zookeeper connection and gracefully shuts down the tapestry node
func (t *Tapestry) GracefulExit() {
	t.zk.Close()
	t.tap.Leave() //# Uncomment this
}
