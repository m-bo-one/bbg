package main

import (
	"flag"
	"net/http"
	"os"
	"sync"

	pb "github.com/DeV1doR/bbg/server/protobufs"
	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

const (
	maxMessageSize = 512
)

var (
	debug    = os.Getenv("BBG_DEBUG")
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	addr                  = flag.String("addr", "127.0.0.1:8888", "http service address")
	mutex                 = &sync.Mutex{}
	protocolVesion uint32 = 1
	redisConf             = &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
)

func init() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.ErrorLevel)
}

func serveWS(hub *Hub, redis *redis.Client, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorln(err)
		return
	}
	client := &Client{
		redis: redis,
		hub:   hub,
		conn:  conn,
		send:  make(chan *pb.BBGProtocol, 256),
	}
	client.hub.register <- client
	go client.writePump()
	client.readPump()
}

func main() {
	redis, err := RedisClient(redisConf)
	if err != nil {
		log.Errorln(err)
	}
	hub := newHub()
	go hub.run()
	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, redis, w, r)
	})
	log.Infof("Starting server on %s \n", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Errorln("ListenAndServe: ", err)
	}
}
