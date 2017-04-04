package main

import (
	"sync/atomic"

	pb "github.com/DeV1doR/bbg/server/protobufs"
	log "github.com/Sirupsen/logrus"
)

var bulletIDCounter uint32

type Bullet struct {
	ID     uint32
	TankID uint32
	X, Y   int32
	Angle  float64
	Speed  int32
}

func NewBullet(tank *Tank) (*Bullet, error) {
	atomic.AddUint32(&bulletIDCounter, 1)
	t := &Bullet{
		ID:     bulletIDCounter,
		TankID: tank.ID,
		X:      tank.Cmd.X,
		Y:      tank.Cmd.Y,
		Speed:  tank.FireRate,
	}
	return t, nil
}

func (b *Bullet) ToProtobuf() *pb.BulletUpdate {
	return &pb.BulletUpdate{
		Id:     &b.ID,
		TankId: &b.TankID,
		X:      &b.X,
		Y:      &b.Y,
		Angle:  &b.Angle,
		Speed:  &b.Speed,
	}
}

func (b *Bullet) Update() error {
	return nil
}

func Updator(c *Client, updator chan *Bullet) {
	select {
	case bullet := <-updator:
		if err := bullet.Update(); err != nil {
			log.Errorln("Bullet Updator error: ", err)
			return
		}
		c.sendProtoData(pb.BBGProtocol_BulletUpdate, bullet.ToProtobuf(), true)
	}
}
