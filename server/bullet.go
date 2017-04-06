package main

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/DeV1doR/bbg/server/protobufs"
	log "github.com/Sirupsen/logrus"
)

var bulletIDCounter uint32

type Bullet struct {
	ID     uint32
	TankID uint32
	X, Y   float64
	Angle  float64
	Speed  int32
	Alive  bool

	sync.Mutex
}

func (b *Bullet) GetX() int32 {
	return int32(b.X)
}

func (b *Bullet) GetY() int32 {
	return int32(b.Y)
}

func (b *Bullet) GetRadius() int32 {
	return 5
}

func (b *Bullet) ToProtobuf() *pb.BulletUpdate {
	return &pb.BulletUpdate{
		Id:     &b.ID,
		TankId: &b.TankID,
		X:      &b.X,
		Y:      &b.Y,
		Angle:  &b.Angle,
		Speed:  &b.Speed,
		Alive:  &b.Alive,
	}
}

func (b *Bullet) IsColide() (*Tank, bool) {
	if b.X < 0 || b.X > MapWidth || b.Y < 0 || b.Y > MapHeight {
		return nil, true
	}
	for _, other := range world.Nearby(b) {
		if tank, ok := other.(*Tank); ok {
			if tank.ID != b.TankID {
				log.Debugf("Collided with: %+v \n", tank)
				return tank, true
			}
		}
	}
	return nil, false
}

func (b *Bullet) Update(c *Client) {
	ticker := time.NewTicker(time.Second / TickRate)

	defer func() {
		c.sendProtoData(pb.BBGProtocol_BulletUpdate, b.ToProtobuf(), true)
		ticker.Stop()
		world.Remove(b)
	}()

	speed := float64(b.Speed)
	for {
		select {
		case <-ticker.C:
			world.Update(b, func() {
				b.Lock()
				defer b.Unlock()
				b.X += math.Cos(b.Angle) * speed
				b.Y += math.Sin(b.Angle) * speed
			})
			if tank, isCollide := b.IsColide(); isCollide {
				if tank != nil {
					tank.GetDamage(5)
					c.sendProtoData(pb.BBGProtocol_TankUpdate, tank.ToProtobuf(), true)
				}
				b.Alive = false
				return
			}
			c.sendProtoData(pb.BBGProtocol_BulletUpdate, b.ToProtobuf(), true)
		}
	}

}

func NewBullet(tank *Tank) (*Bullet, error) {
	atomic.AddUint32(&bulletIDCounter, 1)
	b := &Bullet{
		ID:     bulletIDCounter,
		TankID: tank.ID,
		X:      float64(tank.Cmd.X),
		Y:      float64(tank.Cmd.Y),
		Speed:  10,
		Angle:  tank.Cmd.Angle,
		Alive:  true,
	}
	world.Add(b)
	return b, nil
}
