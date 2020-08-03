package climacell

import (
	"math"

	"github.com/gigawhitlocks/weather/geocoding"
)

type MapTile struct {
	Z    int
	X    int
	Y    int
	Lat  float64
	Long float64
}

func CoordinatesToTile(coordinates *geocoding.Coordinates, zoom int) *MapTile {
	tile := &MapTile{
		Lat:  coordinates.Latitude,
		Long: coordinates.Longitude,
		Z:    zoom,
	}
	tile.X, tile.Y = tile.Deg2num(tile)

	return tile
}

type Conversion interface {
	deg2num(t *MapTile) (x int, y int)
	num2deg(t *MapTile) (lat float64, long float64)
}

func (*MapTile) Deg2num(t *MapTile) (x int, y int) {
	x = int(math.Floor((t.Long + 180.0) / 360.0 * (math.Exp2(float64(t.Z)))))
	y = int(math.Floor((1.0 - math.Log(math.Tan(t.Lat*math.Pi/180.0)+1.0/math.Cos(t.Lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(t.Z)))))
	return
}

func (*MapTile) Num2deg(t *MapTile) (lat float64, long float64) {
	n := math.Pi - 2.0*math.Pi*float64(t.Y)/math.Exp2(float64(t.Z))
	lat = 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long = float64(t.X)/math.Exp2(float64(t.Z))*360.0 - 180.0
	return lat, long
}
