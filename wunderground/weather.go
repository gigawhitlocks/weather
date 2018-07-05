package wunderground

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"image/gif"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var CityStatePattern, _ = regexp.Compile("[A-Za-z]+,?[ \t]+[A-Za-z]+")
var ZipPattern, _ = regexp.Compile("[0-9]{5}")

var APIKey = os.Getenv("WUNDERGROUND_API_KEY")

type IntOrNANString struct {
	value string
}

func (i *IntOrNANString) String() string {
	return i.value
}

func (i *IntOrNANString) UnmarshalJSON(data []byte) (err error) {
	var d interface{}

	if err = json.Unmarshal(data, &d); err != nil {
		return err
	}
	i.value = fmt.Sprintf("%s", d)
	return nil
}

type ResponseFeatures struct {
	Conditions int `json:"conditions"`
}

type ResponseMetadata struct {
	Version          string            `json:"version"`
	TermsofService   string            `json:"termsofService"`
	ResponseFeatures *ResponseFeatures `json:"features"`
}

type CurrentObservation struct {
	Image                 *Image               `json:"image"`
	DisplayLocation       *DisplayLocation     `json:"display_location"`
	ObservationLocation   *ObservationLocation `json:"observation_location"`
	StationId             string               `json:"station_id"`
	ObservationTime       string               `json:"observation_time"`
	ObservationTimeRfc822 string               `json:"observation_time_rfc822"`
	ObservationEpoch      string               `json:"observation_epoch"`
	LocalTimeRfc822       string               `json:"local_time_rfc822"`
	LocalEpoch            string               `json:"local_epoch"`
	LocalTzShort          string               `json:"local_tz_short"`
	LocalTzLong           string               `json:"local_tz_long"`
	LocalTzOffset         string               `json:"local_tz_offset"`
	Weather               string               `json:"weather"`
	TemperatureString     string               `json:"temperature_string"`
	TempF                 float32              `json:"temp_f"`
	TempC                 float32              `json:"temp_c"`
	RelativeHumidity      string               `json:"relative_humidity"`
	WindString            string               `json:"wind_string"`
	WindDir               string               `json:"wind_dir"`
	WindDegrees           float32              `json:"wind_degrees"`
	WindMph               IntOrNANString       `json:"wind_mph"`
	WindGustMph           IntOrNANString       `json:"wind_gust_mph"`
	WindKph               IntOrNANString       `json:"wind_kph"`
	WindGustKph           IntOrNANString       `json:"wind_gust_kph"`
	PressureMb            string               `json:"pressure_mb"`
	PressureIn            string               `json:"pressure_in"`
	PressureTrend         string               `json:"pressure_trend"`
	DewpointString        string               `json:"dewpoint_string"`
	DewpointF             IntOrNANString       `json:"dewpoint_f"`
	DewpointC             IntOrNANString       `json:"dewpoint_c"`
	HeatIndexString       IntOrNANString       `json:"heat_index_string"`
	HeatIndexF            IntOrNANString       `json:"heat_index_f"`
	HeatIndexC            IntOrNANString       `json:"heat_index_c"`
	WindchillString       string               `json:"windchill_string"`
	WindchillF            IntOrNANString       `json:"windchill_f"`
	WindchillC            IntOrNANString       `json:"windchill_c"`
	FeelslikeString       string               `json:"feelslike_string"`
	FeelslikeF            IntOrNANString       `json:"feelslike_f"`
	FeelslikeC            IntOrNANString       `json:"feelslike_c"`
	VisibilityMi          string               `json:"visibility_mi"`
	VisibilityKm          string               `json:"visibility_km"`
	Solarradiation        string               `json:"solarradiation"`
	UV                    string               `json:"UV"`
	Precip1hrIn           string               `json:"precip_1hr_in"`
	Precip1hrMetric       string               `json:"precip_1hr_metric"`
	Precip1hrString       string               `json:"precip_1hr_string"`
	PrecipTodayString     string               `json:"precip_today_string"`
	PrecipTodayIn         string               `json:"precip_today_in"`
	PrecipTodayMetric     string               `json:"precip_today_metric"`
	Icon                  string               `json:"icon"`
	IconUrl               string               `json:"icon_url"`
	ForecastUrl           string               `json:"forecast_url"`
	HistoryUrl            string               `json:"history_url"`
	ObUrl                 string               `json:"ob_url"`
	Nowcast               string               `json:"nowcast"`
}

type Image struct {
	Url   string `json:"url"`
	Title string `json:"title"`
	Link  string `json:"link"`
}

type DisplayLocation struct {
	Full           string `json:"full"`
	City           string `json:"city"`
	State          string `json:"state"`
	StateName      string `json:"state_name"`
	Country        string `json:"country"`
	CountryIso3166 string `json:"country_iso3166"`
	Zip            string `json:"zip"`
	Magic          string `json:"magic"`
	Wmo            string `json:"wmo"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	Elevation      string `json:"elevation"`
}

type ObservationLocation struct {
	Full           string `json:"full"`
	City           string `json:"city"`
	State          string `json:"state"`
	Country        string `json:"country"`
	CountryIso3166 string `json:"country_iso3166"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	Elevation      string `json:"elevation"`
}

func (w *CurrentConditions) String() string {
	t := template.New("CurrentConditions")
	if w.Precip1hrIn == "-999.00" {
		w.Precip1hrIn = "0"
	}
	t, _ = t.Parse(`From {{.ObservationLocation.Full}}
{{.CurrentObservation.ObservationTime}} it was {{.Weather}}
Temperature was {{.TemperatureString}}; felt like {{.FeelslikeString}}
with relative humidity {{.RelativeHumidity}}, Wind {{.WindString}}, and {{.Precip1hrIn}}" of precipitation in the last hour.
Dewpoint {{.DewpointString}}
`)

	// {{ if NAN not in .HeatIndexString }} Heat index: {{.HeatIndexString}} {{end}}
	// {{ if NAN not in .WindchillString }} Wind chill: {{.WindchillString}} {{end}}
	buf := new(bytes.Buffer)
	t.Execute(buf, w)

	return buf.String()
}

// CurrentConditions is the outer type for the current conditions API endpoint
type CurrentConditions struct {
	ResponseMetadata   `json:"response"`
	CurrentObservation `json:"current_observation"`
}

// ForecastDay comes from the Forecast API and represents one day
type ForecastDay struct {
	Period        int32  `json:"period"`
	Icon          string `json:"icon"`
	IconUrl       string `json:"icon_url"`
	Title         string `json:"title"`
	Fcttext       string `json:"fcttext"`
	FcttextMetric string `json:"fcttext_metric"`
	Pop           string `json:"pop"`
}

// TxtForecast is a collection of metadta and ForecastDays representing a week of Forecast API data
type TxtForecast struct {
	Date        string        `json:"date"`
	ForecastDay []ForecastDay `json:"forecastday"`
}

type forecast struct {
	TxtForecast `json:"txt_forecast"`
}

func (w *Weather) String() string {
	return fmt.Sprintf("%s\n%s",
		w.CurrentConditions.String(),
		w.Forecast.String())

}
func (f *Forecast) String() string {
	today := f.ForecastDay[0]
	t := template.New("forecast")
	t, _ = t.Parse(`{{ .Fcttext }}
{{.Pop}}% chance of precipitation`)
	buf := new(bytes.Buffer)
	t.Execute(buf, today)
	return buf.String()
}

// Forecast is the outer container type for a the 5-day forecast API
type Forecast struct {
	forecast `json:"forecast"`
}

type Weather struct {
	CurrentConditions
	Forecast
}

func (c *CurrentConditions) Get(url string, results chan interface{}) {
	var resp *http.Response
	var err error
	if resp, err = http.Get(url); err != nil {
		results <- err
		return
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(c); err != nil {
		results <- err
	}

	results <- c
}

func (f *Forecast) Get(url string, results chan interface{}) {
	var resp *http.Response
	var err error
	if resp, err = http.Get(url); err != nil {
		results <- err
		return
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(f); err != nil {
		results <- err
	}

	results <- f
}

func getWeather(currentConditions, forecastURL string) (w *Weather, err error) {
	results := make(chan interface{})
	go new(Forecast).Get(forecastURL, results)
	go new(CurrentConditions).Get(currentConditions, results)

	var currCond *CurrentConditions
	var forecast *Forecast
	for i := 0; i < 2; i++ {
		r := <-results
		switch r := r.(type) {
		case error:
			return nil, r.(error)
		case *CurrentConditions:
			currCond = r
		case *Forecast:
			forecast = r
		default:
			return nil, fmt.Errorf("Couldn't decode %+v", &r)
		}

	}

	return &Weather{
		CurrentConditions: *currCond,
		Forecast:          *forecast,
	}, nil

}

func CleanCityState(query string) []string {
	location := strings.SplitN(query, ",", 2)
	if len(location) != 2 {
		location = strings.SplitN(query, " ", 2)
	}
	location[0] = strings.TrimSpace(location[0])
	location[1] = strings.TrimSpace(location[1])

	return location
}

func GetRadar(query string) (result *gif.GIF) {

	if !CityStatePattern.MatchString(query) {
		return nil
	}
	fmt.Printf("%s\n", query)

	location := CleanCityState(query)
	// http://api.wunderground.com/api/17234af6deee4427/radar/q/KS/Topeka.gif?width=280&height=280&newmaps=1

	url := fmt.Sprintf(
		"https://api.wunderground.com/api/%s/animatedradar/q/%s/%s.gif?width=400&height=400&newmaps=1",
		APIKey,
		location[1],
		location[0])

	var resp *http.Response
	var err error
	if resp, err = http.Get(url); err != nil {
		return nil
	}
	if result, err = gif.DecodeAll(resp.Body); err != nil {
		return nil
	}
	return
}

func GetWeather(query string) (result *Weather, err error) {

	// results := make(chan interface{})
	// var resp *http.Response
	if CityStatePattern.MatchString(query) {

		location := CleanCityState(query)
		cc := fmt.Sprintf(
			"https://api.wunderground.com/api/%s/conditions/q/%s/%s.json",
			APIKey,
			location[1],
			location[0])

		forecast := fmt.Sprintf(
			"https://api.wunderground.com/api/%s/features/forecast/q/%s/%s.json",
			APIKey,
			location[1],
			location[0])

		return getWeather(cc, forecast)
	} else if ZipPattern.MatchString(query) {
		cc := fmt.Sprintf(
			"https://api.wunderground.com/api/%s/conditions/q/%s.json",
			APIKey,
			query)

		forecast := fmt.Sprintf(
			"https://api.wunderground.com/api/%s/features/forecast/q/%s.json",
			APIKey, query)

		return getWeather(cc, forecast)

	}

	return nil, fmt.Errorf("Query malformed; provide ZIP or City, St.")
}
