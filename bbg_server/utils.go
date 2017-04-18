package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sync/atomic"
	"unsafe"

	pb "github.com/DeV1doR/bbg/bbg_server/protobufs"
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

func random(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getTanksToProtobuf(hub *Hub) (tanks []*pb.TankUpdate) {
	hub.RLock()
	defer hub.RUnlock()
	for client, active := range hub.clients {
		if active && client.tank != nil {
			tanks = append(tanks, client.tank.ToProtobuf())
		}
	}
	return
}

func getBulletsToProtobuf(hub *Hub) (bullets []*pb.BulletUpdate) {
	// TODO
	return
}
