# Gin-Test-Mocker-Server
A mock server (simply)

## Note
**NEED TO UPDATE**
- The mock server is in the directory /mockServer. All other remaining is a distributed file system is for test purpose (maybe of no use).

## Instructions (Only for test purpose!)
### Get Mock Server

```bash
cd $GOPATH
go get -u -v -d github.com/yuluobin/Gin-Test-Mocker-Server/...
cd $GOPATH/src/github.com/yuluobin/Gin-Test-Mocker-Server
git checkout feat
sudo docker build -t mockserver . # Build docker image
sudo docker run -p 8081:8081 mockserver
```

### Test Example

As the default configuration file

```bash
$ curl "http://localhost:8081/login?user=chadli&pwd=123456"
{"msg":"Successfully logged in!","token":"ABC"}
$ curl "http://localhost:8081/login?pwd=qwerty&user=ekopei"
{"msg":"Successfully logged in!","token":"DEF"}
$ curl "http://localhost:8081/get_userinfo?token=ABC"
{"age":20,"gender":"male","msg":"Successfully get user info!"}
$ curl "http://localhost:8081/get_userinfo?token=DEF"
{"age":21,"gender":"male","msg":"Successfully get user info!"}
$ curl "http://localhost:8081/set_userinfo" -X POST -d 'token=ABC&age=20'
{"msg":"Successfully set user info!","ret_code":0}
$ curl "http://localhost:8081/set_userinfo" -X POST -d 'token=DEF&age=21'
{"msg":"Successfully set user info!","ret_code":0}
$ curl "http://localhost:8081/get_userinfo?token=ABD" # Wrong token
null
```
