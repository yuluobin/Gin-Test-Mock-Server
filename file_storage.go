package puddlestore

import (
	"bytes"
	"fmt"
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

// `File` is actually a type of `iNode`, of course with `IsDir` set to `false`
type File struct {
	iNode
}

func CreateFile(FileName string, c *PuddleClient) (*File, error) {
	file, err := newiNode(c.zk, FileName, false, c.BlockSize)
	if err != nil {
		return nil, err
	}
	return &File{*file}, nil
}

/*
Write into a file Procedure
- Examine blah blah whatever...
- Put the data (bytes) I need to write into a buffer (locally)
- Use a loop to create blocks and put data in the buffer into blocks
- In a single process (in the loop):
					- Create a block
					- Fill it with buffer
					- If the block has remaining space, fill it with $0$
					- Add the new block into tapestry
					- Update metadata (the iNode)
- If I need to **modify** a block:
					- Retrieve the block from tapestry
					- Update its GUID (just generate one)
					- Fill it with buffer
					- Add it into tapestry
*/
// Do I need to delete block in tapestry when retrieve?????????????????????????????????????/
func (f *File) WriteFile(offset uint64, data []byte, c *PuddleClient) error {
	if len(data) == 0 {
		return nil
	}
	// Use a buffer to store the data I need to insert into block(s)
	var buf bytes.Buffer
	if offset > f.Size {
		for i := f.Size; i < offset; i++ {
			buf.WriteByte(byte(0))
		}
		offset = f.Size
	}
	buf.Write(data)

	// Create (Reuse) and write into block
	// index is from 0!!!!!!!!!!!!
	for idx := offset / f.BlockSize; buf.Len() > 0; idx++ {
		if idx < uint64(len(f.Blocks)) {
			// Modify the original block
			// Get the original block
			DBTemp, err := f.getBlockFromTapestry(idx, c)
			if err != nil {
				return fmt.Errorf("error: cannot get block from tapestry, reason is %v", err)
			}
			// The original block has original data in it
			DBTemp.GUID = newGUID()
			length := DBTemp.fill(offset%f.BlockSize, &buf)
			offset += length
			err = f.addBlockToTapestry(DBTemp, c)
			if err != nil {
				return fmt.Errorf("error: cannot add block into tapestry, reason is %v", err)
			}
			f.Blocks[int(idx)] = DBTemp.GUID

		} else {
			DBTemp := newDB(f.BlockSize)
			length := DBTemp.fill(0, &buf)
			offset += length
			err := f.addBlockToTapestry(DBTemp, c)
			if err != nil {
				return fmt.Errorf("error: cannot add block into tapestry, reason is %v", err)
			}
			// Update File's metadata
			f.Blocks[int(idx)] = DBTemp.GUID
		}
	}

	if offset > f.Size {
		f.Size = offset
	}

	return nil
}

func (f *File) ReadFile(offset, size uint64, c *PuddleClient) ([]byte, error) {
	//sizeReturn := size
	//if size > f.Size - offset {
	//	sizeReturn = f.Size - offset
	//}
	if offset > f.Size || size == 0 {
		return []byte{}, nil
	}
	// If offset is larger than size of the file, just return an empty bytes with no error
	// ???????????????????????????????????????????????????????????????????????????????????
	//if sizeReturn <= 0 {
	//	return []byte{}, nil
	//}
	var buf bytes.Buffer
	fromBlockIdx := offset / f.BlockSize
	fromInBlock := offset % f.BlockSize
	bytesLeft := size
	if f.Size-offset < bytesLeft {
		bytesLeft = f.Size - offset
	}

	for idx := fromBlockIdx; idx <= ((f.Size - 1) / f.BlockSize); idx++ {
		DBTemp, err := f.getBlockFromTapestry(idx, c)
		if err != nil {
			return nil, err
		}
		//fmt.Printf("Now we see what's inside our block: %v - %v \n", DBTemp.GUID, DBTemp.Data)
		readable := f.BlockSize - fromInBlock
		if readable > bytesLeft {
			buf.Write(DBTemp.Data[fromInBlock:(fromInBlock + bytesLeft)])
			bytesLeft = 0
			break
		} else {
			buf.Write(DBTemp.Data[fromInBlock:])
			bytesLeft -= readable
		}

		fromInBlock = 0
	}
	return buf.Bytes(), nil
}

// We should remove metadata and blocks from tapestry
func (f *File) Remove(c *PuddleClient) error {
	// First, remove blocks from tapestry
	for idx := 0; idx <= int(f.Size/f.BlockSize); idx++ {
		//GUID := f.Blocks[idx]
		//tapClients := ChooseTapestry(c)
		//err := tapClient.Store(GUID, []byte{})
		//if err != nil {
		//	return fmt.Errorf("error: cannot store empty value, reason is %v", err)
		//}
		delete(f.Blocks, idx)
	}
	// Second, remove metadata from zookeeper
	err := c.zk.Delete(f.FileName, -1)
	if err != nil {
		return fmt.Errorf("error: cannot delete metadata from zookeeper, reason is %v", err)
	}
	return nil
}

func (f *File) getBlockFromTapestry(idx uint64, c *PuddleClient) (*DataBlock, error) {
	GUID := f.Blocks[int(idx)]
	DBTemp, err := getBlockByGUID(GUID, c)
	if err != nil {
		return nil, err
	}
	return DBTemp, nil
}

func (f *File) addBlockToTapestry(dataBlock *DataBlock, c *PuddleClient) error {
	tapClients := ChooseTapestrys(c)
	encodedData, err := encodeMsgPack(dataBlock)
	if err != nil {
		return err
	}
	for _, tapClient := range tapClients {
		err = tapClient.Store(dataBlock.GUID, encodedData.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}
