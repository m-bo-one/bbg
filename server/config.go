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
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(pwd)
	}
	fileName := strings.Join([]string{
		pwd,
		"/conf/",
		configName,
		".json",
	}, "")
	log.Printf("Use %s settings file. \n", fileName)
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
