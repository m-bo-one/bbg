package main

import (
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/DeV1doR/bbg/server/protobufs"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
)

const tankDbKey string = "bbg:tanks"

var tLock = &sync.Mutex{}

type Tank struct {
	ID            uint32
	Health        int32
	FireRate      int32
	Bullets       int32
	Speed         int32
	Damage        uint32
	Width, Height uint32
	LastShoot     int64
	Cmd           *Cmd
	WSClient      *Client
}

func (t *Tank) GetX() int32 {
	return t.Cmd.X
}

func (t *Tank) GetY() int32 {
	return t.Cmd.Y
}

func (t *Tank) GetRadius() int32 {
	return int32(t.Width+t.Height) / 2
}

func (t *Tank) GetDamage(d int32) error {
	atomic.AddInt32(&t.Health, -d)
	return nil
}

func (t *Tank) Shoot(axes *pb.MouseAxes) error {
	// if time.Now().UTC().Unix() > t.LastShoot {
	t.LastShoot = time.Now().UTC().Unix()
	t.Cmd.MouseAxes.X = *axes.X
	t.Cmd.MouseAxes.Y = *axes.Y
	// }

	bullet, err := NewBullet(t)
	if err != nil {
		return err
	}
	go bullet.Update(t.WSClient)
	return nil
}

func (t *Tank) Stop() error {
	return nil
}

func (t *Tank) TurretRotate(axes *pb.MouseAxes) error {
	t.Cmd.MouseAxes.X = *axes.X
	t.Cmd.MouseAxes.Y = *axes.Y
	t.Cmd.Angle = math.Atan2(
		t.Cmd.MouseAxes.Y-float64(t.Cmd.Y),
		t.Cmd.MouseAxes.X-float64(t.Cmd.X),
	)
	return nil
}

func (t *Tank) Move(direction *pb.Direction) error {
	t.Cmd.Direction = *direction
	world.Update(t, func() {
		switch *direction {
		case pb.Direction_N:
			t.Cmd.Y -= t.Speed
		case pb.Direction_S:
			t.Cmd.Y += t.Speed
		case pb.Direction_E:
			t.Cmd.X += t.Speed
		case pb.Direction_W:
			t.Cmd.X -= t.Speed
		}
	})
	return nil
}

func (t *Tank) ToProtobuf() *pb.TankUpdate {
	return &pb.TankUpdate{
		Id:        &t.Cmd.ID,
		TankId:    &t.ID,
		X:         &t.Cmd.X,
		Y:         &t.Cmd.Y,
		Health:    &t.Health,
		FireRate:  &t.FireRate,
		Bullets:   &t.Bullets,
		Speed:     &t.Speed,
		Direction: &t.Cmd.Direction,
		Angle:     &t.Cmd.Angle,
		Damage:    &t.Damage,
	}
}

func NewTank(c *Client) (*Tank, error) {
	pk, err := c.redis.Incr(tankDbKey + ":id").Result()
	if err != nil {
		return nil, err
	}
	direction := pb.Direction_N
	t := &Tank{
		ID:       uint32(pk),
		Health:   100,
		FireRate: 100,
		Bullets:  1000,
		Speed:    5,
		Width:    10,
		Height:   10,
		Damage:   10,
		WSClient: c,
		Cmd: &Cmd{
			X:         0,
			Y:         0,
			Direction: direction,
			MouseAxes: &MouseAxes{},
		},
	}
	encoded, err := proto.Marshal(t.ToProtobuf())
	if err != nil {
		return nil, err
	}
	if err := c.redis.HSet(tankDbKey, strconv.FormatInt(pk, 10), encoded).Err(); err != nil {
		return nil, err
	}

	world.Add(t)
	return t, nil
}

func LoadTank(redis *redis.Client, pk *uint32) (*Tank, error) {
	val, err := redis.HGet(tankDbKey, strconv.Itoa(int(*pk))).Result()
	if err != nil {
		return nil, err
	}
	pbMsg := &pb.TankUpdate{}
	if err := proto.Unmarshal([]byte(val), pbMsg); err != nil {
		return nil, err
	}
	return &Tank{
		ID:       *pbMsg.TankId,
		Health:   *pbMsg.Health,
		FireRate: *pbMsg.FireRate,
		Bullets:  *pbMsg.Bullets,
		Speed:    *pbMsg.Speed,
		Cmd: &Cmd{
			X:         *pbMsg.X,
			Y:         *pbMsg.Y,
			Direction: *pbMsg.Direction,
			Angle:     *pbMsg.Angle,
			MouseAxes: &MouseAxes{
				X: 0,
				Y: 0,
			},
		},
	}, nil
}

func RemoveTank(c *Client) (uint32, error) {
	_, err := c.redis.HDel(tankDbKey, strconv.Itoa(int(c.tank.ID))).Result()
	if err != nil {
		return 0, err
	}
	world.Remove(c.tank)
	return c.tank.ID, nil
}
