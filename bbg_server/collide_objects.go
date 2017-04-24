package main

type StaticObject struct {
	ID            uint32
	X, Y          int32
	Height, Width uint32
	Rotation      int32
	Name          string
	Visible       bool
	Type          string
	Refferer      string
}

func (so *StaticObject) GetX() int32 {
	return so.X
}

func (so *StaticObject) GetY() int32 {
	return so.Y
}

func (so *StaticObject) GetWidth() int32 {
	return int32(so.Width)
}

func (so *StaticObject) GetHeight() int32 {
	return int32(so.Height)
}

func (so *StaticObject) AddToMap() {
	world.Add(so)
}

func (so *StaticObject) RemoveFromMap() {
	world.Remove(so)
}

func (so *StaticObject) UpdateOnMap(callback func()) {
	world.Update(so, callback)
}
