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
}

func NewBullet(tank *Tank) (*Bullet, error) {
	atomic.AddUint32(&bulletIDCounter, 1)
	return &Bullet{
		ID:     bulletIDCounter,
		TankID: tank.ID,
		X:      float64(tank.Cmd.X),
		Y:      float64(tank.Cmd.Y),
		Speed:  1000,
		Angle:  tank.Cmd.Angle,
		Alive:  true,
	}, nil
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

func (b *Bullet) OutOfBoundaries() bool {
	// Hardcoded
	if 0 < b.X && b.X < 1024 && 10 < b.Y && b.Y < 768 {
		return false
	}
	return true
}

func (b *Bullet) Update(c *Client) {
	var wg sync.WaitGroup
	ticker := time.NewTicker(time.Second / TickRate)

	defer func() {
		ticker.Stop()
		c.sendProtoData(pb.BBGProtocol_BulletUpdate, b.ToProtobuf(), true)
		log.Infof("Update goroutine for bullet - %d destroyed successfully.", b.ID)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		speed := float64(b.Speed)
		for {
			select {
			case <-ticker.C:
				b.X += math.Cos(b.Angle) * (speed / TickRate)
				b.Y += math.Sin(b.Angle) * (speed / TickRate)

				if b.OutOfBoundaries() {
					b.Alive = false
					return
				}
				c.sendProtoData(pb.BBGProtocol_BulletUpdate, b.ToProtobuf(), true)
			}
		}
	}()

	wg.Wait()
}
