package main

import (
	"errors"
	"fmt"
	"time"

	pb "github.com/DeV1doR/bbg/bbg_server/protobufs"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

const (
	maxMessageSize = 8192
	pongWait       = 60 * time.Second
)

type Client struct {
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
		"Version": &appConf.ProtocolVersion,
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

func (c *Client) mapToProtobuf() *pb.MapUpdate {
	return &pb.MapUpdate{
		Tanks:   getTanksToProtobuf(c.hub),
		Bullets: getBulletsToProtobuf(c.hub),
	}
}

func (c *Client) validateTKey(token string, tKey string) (string, error) {
	rows, err := c.hub.mysql.Query(`SELECT tank.tkey
		FROM authtoken_token AS token
		INNER JOIN core_tank AS tank
		ON tank.player_id = token.user_id
		WHERE token.key=?
		AND tank.tkey=?
		LIMIT 1;`, token, tKey)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var tKey string
		if err := rows.Scan(&tKey); err != nil {
			return "", err
		}
		return tKey, nil
	}
	return "", errors.New("tkey not found")
}

func (c *Client) manageEvent(message *pb.BBGProtocol) error {
	switch pType := message.Type; *pType {
	case pb.BBGProtocol_TankUnreg:
		if c.tank == nil {
			return errors.New("TankUnreg: Tank does not exist")
		}
		pk := c.tank.ID
		if err := c.tank.RemoveTank(); err != nil {
			return fmt.Errorf("TankUnreg: Remove tank error: %s", err)
		}
		c.tank = nil
		var testID uint32
		c.sendProtoData(pb.BBGProtocol_TankRemove, &pb.TankRemove{Id: &testID, TankId: &pk}, true)

	case pb.BBGProtocol_TankReg:
		if c.tank != nil {
			return errors.New("TankReg: Tank already registred for this client")
		}
		tKey, err := c.validateTKey(message.TankReg.GetToken(), message.TankReg.GetTKey())
		if err != nil {
			return fmt.Errorf("TankReg: Invalid tKey: %s", err)
		}
		tank, err := LoadTank(c, c.hub.redis, tKey)
		if err != nil {
			return fmt.Errorf("TankReg: Load tank error: %s", err)
		}
		c.tank = tank
		c.sendProtoData(pb.BBGProtocol_TankNew, c.tank.ToProtobuf(), false)
		c.sendProtoData(pb.BBGProtocol_MapUpdate, c.mapToProtobuf(), false)

	case pb.BBGProtocol_TankMove:
		if c.tank == nil {
			return errors.New("TankMove: Tank does not exist")
		}
		if err := c.tank.Move(message.TankMove.Direction); err != nil {
			return err
		}

	case pb.BBGProtocol_TankRotate:
		if c.tank == nil {
			return errors.New("TankRotate: Tank does not exist")
		}
		if err := c.tank.TurretRotate(message.TankRotate.MouseAxes); err != nil {
			return err
		}

	case pb.BBGProtocol_TankShoot:
		if c.tank == nil {
			return errors.New("TankShoot: Tank does not exist")
		}
		if err := c.tank.Shoot(message.TankShoot.MouseAxes); err != nil {
			return fmt.Errorf("TankShoot: Got error while shooting: %s", err)
		}

	default:
		c.sendProtoData(pb.BBGProtocol_UnhandledType, nil, false)
		return nil
	}

	log.Debugf("BBG: Incomming message: %+v \n", message)

	if c.tank != nil {
		c.sendProtoData(pb.BBGProtocol_TankUpdate, c.tank.ToProtobuf(), true)
	}

	return nil
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	for {
		log.Debugln("readPump GOGOGO")
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Errorf("error: %v", err)
			}
			break
		}

		pbMsg := &pb.BBGProtocol{}
		if err := proto.Unmarshal(message, pbMsg); err != nil {
			log.Errorln("Unmarshaling error: ", err)
			continue
		}

		log.Debugln("readPump - reading...")
		if err := c.manageEvent(pbMsg); err != nil {
			panic(err)
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			log.Debugln("STARTER writePump")
			if !ok {
				// The hub closed the channel.
				msg := "Hub closed."
				c.conn.WriteMessage(websocket.CloseMessage, []byte(msg))
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

			log.Debugln("writePump - writing...")
			w.Write(encoded)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
