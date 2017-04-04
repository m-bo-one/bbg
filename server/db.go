package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/go-redis/redis"
)

func RedisClient(conf *redis.Options) (*redis.Client, error) {
	client := redis.NewClient(conf)
	_, err := client.Ping().Result()
	if err != nil {
		fmt.Println("Error during redis connection: ", err)
		return nil, err
	}
	log.Println("STARTER redis client")
	return client, nil
}
