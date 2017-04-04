package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type conf struct {
	Debug           bool
	CPUCount        int
	Addr            string
	ProtocolVersion uint32
	Db              struct {
		Redis struct {
			Addr     string
			Password string
			DB       int
		}
	}
}

func getConf(configName string) *conf {
	if configName == "" {
		configName = "developer"
		log.Printf("Settings file not found. use %s. \n", configName)
	}
	fileName := strings.Join([]string{
		os.Getenv("GOPATH"),
		"/src/",
		"github.com/DeV1doR/bbg",
		"/server/conf/",
		configName,
		".json",
	}, "")
	c := &conf{}
	jsonFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("jsonFile.Get err   #%v ", err)
	}
	if err = json.Unmarshal(jsonFile, c); err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}
