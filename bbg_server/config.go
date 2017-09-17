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

type dbParams struct {
	Addr     string
	User     string
	Password string
	DB       string
}

type conf struct {
	Debug           bool
	CPUCount        int
	Addr            string
	ProtocolVersion uint32
	SecretKey       string
	Db              struct {
		Redis dbParams
		MySQL dbParams
		Kafka dbParams
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
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost != "" {
		c.Db.Redis.Addr = redisHost + ":6379"
	}
	mysqlHost := os.Getenv("MYSQL_HOST")
	if mysqlHost != "" {
		c.Db.MySQL.Addr = mysqlHost + ":3306"
	}
	kafkaHost := os.Getenv("KAFKA_HOST")
	if kafkaHost != "" {
		c.Db.Kafka.Addr = kafkaHost + ":9092"
	}
	return c
}
