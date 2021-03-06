package geocoding

import (
	"math"
)

type SlippyMapTile struct {
	Z    int
	X    int
	Y    int
	Lat  float64
	Long float64

	PointX float64
	PointY float64
}

const (
	TopLeft = iota
	TopRight
	BottomLeft
	BottomRight
)

func CoordinatesToTile(coordinates *Coordinates, zoom int) *SlippyMapTile {
	tile := &SlippyMapTile{
		Lat:  coordinates.Latitude,
		Long: coordinates.Longitude,
		Z:    zoom,
	}

	x, y := tile.Deg2num(tile)
	tile.X = int(math.Floor(x))
	tile.Y = int(math.Floor(y))
	tile.PointX = math.Mod(x, 1.0) * 256
	tile.PointY = math.Mod(y, 1.0) * 256

	return tile
}

func (m *SlippyMapTile) Corner() int {
	if m.PointX > 128 && m.PointY < 128 {
		return BottomLeft
	} else if m.PointX < 128 && m.PointY < 128 {
		return BottomRight
	} else if m.PointX < 128 && m.PointY > 128 {
		return TopRight
	} else {
		return TopLeft
	}
}

type Conversion interface {
	deg2num(t *SlippyMapTile) (x int, y int)
	num2deg(t *SlippyMapTile) (lat float64, long float64)
}

func (*SlippyMapTile) Deg2num(t *SlippyMapTile) (x float64, y float64) {
	x = (t.Long + 180.0) / 360.0 * (math.Exp2(float64(t.Z)))
	y = (1.0 - math.Log(math.Tan(t.Lat*math.Pi/180.0)+1.0/math.Cos(t.Lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(t.Z)))
	return
}

func (*SlippyMapTile) Num2deg(t *SlippyMapTile) (lat float64, long float64) {
	n := math.Pi - 2.0*math.Pi*float64(t.Y)/math.Exp2(float64(t.Z))
	lat = 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long = float64(t.X)/math.Exp2(float64(t.Z))*360.0 - 180.0
	return lat, long
}

func CoordsToSlippyMapTiles(coords *Coordinates, zoom int) [4]*SlippyMapTile {
	tile := CoordinatesToTile(coords, zoom)
	var tiles [4]*SlippyMapTile
	corner := tile.Corner()
	switch corner {
	case TopLeft:
		tiles = [4]*SlippyMapTile{
			tile, {X: tile.X + 1, Y: tile.Y},
			{X: tile.X, Y: tile.Y + 1}, {X: tile.X + 1, Y: tile.Y + 1},
		}
	case TopRight:
		tiles = [4]*SlippyMapTile{
			{X: tile.X - 1, Y: tile.Y}, tile,
			{X: tile.X - 1, Y: tile.Y + 1}, {X: tile.X, Y: tile.Y + 1},
		}
	case BottomLeft:
		tiles = [4]*SlippyMapTile{
			{X: tile.X, Y: tile.Y - 1}, {X: tile.X + 1, Y: tile.Y - 1},
			tile, {X: tile.X + 1, Y: tile.Y},
		}
	case BottomRight:
		tiles = [4]*SlippyMapTile{
			{X: tile.X - 1, Y: tile.Y - 1}, {X: tile.X, Y: tile.Y - 1},
			{X: tile.X - 1, Y: tile.Y}, tile,
		}
	}

	for _, t := range tiles {
		t.Z = zoom
	}

	return tiles
}
