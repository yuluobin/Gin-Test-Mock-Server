package system

import (
	"fmt"
	"github.com/yuluobin/Gin-Test-Mock-Server/mockServer/conf"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
)

// LoadConfigInformation load config information for application

func LoadConfigInformation(configPath string) (err error) {
	var (
		filePath string
		wr       string
	)

	if configPath == "" {
		wr, _ = os.Getwd()
		wr = path.Join(wr, "conf")
	} else {
		wr = configPath
	}

	conf.WorkSpace = wr
	filePath = path.Join(conf.WorkSpace, "debug.yml")
	configData, err := ioutil.ReadFile(filePath)

	if err != nil {
		fmt.Printf(" config file read failed: %s", err)
		os.Exit(-1)
	}

	err = yaml.Unmarshal(configData, &conf.ConfigInfo)

	if err != nil {
		fmt.Printf(" config parse failed: %s", err)
		os.Exit(-1)
	}
	// server information
	conf.ServerInfo = conf.ConfigInfo.Server
	return nil
}

//func LoadRoutes(path string) []conf.RouteModel {
//	res := make([]conf.RouteModel, 0)
//	routesConf, err := ioutil.ReadFile(path)
//	if err != nil {
//		log.Fatalf("Fatal: getting config file failure")
//	}
//	err = yaml.Unmarshal(routesConf, &res)
//	if err != nil {
//		log.Fatalf("Fatal: decode conf file failure")
//	}
//
//	return res
//}
