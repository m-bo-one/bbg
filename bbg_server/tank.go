package main

import (
	"strconv"
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
	Distance        float64
	needRecharge    bool
	reloaderStarted bool
}

type Tank struct {
	ID            string
	Name          string
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

func (t *Tank) GetDamage(b *Bullet) error {
	atomic.AddInt32(&t.Health, -int32(b.Tank.Damage))
	if t.IsDead() {
		world.Remove(t)
		t.WSClient.hub.sendToPushService("tank_stat", strconv.Itoa(int(pb.StatStatus_Death)), t.ID)
		if b.Tank != nil {
			b.Tank.WSClient.hub.sendToPushService("tank_stat", strconv.Itoa(int(pb.StatStatus_Kill)), b.Tank.ID)
		}
	}
	t.Save()
	return nil
}

func (t *Tank) restoreBullet() bool {
	log.Debugf("Bullets: %+v, bulletsToReload: %+v, Need recharge: %+v \n", t.Bullets, bulletsToReload, !t.needRecharge)
	if t.isFullReloaded() {
		t.Lock()
		{
			t.needRecharge = false
		}
		t.Unlock()
		return true
	}
	atomic.AddInt32(&t.Bullets, 1)
	t.Save()
	return false
}

func (t *Tank) reloader() {
	ticker := time.NewTicker(TickRate * time.Millisecond)
	defer func() {
		ticker.Stop()
		t.Lock()
		{
			t.reloaderStarted = false
		}
		t.Unlock()
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
		log.Infof("Can't make a shoot. Tank #%s is dead.", t.ID)
		return nil
	}
	if !t.reloaderStarted && !t.isFullReloaded() {
		t.Lock()
		{
			t.reloaderStarted = true
		}
		t.Unlock()

		go t.reloader()
	}
	if t.Bullets == 0 && !t.needRecharge {
		t.needRecharge = true
		return nil
	} else if t.needRecharge || t.Bullets < 0 {
		return nil
	}

	t.Lock()
	{
		t.LastShoot = time.Now().UTC().Unix()
		t.Cmd.MouseAxes.X = *axes.X
		t.Cmd.MouseAxes.Y = *axes.Y
	}
	t.Unlock()

	atomic.AddInt32(&t.Bullets, -1)
	bullet, err := NewBullet(t)
	if err != nil {
		atomic.AddInt32(&t.Bullets, 1)
		return err
	}
	go bullet.Update(t.WSClient)
	t.Save()

	return nil
}

// Stop command for tank
func (t *Tank) Stop() error {
	return nil
}

// UpdateAngle update agnle of turret
func (t *Tank) UpdateAngle() {
	t.Cmd.Angle = AngleFromP2P(float64(t.Cmd.X), float64(t.Cmd.Y), t.Cmd.MouseAxes.X, t.Cmd.MouseAxes.Y)
}

// IsDead show tank status if health of tank less than zero
func (t *Tank) IsDead() bool {
	if t.Health <= 0 {
		return true
	}
	return false
}

// TurretRotate make tank turret rotating around its axis
func (t *Tank) TurretRotate(axes *pb.MouseAxes) error {
	if t.IsDead() {
		log.Infof("Can't make a turret rotation. Tank #%s is dead.", t.ID)
		return nil
	}
	t.Lock()
	{
		t.Cmd.MouseAxes.X = *axes.X
		t.Cmd.MouseAxes.Y = *axes.Y
		t.UpdateAngle()
	}
	t.Unlock()
	t.Save()
	return nil
}

func (t *Tank) Move(direction *pb.Direction) error {
	if t.IsDead() {
		log.Infof("Can't make a move. Tank #%s is dead.", t.ID)
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
	t.Save()
	return nil
}

func (t *Tank) ToProtobuf() *pb.TankUpdate {
	t.Lock()
	defer t.Unlock()

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
		Name:      &t.Name,
		FireRate:  &t.FireRate,
		Bullets:   &t.Bullets,
		Speed:     &t.Speed,
		Direction: &t.Cmd.Direction,
		Angle:     &t.Cmd.Angle,
		Damage:    &t.Damage,
		Status:    &status,
	}
}

func LoadTank(c *Client, redis *redis.Client, tKey string) (*Tank, error) {
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
		Name:     pbMsg.GetName(),
		Health:   pbMsg.GetHealth(),
		FireRate: pbMsg.GetFireRate(),
		Speed:    pbMsg.GetSpeed(),
		Width:    pbMsg.GetWidth(),
		Height:   pbMsg.GetHeight(),
		TGun: TGun{
			Bullets:  pbMsg.Gun.GetBullets(),
			Damage:   pbMsg.Gun.GetDamage(),
			Distance: pbMsg.Gun.GetDistance(),
		},
		Cmd: &Cmd{
			X:         pbMsg.GetX(),
			Y:         pbMsg.GetY(),
			Direction: pbMsg.GetDirection(),
			Angle:     pbMsg.GetAngle(),
			MouseAxes: &MouseAxes{},
		},
		WSClient: c,
	}
	world.Add(tank)
	return tank, nil
}

func (t *Tank) Update(pbMsg *pb.Tank) error {
	t.Lock()
	{
		t.ID = pbMsg.GetId()
		t.Cmd.X = pbMsg.GetX()
		t.Cmd.Y = pbMsg.GetY()
		t.Name = pbMsg.GetName()
		t.Health = pbMsg.GetHealth()
		t.Speed = pbMsg.GetSpeed()
		t.FireRate = pbMsg.GetFireRate()
		t.Width = pbMsg.GetWidth()
		t.Height = pbMsg.GetHeight()
		t.TGun.Bullets = pbMsg.Gun.GetBullets()
		t.TGun.Damage = pbMsg.Gun.GetDamage()
		t.TGun.Distance = pbMsg.Gun.GetDistance()
		t.Cmd.Angle = pbMsg.GetAngle()
		t.Cmd.Direction = pbMsg.GetDirection()
	}
	t.Unlock()

	encoded, err := proto.Marshal(pbMsg)
	if err != nil {
		return err
	}
	if _, err := t.WSClient.hub.redis.HSet(tHash, t.ID, encoded).Result(); err != nil {
		return err
	}
	log.Debugf("Updateing tank #%s to redis...", t.ID)
	return nil
}

func (t *Tank) Save() error {
	var pbMsg *pb.Tank
	t.Lock()
	{
		pbMsg = &pb.Tank{
			Id:       &t.ID,
			X:        &t.Cmd.X,
			Y:        &t.Cmd.Y,
			Name:     &t.Name,
			Health:   &t.Health,
			Speed:    &t.Speed,
			FireRate: &t.FireRate,
			Width:    &t.Width,
			Height:   &t.Height,
			Gun: &pb.TankGun{
				Damage:   &t.TGun.Damage,
				Bullets:  &t.TGun.Bullets,
				Distance: &t.TGun.Distance,
			},
			Angle:     &t.Cmd.Angle,
			Direction: &t.Cmd.Direction,
		}
	}
	t.Unlock()

	encoded, err := proto.Marshal(pbMsg)
	if err != nil {
		return err
	}
	if _, err := t.WSClient.hub.redis.HSet(tHash, t.ID, encoded).Result(); err != nil {
		return err
	}
	log.Debugf("Saving tank #%s to redis...", t.ID)
	return nil
}

func (t *Tank) RemoveTank() error {
	world.Remove(t)
	t.Save()
	return nil
}
