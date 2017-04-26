package engine

type MapObject struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Type   string
}

func (mo *MapObject) GetX() int32 {
	return int32(mo.X)
}

func (mo *MapObject) GetY() int32 {
	return int32(mo.Y)
}

func (mo *MapObject) GetWidth() int32 {
	return int32(mo.Width)
}

func (mo *MapObject) GetHeight() int32 {
	return int32(mo.Height)
}
