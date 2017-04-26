package main

import (
	"math"
	"strconv"
	"sync/atomic"

	pb "github.com/DeV1doR/bbg/bbg_server/protobufs"
)

var bulletIDCounter uint32

type Bullet struct {
	ID            uint32
	Tank          *Tank
	X, Y          float64
	Angle         float64
	Speed         int32
	Alive         bool
	Distance      float64
	TotalDistance float64

	ws *Client
}

func (b *Bullet) GetX() int32 {
	return int32(b.X)
}

func (b *Bullet) GetY() int32 {
	return int32(b.Y)
}

func (b *Bullet) GetWidth() int32 {
	return 5
}

func (b *Bullet) GetHeight() int32 {
	return 5
}

func (b *Bullet) ToProtobuf() *pb.BulletUpdate {
	return &pb.BulletUpdate{
		Id:       b.ID,
		TankId:   b.Tank.ID,
		X:        b.X,
		Y:        b.Y,
		Angle:    b.Angle,
		Speed:    b.Speed,
		Alive:    b.Alive,
		Distance: b.Distance,
	}
}

func (b *Bullet) IsColide() bool {
	for _, i := range world.Nearby(b) {
		switch object := i.(type) {
		case *Tank:
			if object.ID != b.Tank.ID {
				if !object.IsDead() {
					object.GetDamage(b)
				}
				return true
			}
		case *Bullet:
			object.Alive = false
			return true
		}
	}
	return false
}

func (b *Bullet) IsOutOfRange() bool {
	return b.TotalDistance >= b.Distance
}

func (b *Bullet) UpdateDistance(x float64, y float64) {
	b.TotalDistance += math.Sqrt(math.Pow(x-b.X, 2) + math.Pow(y-b.Y, 2))
}

func (b *Bullet) Update() bool {
	speed := float64(b.Speed)
	world.Update(b, func() {
		nX := b.X + math.Cos(b.Angle)*speed
		nY := b.Y + math.Sin(b.Angle)*speed
		b.UpdateDistance(nX, nY)
		b.X = nX
		b.Y = nY
	})
	if b.IsOutOfRange() {
		b.Alive = false
		world.Remove(b)
		return false
	} else if b.IsColide() {
		go b.ws.hub.sendToPushService("tank_stat", strconv.Itoa(int(pb.StatStatus_Hit)), b.Tank.ID)
		b.Alive = false
		world.Remove(b)
		return false
	}
	return true
}

func NewBullet(tank *Tank) (*Bullet, error) {
	atomic.AddUint32(&bulletIDCounter, 1)
	b := &Bullet{
		ID:       bulletIDCounter,
		Tank:     tank,
		X:        float64(tank.Cmd.X),
		Y:        float64(tank.Cmd.Y),
		Speed:    8,
		Angle:    tank.Cmd.Angle,
		Alive:    true,
		Distance: tank.TGun.Distance,
		ws:       tank.ws,
	}
	world.Add(b)
	return b, nil
}
