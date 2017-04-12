package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
)

type socialConf struct {
	ClientID     string
	ClientSecret string
}

type conf struct {
	Debug           bool
	CPUCount        int
	Addr            string
	ProtocolVersion uint32
	SecretKey       string
	ProxyHost       string
	Db              struct {
		Redis struct {
			Addr     string
			Password string
			DB       int
		}
	}
	Oauth2 struct {
		Facebook socialConf
		Github   socialConf
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
	_, fileName, _, _ := runtime.Caller(1)
	fileName = strings.Join([]string{
		path.Dir(fileName),
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
