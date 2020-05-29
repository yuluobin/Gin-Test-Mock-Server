# Gin-Test-Mocker-Server
A mock server (simply)

## Note
**NEED TO UPDATE**
- The mock server is in the directory /mockServer. All other remaining is a distributed file system for my personal final project, for test purpose.
- Since the file system is running on Zookeeper, for testing you should run Zookeeper server on the local machine (the default zk port is `127.0.0.1:2181`).

## Instructions (Only for test purpose!)
**NEED TO UPDATE**

### Get Zookeeper running in the background (Or use a zk image)

```bash
wget https://downloads.apache.org/zookeeper/zookeeper-3.6.0/apache-zookeeper-3.6.0-bin.tar.gz
tar -xvzf apache-zookeeper-3.6.0-bin.tar.gz
cd apache-zookeeper-3.6.0-bin
cp conf/zoo_sample.cfg conf/zoo.cfg
bin/zkServer.sh start-foreground # Running zk server
bin/zkCli.sh -server 127.0.0.1:2181
```

### Get Mock Server

```bash
cd $GOPATH
go get -u -v -d github.com/yuluobin/Gin-Test-Mocker-Server/...
cd $GOPATH/src/github.com/yuluobin/Gin-Test-Mocker-Server
sudo docker build . # Build docker image
sudo docker run --network host e713
```

### Test Example

```bash
$ curl http://localhost:8081/post -X POST -d 'path=/a&content=123456'
{"Content":"123456","Path":"/a"}
$ curl http://localhost:8081/post -X POST -d 'path=/b&content=qwerty'
{"Content":"qwerty","Path":"/b"}
$ curl http://localhost:8081/get?path=/a
{"Content":"123456","Path":"/a"}
$ curl http://localhost:8081/get?path=/b
{"Content":"qwerty","Path":"/b"}
$ curl http://localhost:8081/get?path=/c
{"Content":"NULL","Path":"/c"}
$ curl http://localhost:8081/post -X POST -d 'path=/b&delete=true'
{"Content":"NULL","Exists":"false","Path":"/a"}
$ curl http://localhost:8081/get?path=/a
{"Content":"NULL","Path":"/a"}
```