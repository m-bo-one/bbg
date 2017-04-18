package main

import (
	"database/sql"
	"fmt"
	"strconv"

	sarama "gopkg.in/Shopify/sarama.v1"

	log "github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"

	"github.com/go-redis/redis"
)

type kafkaClient struct {
	// kafka producer
	kafkaProducer sarama.AsyncProducer

	// kafka consumer
	kafkaConsumer sarama.Consumer

	// kafka partion consumer
	partitionConsumer map[string]sarama.PartitionConsumer
}

type DBClient struct {
	// mysql client
	mysql *sql.DB

	// redis client
	redis *redis.Client

	// kafka client
	kafkaClient
}

func (c *DBClient) Close() error {
	if c.mysql != nil {
		c.mysql.Close()
	}
	if c.redis != nil {
		c.redis.Close()
	}
	if c.kafkaProducer != nil {
		c.kafkaProducer.Close()
	}
	if c.kafkaConsumer != nil {
		c.kafkaConsumer.Close()
	}
	return nil
}

func RedisClient(appConf *conf) (*redis.Client, error) {
	dbName, _ := strconv.ParseInt(appConf.Db.Redis.DB, 10, 32)
	client := redis.NewClient(&redis.Options{
		Addr:     appConf.Db.Redis.Addr,
		Password: appConf.Db.Redis.Password,
		DB:       int(dbName),
	})
	_, err := client.Ping().Result()
	if err != nil {
		log.Errorln("Redis: Error during connection: ", err)
		return nil, err
	}
	log.Infoln("Redis: Started")
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
		log.Errorln("MySQL: Error during connection: ", err)
		return nil, err
	}
	log.Infoln("MySQL: Started")
	return conn, nil
}

func KafkaProducer(appConf *conf) (sarama.AsyncProducer, error) {
	producer, err := sarama.NewAsyncProducer([]string{appConf.Db.Kafka.Addr}, nil)
	if err != nil {
		log.Errorln("Kafka: Error during connection: ", err)
		return nil, err
	}
	log.Infoln("Kafka: Producer started")
	return producer, nil
}

func KafkaConsumer(appConf *conf) (sarama.Consumer, error) {
	consumer, err := sarama.NewConsumer([]string{appConf.Db.Kafka.Addr}, nil)
	if err != nil {
		log.Errorln("Kafka: Error during connection: ", err)
		return nil, err
	}
	log.Infoln("Kafka: Consumer started")
	return consumer, nil
}

func NewDBClient(appConf *conf) (*DBClient, error) {
	var err error
	dbClient := new(DBClient)
	// Initialize mysql db
	dbClient.mysql, err = MySQLClient(appConf)
	if err != nil {
		return nil, err
	}

	// Initialize redis db
	dbClient.redis, err = RedisClient(appConf)
	if err != nil {
		return nil, err
	}

	// Initialize kafka producer
	dbClient.kafkaProducer, err = KafkaProducer(appConf)
	if err != nil {
		return nil, err
	}

	// Initialize kafka consumer
	dbClient.kafkaConsumer, err = KafkaConsumer(appConf)
	if err != nil {
		return nil, err
	}

	dbClient.partitionConsumer = make(map[string]sarama.PartitionConsumer)

	return dbClient, nil
}
