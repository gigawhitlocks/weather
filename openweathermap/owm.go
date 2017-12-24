package openweathermap

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"net/http"
	"os"
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

var APIKEY = os.Getenv("OWM_API_KEY")

func GetSatellite(query string) (*image.Image, error) {
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
	s := fmt.Sprintf("https://sat.owm.io/sql/%d/%d/%d?APPID=%s&op=rgb&from=s2&select=b4,b3,b2&order=best", zoom, xtile, ytile, APIKEY)

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
