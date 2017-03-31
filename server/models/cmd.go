package models

import (
	pb "github.com/DeV1doR/bbg/server/protobufs"
)

type MouseAxes struct {
	X, Y int
}

type Cmd struct {
	ID        uint32
	X, Y      int32
	Action    string
	Direction pb.Direction
	PrevID    int32
	// MouseAxes *MouseAxes
}
