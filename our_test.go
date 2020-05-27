package puddlestore

import (
	"fmt"
	"testing"
	"time"
	// "github.com/samuel/go-zookeeper/zk"
)

func writeFile(client Client, path string, offset uint64, data []byte) error {
	fd, err := client.Open(path, true, true)
	// Double open
	fd, err = client.Open(path, true, true)
	if err != nil {
		return err
	}
	defer client.Close(fd)

	return client.Write(fd, offset, data)
}

func readFile(client Client, path string, offset, size uint64) ([]byte, error) {
	fd, err := client.Open(path, true, false)
	if err != nil {
		return nil, err
	}
	defer client.Close(fd)

	return client.Read(fd, offset, size)
}

func TestReadWrite(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	in := "testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest"
	if err := writeFile(client, "/a", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}

	var out []byte
	if out, err = readFile(client, "/a", 0, 80); err != nil {
		t.Fatal(err)
	}
	temp := string(out)
	t.Logf("output is %v", temp)
	if in != string(out) {
		t.Fatalf("Expected: %v, Got: %v", in, string(out))
	}

	err = client.Remove("/a")
	if err != nil {
		t.Fatal(err)
	}
	out, err = readFile(client, "/a", 0, 5)
	if out != nil {
		t.Fatalf("Should not get any value")
	}
}

func TestCodec(t *testing.T) {
	inode := &iNode{
		FileName:  "/a/hi",
		Size:      64,
		BlockSize: 64,
		Blocks:    nil,
		IsDir:     false,
		Conn:      nil,
	}
	data, err := encodeMsgPack(inode)
	if err != nil {
		t.Errorf("Cannot encode")
	}
	var retNode *iNode
	err = decodeMsgPack(data.Bytes(), &retNode)
	if err != nil {
		t.Errorf("Cannot decode")
	}
	if retNode.FileName != inode.FileName {
		t.Fatalf("Failure encode or decode")
	}
}

func TestSeqCompare1(t *testing.T) {
	//str1 := "123456-StandingBook:1234-0000000001"
	//str2 := "123456-StandingBook:1234-0000000005"
	//str3 := "123456-StandingBook:1234-0000000019"
	//str4 := "123456-StandingBook:1234-0000000234"
	//str5 := "123456-StandingBook:1234-0000124687"
	str := "_c_49d4f619cf9fa9a19ea7ef9970c7c027-write-0000000000"
	this := "/lock/c9bb77e0-dce4-42bd-bd35-6409cb95a2e9/_c_49d4f619cf9fa9a19ea7ef9970c7c027-write-0000000001"

	var strs []string
	//strs = append(strs, str5)
	//strs = append(strs, str3)
	//strs = append(strs, str4)
	//strs = append(strs, str1)
	//strs = append(strs, str2)
	strs = append(strs, str)

	isMin, strs, _ := isSeqMinimum(this, strs)
	t.Logf("bool is %v, strs are %v", isMin, strs)
}

func TestSeqCompare2(t *testing.T) {
	str1 := "123456-StandingBook:write-0000000001"
	str2 := "123456-StandingBook:read-0000000005"
	str3 := "123456-StandingBook:write-0000000019"
	str4 := "123456-StandingBook:read-0000000234"
	str5 := "123456-StandingBook:read-0000124687"
	str6 := "_c_49d4f619cf9fa9a19ea7ef9970c7c027-write-0000000003"
	this := "/lock/c9bb77e0-dce4-42bd-bd35-6409cb95a2e9/_c_49d4f619cf9fa9a19ea7ef9970c7c027-read-0000000000"

	var strs []string
	strs = append(strs, str5)
	strs = append(strs, str3)
	strs = append(strs, str4)
	strs = append(strs, str1)
	strs = append(strs, str2)
	strs = append(strs, str6)

	isMin, strs, idx := isSeqMinimumRead(this, strs)
	t.Logf("bool is %v, strs are %v, idx is %v", isMin, strs, idx)
}

func TestWrite1(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	//write less than 64.
	in := "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabc"
	if err = writeFile(client, "/b", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}

	err = client.Remove("/b")
	if err != nil {
		t.Fatal(err)
	}
	out, err := readFile(client, "/b", 0, 5)
	if out != nil {
		t.Fatalf("Should not get any value")
	}

}

func TestWrite2(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	//write more than 64.
	in := "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcda"
	if err = writeFile(client, "/c", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}
	err = client.Remove("/c")
	if err != nil {
		t.Fatal(err)
	}
	out, err := readFile(client, "/c", 0, 5)
	if out != nil {
		t.Fatalf("Should not get any value")
	}
}

func TestWrite3(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	//write more than 128.
	in := "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdaabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcda"
	if err = writeFile(client, "/d", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}
	err = client.Remove("/d")
	if err != nil {
		t.Fatal(err)
	}
	out, err := readFile(client, "/d", 0, 5)
	if out != nil {
		t.Fatalf("Should not get any value")
	}
}

//A client’s view of the file (and the fd) will not change after Open.
//Do not re-read inode from zk after Open

func TestRead1(t *testing.T) {
	//write less than 64, read 0-64
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	//write more than 64.
	in := "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabc"
	if err = writeFile(client, "/e", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}

	var out []byte
	if out, err = readFile(client, "/e", 0, 64); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Here!!!!!!")
	temp := string(out)
	t.Logf("output is %v", temp)
	if "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabc" != string(out) {
		t.Fatalf("Expected: %v, Got: %v", "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd", string(out))
	}

	err = client.Remove("/e")
	if err != nil {
		t.Fatal(err)
	}
	out, err = readFile(client, "/e", 0, 5)
	if out != nil {
		t.Fatalf("Should not get any value")
	}
}

func TestRead2(t *testing.T) {
	//write less than 64, read 10-64
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	//write more than 64.
	in := "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcl"
	if err = writeFile(client, "/f", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}

	var out []byte
	if out, err = readFile(client, "/f", 10, 128); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Here!!!!!!")
	temp := string(out)
	t.Logf("output is %v", temp)
	shouldbe := "cdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcl"
	t.Logf("length is %v", len(shouldbe))
	if shouldbe != string(out) {
		t.Fatalf("Expected: %v, Got: %v", shouldbe, string(out))
	}

	err = client.Remove("/f")
	if err != nil {
		t.Fatal(err)
	}
	out, err = readFile(client, "/f", 0, 5)
	if out != nil {
		t.Fatalf("Should not get any value")
	}
}

func TestRead3(t *testing.T) {
	//write more than 128, read 0-130
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	in := "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdaabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcda"
	if err = writeFile(client, "/g", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}

	var out []byte
	if out, err = readFile(client, "/g", 0, 130); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Here!!!!!!")

	err = client.Remove("/g")
	if err != nil {
		t.Fatal(err)
	}
	out, err = readFile(client, "/g", 0, 5)
	if out != nil {
		t.Fatalf("Should not get any value")
	}
}

func TestRead4(t *testing.T) {
	//write more than 128, read 10-120
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	in := "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdaabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcda"
	if err = writeFile(client, "/h", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}

	var out []byte
	if out, err = readFile(client, "/h", 10, 120); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Here!!!!!!")

	err = client.Remove("/h")
	if err != nil {
		t.Fatal(err)
	}
	out, err = readFile(client, "/h", 0, 5)
	if out != nil {
		t.Fatalf("Should not get any value")
	}
}

func TestRead5(t *testing.T) {
	//write more than 128, read 128-160
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	in := "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdaabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdaabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd"
	if err = writeFile(client, "/i", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}

	var out []byte
	if out, err = readFile(client, "/i", 128, 160); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Here!!!!!!")

	err = client.Remove("/i")
	if err != nil {
		t.Fatal(err)
	}
	out, err = readFile(client, "/i", 0, 5)
	if out != nil {
		t.Fatalf("Should not get any value")
	}
}

func TestDir1(t *testing.T) {
	//mkDir, create file, write and read.
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Mkdir("/j")
	if err != nil {
		t.Fatal(err)
	}
	in := "abcdabcd"
	if err = writeFile(client, "/j/c", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}
	var out []byte
	if out, err = readFile(client, "/j/c", 0, 10); err != nil {
		t.Fatal(err)
	}
	temp := string(out)
	t.Logf("output is %v", temp)
	if in != string(out) {
		t.Fatalf("Expected: %v, Got: %v", in, string(out))
	}
	err = client.Remove("/j/c")
	if err != nil {
		t.Fatal(err)
	}
	// err = client.Remove("/j")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	err = client.Remove("/j")
	if err != nil {
		t.Fatalf("cannot remove /j, err is %v", err)
	}
}

func TestDir2(t *testing.T) {
	//mkDir, create file, write and read.
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Mkdir("/k")
	if err != nil {
		t.Fatal(err)
	}
	in := "abcdabcd"
	//didn't create the dir
	if err = writeFile(client, "/k/b/file", 0, []byte(in)); err == nil {
		t.Fatal("There should be error, cuz didn't create the dir yet.")
	}

	// create the dir.
	err = client.Mkdir("/k/b")
	if err != nil {
		t.Fatal(err)
	}
	if err = writeFile(client, "/k/b/file", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}
	var out []byte
	if out, err = readFile(client, "/k/b/file", 0, 10); err != nil {
		t.Fatal(err)
	}
	temp := string(out)
	t.Logf("output is %v", temp)
	if in != string(out) {
		t.Fatalf("Expected: %v, Got: %v", in, string(out))
	}

	err = client.Remove("/k")
	if err != nil {
		t.Fatalf("cannot remove /k, err is %v", err)
	}

}

func TestDir3(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Mkdir("/ka/")
	if err != nil {
		t.Fatal(err)
	}
	// if err==nil{
	// 	t.Error("Should have an error in making directories!")
	// }
	err = client.Remove("/ka")
	if err != nil {
		t.Fatalf("cannot remove /j, err is %v", err)
	}
}

func TestDirSlash(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Mkdir("/kb/")
	if err != nil {
		t.Fatal(err)
	}
	in := "abcdabcd"
	if err = writeFile(client, "/kb/c/", 0, []byte(in)); err == nil {
		// t.Error("Should have an error, file cannot end with /")
		client.Remove("/kb")
		t.Fatal(err)
	}
	client.Remove("/kb")
}

//Design Principles:

//A client’s view of the file (and the fd) will not change after Open.
//Do not re-read inode from zk after Open

func TestViewNotChange(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()
	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Mkdir("/l")
	if err != nil {
		t.Fatal(err)
	}
	in := "abcdabcdabcdabcdabcd"
	if err = writeFile(client, "/l/file", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}
	fd, err := client.Open("/l/file", true, true)
	if err != nil {
		t.Fatal(err)
	}
	var out []byte
	out, err = client.Read(fd, 0, 20)
	if err != nil {
		t.Fatal(err)
	}

	temp := string(out)
	t.Logf("output is %v", temp)
	if in != string(out) {
		t.Fatalf("Expected: %v, Got: %v", in, string(out))
	}
	err = client.Remove("/l")
	if err != nil {
		t.Fatalf("cannot remove /l, err is %v", err)
	}
}

//Test List
func TestListRemove(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()
	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Mkdir("/m")
	if err != nil {
		t.Fatal(err)
	}
	in := "abcdabcdabcdabcdabcd"
	if err = writeFile(client, "/m/file", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}
	// var dir []string
	dirs, err := client.List("/m")
	if dirs[0] != "file" {
		t.Fatal(err)
	}
	// fmt.Print(dirs)
	err = client.Remove("/m/file")
	dirs, err = client.List("/m")
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 0 {
		t.Fatal(err)
	}
	err = client.Remove("/m")
	if err != nil {
		t.Fatalf("cannot remove /m, err is %v", err)
	}
}

//Can file be created under file???
func TestFileUnderFile(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()
	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Mkdir("/n")
	in := "abcdabcdabcdabcdabcd"
	if err = writeFile(client, "/n/file", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}
	fd, err := client.Open("/n/file", true, true)
	if err != nil {
		t.Fatal(err)
	}
	var out []byte
	out, err = client.Read(fd, 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	temp := string(out)
	t.Logf("output is %v", temp)
	if in != string(out) {
		t.Fatalf("Expected: %v, Got: %v", in, string(out))
	}
	// if err = writeFile(client, "/n/file/file1", 0, []byte(in)); err != nil {
	// 	t.Fatal(err)
	// }
	// fd, err = client.Open("/n/file/file1", true, true)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// out, err = client.Read(fd, 0, 20)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// temp = string(out)
	// t.Logf("output is %v", temp)
	// if in != string(out) {
	// 	t.Fatalf("Expected: %v, Got: %v", in, string(out))
	// }

	fd, _ = client.Open("/file", true, true)
	client.Write(fd, 0, []byte("i'm a file"))
	client.Close(fd)

	fd, err = client.Open("/file/file2", true, true)
	if err == nil {
		t.Fatalf("Should throw error here 'cause file cannot be created under file")
	}
	client.Remove("/n")
	client.Remove("/file")
}

func TestRemove(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()
	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	client.Remove("/na")
	client.Mkdir("/na")
	client.Mkdir("/na/a")
	client.Mkdir("/na/a/b")
	client.Mkdir("/na/a/b/c")
	client.Mkdir("/na/a/b/c/d")
	client.Mkdir("/na/a/b/c/d/e")
	client.Mkdir("/na/a/b/c/d/e/f")
	client.Mkdir("/na/a/b/c/d/e/f/g")
	in := "abcdabcdabcdabcdabcd"
	if err = writeFile(client, "/na/a/b/c/d/e/f/g/file", 0, []byte(in)); err != nil {
		t.Fatal(err)
	}
	fd, err := client.Open("/na/a/b/c/d/e/f/g/file", true, true)
	if err != nil {
		t.Fatal(err)
	}
	var out []byte
	out, err = client.Read(fd, 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	temp := string(out)
	t.Logf("output is %v", temp)
	if in != string(out) {
		t.Fatalf("Expected: %v, Got: %v", in, string(out))
	}
	client.Close(fd)

	//client.Remove("/na/a/b/c/d/e")
	client.Remove("/na")
	fd, err = client.Open("/na/a/b/c/d/e/f/g/file", true, true)
	if err == nil {
		t.Fatalf("There should be no file")
	}
	//out, err = client.Read(fd, 0, 20)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//temp = string(out)
	//t.Logf("output is %v", temp)
	//if len(temp) != 0 {
	//	t.Fatalf("No content can be left! Error!")
	//}
	//t.Logf("output is %v", temp)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//client.Remove("/na")
}

func TestRemove2(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()
	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	client.Mkdir("/Hideo")
	client.Mkdir("/Hideo/Kojima")
	fd, err := client.Open("/Hideo/Kojima/MG1", true, true)
	client.Write(fd, 0, []byte("MG1 is not on PC"))
	client.Close(fd)
	client.Mkdir("/Hideo/Kojima/MGSeries")
	fd, err = client.Open("/Hideo/Kojima/MGSeries/MGS1", true, true)
	client.Write(fd, 0, []byte("MGS1 is not on PC"))
	client.Close(fd)
	fd, err = client.Open("/Hideo/Kojima/MGSeries/MGS2", true, true)
	client.Write(fd, 0, []byte("MGS2 is on PC"))
	client.Close(fd)
	fd, err = client.Open("/Hideo/production", true, true)
	client.Write(fd, 0, []byte("MGS is Hideo Kojima Production"))
	client.Close(fd)

	list, _ := client.List("/Hideo/Kojima/MGSeries")
	t.Logf("%v", list)

	list, _ = client.List("/Hideo/Kojima/MGSeries/")
	t.Logf("%v", list)

	err = client.Remove("/Hideo")
	if err != nil {
		t.Fatalf("remove error")
	}
}

func TestFailNode(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	//defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	fd, _ := client.Open("/a", true, true)
	data := "This is a tapestry fail test"
	client.Write(fd, 0, []byte(data))
	client.Close(fd)

	cluster.nodes[0].GracefulExit()
	time.Sleep(time.Millisecond * 1000)

	fd, _ = client.Open("/a", false, false)
	data2, _ := client.Read(fd, 0, 64)
	t.Logf("%v", string(data2))
	if string(data2) != data {
		t.Fatalf("After a node fail, client cannot get correct answer")
	}

	client.Remove("/a")

	cluster.nodes[1].GracefulExit()
}

func TestFindParentPath(t *testing.T) {
	parentDir, err := findParentDir("/")
	if err == nil {
		t.Fatalf("There should be an error")
	}

	parentDir, err = findParentDir("/miku")
	if err != nil {
		t.Fatalf("There should be no error")
	}
	if parentDir != "/" {
		t.Fatalf("\"/miku\"'s parent path should be \"/\"")
	}

	parentDir, err = findParentDir("/Hideo/Kojima")
	if err != nil {
		t.Fatalf("There should be no error")
	}
	if parentDir != "/Hideo" {
		t.Fatalf("\"/Hideo/Kojima\"'s parent path should be \"/Hideo\"")
	}

	parentDir, err = findParentDir("/Hideo/Kojima/")
	if err != nil {
		t.Fatalf("There should be no error")
	}
	if parentDir != "/Hideo" {
		t.Fatalf("\"/Hideo/Kojima/\"'s parent path should be \"/Hideo\"")
	}
}

func TestReadwithCreate(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()
	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	fd, err := client.Open("/a", true, true)
	client.Write(fd, 0, []byte("hi, it is test TestReadwithCreate"))
	client.Close(fd)

	fd, err = client.Open("/a", true, false)
	err = client.Write(fd, 0, []byte("i'm consuming you"))
	if err == nil {
		t.Fatalf("you cannot write!")
	}
	client.Close(fd)

	client.Remove("/a")
}
