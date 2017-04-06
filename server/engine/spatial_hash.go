package engine

import "math"

type object interface {
	GetX() int32
	GetY() int32
	GetRadius() int32
}

type Vector struct {
	X int32
	Y int32
}

type spatialHash struct {
	Width    int32
	Height   int32
	CellSize int32
	Objects  map[int32][]object
}

func (sh *spatialHash) hashID(v *Vector) int32 {
	return int32(math.Floor(float64(v.X/sh.CellSize))) +
		int32(math.Floor(float64(v.Y/sh.CellSize)))*sh.Width
}

func (sh *spatialHash) HashIds(o object) []int32 {
	ids := []int32{}
	min := &Vector{o.GetX() - o.GetRadius(), o.GetY() - o.GetRadius()}
	max := &Vector{o.GetX() + o.GetRadius(), o.GetY() + o.GetRadius()}

	_append := func(slice []int32, i int32) []int32 {
		for _, other := range slice {
			if other == i {
				return slice
			}
		}

		return append(slice, i)
	}

	add := func(v *Vector) {
		ids = _append(ids, sh.hashID(v))
	}

	// make a list of all hash IDs that
	// are hit by the four corners of the
	// Object's bounding box
	add(&Vector{min.X, max.Y}) // top left
	add(&Vector{max.X, max.Y}) // top right
	add(&Vector{max.X, min.Y}) // bottom right
	add(min)                   // bottom left

	return ids
}

func (sh *spatialHash) Add(o object) {
	ids := sh.HashIds(o)
	for _, id := range ids {
		sh.Objects[id] = append(sh.Objects[id], o)
	}
}

func (sh *spatialHash) Remove(o object) {
	ids := sh.HashIds(o)
	for _, id := range ids {
		for j, other := range sh.Objects[id] {
			if o == other {
				sh.Objects[id] = append(sh.Objects[id][:j], sh.Objects[id][j+1:]...)
			}
		}
	}
}

func (sh *spatialHash) Nearby(o object) []object {
	objects := []object{}
	ids := sh.HashIds(o)

	_append := func(slice []object, o object) []object {
		for _, other := range slice {
			if other == o {
				return slice
			}
		}

		return append(slice, o)
	}

	for _, id := range ids {
		for _, object := range sh.Objects[id] {
			if object != o {
				objects = _append(objects, object)
			}
		}
	}

	return objects
}

func NewSpatialHash(width int32, height int32) *spatialHash {
	return &spatialHash{
		Width:   width,
		Height:  height,
		Objects: make(map[int32][]object),
	}
}
