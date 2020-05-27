package puddlestore

import (
	"testing"
	"time"
)

func TestDoubleWrite(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	fd, _ := client.Open("/a", true, true)
	data := "This is a clients concurrency test"
	client.Write(fd, 0, []byte(data))
	client.Close(fd)

	fd, _ = client.Open("/a", true, true)
	data = "This is not a client concurrency test"
	client.Write(fd, 10, []byte(data))
	client.Close(fd)

	fd3, _ := client.Open("/a", false, false)
	data3, _ := client.Read(fd3, 0, 64)
	t.Logf("%v", string(data3))

	//if string(data3) != "This is not a client concurrency test" {
	//	t.Fatalf("output should be as the second time's")
	//}

	client.Remove("/a")

}

func TestTwoClientWriteWrite(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client1, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	client2, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	client3, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	// Create a file first
	fd, _ := client1.Open("/a", true, true)
	data := "This is a clients concurrency test"
	client1.Write(fd, 0, []byte(data))
	client1.Close(fd)

	go func() {
		fd1, err := client1.Open("/a", false, true)
		if err != nil {
			t.Fatalf("client1 open fail! %v", err)
		}
		data1 := "!no gnikrow si ecilA"
		err = client1.Write(fd1, 14, []byte(data1))
		if err != nil {
			t.Fatalf("client1 write has problem!, %v", err)
		}
		err = client1.Close(fd1)
		t.Logf("client1 has closed safe and sound")
	}()
	//go func() {
	fd2, err := client2.Open("/a", false, true)
	if err != nil {
		t.Fatalf("client2 open fail! %v", err)
	}
	data2 := "Eve has taken the power!"
	err = client2.Write(fd2, 0, []byte(data2))
	if err != nil {
		t.Fatalf("client2 write has problem!, %v", err)
	}
	err = client2.Close(fd2)
	t.Logf("client2 has closed safe and sound")
	//}()

	// Wait for two gorountines to complete working
	time.Sleep(time.Millisecond * 1000)
	fd3, _ := client3.Open("/a", false, false)
	data3, _ := client3.Read(fd3, 0, 64)
	//if string(data3) == data {
	//	t.Fatalf("error: original data has not been modified")
	//}
	t.Logf("%v", string(data3))
	if string(data3) == "Eve has taken the power!w si ecilA" {
		t.Fatalf("Should be either Alice or Eve, but not BOTH!")
	}

	client3.Close(fd3)
	client1.Remove("/a")
}

func TestTwoClientWriteRead(t *testing.T) {
	cluster, err := CreateCluster(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Shutdown()

	client1, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	client2, err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	// Create a file first
	fd, _ := client1.Open("/a", true, true)
	data := "This is a clients concurrency test"
	err = client1.Write(fd, 0, []byte(data))
	if err != nil {
		t.Fatalf("write fail")
	}
	err = client1.Close(fd)
	if err != nil {
		t.Fatalf("client1 close fail")
	}

	go func() {
		fd1, err := client1.Open("/a", false, true)
		if err != nil {
			t.Fatalf("client1 open fail! %v", err)
		}
		data1 := "!no gnikrow si ecilA"
		err = client1.Write(fd1, 14, []byte(data1))
		if err != nil {
			t.Fatalf("client1 write has problem!, %v", err)
		}
		err = client1.Close(fd1)
		t.Logf("client1 has closed safe and sound")
	}()
	//go func() {
	fd2, err := client2.Open("/a", false, false)
	if err != nil {
		t.Fatalf("client2 open fail! %v", err)
	}
	outData, err := client2.Read(fd2, 0, 64)
	if err != nil {
		t.Fatalf("client2 read has problem!, %v", err)
	}
	err = client2.Close(fd2)
	t.Logf("client2 has closed safe and sound")
	//}()

	time.Sleep(time.Millisecond * 1000)

	t.Logf("%v", string(outData))
	str := string(outData)
	if str != "This is a clients concurrency test" {
		t.Fatalf("read and write have concurrency problem")
	}
	//if str != "This is a clie!no gnikrow si ecilA" {
	//	t.Fatalf("read and write have concurrency problem")
	//}

	client1.Remove("/a")
}
