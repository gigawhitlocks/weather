package openweathermap

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"net/http"
	"regexp"
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

func GetSatellite(query string) (*image.Image, error) {
	var err error
	var url string

	switch {
	// case CityStatePattern.MatchString(query):
	// url = fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=bdedb4052d36b03e61a5768dfdaff8a5", query)
	case ZipPattern.MatchString(query):
		url = fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?zip=%s,us&appid=bdedb4052d36b03e61a5768dfdaff8a5", query)
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
	// s := fmt.Sprintf("https://tile.openweathermap.org/map/precipitation_new/%d/%d/%d.png?appid=bdedb4052d36b03e61a5768dfdaff8a5", zoom, xtile, ytile)
	// s = fmt.Sprintf("%s\nhttps://sat.owm.io/sql/%d/%d/%d?APPID=bdedb4052d36b03e61a5768dfdaff8a5&op=rgb&from=l8&select=b4,b3,b2&order=first", s, zoom, xtile, ytile)
	// https: //{s}.sat.owm.io/sql/{z}/{x}/{y}?appid={APIKEY}&op=rgb&from=s2&select=b4,b3,b2&order=best
	s := fmt.Sprintf("https://sat.owm.io/sql/%d/%d/%d?APPID=bdedb4052d36b03e61a5768dfdaff8a5&op=rgb&from=s2&select=b4,b3,b2&order=best", zoom, xtile, ytile)

	if resp, err = http.Get(s); err != nil {
		return nil, err
	}

	var satellite image.Image
	satellite, err = png.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return &satellite, nil
	// if resp, err = http.Get(); err != nil {
	// 	return nil, err
	// }

	// p := new(PrecipitationMap)
	// decoder = json.NewDecoder(resp.Body)
	// if err = decoder.Decode(c); err != nil {
	// 	return nil, err
	// }

}
