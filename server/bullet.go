package main

import (
	"math/rand"
	"sync/atomic"
	"time"

	pb "github.com/DeV1doR/bbg/server/protobufs"
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

func (b *Bullet) Update(c *Client) {
	ticker := time.NewTicker(time.Second / 25)
	done := make(chan struct{})
	defer close(done)
	for {
		select {
		case <-ticker.C:
			if b.X >= 1000 {
				done <- struct{}{}
			} else {
				b.X += rand.Int31n(100)
				c.sendProtoData(pb.BBGProtocol_BulletUpdate, b.ToProtobuf(), true)
			}
		case <-done:
			ticker.Stop()
			return
		}
	}
}

// func Updator(c *Client, updator chan *Bullet) {
// 	ticker := time.Tick(100 * time.Millisecond)
// 	for sec := range ticker {
// 		select {
// 		case bullet := <-updator:
// 			log.Errorf("%v\n", sec)
// 			if err := bullet.Update(); err != nil {
// 				log.Errorln("Bullet Updator error: ", err)
// 				return
// 			}
// 			c.sendProtoData(pb.BBGProtocol_BulletUpdate, bullet.ToProtobuf(), true)
// 			updator <- bullet
// 		}
// 	}
// }
