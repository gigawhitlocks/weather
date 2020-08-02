package climacell

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

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

type ClimaCellObservation []struct {
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	Temp struct {
		Value float64 `json:"value"`
		Units string  `json:"units"`
	} `json:"temp"`
	FeelsLike struct {
		Value float64 `json:"value"`
		Units string  `json:"units"`
	} `json:"feels_like"`
	Dewpoint struct {
		Value float64 `json:"value"`
		Units string  `json:"units"`
	} `json:"dewpoint"`
	WindGust struct {
		Value float64 `json:"value"`
		Units string  `json:"units"`
	} `json:"wind_gust"`
	BaroPressure struct {
		Value float64 `json:"value"`
		Units string  `json:"units"`
	} `json:"baro_pressure"`
	Visibility struct {
		Value float64 `json:"value"`
		Units string  `json:"units"`
	} `json:"visibility"`
	Precipitation struct {
		Value float64 `json:"value"`
		Units string  `json:"units"`
	} `json:"precipitation"`
	CloudCover struct {
		Value float64 `json:"value"`
		Units string  `json:"units"`
	} `json:"cloud_cover"`
	CloudCeiling struct {
		Value interface{} `json:"value"`
		Units string      `json:"units"`
	} `json:"cloud_ceiling"`
	CloudBase struct {
		Value interface{} `json:"value"`
		Units string      `json:"units"`
	} `json:"cloud_base"`
	SurfaceShortwaveRadiation struct {
		Value float64 `json:"value"`
		Units string  `json:"units"`
	} `json:"surface_shortwave_radiation"`
	Humidity struct {
		Value float64 `json:"value"`
		Units string  `json:"units"`
	} `json:"humidity"`
	WindDirection struct {
		Value float64 `json:"value"`
		Units string  `json:"units"`
	} `json:"wind_direction"`
	PrecipitationType struct {
		Value string `json:"value"`
	} `json:"precipitation_type"`
	Sunrise struct {
		Value time.Time `json:"value"`
	} `json:"sunrise"`
	Sunset struct {
		Value time.Time `json:"value"`
	} `json:"sunset"`
	ObservationTime struct {
		Value time.Time `json:"value"`
	} `json:"observation_time"`
	WeatherCode struct {
		Value string `json:"value"`
	} `json:"weather_code"`
}

func CurrentConditions(location string) (string, error) {
	coords, err := geocoding.Geocode(location)
	if err != nil {
		return "", errors.Wrapf(err, "failed to find latitude and longitude for '%s'", location)
	}
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
	fmt.Println(q)
	if err != nil {
		return "", errors.Wrap(err, "couldn't get current weather from ClimaCell")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "couldn't read body from response")
	}

	cco := &ClimaCellObservation{}
	err = json.Unmarshal(body, cco)
	if err != nil {
		return "", errors.Wrap(err, "couldn't unmarshal JSON from body")
	}
	fmt.Printf("XXXX CCO:\n%+v\n", cco)
	return string(body), nil
}

type QueryParams struct {
	flags  map[string]string
	fields []string
}

func (q QueryParams) String() string {
	flags := ""
	fields := fmt.Sprintf("fields=%s", strings.Join(q.fields, `%2C`))
	for key, value := range q.flags {
		flags = fmt.Sprintf("%s&%s=%s", flags, key, value)
	}
	return fmt.Sprintf("%s&%s", flags, fields)
}

func buildURL(endpoint string, queryParams *QueryParams) string {
	return fmt.Sprintf("%s%s?apikey=%s%s", apiURL, endpoint, apiKey, queryParams)
}
