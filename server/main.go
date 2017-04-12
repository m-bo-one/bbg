package main

import (
	"net/http"
	"os"
	"runtime"

	"github.com/gorilla/mux"

	"github.com/DeV1doR/bbg/server/engine"
	pb "github.com/DeV1doR/bbg/server/protobufs"
	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

const (
	TickRate  = 100
	MapWidth  = 1024
	MapHeight = 768
	CellSize  = 10
)

var (
	configName = os.Getenv("BBG_CONFIG")
	appConf    = getConf(configName)
	upgrader   = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	world = engine.NewSpatialHash(MapWidth, MapHeight, CellSize)
)

func init() {
	runtime.GOMAXPROCS(appConf.CPUCount)

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	if !appConf.Debug {
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
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
	redis, err := RedisClient(appConf)
	if err != nil {
		log.Errorln(err)
	}

	hub := newHub()
	go hub.run()

	rtr := mux.NewRouter()
	rtr.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, redis, w, r)
	})
	rtr.HandleFunc("/login/{social:[a-z]+}/", serveSocial).Methods("GET")

	log.Infof("Starting server on %s \n", appConf.Addr)
	log.Errorln(http.ListenAndServe(appConf.Addr, rtr))
}
