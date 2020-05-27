package puddlestore

import (
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

// connectZk sets up the zookeeper connection
func connectZk(zkAddr string) (*zk.Conn, error) {
	conn, _, err := zk.Connect([]string{zkAddr}, 1*time.Second)
	return conn, err
}

// createEphSeq creates an ephemeral|sequential znode
func createEphSeq(conn *zk.Conn, path string, data []byte) error {
	_, err := conn.CreateProtectedEphemeralSequential(
		path,
		data,
		zk.WorldACL(zk.PermAll),
	)

	return err
}
