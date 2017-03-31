package models

import (
	"strconv"

	pb "github.com/DeV1doR/bbg/server/protobufs"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
)

const dbKey string = "bbg:tanks"

type Tank struct {
	ID       uint32
	Health   int32
	FireRate int32
	Bullets  int32
	Speed    int32
	Cmd      *Cmd
}

func (t *Tank) Shoot() error {
	return nil
}

func (t *Tank) Stop() error {
	return nil
}

func (t *Tank) Move(direction *pb.Direction) error {
	t.Cmd.Direction = *direction
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
	}
}

func NewTank(redis *redis.Client) (*Tank, error) {
	pk, err := redis.Incr("tanks:id").Result()
	if err != nil {
		return nil, err
	}
	direction := pb.Direction_N
	t := &Tank{
		ID:       uint32(pk),
		Health:   100,
		FireRate: 100,
		Bullets:  100,
		Speed:    5,
		Cmd: &Cmd{
			X:         200,
			Y:         200,
			Direction: direction,
			// MouseAxes: &MouseAxes{
			// 	X: 0,
			// 	Y: 0,
			// },
		},
	}
	encoded, err := proto.Marshal(t.ToProtobuf())
	if err != nil {
		return nil, err
	}
	if err := redis.HSet(dbKey, strconv.FormatInt(pk, 10), encoded).Err(); err != nil {
		return nil, err
	}
	return t, nil
}

func LoadTank(redis *redis.Client, pk *uint32) (*Tank, error) {
	val, err := redis.HGet(dbKey, strconv.Itoa(int(*pk))).Result()
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
			// MouseAxes: &MouseAxes{
			// 	X: 0,
			// 	Y: 0,
			// },
		},
	}, nil
}

func RemoveTank(redis *redis.Client, pk *uint32) (uint32, error) {
	_, err := redis.HDel(dbKey, strconv.Itoa(int(*pk))).Result()
	if err != nil {
		return 0, err
	}
	return *pk, nil
}
