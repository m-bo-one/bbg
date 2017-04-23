package main

import (
	"strconv"
	"sync"
	"time"

	pb "github.com/DeV1doR/bbg/bbg_server/protobufs"
	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
)

const tHash string = "bbg:tanks"

type TGun struct {
	Damage   uint32
	Bullets  int32
	Distance float64
}

type Tank struct {
	ID            string
	Name          string
	Health        int32
	Speed         int32
	Width, Height int32
	LastShoot     int64
	Cmd           *Cmd
	ws            *Client

	TGun
	sync.RWMutex
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

func (t *Tank) IsColide() bool {
	if t.Cmd.X-t.Speed < 0 || t.Cmd.X+t.Speed > MapWidth || t.Cmd.Y-t.Speed < 0 || t.Cmd.Y+t.Speed > MapHeight {
		return true
	}
	return false
}

func (t *Tank) GetDamage(b *Bullet) error {
	t.Health -= int32(b.Tank.Damage)
	if t.IsDead() {
		t.ws.hub.sendToPushService("tank_stat", strconv.Itoa(int(pb.StatStatus_Death)), t.ID)
		b.Tank.ws.hub.sendToPushService("tank_stat", strconv.Itoa(int(pb.StatStatus_Kill)), b.Tank.ID)
		go func() {
			time.Sleep(time.Second * 3)
			t.Resurect()
			t.ws.sendProtoData(pb.BBGProtocol_TTankUpdate, t.ToProtobuf(), true)
		}()
	} else {
		if err := t.Save(); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tank) Resurect() error {
	if !t.IsDead() {
		log.Infof("Can't make a resurect. Tank #%s is alive.", t.ID)
		return nil
	}
	world.Update(t, func() {
		t.Health = 100
		t.Cmd.X = int32(random(10, MapWidth))
		t.Cmd.Y = int32(random(10, MapHeight))
	})
	if err := t.Save(); err != nil {
		return err
	}
	return nil
}

func (t *Tank) isFullReloaded() bool {
	return time.Now().UTC().Unix() < t.LastShoot+2
}

// Shoot command for tank
func (t *Tank) Shoot(pbMsg *pb.TankShoot) error {
	if t.IsDead() {
		log.Infof("Can't make a shoot. Tank #%s is dead.", t.ID)
		return nil
	}

	if t.isFullReloaded() {
		return nil
	}

	t.LastShoot = time.Now().UTC().Unix()
	t.Cmd.MouseAxes.X = pbMsg.MouseAxes.GetX()
	t.Cmd.MouseAxes.Y = pbMsg.MouseAxes.GetY()

	bullet, err := NewBullet(t)
	if err != nil {
		return err
	}
	go bullet.Update()

	if err := t.Save(); err != nil {
		return err
	}

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
func (t *Tank) TurretRotate(pbMsg *pb.TankRotate) error {
	if t.IsDead() {
		log.Infof("Can't make a turret rotation. Tank #%s is dead.", t.ID)
		return nil
	}
	t.Cmd.MouseAxes.X = pbMsg.MouseAxes.GetX()
	t.Cmd.MouseAxes.Y = pbMsg.MouseAxes.GetY()
	t.UpdateAngle()
	if err := t.Save(); err != nil {
		return err
	}
	return nil
}

func (t *Tank) Move(pbMsg *pb.TankMove) error {
	if t.IsDead() {
		log.Infof("Can't make a move. Tank #%s is dead.", t.ID)
		return nil
	}
	world.Update(t, func() {
		t.Cmd.Direction = pbMsg.GetDirection()

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

		if t.IsColide() {
			switch t.Cmd.Direction {
			case pb.Direction_N:
				t.Cmd.Y += t.Speed
			case pb.Direction_S:
				t.Cmd.Y -= t.Speed
			case pb.Direction_E:
				t.Cmd.X -= t.Speed
			case pb.Direction_W:
				t.Cmd.X += t.Speed
			}
		}
	})

	if err := t.Save(); err != nil {
		return err
	}
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
		Id:        t.Cmd.ID,
		TankId:    t.ID,
		X:         t.Cmd.X,
		Y:         t.Cmd.Y,
		Health:    t.Health,
		Name:      t.Name,
		Bullets:   t.Bullets,
		Speed:     t.Speed,
		Direction: t.Cmd.Direction,
		Angle:     t.Cmd.Angle,
		Damage:    t.Damage,
		Status:    status,
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
		ID:     pbMsg.GetId(),
		Name:   pbMsg.GetName(),
		Health: pbMsg.GetHealth(),
		Speed:  pbMsg.GetSpeed(),
		Width:  pbMsg.GetWidth(),
		Height: pbMsg.GetHeight(),
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
		ws: c,
	}
	world.Add(tank)
	// restore if dead
	tank.Resurect()
	return tank, nil
}

func (t *Tank) Update(pbMsg *pb.Tank) error {
	world.Update(t, func() {
		t.Lock()
		{
			t.ID = pbMsg.GetId()
			t.Cmd.X = pbMsg.GetX()
			t.Cmd.Y = pbMsg.GetY()
			t.Name = pbMsg.GetName()
			t.Health = pbMsg.GetHealth()
			t.Speed = pbMsg.GetSpeed()
			t.Width = pbMsg.GetWidth()
			t.Height = pbMsg.GetHeight()
			t.TGun.Bullets = pbMsg.Gun.GetBullets()
			t.TGun.Damage = pbMsg.Gun.GetDamage()
			t.TGun.Distance = pbMsg.Gun.GetDistance()
			t.Cmd.Angle = pbMsg.GetAngle()
			t.Cmd.Direction = pbMsg.GetDirection()
		}
		t.Unlock()
	})
	t.Save()
	return nil
}

func (t *Tank) Save() error {
	pbMsg := &pb.Tank{
		Id:     t.ID,
		X:      t.Cmd.X,
		Y:      t.Cmd.Y,
		Name:   t.Name,
		Health: t.Health,
		Speed:  t.Speed,
		Width:  t.Width,
		Height: t.Height,
		Gun: &pb.TankGun{
			Damage:   t.TGun.Damage,
			Bullets:  t.TGun.Bullets,
			Distance: t.TGun.Distance,
		},
		Angle:     t.Cmd.Angle,
		Direction: t.Cmd.Direction,
	}

	encoded, err := proto.Marshal(pbMsg)
	if err != nil {
		return err
	}
	if _, err := t.ws.hub.redis.HSet(tHash, t.ID, encoded).Result(); err != nil {
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
