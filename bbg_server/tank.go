package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/DeV1doR/bbg/bbg_server/protobufs"
	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
)

const (
	tHash           string = "bbg:tanks"
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
	UID           uint32

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
	t.UpdateTank(t.WSClient.redis, t.UID, t.ID)
	return nil
}

func (t *Tank) restoreBullet() bool {
	log.Debugf("Bullets: %+v, bulletsToReload: %+v, Need recharge: %+v \n", t.Bullets, bulletsToReload, !t.needRecharge)
	if t.isFullReloaded() {
		t.needRecharge = false
		return true
	}
	atomic.AddInt32(&t.Bullets, 1)
	t.UpdateTank(t.WSClient.redis, t.UID, t.ID)
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
	if t.IsDead() {
		log.Infof("Can't make a shoot. Tank #%d is dead.", t.ID)
		return nil
	}
	if t.Bullets == 0 && !t.needRecharge {
		t.needRecharge = true
		return nil
	} else if t.needRecharge || t.Bullets < 0 {
		return nil
	}
	t.LastShoot = time.Now().UTC().Unix()
	t.Cmd.MouseAxes.X = *axes.X
	t.Cmd.MouseAxes.Y = *axes.Y

	atomic.AddInt32(&t.Bullets, -1)
	bullet, err := NewBullet(t)
	if err != nil {
		atomic.AddInt32(&t.Bullets, 1)
		return err
	}
	go bullet.Update(t.WSClient)
	go t.reloader()

	t.UpdateTank(t.WSClient.redis, t.UID, t.ID)

	return nil
}

func (t *Tank) Stop() error {
	return nil
}

func (t *Tank) UpdateAngle() {
	t.Cmd.Angle = AngleFromP2P(float64(t.Cmd.X), float64(t.Cmd.Y), t.Cmd.MouseAxes.X, t.Cmd.MouseAxes.Y)
}

func (t *Tank) IsDead() bool {
	if t.Health <= 0 {
		return true
	}
	return false
}

func (t *Tank) TurretRotate(axes *pb.MouseAxes) error {
	if t.IsDead() {
		log.Infof("Can't make a turret rotation. Tank #%d is dead.", t.ID)
		return nil
	}
	t.Cmd.MouseAxes.X = *axes.X
	t.Cmd.MouseAxes.Y = *axes.Y
	t.UpdateAngle()
	t.UpdateTank(t.WSClient.redis, t.UID, t.ID)
	return nil
}

func (t *Tank) Move(direction *pb.Direction) error {
	if t.IsDead() {
		log.Infof("Can't make a move. Tank #%d is dead.", t.ID)
		return nil
	}
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
	t.UpdateTank(t.WSClient.redis, t.UID, t.ID)
	return nil
}

func (t *Tank) ToProtobuf() *pb.TankUpdate {
	var status pb.TankUpdate_Status
	if t.IsDead() {
		status = pb.TankUpdate_Dead
	} else {
		status = pb.TankUpdate_Alive
	}
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
		Status:    &status,
	}
}

func LoadTank(c *Client, redis *redis.Client, UID uint32, pk uint32) (*Tank, error) {
	tKey := fmt.Sprintf("uid:%d:tank:%d", UID, pk)
	val, err := redis.HGet(tHash, tKey).Result()
	if err != nil {
		return nil, err
	}
	pbMsg := &pb.Tank{}
	if err := proto.Unmarshal([]byte(val), pbMsg); err != nil {
		return nil, err
	}
	tank := &Tank{
		ID:       pbMsg.GetId(),
		Health:   pbMsg.GetHealth(),
		FireRate: pbMsg.GetFireRate(),
		Speed:    pbMsg.GetSpeed(),
		TGun: TGun{
			Bullets: pbMsg.Gun.GetBullets(),
			Damage:  pbMsg.Gun.GetDamage(),
		},
		Cmd: &Cmd{
			X:         pbMsg.GetX(),
			Y:         pbMsg.GetY(),
			Direction: pbMsg.GetDirection(),
			Angle:     pbMsg.GetAngle(),
			MouseAxes: &MouseAxes{},
		},
		WSClient: c,
		UID:      UID,
	}
	world.Add(tank)
	return tank, nil
}

func (t *Tank) UpdateTank(redis *redis.Client, UID uint32, pk uint32) error {
	tKey := fmt.Sprintf("uid:%d:tank:%d", UID, pk)
	if _, err := redis.HGet(tHash, tKey).Result(); err != nil {
		return err
	}
	pbMsg := &pb.Tank{
		Id:       &t.ID,
		X:        &t.Cmd.X,
		Y:        &t.Cmd.Y,
		Health:   &t.Health,
		Speed:    &t.Speed,
		FireRate: &t.FireRate,
		Width:    &t.Width,
		Height:   &t.Height,
		Gun: &pb.TankGun{
			Damage:  &t.TGun.Damage,
			Bullets: &t.TGun.Bullets,
		},
		Angle:     &t.Cmd.Angle,
		Direction: &t.Cmd.Direction,
	}
	encoded, err := proto.Marshal(pbMsg)
	if err != nil {
		return err
	}
	if _, err := redis.HSet(tHash, tKey, encoded).Result(); err != nil {
		return err
	}
	log.Debugf("Send update for tank #%d to redis...", t.ID)
	return nil
}

func (t *Tank) RemoveTank() error {
	world.Remove(t)
	t.UpdateTank(t.WSClient.redis, t.UID, t.ID)
	return nil
}
