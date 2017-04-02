package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/DeV1doR/bbg/server/models"
	pb "github.com/DeV1doR/bbg/server/protobufs"
	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

const maxMessageSize = 512

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	addr                  = flag.String("addr", ":8888", "http service address")
	mutex                 = &sync.Mutex{}
	tanks                 = make(map[uint]*models.Tank)
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
	log.SetLevel(log.DebugLevel)
}

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

type Client struct {
	// redis client
	redis *redis.Client

	// Base ws hub which manages channels
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan *pb.BBGProtocol

	// Client tank model
	tank *models.Tank
}

func (c *Client) sendProtoData(wsType pb.BBGProtocol_Type, data interface{}, all bool) error {
	pbMsg := new(pb.BBGProtocol)
	dict2proto := map[string]interface{}{
		"Type":    &wsType,
		"Version": &protocolVesion,
	}

	if data != nil {
		dict2proto[pb.BBGProtocol_Type_name[int32(wsType)]] = data
	}

	if err := FillStruct(dict2proto, pbMsg); err != nil {
		return fmt.Errorf("Send proto error: %s", err)
	}

	if all {
		log.Debugf("Mass send message: %+v \n", pbMsg)
		c.hub.broadcast <- pbMsg
	} else {
		log.Debugf("Single send message: %+v \n", pbMsg)
		c.send <- pbMsg
	}

	return nil
}

func (c *Client) manageEvent(message *pb.BBGProtocol) {
	switch pType := message.Type; *pType {
	case pb.BBGProtocol_TankUnreg:
		if c.tank == nil {
			return
		}
		pk, err := models.RemoveTank(c.redis, &c.tank.ID)
		if err != nil {
			log.Errorln("TankUreg error: ", err)
			return
		}
		c.tank = nil
		var testID uint32
		c.sendProtoData(pb.BBGProtocol_TankRemove, &pb.TankRemove{Id: &testID, TankId: &pk}, true)

	case pb.BBGProtocol_TankReg:
		if c.tank != nil {
			return
		}
		tank, err := models.NewTank(c.redis)
		if err != nil {
			log.Errorln("TankReg error: ", err)
			return
		}
		c.tank = tank
		c.sendProtoData(pb.BBGProtocol_TankNew, c.tank.ToProtobuf(), false)

	case pb.BBGProtocol_TankMove:
		if c.tank == nil {
			return
		}
		if err := c.tank.Move(message.TankMove.Direction); err != nil {
			log.Errorln(err)
			return
		}

	case pb.BBGProtocol_TankRotate:
		if c.tank == nil {
			return
		}
		if err := c.tank.TurretRotate(message.TankRotate.MouseAxes); err != nil {
			log.Errorln(err)
			return
		}

	default:
		c.sendProtoData(pb.BBGProtocol_UnhandledType, nil, false)
		return
	}

	if c.tank != nil {
		c.sendProtoData(pb.BBGProtocol_TankUpdate, c.tank.ToProtobuf(), true)
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	log.Debugln("STARTER readPump")
	for {
		log.Infoln("readPump GOGOGO")
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Errorf("error: %v \n", err)
			}
			break
		}

		pbMsg := &pb.BBGProtocol{}
		if err := proto.Unmarshal(message, pbMsg); err != nil {
			log.Errorln("Unmarshaling error: ", err)
			continue
		}

		log.Debugf("Incomming message: %+v \n", pbMsg)

		c.manageEvent(pbMsg)
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	log.Debugln("STARTER writePump")
	for {
		select {
		case message, ok := <-c.send:
			log.Infoln("writePump GOGOGO")
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				log.Errorln("Hub closed.")
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}

			if encoded, err := proto.Marshal(message); err == nil {
				w.Write(encoded)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
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
