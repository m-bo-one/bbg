package main

import (
	"database/sql"
	"fmt"
	"strconv"

	log "github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"

	"github.com/go-redis/redis"
)

func RedisClient(appConf *conf) (*redis.Client, error) {
	dbName, _ := strconv.ParseInt(appConf.Db.Redis.DB, 10, 32)
	client := redis.NewClient(&redis.Options{
		Addr:     appConf.Db.Redis.Addr,
		Password: appConf.Db.Redis.Password,
		DB:       int(dbName),
	})
	_, err := client.Ping().Result()
	if err != nil {
		fmt.Println("Error during redis connection: ", err)
		return nil, err
	}
	log.Println("STARTER redis client")
	return client, nil
}

func MySQLClient(appConf *conf) (*sql.DB, error) {
	url := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8",
		appConf.Db.MySQL.User,
		appConf.Db.MySQL.Password,
		appConf.Db.MySQL.Addr,
		appConf.Db.MySQL.DB,
	)
	conn, err := sql.Open("mysql", url)
	if err != nil {
		fmt.Println("Error during mysql connection: ", err)
		return nil, err
	}
	log.Println("STARTER mysql connection")
	return conn, nil
}
