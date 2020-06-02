package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yuluobin/Gin-Test-Mocker-Server/mockServer/conf"
	"github.com/yuluobin/Gin-Test-Mocker-Server/mockServer/system"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
)

var responses map[string]conf.Response

//var client puddlestore.Client

func main() {

	// Initialize with config file
	fPath, _ := os.Getwd()
	fPath = path.Join(fPath, "conf")
	configPath := flag.String("c", fPath, "config file path")
	flag.Parse()
	responses = make(map[string]conf.Response)
	err := system.LoadConfigInformation(*configPath)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("%+v\n",conf.ConfigInfo.Server)
	//
	//// Initialize cluster
	//config := puddlestore.Config{
	//	BlockSize:   conf.ConfigInfo.Server.BlockSize,
	//	NumReplicas: conf.ConfigInfo.Server.NumReplicas,
	//	NumTapestry: conf.ConfigInfo.Server.NumTapestry,
	//	ZkAddr:      conf.ConfigInfo.Server.ZKPort,
	//}
	//cluster, err := puddlestore.CreateCluster(config)
	//if err != nil {
	//	panic(err)
	//}
	//defer cluster.Shutdown()
	//
	//client, err = cluster.NewClient()
	//if err != nil {
	//	panic(err)
	//}

	r := gin.Default()
	for _, route := range conf.ConfigInfo.Func {
		//method := route.Method
		//for _, res := range route.Responses {
		if route.Method == "GET" {
			// Extract info from `route.Route` and insert it into a map?
			// to let `GET` have access to res information

			r.GET(route.Route, HandleGet(route))
		} else if route.Method == "POST" {
			// Extract info from `route.Route` and insert it into a map?
			// to let `POST` have access to res information
			r.POST(route.Route, HandlePOST(route))
		} else {
			// Error message
		}
		//}
	}
	//r.GET("/get", HandleGet)
	//r.GET("/list", HandleList)
	//r.POST("/post", HandlePost)
	//r.POST("/mkdir", HandleMkdir)
	r.Run( /*":8081"*/ conf.ConfigInfo.Server.Port)

}

func HandleGet(route *conf.RouteModel) gin.HandlerFunc {
	// This can be added as a global variable later
	//var responses map[string]conf.Response
	//responses := make(map[string]conf.Response)
	// Insert values as a plain string and corresponding responses as values
	for _, response := range route.Responses {
		// Link params and values as a string
		// Get component at right side of "?"
		uri, err := url.Parse(response.URI)
		if err != nil {
			// Handle error
			panic("Panic: config url cannot be decoded")
		}
		values, err := url.ParseQuery(uri.RawQuery)
		if err != nil {
			// Error message
			panic("Panic: config url cannot be decoded")
		}
		var keys []string
		for k := range values {
			keys = append(keys, k)
			//fmt.Printf("%v\n", k)
		}
		sort.Strings(keys)
		var params string
		for _, k := range keys {
			params += k + "=" + values[k][0]
		}

		key := fmt.Sprintf("%s|%s|%s", route.Route, route.Method, params)
		responses[key] = response
		fmt.Printf("%v\n", key)
	}

	fn := func(c *gin.Context) {
		values, err := url.ParseQuery(c.Request.URL.RawQuery)
		if err != nil {
			// Error message
			render(c, gin.H{
				"error": "Request invalid",
			}, conf.Response{}, http.StatusBadRequest)
		}
		var keys []string
		for k := range values {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var params string
		for _, k := range keys {
			params += k + "=" + values[k][0]
		}
		key := fmt.Sprintf("%s|%s|%s", route.Route, route.Method, params)
		fmt.Printf("%v\n", key)

		if response, ok := responses[key]; ok {
			// There exists response
			render(c, response.RetBody, response, http.StatusOK)
		} else {
			render(c, response.ErrBody, response, http.StatusNotFound)
		}
	}

	return fn
}

func HandlePOST(route *conf.RouteModel) gin.HandlerFunc {

	fn := func(c *gin.Context) {
		// Your handler code goes in here - e.g.

	}

	return fn
}

// Create response in the format of JSON
//func createResponse(res conf.Response) interface{} {
//
//}

func render(c *gin.Context, data gin.H, res conf.Response, status int) {
	switch res.Header {
	case "application/json":
		c.JSON(status, data)
	case "application/xml":
		c.XML(status, data)
	default:
		c.JSON(status, data)
	}
}

//func HandleGet(c *gin.Context) {
//	path := c.Query("path")
//	//file := c.DefaultQuery("role", "")
//	fd, err := client.Open(path, false, false)
//	if err != nil {
//		client.Close(fd)
//		c.JSON(http.StatusBadRequest, gin.H{
//			"Path":    path,
//			"Content": "NULL",
//		})
//		return
//	}
//	data, _ := client.Read(fd, 0, 1000)
//	client.Close(fd)
//
//	c.JSON(200, gin.H{
//		"Path":    path,
//		"Content": string(data),
//	})
//}
//
//// Put and Delete
//func HandlePost(c *gin.Context) {
//	//body, _ := ioutil.ReadAll(c.Request.Body)
//	//c.String(http.StatusOK, "Put,%s", body)
//	path := c.PostForm("path")
//	isRemove := c.DefaultPostForm("delete", "false")
//	content := c.DefaultPostForm("content", "NULL")
//
//	if isRemove == "true" {
//		err := client.Remove(path)
//		if err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{
//				"Path":    path,
//				"Content": content,
//			})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{
//			"Path":    path,
//			"Content": content,
//			"Exists":  "false",
//		})
//		return
//	}
//
//	fd, err := client.Open(path, true, true)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{
//			"Path":    path,
//			"Content": content,
//		})
//		return
//	}
//
//	err = client.Write(fd, 0, []byte(content))
//	if err != nil {
//		client.Close(fd)
//		c.JSON(http.StatusBadRequest, gin.H{
//			"Path":    path,
//			"Content": content,
//		})
//		return
//	}
//	data, _ := client.Read(fd, 0, 1000)
//	client.Close(fd)
//	c.JSON(200, gin.H{
//		"Path":    path,
//		"Content": string(data),
//	})
//}
//
//func HandleMkdir(c *gin.Context) {
//	path := c.PostForm("path")
//
//	err := client.Mkdir(path)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{
//			"Path":    path,
//			"Success": "false",
//		})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{
//		"Path":    path,
//		"Success": "true",
//	})
//}
//
//func HandleList(c *gin.Context) {
//	path := c.PostForm("path")
//
//	list, err := client.List(path)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{
//			"Path":    path,
//			"Success": "false",
//			"list":    "",
//		})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{
//		"Path":    path,
//		"Success": "true",
//		"list":    list,
//	})
//}
//
//func HandleDelete(c *gin.Context) {
//	body, _ := ioutil.ReadAll(c.Request.Body)
//	c.String(http.StatusOK, "Delete,%s", body)
//
//}
//
//func HandleOptions(c *gin.Context) {
//	body, _ := ioutil.ReadAll(c.Request.Body)
//	c.String(http.StatusOK, "Options,%s", body)
//
//}
//
//func HandlePatch(c *gin.Context) {
//	body, _ := ioutil.ReadAll(c.Request.Body)
//	c.String(http.StatusOK, "patch,%s", body)
//
//}
//
//func HandleHead(c *gin.Context) {
//	// http head only response  header
//	fmt.Printf("head http \r\n")
//}
