package main

import (
	"net/http"
	"os"
	"runtime"

	_ "net/http/pprof"

	"github.com/DeV1doR/bbg/bbg_server/engine"
	pb "github.com/DeV1doR/bbg/bbg_server/protobufs"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

const (
	// Base game tick rate
	TickRate = 100
	// Canvas width
	MapWidth = 1600
	// Canvas height
	MapHeight = 1600
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
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
}

func serveWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorln(err)
		return
	}
	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan *pb.BBGProtocol, 1024),
	}
	client.hub.register <- client
	go client.writePump()
	client.readPump()
}

func main() {
	// Initialize new db client
	dbClient, err := NewDBClient(appConf)
	checkErr(err)
	defer dbClient.Close()

	// Initialize web socket hub
	hub := NewHub(*dbClient)
	go hub.run()
	go hub.listenPushService("tank_update", 0)

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, w, r)
	})

	log.Infof("Starting server on %s \n", appConf.Addr)
	log.Errorln(http.ListenAndServe(appConf.Addr, nil))
}
