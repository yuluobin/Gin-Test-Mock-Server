package main

import (
	"flag"
	"github.com/yuluobin/Gin-Test-Mock-Server/mockServer/system"
	"os"
	"path"
	"testing"
)

func TestConfLoad(t *testing.T) {
	fPath, _ := os.Getwd()
	fPath = path.Join(fPath, "conf")
	configPath := flag.String("c", fPath, "config file path")
	flag.Parse()
	err := system.LoadConfigInformation(*configPath)
	if err != nil {
		t.Errorf("should load successfully")
	}
	//fmt.Printf("%v", conf.ConfigInfo)
}
