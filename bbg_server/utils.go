package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/DeV1doR/bbg/bbg_server/engine"
	"github.com/DeV1doR/bbg/bbg_server/engine/tmx"
	pb "github.com/DeV1doR/bbg/bbg_server/protobufs"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
)

func Keys(v interface{}) ([]string, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil, errors.New("not a map")
	}
	t := rv.Type()
	if t.Key().Kind() != reflect.String {
		return nil, errors.New("not string key")
	}
	var result []string
	for _, kv := range rv.MapKeys() {
		result = append(result, kv.String())
	}
	return result, nil
}

func FillStruct(m map[string]interface{}, s interface{}) error {
	structValue := reflect.ValueOf(s).Elem()

	for name, value := range m {
		structFieldValue := structValue.FieldByName(name)

		if !structFieldValue.IsValid() {
			return fmt.Errorf("No such field: %s in obj", name)
		}

		if !structFieldValue.CanSet() {
			return fmt.Errorf("Cannot set %s field value", name)
		}

		val := reflect.ValueOf(value)
		if structFieldValue.Type() != val.Type() {
			return errors.New("Provided value type didn't match obj field type")
		}

		structFieldValue.Set(val)
	}
	return nil
}

func AddFloat64(val *float64, delta float64) (new float64) {
	for {
		old := *val
		new = old + delta
		if atomic.CompareAndSwapUint64(
			(*uint64)(unsafe.Pointer(val)),
			math.Float64bits(old),
			math.Float64bits(new),
		) {
			break
		}
	}
	return
}

func AngleFromP2P(iX float64, iY float64, tX float64, tY float64) float64 {
	return math.Atan2(tY-iY, tX-iX)
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getTanksToProtobuf(hub *Hub) (tanks []*pb.TankUpdate) {
	for client, active := range hub.clients {
		if active && client.tank != nil {
			tanks = append(tanks, client.tank.ToProtobuf())
		}
	}
	return
}

func getBulletsToProtobuf(hub *Hub) (bullets []*pb.BulletUpdate) {
	for client, active := range hub.clients {
		if active && client.tank != nil && client.tank.bullets != nil {
			for _, bullet := range client.tank.bullets {
				bullets = append(bullets, bullet.ToProtobuf())
			}
		}
	}
	return
}

func getScoresToProtobuf(hub *Hub) (scores []*pb.ScoreUpdate) {
	data, err := hub.redis.HGetAll("bbg:scores").Result()
	if err != nil {
		log.Errorln(err)
		return
	}
	for _, val := range data {
		pbMsg := &pb.ScoreUpdate{}
		if err := proto.Unmarshal([]byte(val), pbMsg); err != nil {
			log.Errorln(err)
			return
		}
		scores = append(scores, pbMsg)
	}
	// log.Fatalf("%+v\n", pbMsg)
	// scores = []string{"DeV1doR - 14k", "Niga - 12k"}
	return
}

func ReadTmxAndUpdateMap(fName string) error {
	_, fileName, _, _ := runtime.Caller(1)
	fName = strings.Join([]string{path.Dir(fileName), "/", fName}, "")
	r, err := os.Open(fName)
	if err != nil {
		log.Errorln("Tmx: Error during open: ", err)
		return err
	}
	defer r.Close()

	m, err := tmx.Read(r)
	if err != nil {
		log.Errorln("Tmx: Error during read: ", err)
		return err
	}

	for _, objectGroup := range m.ObjectGroups {
		for _, object := range objectGroup.Objects {
			world.Add(&engine.MapObject{
				X:      object.X,
				Y:      object.Y,
				Width:  object.Width,
				Height: object.Height,
				Type:   objectGroup.Name,
			})
		}
	}
	log.Infoln("Tmx: Successfully read")
	return nil
}
