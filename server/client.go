package main

import (
	"fmt"

	pb "github.com/DeV1doR/bbg/server/protobufs"
	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

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
	tank *Tank
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
		pk, err := RemoveTank(c)
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
		tank, err := NewTank(c)
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

	case pb.BBGProtocol_TankShoot:
		if c.tank == nil {
			return
		}
		if err := c.tank.Shoot(message.TankShoot.MouseAxes); err != nil {
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

		log.Debugln("readPump - reading...")
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
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				log.Errorln("Hub closed.")
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				log.Errorln(err)
				return
			}

			encoded, err := proto.Marshal(message)
			if err != nil {
				log.Errorln(err)
				return
			}

			log.Infoln("writePump - writing...")
			w.Write(encoded)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
