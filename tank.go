package main

import (
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/DeV1doR/bbg/protobufs"
	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
)

const (
	tankDbKey       string = "bbg:tanks"
	bulletsToReload        = 3
)

type TGun struct {
	Damage          uint32
	Bullets         int32
	needRecharge    bool
	reloaderStarted bool
}

type Tank struct {
	ID            uint32
	Health        int32
	FireRate      int32
	Speed         int32
	Width, Height uint32
	LastShoot     int64
	Cmd           *Cmd
	WSClient      *Client

	TGun
	sync.Mutex
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

func (t *Tank) restoreBullet() bool {
	log.Debugf("Bullets: %+v, bulletsToReload: %+v, Need recharge: %+v \n", t.Bullets, bulletsToReload, !t.needRecharge)
	if t.isFullReloaded() {
		t.needRecharge = false
		return true
	}
	atomic.AddInt32(&t.Bullets, 1)
	return false
}

func (t *Tank) reloader() {
	if t.reloaderStarted {
		return
	}
	t.reloaderStarted = true
	ticker := time.NewTicker(TickRate * time.Millisecond)
	defer func() {
		ticker.Stop()
		t.reloaderStarted = false
	}()
	for {
		select {
		case <-ticker.C:
			if ok := t.restoreBullet(); ok {
				return
			}
		}
	}
}

func (t *Tank) isFullReloaded() bool {
	return t.Bullets == bulletsToReload
}

func (t *Tank) Shoot(axes *pb.MouseAxes) error {
	if t.Bullets == 0 && !t.needRecharge {
		t.needRecharge = true
		return nil
	} else if t.needRecharge {
		return nil
	}
	atomic.AddInt32(&t.Bullets, -1)
	t.LastShoot = time.Now().UTC().Unix()
	t.Cmd.MouseAxes.X = *axes.X
	t.Cmd.MouseAxes.Y = *axes.Y

	bullet, err := NewBullet(t)
	if err != nil {
		atomic.AddInt32(&t.Bullets, 1)
		return err
	}
	go bullet.Update(t.WSClient)
	go t.reloader()
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
	world.Update(t, func() {
		t.Lock()
		{
			t.Cmd.Direction = *direction

			switch t.Cmd.Direction {
			case pb.Direction_N:
				t.Cmd.Y -= t.Speed
			case pb.Direction_S:
				t.Cmd.Y += t.Speed
			case pb.Direction_E:
				t.Cmd.X += t.Speed
			case pb.Direction_W:
				t.Cmd.X -= t.Speed
			}
		}
		t.Unlock()
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
		Speed:    5,
		Width:    10,
		Height:   10,
		WSClient: c,
		TGun: TGun{
			Bullets: bulletsToReload,
			Damage:  10,
		},
		Cmd: &Cmd{
			X:         MapWidth / 2,
			Y:         MapHeight / 2,
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
		Speed:    *pbMsg.Speed,
		TGun: TGun{
			Bullets: *pbMsg.Bullets,
			Damage:  *pbMsg.Damage,
		},
		Cmd: &Cmd{
			X:         *pbMsg.X,
			Y:         *pbMsg.Y,
			Direction: *pbMsg.Direction,
			Angle:     *pbMsg.Angle,
			MouseAxes: &MouseAxes{},
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
