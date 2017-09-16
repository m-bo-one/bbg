package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/DeV1doR/bbg/bbg_server/engine"
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
	LastShoot     float64
	Cmd           *Cmd
	ws            *Client
	bullets       []*Bullet

	TGun
	sync.RWMutex
}

func (t *Tank) GetX() int32 {
	return t.Cmd.X - t.GetWidth()/2
}

func (t *Tank) GetY() int32 {
	return t.Cmd.Y - t.GetHeight()/2
}

func (t *Tank) GetWidth() int32 {
	return int32(t.Width)
}

func (t *Tank) GetHeight() int32 {
	return int32(t.Height)
}

func (t *Tank) IsColide() bool {
	if t.Cmd.X-t.Speed < 0 ||
		t.Cmd.X+t.Speed > MapWidth ||
		t.Cmd.Y-t.Speed < 0 ||
		t.Cmd.Y+t.Speed > MapHeight {
		return true
	}
	for _, i := range world.Nearby(t) {
		switch i.(type) {
		case *engine.MapObject:
			return true
		}
	}
	return false
}

func (t *Tank) GetDamage(b *Bullet) error {
	t.Health -= int32(b.Tank.Damage)
	if t.IsDead() {
		world.Remove(t)
		go t.ws.hub.sendToPushService("tank_stat", strconv.Itoa(int(pb.StatStatus_Death)), t.ID)
		go b.Tank.ws.hub.sendToPushService("tank_stat", strconv.Itoa(int(pb.StatStatus_Kill)), b.Tank.ID)
		go func() {
			time.Sleep(time.Second * 3)
			t.Resurect()
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
		t.Cmd.X = int32(random(10, MapWidth))
		t.Cmd.Y = int32(random(10, MapHeight))
	})

	if len(world.Nearby(t)) != 0 {
		t.Resurect()
	}

	t.Health = 100
	if err := t.Save(); err != nil {
		return err
	}
	return nil
}

func (t *Tank) isFullReloaded() bool {
	return float64(time.Now().UTC().Unix()) >= t.LastShoot
}

// Shoot command for tank
func (t *Tank) Shoot(pbMsg *pb.TankShoot) error {
	if t.IsDead() {
		log.Infof("Can't make a shoot. Tank #%s is dead.", t.ID)
		return nil
	}
	t.Lock()
	if !t.isFullReloaded() {
		t.Unlock()
		return nil
	}
	t.LastShoot = float64(time.Now().UTC().Unix()) + 0.02
	t.Cmd.MouseAxes.X = pbMsg.MouseAxes.X
	t.Cmd.MouseAxes.Y = pbMsg.MouseAxes.Y
	t.Unlock()

	bullet, err := NewBullet(t)
	if err != nil {
		return err
	}
	t.bullets = append(t.bullets, bullet)

	if err := t.Save(); err != nil {
		return err
	}

	go t.ws.hub.sendToPushService("tank_stat", strconv.Itoa(int(pb.StatStatus_Shoot)), t.ID)

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
	t.Cmd.MouseAxes.X = pbMsg.MouseAxes.X
	t.Cmd.MouseAxes.Y = pbMsg.MouseAxes.Y
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
		t.Cmd.Direction = pbMsg.Direction

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
		ID:     pbMsg.Id,
		Name:   pbMsg.Name,
		Health: pbMsg.Health,
		Speed:  pbMsg.Speed,
		Width:  pbMsg.Width,
		Height: pbMsg.Height,
		TGun: TGun{
			Bullets:  pbMsg.Gun.Bullets,
			Damage:   pbMsg.Gun.Damage,
			Distance: pbMsg.Gun.Distance,
		},
		Cmd: &Cmd{
			X:         pbMsg.X,
			Y:         pbMsg.Y,
			Direction: pbMsg.Direction,
			Angle:     pbMsg.Angle,
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
			t.ID = pbMsg.Id
			t.Cmd.X = pbMsg.X
			t.Cmd.Y = pbMsg.Y
			t.Name = pbMsg.Name
			t.Health = pbMsg.Health
			t.Speed = pbMsg.Speed
			t.Width = pbMsg.Width
			t.Height = pbMsg.Height
			t.TGun.Bullets = pbMsg.Gun.Bullets
			t.TGun.Damage = pbMsg.Gun.Damage
			t.TGun.Distance = pbMsg.Gun.Distance
			t.Cmd.Angle = pbMsg.Angle
			t.Cmd.Direction = pbMsg.Direction
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
