package puddlestore

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"path/filepath"
)

/*
type File struct {
	FileName string
	Size     uint64
	Blocks   map[int]string
	IsDir    bool
	conn     *zk.Conn
}
*/

type Dir struct {
	iNode
}

func CreateDir(FileName string, c *PuddleClient) (*Dir, error) {
	dir, err := newiNode(c.zk, FileName, true, c.BlockSize)
	if err != nil {
		return nil, err
	}
	data, err := encodeMsgPack(dir)
	if err != nil {
		return nil, fmt.Errorf("createDir: error when encode dir")
	}
	_, err = c.zk.Create(dir.FileName, data.Bytes(), 0, zk.WorldACL(zk.PermAll))
	return &Dir{*dir}, nil
}

// A recursive remove function
func (f *Dir) Remove(c *PuddleClient) error {
	// Examine the target directory
	children, _, err := c.zk.Children(f.FileName)
	if err != nil {
		return err
	}
	for _, childName := range children {
		childName = filepath.Join(f.FileName, childName)
		codedBytes, _, err := c.zk.Get(childName)
		if err != nil {
			return fmt.Errorf("error: when zookeeper is trying to find target, %v", err)
		}
		var target *iNode
		err = decodeMsgPack(codedBytes, &target)
		if err != nil {
			return fmt.Errorf("error: when decoding... %v", err)
		}
		err = target.Remove(c)
		if err != nil {
			return fmt.Errorf("error: remove error, %v", err)
		}
	}
	// Last, **this** directory metadata should be removed from zookeeper
	return c.zk.Delete(f.FileName, -1)
}
