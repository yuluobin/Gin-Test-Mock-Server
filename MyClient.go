package puddlestore

import (
	"fmt"
	"github.com/brown-csci1380-s20/puddlestorenew-puddlestorenew-cwang147-byu18-mxu57/tapestry"
	"github.com/samuel/go-zookeeper/zk"
	"path/filepath"
	"sort"
	"strings"
)

// https://piazza.com/class/k5illljohg02m8?cid=1097
type PuddleClient struct {
	count       int // https://piazza.com/class/k5illljohg02m8?cid=1090
	zk          *zk.Conn
	tapClients  []*tapestry.Client
	files       map[int]*File
	writable    map[int]bool
	lockPaths   map[int]string
	BlockSize   uint64
	NumReplicas int
	//FileMapLock *zk.Lock
}

func NewPuddleClient(c *Cluster) (*PuddleClient, error) {
	var resClient PuddleClient

	// - A puddle client does not know any tapestry nodes to begin with.
	// It will discover tapestry nodes through zookeeper.
	conn, err := connectZk(c.config.ZkAddr)
	if err != nil {
		return nil, err
	}
	resClient.zk = conn
	resClient.count = 0
	resClient.BlockSize = c.config.BlockSize
	resClient.files = make(map[int]*File)
	resClient.writable = make(map[int]bool)
	resClient.lockPaths = make(map[int]string)
	resClient.NumReplicas = c.config.NumReplicas

	// -------------------------------- Create lock directory --------------------------
	exists, _, err := conn.Exists("/lock")
	if err != nil {
		return nil, fmt.Errorf("error: zookeeper fail to find lock directory, reason is %v", err)
	}
	if !exists {
		_, err = conn.Create("/lock", nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			return nil, err
		}
	}
	// Create a zk ordinary lock file
	//l1 := zk.NewLock(conn, "/fileLock", zk.WorldACL(zk.PermAll))
	//resClient.FileMapLock = l1
	// Create Tapestry lock file
	//l2 := zk.NewLock(conn, "/tapClientLock", zk.WorldACL(zk.PermAll))
	//resClient.tapLock = l2
	// Create dir lock file
	//l3 := zk.NewLock(conn, "/dirLock", zk.WorldACL(zk.PermAll))
	//resClient.dirLock = l3
	// -------------------------------- Create lock directory End -------------------------

	// So `zkEvent` can "watch"??????????????????????????????????????????????
	children, _, _, err := resClient.zk.ChildrenW("/tapestry")
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(children); i++ {
		// Create tap client
		port, _, err := resClient.zk.Get(filepath.Join("/tapestry", children[i]))
		if err != nil {
			return nil, err
		}
		tapClient, err := tapestry.Connect(string(port))
		if err != nil {
			return nil, err
		}
		resClient.tapClients = append(resClient.tapClients, tapClient)
	}
	//go resClient.WatchChildren()
	//watchChan := make(chan bool)
	go resClient.WatcherTap()

	return &resClient, nil
}

// Here may has concurrency problem
// You can ignore the error message at the end of test, since Tapestry node exit earlier than PuddleClient
func (c *PuddleClient) WatcherTap() {
	for {
		_, _, zkEvent, err := c.zk.ChildrenW("/tapestry")
		if err != nil {
			fmt.Printf("watch error: path is %v and reason is %v", "/tapestry\n", err)
			continue
		}
		<-zkEvent
		// Update tapestry client connection list
		//fmt.Printf("I get here\n")
		c.tapClients = nil
		children, _, _, err := c.zk.ChildrenW("/tapestry")
		if err != nil {
			fmt.Printf("watch error: path is %v and reason is %v", "/tapestry\n", err)
			continue
		}
		for i := 0; i < len(children); i++ {
			// Create tap client
			port, _, err := c.zk.Get(filepath.Join("/tapestry", children[i]))
			if err != nil {
				fmt.Printf("watch error: path is %v and reason is %v\n", "/tapestry", err)
				continue
			}
			tapClient, err := tapestry.Connect(string(port))
			if err != nil {
				fmt.Printf("watch error: path is %v and reason is %v\n", "/tapestry", err)
				continue
			}
			c.tapClients = append(c.tapClients, tapClient)
		}
		//fmt.Printf("Now client's tapClient has %v clients\n", len(c.tapClients))
	}
}

func (c *PuddleClient) Open(path string, create, write bool) (int, error) {
	// if it has opened
	// This part may need to modify!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	for k, v := range c.files {
		if v.FileName == path {
			return k, nil
		}
	}
	// I mean the part above!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	// If there's no parent path, fail
	if path[len(path)-1] == '/' {
		return -1, fmt.Errorf("open: file name contains invalid character")
	}
	parentpath, _ := findParentDir(path)
	exists, _, err := c.zk.Exists(parentpath)
	if err != nil {
		return -1, err
	}
	if !exists {
		// No parent path
		return -1, fmt.Errorf("open: the target file has no parent path")
	}

	exists, _, err = c.zk.Exists(path)
	if err != nil {
		return -1, err
	}

	if create && !exists {
		// Examine the file path first
		if parentpath != "/" {
			data, _, err := c.zk.Get(parentpath)
			if err != nil {
				return -1, fmt.Errorf("open: cannot GET, zk connection might have problem")
			}
			var inode *iNode
			err = decodeMsgPack(data, &inode)
			if err != nil {
				return fmt.Printf("error: cannot decode file, reason is %v", err)
			}
			if !inode.IsDir {
				return -1, fmt.Errorf("open: cannot create file 'cause parent path is not a disrectory")
			}
		}

		// Create a new file
		// NOTE: although `exists` can be outdated, there's still no concurrency problem since `createFile` will
		// return an error "file already exists". It's normal actually
		file, err := CreateFile(path, c) // But doesn't write into zookeeper
		if err != nil {
			return -1, err
		}
		// Create lock directory
		// file creation happens immediately
		_, err = c.zk.Create(filepath.Join("/lock", file.GUID), nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			return -1, fmt.Errorf("error: zookeeper create lock dir error, reason is %v", err)
		}
		// lock writelock or
		if write {
			err = c.LockWrite(file)
			if err != nil {
				return -1, err
			}
		} else {
			// NOT write, so Read!
			err = c.LockRead(file)
			if err != nil {
				return -1, err
			}
		}
		c.files[c.count] = file
		c.writable[c.count] = write
		c.count++
		return c.count - 1, nil
		// return something
	} else if !create && !exists {
		// Unlock zk lock
		return -1, fmt.Errorf("open: No file exits and has no rights to create new file")
	}
	// if exists
	// Unlock zk lock
	if err != nil {
		panic(err)
	}
	data, _, err := c.zk.Get(path)
	if err != nil {
		return fmt.Printf("open: cannot get from zookeeper, reason is %v", err)
	}
	var inode *iNode
	err = decodeMsgPack(data, &inode)
	if err != nil {
		return fmt.Printf("error: cannot decode file, reason is %v", err)
	}
	// if it is not a file...
	if inode.IsDir {
		return fmt.Printf("open: cannot open a directory")
	}
	file := inode.toFile()
	if write {
		err = c.LockWrite(file)
		if err != nil {
			return -1, err
		}
	} else {
		// NOT write, so Read!
		err = c.LockRead(file)
		if err != nil {
			return -1, err
		}
	}
	c.files[c.count] = inode.toFile()
	c.writable[c.count] = write
	c.count++
	return c.count - 1, nil
}

func (c *PuddleClient) Close(fd int) error {
	if file, find := c.files[fd]; find {
		// if writable, flush; if not, just remove from map
		if !c.writable[fd] {
			delete(c.files, fd)
			delete(c.writable, fd)
			err := c.Unlock(file, c.lockPaths[fd])
			if err != nil {
				return fmt.Errorf("close: unlock (read) error, reason is %v", err)
			}
			delete(c.lockPaths, fd)
		} else {
			// Writable
			data, err := encodeMsgPack(file.iNode)
			if err != nil {
				return fmt.Errorf("error: cannot encode")
			}
			exist, _, err := c.zk.Exists(file.FileName)
			if exist {
				_, err = c.zk.Set(file.FileName, data.Bytes(), -1)
			} else {
				_, err = c.zk.Create(file.FileName, data.Bytes(), 0, zk.WorldACL(zk.PermAll))
			}
			if err != nil {
				return fmt.Errorf("close: zookeeper cannot set or create, reason is %v", err)
			}
			//err = c.Unlock(file, c.lockPaths[fd])
			//if err != nil {
			//	return fmt.Errorf("close: unlock (writable) error, reason is %v", err)
			//}
			//delete(c.lockPaths, fd)
			delete(c.files, fd)
			delete(c.writable, fd)
			err = c.Unlock(file, c.lockPaths[fd])
			//fmt.Printf("write has deleted%v, c.lockPaths[fd] = %v, fd = %v\n", err, c.lockPaths[fd], fd)
			if err != nil {
				return fmt.Errorf("close: unlock (writable) error, reason is %v", err)
			}
			delete(c.lockPaths, fd)
		}
	} else {
		return fmt.Errorf("close: this file has not been opened or not exists")
	}
	return nil
}

func (c *PuddleClient) Read(fd int, offset, size uint64) ([]byte, error) {
	if file, find := c.files[fd]; find {
		bytes, err := file.ReadFile(offset, size, c)
		if err != nil {
			return nil, fmt.Errorf("read: cannot read file, reason is %v", err)
		}
		return bytes, nil
	} else {
		return nil, fmt.Errorf("read: this file has not been opened")
	}
}

// So... `fd` is updated when a file is "opened" or "closed"??????????????????????????????????????????????
// in the map?????????????????????//
func (c *PuddleClient) Write(fd int, offset uint64, data []byte) error {
	// Write Lock:
	// In our strategy, `Read` operation **can** read half `Write` file (between two operations of `Write`), but
	// there should not be any data race. Since PuddleStore adopts copy-on-write,`Read` operation should not has
	// any data race.

	if file, find := c.files[fd]; find {
		if !c.writable[fd] {
			return fmt.Errorf("write: cannot write to an unwritable file")
		}
		err := file.WriteFile(offset, data, c)
		if err != nil {
			return fmt.Errorf("write: write into file error, reason is %v", err)
		}
	} else {
		return fmt.Errorf("write: this file has not been opened")
	}

	return nil
}

func (c *PuddleClient) Mkdir(path string) error {
	// `findParentDir` also can examine if the path is valid
	parentDir, err := findParentDir(path)
	if err != nil {
		return fmt.Errorf("mkdir: error when find parent dir, reason is %v, zk connection may get lost", err)
	}
	// Check if there exists parent path
	exists, _, err := c.zk.Exists(parentDir)
	if err != nil {
		return fmt.Errorf("mkdir: error when find parent dir, reason is %v, zk connection may get lost", err)
	}
	if !exists {
		return fmt.Errorf("mkdir: There's no parent directory")
	}
	// Check if there exits same file
	// If directory name has "/" at the last character
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	exists, _, err = c.zk.Exists(path)
	if err != nil {
		return err
	}
	if exists {
		// Duplicate dir
		return fmt.Errorf("mkdir: cannot create directory ‘%v/’: File exists\n", path)
	} else {
		// Create dir
		_, err := CreateDir(path, c)
		if err != nil {
			return err
		}
		return nil
	}
}

func (c *PuddleClient) Remove(path string) error {
	exists, _, err := c.zk.Exists(path)
	if err != nil {
		return fmt.Errorf("remove: cannot find the target, reason is %v", err)
	}
	if !exists {
		return fmt.Errorf("remove: cannot find target object, maybe not exists")
	}
	codedBytes, _, err := c.zk.Get(path)
	if err != nil {
		return fmt.Errorf("remove: error when zookeeper is trying to find target, %v", err)
	}
	var target *iNode
	err = decodeMsgPack(codedBytes, &target)
	if err != nil {
		return fmt.Errorf("remove: error when decoding... %v", err)
	}
	// If it is open????????????????????????????????????????????????????????

	return target.Remove(c)
}

func (c *PuddleClient) List(path string) ([]string, error) {
	// If path has slash
	if len(path) != 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	children, _, err := c.zk.Children(path)
	if err != nil {
		return nil, fmt.Errorf("list: error when get children, reason is %v", err)
	}
	return children, nil
}

/****************** Helper functions *********************/
func (c *PuddleClient) LockWrite(file *File) error {
	// If writable and ready to write, add lock
	exists, _, err := c.zk.Exists(filepath.Join("/lock", file.GUID))
	if err != nil {
		return fmt.Errorf("error: zookeeper fail to find target GUID dir, reason is %v", err)
	}
	if !exists {
		return fmt.Errorf("error: lock directory missing")
	}
	// Create lock node https://zookeeper.apache.org/doc/r3.6.0/recipes.html#sc_recipes_Locks
	// Careful here! `thisLock` has full path name!
	thisLock, err := c.zk.CreateProtectedEphemeralSequential(filepath.Join("/lock", file.GUID, "write-"),
		nil, zk.WorldACL(zk.PermAll))
	if err != nil {
		return fmt.Errorf("error: cannot create lock znode, reason is %v", err)
	}
Loop:
	locks, _, err := c.zk.Children(filepath.Join("/lock", file.GUID))
	if err != nil {
		// Reverse lock
		c.zk.Delete(thisLock, -1)
		return fmt.Errorf("error: cannot get locks, reason is %v", err)
	}
	isMin, locks, idx := isSeqMinimum(thisLock, locks)
	if !isMin {
		//fmt.Printf("locks is %v\n", locks)
		//fmt.Printf("I'm watching %v\n", locks[idx-1])
		exist, _, watcher, err := c.zk.ExistsW(filepath.Join("/lock", file.GUID, locks[idx-1]))
		if err != nil {
			// Reverse lock
			c.zk.Delete(thisLock, -1)
			return err
		}
		if !exist {
			goto Loop
		}
		// Not sure below
		<-watcher
		goto Loop // this line could be omitted
	}
	c.lockPaths[c.count] = thisLock
	return nil
}

func (c *PuddleClient) LockRead(file *File) error {
	// If not writable and ready to read, add lock
	exists, _, err := c.zk.Exists(filepath.Join("/lock", file.GUID))
	if err != nil {
		return fmt.Errorf("error: zookeeper fail to find target GUID dir, reason is %v", err)
	}
	if !exists {
		return fmt.Errorf("error: lock disrectory missing")
	}
	// Create lock node https://zookeeper.apache.org/doc/r3.6.0/recipes.html#sc_recipes_Locks
	// Careful here! `thisLock` has full path name!
	thisLock, err := c.zk.CreateProtectedEphemeralSequential(filepath.Join("/lock", file.GUID, "read-"),
		nil, zk.WorldACL(zk.PermAll))
	if err != nil {
		return fmt.Errorf("error: cannot create lock znode, reason is %v", err)
	}
Loop:
	locks, _, err := c.zk.Children(filepath.Join("/lock", file.GUID))
	if err != nil {
		// Reverse lock
		c.zk.Delete(thisLock, -1)
		return fmt.Errorf("error: cannot get locks, reason is %v", err)
	}
	isMin, locks, _ := isSeqMinimumRead(thisLock, locks)
	if !isMin {
		//fmt.Printf("file GUID is %v\n", file.GUID)
		//fmt.Printf("len(locks) = %v, idx = %v\n", len(locks), /*idx - 1*/0)
		//fmt.Printf("I'm watching %v\n", locks[0])
		exist, _, watcher, err := c.zk.ExistsW(filepath.Join("/lock", file.GUID, locks[ /*idx-1*/ 0]))
		if err != nil {
			// Reverse lock
			c.zk.Delete(thisLock, -1)
			return err
		}
		if !exist {
			goto Loop
		}
		// Not sure below
		fmt.Printf("!\n")
		<-watcher
		fmt.Printf("!\n")
		goto Loop // this line could be omitted
	}
	c.lockPaths[c.count] = thisLock
	return nil
}

func (c *PuddleClient) Unlock(file *File, thisLock string) error {
	err := c.zk.Delete(thisLock, -1)
	if err != nil {
		return err
	}
	return nil
}

func findParentDir(path string) (string, error) {
	if path[0] != '/' {
		return path, fmt.Errorf("error: File path invalid. Hint: must be absolute path")
	} else if path[0] == '/' && len(path) == 1 {
		return path, fmt.Errorf("error: File path invalid. Cannot create root path")
	}

	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return path[0 : i+1], nil
			}
			return path[0:i], nil
		}
	}
	return path, fmt.Errorf("error: File path invalid")
}

func isSeqMinimum(thisLock string, locks []string) (bool, []string, int) {
	sort.Slice(locks, func(i, j int) bool {
		return strings.Compare(locks[i][len(locks[i])-10:], locks[j][len(locks[j])-10:]) <= 0
	})

	//if locks[0] == thisLock {
	if strings.Compare(locks[0][len(locks[0])-10:], thisLock[len(thisLock)-10:]) == 0 {
		return true, locks, 0
	} else {
		var idx int
		for i, value := range locks {
			if strings.Compare(value[len(value)-10:], thisLock[len(thisLock)-10:]) == 0 {
				idx = i
				break
			}
		}
		return false, locks, idx
	}
}

func isSeqMinimumRead(thisLock string, locks []string) (bool, []string, int) {
	idx := 0
	for i := 0; i < len(locks); i++ {
		if locks[i][len(locks[i])-12] == 'e' {
			locks[idx] = locks[i]
			idx++
		}
	}
	locks = locks[:idx]
	sort.Slice(locks, func(i, j int) bool {
		return strings.Compare(locks[i][len(locks[i])-10:], locks[j][len(locks[j])-10:]) <= 0
	})

	//if locks[0] == thisLock OR currently there's no write lock {
	if len(locks) == 0 || strings.Compare(locks[0][len(locks[0])-10:], thisLock[len(thisLock)-10:]) == 0 {
		return true, locks, 0
	} else {
		var idx int
		for i, value := range locks {
			if strings.Compare(value[len(value)-10:], thisLock[len(thisLock)-10:]) == 0 {
				idx = i
				break
			}
		}
		if strings.Compare(locks[0][len(locks[0])-10:], thisLock[len(thisLock)-10:]) > 0 {
			return true, locks, idx
		}
		return false, locks, idx
	}
}
