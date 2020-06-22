# Gin-Test-Mock-Server
## Introduction

This is a simple web API mocking server based on Gin. Simply modify configuration file and map API requests and responses in YAML, you are ready to test front-end API without an actual server.

## Features

- Support server port change
- Support custom route response
- Support lua scripts to test mock server

## Installation
### Get Mock Server and Run

If you haven't modify or create your new config file, the program will load `debug.yml` as default configuration file.

```bash
cd $GOPATH
go get -u -v -d github.com/yuluobin/Gin-Test-Mocker-Server/...
cd $GOPATH/src/github.com/yuluobin/Gin-Test-Mocker-Server
git checkout feat
sudo docker build -t mockserver . # Build docker image
sudo docker run -p 8081:8081 mockserver
```

In this way the mock server will run in docker. But if you want to build it in your own machine and start to work instantly, simply do as follows after checking out the `feat` branch:

```bash
go build -o mockserver; ./mockserver
```

If you see `Listening and serving HTTP on :8081`, the mock server is successfully running in the background.

## Getting Started

This simple guide will help you go through the process of using the mock server. Make sure you have stopped the original mock server before continue.

1. Create a configuration file wherever you like (we recommend it be in the `mockServer` directory). You can find the configuration format guide below. If you are still confused about how to write a YAML file, please refer to the default file `debug.yml`.

2. Replace the default file name `debug.yml` with your new config file in `mockServer/system/config.go`.

3. Build the mockServer in docker and run it

   ```bash
   sudo docker build -t mockserver . # Build docker image
   sudo docker run -p 8081:8081 mockserver
   ```

   Make sure the first port number is the same as the one in your configuration file.

   Or of course you can run it without docker:

   ```bash
   go build -o mockserver
   ./mockserver
   ```

4. Use *curl* or other service test methods to test your API. 

   (**Don't** know how to use *curl*? You can refer to the test examples below.)

5. (Optional) There's a lua scripts provided to test mock server performance. 

## Configuration File Format Guide

The format of configuration file is YAML.

- `route`: the basic URI of *this* request
- `method`: the *http* request method (up to now the mock server supports `GET` and `POST` request)
- `res`: is a respond array containing specific configuration of each request
  - `uri`: a URI contains variables
  - `header`: format of respond message
  - `post_body`: contains post body. Can be omitted in `GET` request
  - `ret_body`: contains respond message body if the request is successful
  - `err_body`: contains error message if the request is unsuccessful.

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

## wrk Test Guide

Under `mockServer` directory

```bash
wrk -t5 -c400 -d20s -s wrk_test.lua http://localhost:8081
```
