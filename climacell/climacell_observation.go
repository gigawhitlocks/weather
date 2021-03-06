package climacell

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

type ClimaCellObservation struct {
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

func (c *ClimaCellObservation) String() string {
	t, _ := template.
		New("ClimaCellObservation").
		Parse(
			`| Latitude | {{.Lat}} | Longitude | {{.Lon}} |
| Temperature | {{.Temp.Value}} °{{.Temp.Units}} | Feels Like | {{.FeelsLike.Value}} °{{.FeelsLike.Units}} |{{if (ne .Precipitation.Value 0.0)}}
| Precipitation | {{.Precipitation.Value}} {{.Precipitation.Units}} | Type of Precipitation | {{.PrecipitationType.Value }} |{{end}}
| Wind Gust | {{.WindGust.Value}} {{.WindGust.Units}} | Barometric Pressure | {{.BaroPressure.Value}} {{.BaroPressure.Units}} |
| Humidity | {{.Humidity.Value}}{{.Humidity.Units}} | Cloud Cover | {{.CloudCover.Value}}{{.CloudCover.Units}} |
`)
	buffer := new(bytes.Buffer)
	err := t.Execute(buffer, c)
	if err != nil {
		fmt.Println(err.Error())
	}
	return buffer.String()
}
