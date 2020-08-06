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
	"os"

	"github.com/disintegration/imaging"
	"github.com/gigawhitlocks/weather/geocoding"
	"github.com/pkg/errors"
)

var apiKey = os.Getenv("WEATHER_KEY")
var apiURL string = "https://api.climacell.co/v3"

func init() {
	if apiKey == "" {
		panic("must provide ClimaCell API key (export WEATHER_KEY) to use this package")
	}
}

func CurrentConditions(location string) (string, []byte, error) {
	geocoder, err := geocoding.NewOpenCageData(location)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to find geocoding information for '%s'", location)
	}
	coords := geocoder.Latlong()
	parsedLocation := geocoder.ParsedLocation()

	q := buildURL("/weather/nowcast",
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
		return "", nil, errors.Wrap(err, "failed to get current weather from ClimaCell")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to read body from response")
	}

	cco := []*ClimaCellObservation{}
	err = json.Unmarshal(body, &cco)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to unmarshal JSON from body")
	}
	if len(cco) == 0 {
		return "", nil, errors.New("unmarshaled ClimaCell observations from JSON without error but failed to get results")
	}

	zoom := 7
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

	mapImage, err := getOpenStreetMapLayers(tiles)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to get open street map layers")
	}

	weatherLayers, err := getPrecipitationLayer(tiles)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to get precipitation map")
	}

	image := imaging.Overlay(mapImage, weatherLayers, image.Point{X: 0, Y: 0}, 1)
	buf := new(bytes.Buffer)
	err = png.Encode(buf, image)

	return fmt.Sprintf("| Current Conditions | %s | Location  | %s |\n| :--- | ---: | :--- | ---: |\n%s", cco[0].Title(), parsedLocation, cco[0].String()),
		buf.Bytes(), nil
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

func getWeatherLayer(tiles [4]*SlippyMapTile, feature string) (image.Image, error) {
	images := [4]*image.Image{}
	for i := 0; i < 4; i++ {
		q := buildURL(fmt.Sprintf("/weather/layers/%s/now/%d/%d/%d.png", feature, tiles[i].Z, tiles[i].X, tiles[i].Y),
			&QueryParams{
				flags:  map[string]string{},
				fields: []string{},
			})
		resp, err := http.Get(q)
		if err != nil {
			return nil, err
		}

		weatherLayer, err := png.Decode(resp.Body)
		if err != nil {
			return nil, err
		}
		images[i] = &weatherLayer
	}
	return assembleMapTiles(images), nil
}

func getPrecipitationLayer(tiles [4]*SlippyMapTile) (image.Image, error) {
	return getWeatherLayer(tiles, "precipitation")
}

func getOpenStreetMapLayers(tiles [4]*SlippyMapTile) (image.Image, error) {
	images := [4]*image.Image{}
	for i := 0; i < 4; i++ {
		server := []string{"a", "b", "c"}[rand.Int()%3]
		url := fmt.Sprintf("https://%s.tile.openstreetmap.org/%d/%d/%d.png",
			server, tiles[i].Z, tiles[i].X, tiles[i].Y)
		fmt.Println("XXX " + url)
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
