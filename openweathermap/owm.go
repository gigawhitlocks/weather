package openweathermap

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"os"
	"regexp"

	"github.com/disintegration/imaging"
)

var CityStatePattern, _ = regexp.Compile("[A-Z a-z]+(,?[ \t]+[A-Za-z]+)?")
var ZipPattern, _ = regexp.Compile("[0-9]{5}")

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

func GetSatellite(query string) (*image.NRGBA, error) {
	var err error
	var url string

	switch {
	case ZipPattern.MatchString(query):
		url = fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?zip=%s,us&appid=%s",
			query, APIKEY)
	default:
		return nil, fmt.Errorf("ah shit")

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

	zoom := 12
	xtile, ytile := tileNumbers(c.Lat, c.Long, zoom)

	// get the images
	type imagePos struct {
		Image *image.Image
		X     int
		Y     int
	}

	imagesChannel := make(chan interface{}, 9)
	for x := 0; x < 3; x++ {
		for y := 0; y < 3; y++ {
			go func(x, y int) {
				url := fmt.Sprintf(
					"https://sat.owm.io/sql/%d/%d/%d?APPID=%s&op=rgb&from=l8t&select=b4,b3,b2&order=best",
					zoom, xtile+(x-1), ytile+(y-1), APIKEY)
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
			}(x, y)
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

	// stitch together the tiles
	dst := imaging.New(768, 768, color.NRGBA{0, 0, 0, 0})
	for x := 0; x < len(images); x++ {
		for y := 0; y < len(images[x]); y++ {
			dst = imaging.Paste(dst, *images[x][y], image.Pt(256*x, 256*y))
		}
	}
	return dst, nil
}
