package main

import (
	"fmt"

	redistore "gopkg.in/boj/redistore.v1"

	log "github.com/Sirupsen/logrus"

	"github.com/go-redis/redis"
)

func RedisClient(appConf *conf) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     appConf.Db.Redis.Addr,
		Password: appConf.Db.Redis.Password,
		DB:       appConf.Db.Redis.DB,
	})
	_, err := client.Ping().Result()
	if err != nil {
		fmt.Println("Error during redis connection: ", err)
		return nil, err
	}
	log.Println("STARTER redis client")
	return client, nil
}

func RedisStore(appConf *conf) (*redistore.RediStore, error) {
	return redistore.NewRediStore(10, "tcp", appConf.Db.Redis.Addr, "", []byte(appConf.SecretKey))
}
