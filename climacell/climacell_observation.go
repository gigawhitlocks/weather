package climacell

import (
	"bytes"
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
		Parse(`
Temperature: {{.Temp.Value}} °{{.Temp.Units}}
Feels Like: {{.FeelsLike.Value}} °{{.FeelsLike.Units}}
Dewpoint: {{.Dewpoint.Value}} °{{.Dewpoint.Units}}
Wind Gust: {{.WindGust.Value}} {{.WindGust.Units}}
Barometric Pressure: {{.BaroPressure.Value}} {{.BaroPressure.Units}}
Visibility: {{.Visibility.Value}} {{.Visibility.Units}}
Precipitation: {{.Precipitation.Value}} {{.Precipitation.Units}}
Cloud Cover: {{.CloudCover.Value}}{{.CloudCover.Units}}{{if .CloudCeiling.Value }}
Cloud Ceiling: {{.CloudCeiling.Value}} {{.CloudCeiling.Units}} 
{{end}}{{if .CloudBase.Value }}
Cloud Base: {{.CloudBase.Value}}{{.CloudBase.Units}}
{{end}}Humidity: {{.Humidity.Value}}{{.Humidity.Units}}
Latitude and Longitude: {{.Lat}}, {{.Lon}}`)
	buffer := new(bytes.Buffer)
	_ = t.Execute(buffer, c)
	return buffer.String()
}
