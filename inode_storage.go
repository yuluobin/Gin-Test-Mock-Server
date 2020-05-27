package puddlestore

import (
	"bytes"
	"fmt"
	"github.com/brown-csci1380-s20/puddlestorenew-puddlestorenew-cwang147-byu18-mxu57/tapestry"
	"github.com/samuel/go-zookeeper/zk"
	uuid "github.com/satori/go.uuid"
	"math/rand"
)

// Implement "what is a file"
// So... iNode includes (several) DataBlock?
type DataBlock struct {
	Capacity uint64
	GUID     string
	Data     []byte
}

type iNode struct {
	FileName  string // This is unique ID. Don't forget to add path! (So this is also a `path`)
	Size      uint64
	BlockSize uint64
	Blocks    map[int]string // How to make it a map???????????????????????????????? (Maybe just 1, 2, 3)
	IsDir     bool           // Judge if the current iNode is a `Directory` or a `File`
	Conn      *zk.Conn
	GUID      string
}

func (node *iNode) toFile() *File {
	return &File{*node}
}

func (node *iNode) toDir() *Dir {
	return &Dir{*node}
}

func newGUID() string {
	// How to choose from uuid??????????????????????????????? V1, V2, V3, V4, V5 Update: Confirmed!
	guid, err := uuid.NewV4()
	if err != nil {
		fmt.Errorf("error: fail to generate GUID")
	}
	return guid.String()
}

func newDB(BlockSize uint64) *DataBlock {
	var resDB DataBlock
	resDB.GUID = newGUID()
	// `BlockSize` should be delivered in by `newiNode` (or something else?)
	resDB.Data = make([]byte, BlockSize)
	resDB.Capacity = BlockSize
	return &resDB
}

func newiNode(conn *zk.Conn, fileName string, isDir bool, BlockSize uint64) (*iNode, error) {
	var resiNode iNode
	// Since `fileName` is actually a file path, examine if it is valid (Should I do that????????) SKIP it for now
	resiNode.FileName = fileName
	resiNode.IsDir = isDir
	resiNode.Size = 0
	resiNode.Conn = conn
	resiNode.Blocks = make(map[int]string)
	resiNode.BlockSize = BlockSize
	resiNode.GUID = newGUID()
	// Store **this** iNode in tapestry // Don't do that! (And I'm sure sure here)
	//tapClient, err := ChooseTapestry(conn)
	//if err != nil {
	//	return nil, err
	//}
	//data, err := encodeMsgPack(fileName)
	//if err != nil {
	//	return nil, err
	//}
	//// don't put inode into tapestry!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//err = tapClient.Store(fileName /*path | ID*/, data.Bytes())
	//if err != nil {
	//	return nil, err
	//}
	// Register **this** iNode in zk
	// Well... I still not sure how to use `Create`
	// I think `1` is **seq**???????????????????????
	// According to the slides, there should be something with **version**?????????????????
	//data, err := encodeMsgPack(resiNode)
	//if err != nil {
	//	return nil, err
	//}
	//_, err = conn.Create(fileName, data.Bytes(), 0, zk.WorldACL(zk.PermAll)) // Should `data` be nil?????????
	//if err != nil {
	//	return nil, err
	//}
	return &resiNode, nil
}

func (f *iNode) Remove(c *PuddleClient) error {
	if f.IsDir {
		return f.toDir().Remove(c)
	} else {
		return f.toFile().Remove(c)
	}
}

// Evenly distribute load balance
func ChooseTapestrys(c *PuddleClient) []*tapestry.Client {
	//children, _, err := conn.Children("/tapestry")
	//if err != nil {
	//	return nil, err
	//}
	// Load balance. Simply generate a random int
	var res []*tapestry.Client
	tapSet := make(map[int]*tapestry.Client)
	for len(tapSet) < c.NumReplicas {
		idx := rand.Intn(len(c.tapClients))
		if _, ok := tapSet[idx]; ok {
			continue
		}
		tapSet[idx] = c.tapClients[idx]
		res = append(res, c.tapClients[idx])
	}
	//fmt.Printf("When I store, I store to %v nodes\n", len(tapSet))
	//port, _, err := conn.Get(filepath.Join("/tapestry",children[idx]))
	//if err != nil {
	//	return nil, err
	//}
	//tapClient, err := tapestry.Connect(string(port))
	//if err != nil {
	//	return nil, err
	//}
	//return tapClient, nil
	return res
}

func ChooseTapestry(c *PuddleClient) *tapestry.Client {
	idx := rand.Intn(len(c.tapClients))
	//fmt.Printf("Guess what I choose? %v \n", idx)
	return c.tapClients[idx]
}

func getBlockByGUID(GUID string, c *PuddleClient) (*DataBlock, error) {
	tapClient := ChooseTapestry(c)
	codedData, err := tapClient.Get(GUID)
	if err != nil {
		//fmt.Printf("hi\n\n\n")
		return nil, err
	}
	// Deleted block will be... empty?????????????????????????????????
	// Need to modify..................................................
	if len(codedData) == 0 {
		return nil, fmt.Errorf("error: the datablock you want to find has been deleted")
	}
	var BDTemp DataBlock
	err = decodeMsgPack(codedData, &BDTemp)
	if err != nil {
		return nil, err
	}
	return &BDTemp, nil
}

// Fill an empty block (or not completely empty) with bytes from buffer
// Since `Write` doesn't truncate files, we should fill the original data after????????????
// https://piazza.com/class/k5illljohg02m8?cid=1174
func (dataBlock *DataBlock) fill(offset uint64, buf *bytes.Buffer) uint64 {
	byteToFill := dataBlock.Capacity - offset
	byteActualFill := buf.Next(int(byteToFill)) // API user...
	if len(byteActualFill) < int(byteToFill) {
		// Not sure if the byte at `offset` would be covered by new value...???????????????
		leftover := dataBlock.Data[(int(offset) + len(byteActualFill)):]
		dataBlock.Data = append(dataBlock.Data[:offset], append(byteActualFill, leftover...)...)
	} else {
		dataBlock.Data = append(dataBlock.Data[:offset], byteActualFill...)
	}
	return uint64(len(byteActualFill))
}
