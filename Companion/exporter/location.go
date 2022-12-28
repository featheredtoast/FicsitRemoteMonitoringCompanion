package exporter

import (
	"math"
	"strconv"
)

type Location struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	Rotation float64 `json:"rotation"`
}

// Calculates if a location is nearby another.
// From observation, 5000 units is "good enough" to be considered nearby.
func (l *Location) isNearby(other Location) bool {
	x := l.X - other.X
	y := l.Y - other.Y
	z := l.Z - other.Z

	dist := math.Sqrt(math.Pow(x, 2) + math.Pow(y, 2) + math.Pow(z, 2))
	return dist <= 5000
}

// Calculates if this location is roughly facing the same way as another
func (l *Location) isSameDirection(other Location) bool {
	diff := math.Abs(float64(l.Rotation - other.Rotation))
	return diff <= 90
}

func toString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
func (l *Location) XString() string {
	return toString(l.X)
}
func (l *Location) YString() string {
	return toString(l.Y)
}
func (l *Location) ZString() string {
	return toString(l.Z)
}
func (l *Location) RotationString() string {
	return toString(l.Rotation)
}
