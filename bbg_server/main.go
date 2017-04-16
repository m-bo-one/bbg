package main

import (
	"database/sql"
	"net/http"
	"os"
	"runtime"

	"github.com/DeV1doR/bbg/bbg_server/engine"
	pb "github.com/DeV1doR/bbg/bbg_server/protobufs"
	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

const (
	// Base game tick rate
	TickRate = 100
	// Canvas width
	MapWidth = 1024
	// Canvas height
	MapHeight = 768
	// Cell size
	CellSize = 10
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

func serveWS(hub *Hub, db *sql.DB, redis *redis.Client, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorln(err)
		return
	}
	client := &Client{
		db:    db,
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
	// Initialize mysql db
	db, err := MySQLClient(appConf)
	checkErr(err)
	defer db.Close()

	// Initialize redis db
	redis, err := RedisClient(appConf)
	checkErr(err)
	defer redis.Close()

	// Initialize web socket hub
	hub := newHub()
	go hub.run()

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, db, redis, w, r)
	})

	log.Infof("Starting server on %s \n", appConf.Addr)
	log.Errorln(http.ListenAndServe(appConf.Addr, nil))
}
