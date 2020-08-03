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

	// tile := CoordinatesToTile(coords, 16)
	// q = buildURL(fmt.Sprintf("weather/layers/field/now/%d/%d/%d.png", tile.Z, tile.X, tile.Y),
	// 	&QueryParams{
	// 		flags: map[string]string{},
	// 		fields: []string{
	// 			"precipitation",
	// 		},
	// 	})
	// resp, err = http.Get(q)

	return fmt.Sprintf("| Current Conditions | %s | Location  | %s |\n| :--- | ---: | :--- | ---: |\n%s", cco[0].Title(), parsedLocation, cco[0].String()), nil
}

func (c *ClimaCellObservation) Title() (titleText string) {
	switch c.WeatherCode.Value {
	case "freezing_rain_heavy":
		titleText = "Heavy Freezing Rain"
	case "freezing_rain":
		titleText = "Freezing Rain"
	case "freezing_rain_light":
		titleText = "Light Freezing Rain"
	case "freezing_drizzle":
		titleText = "Freezing Drizzle"
	case "ice_pellets_heavy":
		titleText = "Heavy Ice Pellets"
	case "ice_pellets":
		titleText = "Ice Pellets"
	case "ice_pellets_light":
		titleText = "Light Ice Pellets"
	case "snow_heavy":
		titleText = "Heavy Snow"
	case "snow":
		titleText = "Snow"
	case "snow_light":
		titleText = "Light Snow"
	case "flurries":
		titleText = "Flurries"
	case "tstorm":
		titleText = "Thunderstorm"
	case "rain_heavy":
		titleText = "Downpour"
	case "rain":
		titleText = "Rain"
	case "rain_light":
		titleText = "Light Rain"
	case "drizzle":
		titleText = "Drizzle"
	case "fog_light":
		titleText = "Light Fog"
	case "fog":
		titleText = "Fog"
	case "cloudy":
		titleText = "Cloudy"
	case "mostly_cloudy":
		titleText = "Mostly Cloudy"
	case "partly_cloudy":
		titleText = "Partly Cloudy"
	case "mostly_clear":
		titleText = "Mostly Clear"
	case "clear":
		titleText = "Clear"
	default:
		titleText = ""
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
