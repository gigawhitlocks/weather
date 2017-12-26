package openweathermap

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"net/http"
	"os"
	"regexp"

	"github.com/disintegration/imaging"
)

var CityStatePattern, _ = regexp.Compile("[A-Z a-z]+(,?[ \t]+[A-Za-z]+)?")
var ZipPattern, _ = regexp.Compile("[0-9]{5}")

const zoom = 7

type PrecipitationMap struct {
}

type Coordinates struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"lon"`
}
type Location struct {
	Coordinates `json:"coord"`
}

func tileNumbers(lat, long float64, zoom int) (int, int) {
	latRad := lat * (math.Pi / 180)
	n := math.Pow(2.0, float64(zoom))
	xtile := int((long + 180.0) / 360.0 * n)
	ytile := int((1.0 - math.Log(math.Tan(latRad)+(1/math.Cos(latRad)))/math.Pi) / 2.0 * n)
	return xtile, ytile
}

var APIKEY = os.Getenv("OWM_API_KEY")

type Map int

const (
	_        = iota
	Base Map = 1 + iota
	Clouds
	Precipitation
)

func GetTiles(xtile, ytile int, mt Map) (*image.NRGBA, error) {
	var baseurl string
	var querystring string

	if mt == Base {
		baseurl = "https://sat.owm.io/sql"
		// querystring = fmt.Sprintf("?APPID=%s&op=rgb&from=l8&select=b4,b3,b2&order=best", APIKEY)
		querystring = fmt.Sprintf("?APPID=%s&op=rgb&from=cloudless&select=red,green,blue&order=best", APIKEY)
	} else if mt == Clouds || mt == Precipitation {
		baseurl = "https://tile.openweathermap.org/map"
		var layer string
		if mt == Clouds {
			layer = "/clouds_new"
		} else {
			layer = "/precipitation_new"
		}
		querystring = fmt.Sprintf(".png?cities=true&appid=%s", APIKEY)
		baseurl = fmt.Sprintf("%s%s", baseurl, layer)
	} else {
		return nil, fmt.Errorf("Unrecognized map type requested")
	}

	// get the images
	type imagePos struct {
		Image *image.Image
		X     int
		Y     int
	}

	imagesChannel := make(chan interface{}, 9)
	for x := 0; x < 3; x++ {
		for y := 0; y < 3; y++ {
			url := fmt.Sprintf(
				"%s/%d/%d/%d%s",
				baseurl,
				zoom,
				xtile+(x-1),
				ytile+(y-1),
				querystring)
			fmt.Printf("%s\n", url)

			go func(x, y int, url string) {
				var resp *http.Response
				var err error
				if resp, err = http.Get(url); err != nil {
					imagesChannel <- err
					return
				}
				var satellite image.Image
				satellite, err = png.Decode(resp.Body)
				if err != nil {
					imagesChannel <- err
					return
				}

				imagesChannel <- imagePos{
					Image: &satellite,
					X:     x,
					Y:     y,
				}
			}(x, y, url)
		}
	}

	images := [3][3]*image.Image{
		{nil, nil, nil},
		{nil, nil, nil},
		{nil, nil, nil},
	}

	for i := 0; i < 9; i++ {
		im := <-imagesChannel
		switch im := im.(type) {
		case error:
			return nil, im
		case imagePos:
			if im.Image == nil {
				return nil, fmt.Errorf("Image was nil")
			}
			images[im.X][im.Y] = im.Image
		}
	}

	return stitchTiles(images), nil
}

func GetTileNumbers(query string) (*Location, error) {
	var err error
	var url string

	switch {
	case ZipPattern.MatchString(query):
		url = fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?zip=%s,us&appid=%s",
			query, APIKEY)
	default:
		return nil, fmt.Errorf("Satellite endpoint only accepts zip codes :(")
	}

	var resp *http.Response
	if resp, err = http.Get(url); err != nil {
		return nil, err
	}

	c := new(Location)
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(c); err != nil {
		return nil, err
	}

	return c, nil
}

func GetSatellite(c *Location) (*image.NRGBA, error) {
	xtile, ytile := tileNumbers(c.Lat, c.Long, zoom)
	return GetTiles(xtile, ytile, Base)
}

func GetComposite(query string) (*image.NRGBA, error) {
	l, err := GetTileNumbers(query)
	if err != nil {
		return nil, err
	}
	type MapResult struct {
		self    *image.NRGBA
		mapType Map
	}
	xtile, ytile := tileNumbers(l.Lat, l.Long, zoom)
	results := make(chan interface{}, 3)

	go func() {
		var basemap *image.NRGBA
		if basemap, err = GetSatellite(l); err != nil {
			results <- err
		}
		results <- &MapResult{
			self:    basemap,
			mapType: Base,
		}
	}()

	// go func() {
	// 	var basemap *image.NRGBA
	// 	if basemap, err = GetTiles(xtile, ytile, Clouds); err != nil {
	// 		results <- err
	// 	}
	// 	results <- &MapResult{
	// 		self:    basemap,
	// 		mapType: Clouds,
	// 	}
	// }()

	go func() {
		var basemap *image.NRGBA
		if basemap, err = GetTiles(xtile, ytile, Precipitation); err != nil {
			results <- err
		}
		results <- &MapResult{
			self:    basemap,
			mapType: Precipitation,
		}

	}()

	im := make(map[Map]*image.NRGBA)
	for x := 0; x < 2; x++ {
		r := <-results
		switch r := r.(type) {
		case *MapResult:
			im[r.mapType] = r.self
		case error:
			return nil, r
		}
	}

	result := im[Base]
	// result = imaging.Overlay(result, im[Precipitation], image.Pt(0, 0), .5)
	m := image.NewRGBA(image.Rect(0, 0, 512*3, 512*3))
	green := color.RGBA{0, 255, 0, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{green}, image.ZP, draw.Src)
	draw.DrawMask(m, m.Bounds(), m, image.Pt(0, 0), im[Precipitation], image.Pt(0, 0), draw.Src)

	result = imaging.Overlay(result, m, image.Pt(0, 0), .9)
	return result, nil
}

func stitchTiles(images [3][3]*image.Image) *image.NRGBA {
	// stitch together the tiles
	dst := imaging.New(768, 768, color.NRGBA{0, 0, 0, 0})
	for x := 0; x < len(images); x++ {
		for y := 0; y < len(images[x]); y++ {
			dst = imaging.Paste(dst, *images[x][y], image.Pt(256*x, 256*y))
		}
	}
	return dst
}
