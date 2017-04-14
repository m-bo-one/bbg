package main

import (
	pb "github.com/DeV1doR/bbg/protobufs"
)

type MouseAxes struct {
	X, Y float64
}

type Bounds struct {
	Width, Height uint32
}

type Cmd struct {
	ID        uint32
	X, Y      int32
	Angle     float64
	Action    string
	Direction pb.Direction
	PrevID    int32
	MouseAxes *MouseAxes
}
