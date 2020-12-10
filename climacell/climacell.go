package climacell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gigawhitlocks/weather/geocoding"
	geo "github.com/gigawhitlocks/weather/geocoding"
	"github.com/pkg/errors"
)

type ClimaCell struct {
	ApiKey          string
	GeocodingApiKey string
}

const apiURL string = "https://api.climacell.co/v3"

func NewClimaCell(apiKey, geocodingApiKey string) *ClimaCell {
	return &ClimaCell{ApiKey: apiKey, GeocodingApiKey: geocodingApiKey}
}

type Observation struct {
	ParsedLocation string

	*ClimaCellObservation
}

func (c *ClimaCell) CurrentConditions(location string) (*Observation, error) {
	geocoder, err := geo.NewOpenCageData(location, c.GeocodingApiKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find geocoding information for '%s'", location)
	}
	coords := geocoder.Latlong()
	parsedLocation := geocoder.ParsedLocation()

	q := c.buildURL("/weather/nowcast",
		&QueryParams{
			flags: map[string]string{
				"start_time":  "now",
				"timestep":    "5",
				"unit_system": "us",
				"lat":         fmt.Sprintf("%0.4f", coords.Latitude),
				"lon":         fmt.Sprintf("%0.4f", coords.Longitude),
			},
			fields: []string{
				"baro_pressure",
				"cloud_base",
				"cloud_ceiling",
				"cloud_cover",
				"dewpoint",
				"feels_like",
				"humidity",
				"precipitation",
				"precipitation_type",
				"sunrise",
				"sunset",
				"surface_shortwave_radiation",
				"visibility",
				"weather_code",
				"wind_direction",
				"wind_gust",
				"temp",
			},
		})

	resp, err := http.Get(q)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current weather from ClimaCell")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body from response")
	}

	cco := []*ClimaCellObservation{}
	err = json.Unmarshal(body, &cco)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON from body")
	}
	if len(cco) == 0 {
		return nil, errors.New("unmarshaled ClimaCell observations from JSON without error but failed to get results")
	}
	return &Observation{ClimaCellObservation: cco[0], ParsedLocation: parsedLocation}, nil

}
func (c *ClimaCell) MarkdownCurrentConditions(location string) (string, error) {

	cco, err := c.CurrentConditions(location)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("| Current Conditions | %s | Location  | %s |\n| :--- | ---: | :--- | ---: |\n%s", cco.Title(), cco.ParsedLocation, cco.String()), nil
}

func isValidFeature(feature string) bool {
	_, ok := map[string]interface{}{
		"precipitation":   nil,
		"temp":            nil,
		"wind_speed":      nil,
		"wind_direction":  nil,
		"wind_gust":       nil,
		"visibility":      nil,
		"baro_pressure ":  nil,
		"dewpoint":        nil,
		"humidity":        nil,
		"cloud_cover":     nil,
		"cloud_base":      nil,
		"cloud_ceiling":   nil,
		"cloud_satellite": nil,
	}[strings.ToLower(feature)]
	return ok
}

func (c *ClimaCell) BuildMap(location string, features ...string) ([]byte, error) {
	geocoder, err := geo.NewOpenCageData(location, c.GeocodingApiKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find geocoding information for '%s'", location)
	}
	validFeatures := []string{}

	for _, feature := range features {
		if isValidFeature(feature) {
			validFeatures = append(validFeatures, feature)
		}
	}

	coords := geocoder.Latlong()
	zoom := 7
	tiles := geocoding.CoordsToSlippyMapTiles(coords, zoom)
	mapImage, err := getOpenStreetMapLayers(tiles)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get open street map layers")
	}
	first := true
	var img image.Image
	if len(validFeatures) == 0 {
		validFeatures = []string{"precipitation"}
	}

	for _, feature := range validFeatures {
		weatherLayers, err := c.getWeatherLayer(tiles, feature)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get %s layer", feature)
		}
		if first {
			img = imaging.Overlay(mapImage, weatherLayers, image.Point{X: 0, Y: 0}, .7)
			first = false
		} else {
			img = imaging.Overlay(img, weatherLayers, image.Point{X: 0, Y: 0}, .7)
		}
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	return buf.Bytes(), err
}

var titleTextMap map[string]string = map[string]string{
	"freezing_rain_heavy": "Heavy Freezing Rain",
	"freezing_rain":       "Freezing Rain",
	"freezing_rain_light": "Light Freezing Rain",
	"freezing_drizzle":    "Freezing Drizzle",
	"ice_pellets_heavy":   "Heavy Ice Pellets",
	"ice_pellets":         "Ice Pellets",
	"ice_pellets_light":   "Light Ice Pellets",
	"snow_heavy":          "Heavy Snow",
	"snow":                "Snow",
	"snow_light":          "Light Snow",
	"flurries":            "Flurries",
	"tstorm":              "Thunderstorm",
	"rain_heavy":          "Downpour",
	"rain":                "Rain",
	"rain_light":          "Light Rain",
	"drizzle":             "Drizzle",
	"fog_light":           "Light Fog",
	"fog":                 "Fog",
	"cloudy":              "Cloudy",
	"mostly_cloudy":       "Mostly Cloudy",
	"partly_cloudy":       "Partly Cloudy",
	"mostly_clear":        "Mostly Clear",
	"clear":               "Clear",
}

func (c *ClimaCellObservation) Title() (titleText string) {
	titleText, ok := titleTextMap[c.WeatherCode.Value]
	if !ok {
		return c.WeatherCode.Value
	}
	return
}

// possible weather fields
// 		"baro_pressure",
// 		"cloud_base",
// 		"cloud_ceiling",
// 		"cloud_cover",
// 		"cloud_satellite",
// 		"dewpoint",
// 		"feels_like",
// 		"humidity",
// 		"moon_phase",
// 		"precipitation",
// 		"precipitation_accumulation",
// 		"precipitation_probability",
// 		"precipitation_type",
// 		"sunrise",
// 		"sunset",
// 		"surface_shortwave_radiation",
// 		"visibility",
// 		"weather_code",
// 		"wind_direction",
// 		"wind_gust",
// 		"temp",

func (c *ClimaCell) getWeatherLayer(tiles [4]*geo.SlippyMapTile, feature string) (image.Image, error) {
	images := [4]*image.Image{}
	for i := 0; i < 4; i++ {
		q := c.buildURL(fmt.Sprintf("/weather/layers/%s/now/%d/%d/%d.png", feature, tiles[i].Z, tiles[i].X, tiles[i].Y),
			&QueryParams{
				flags:  map[string]string{},
				fields: []string{},
			})
		resp, err := http.Get(q)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, errors.Errorf("got status code %d", resp.StatusCode)
		}

		weatherLayer, err := png.Decode(resp.Body)
		if err != nil {
			return nil, err
		}
		images[i] = &weatherLayer
	}
	return assembleMapTiles(images), nil
}

func (c *ClimaCell) getPrecipitationLayer(tiles [4]*geo.SlippyMapTile) (image.Image, error) {
	return c.getWeatherLayer(tiles, "precipitation")
}

func getOpenStreetMapLayers(tiles [4]*geo.SlippyMapTile) (image.Image, error) {
	images := [4]*image.Image{}
	for i := 0; i < 4; i++ {
		server := []string{"a", "b", "c"}[rand.Int()%3]
		url := fmt.Sprintf("https://%s.tile.openstreetmap.org/%d/%d/%d.png",
			server, tiles[i].Z, tiles[i].X, tiles[i].Y)
		resp, err := http.Get(url)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get map tile from osm")
		}

		mapImage, err := png.Decode(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode map tile from osm")
		}
		images[i] = &mapImage
	}

	return assembleMapTiles(images), nil
}

func assembleMapTiles(tiles [4]*image.Image) *image.NRGBA {
	dst := imaging.New(512, 512, color.NRGBA{0, 0, 0, 0})
	dst = imaging.Paste(dst, *tiles[0], image.Pt(0, 0))
	dst = imaging.Paste(dst, *tiles[1], image.Pt(256, 0))
	dst = imaging.Paste(dst, *tiles[2], image.Pt(0, 256))
	dst = imaging.Paste(dst, *tiles[3], image.Pt(256, 256))
	return dst
}
