package climacell

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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

func CurrentConditions(location string) (string, error) {
	geocoder, err := geocoding.NewOpenCageData(location)
	if err != nil {
		return "", errors.Wrapf(err, "failed to find geocoding information for '%s'", location)
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
		return "", errors.Wrap(err, "failed to get current weather from ClimaCell")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read body from response")
	}

	cco := []*ClimaCellObservation{}
	err = json.Unmarshal(body, &cco)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal JSON from body")
	}
	if len(cco) == 0 {
		return "", errors.New("unmarshaled ClimaCell observations from JSON without error but failed to get results")
	}
	return fmt.Sprintf("| Found Location  | %s | Query | %s |\n| :--- | ---: | :--- | ---: |\n%s", parsedLocation, location, cco[0].String()), nil
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
