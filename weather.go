package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

type Result struct {
	BarometricPressure    float32
	Conditions            string
	HeatIndex             string
	Name                  string
	PrecipitationLastHour float32
	RelativeHumidity      string
	Station               string
	Temperature           string
	TemperatureValue      string
	Timestamp             string
	WindGust              float32
	WindSpeed             float32
}

type zipCode string
type latLong [2]string

var zipMap map[zipCode]latLong

const NWSAPI string = "https://api.weather.gov"

type StationProperties struct {
	StationIdentifier string `json:"stationIdentifier"`
	Name              string `json:"name"`
}

type StationFeature struct {
	Id         string            `json:"id"`
	Properties StationProperties `json:"properties"`
}

type Station struct {
	Features []StationFeature `json:"features"`
}

func (s *Station) ID() string {
	if len(s.Features) < 1 {
		return ""
	}
	return s.Features[0].Properties.StationIdentifier
}

type ObservationProperty struct {
	Value          float32 `json:"value"`
	UnitCode       string  `json:"unitCode"`
	QualityControl string  `json:"qualityControl"`
}

type ObservationProperties struct {
	Station                   string              `json:"station"`
	Timestamp                 string              `json:"timestamp"`
	Icon                      string              `json:"icon"`
	TextDescription           string              `json:"textDescription"`
	Temperature               ObservationProperty `json:"temperature"`
	Dewpoint                  ObservationProperty `json:"dewpoint"`
	WindDirection             ObservationProperty `json:"windDirection"`
	WindSpeed                 ObservationProperty `json:"windSpeed"`
	WindGust                  ObservationProperty `json:"windGust"`
	BarometricPressure        ObservationProperty `json:"barometricPressure"`
	SeaLevelPressure          ObservationProperty `json:"seaLevelPressure"`
	Visibility                ObservationProperty `json:"visibility"`
	MaxTemperatureLast24Hours ObservationProperty `json:"maxTemperatureLast24Hours"`
	MinTemperatureLast24Hours ObservationProperty `json:"minTemperatureLast24Hours"`
	PrecipitationLastHour     ObservationProperty `json:"precipitationLastHour"`
	PrecipitationLast3Hours   ObservationProperty `json:"precipitationLast3Hours"`
	PrecipitationLast6Hours   ObservationProperty `json:"precipitationLast6Hours"`
	RelativeHumidity          ObservationProperty `json:"relativeHumidity"`
	WindChill                 ObservationProperty `json:"windChill"`
	HeatIndex                 ObservationProperty `json:"heatIndex"`
}

type Observation struct {
	ObservationProperties `json:"properties"`
}

type NWSRequest struct {
	Client  *http.Client
	Request *http.Request
}

func NewRequest(uri string) (n *NWSRequest) {
	n = new(NWSRequest)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s", NWSAPI, uri), nil)
	req.Proto = "HTTP/1.1"
	req.Header.Set("Accept", "*/*")
	n.Client = client
	n.Request = req
	return
}

func (n *NWSRequest) Do() (*http.Response, error) {
	return n.Client.Do(n.Request)
}

func zipToLatLong(z zipCode) (latLong, error) {
	if l, ok := zipMap[z]; ok {
		return l, nil
	}
	return latLong{}, fmt.Errorf("zip code not found")
}

func readZips() map[zipCode]latLong {
	var zipMap = make(map[zipCode]latLong)
	ziptext, err := ioutil.ReadFile("zip-data.csv")
	if err != nil {
		fmt.Println("no file")
	}
	r := csv.NewReader(strings.NewReader(string(ziptext)))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("%s", err)
			return nil
		}
		if record[0] == "ZIP" {
			continue
		}
		trimRecord := func(record string) string {
			record = strings.Trim(record, " ")
			s, _ := strconv.ParseFloat(record, 64)
			return fmt.Sprintf("%.1f", s)
		}
		zipMap[zipCode(record[0])] = latLong([2]string{
			trimRecord(record[1]),
			trimRecord(record[2])})
	}
	return zipMap
}

func stationFromZip(z zipCode) (output *Station, err error) {
	l, err := zipToLatLong(z)
	if err != nil {
		return nil, err
	}
	fmt.Println(l)
	n := NewRequest(fmt.Sprintf(
		"points/%s,%s/stations",
		l[0], l[1]))

	resp, err := n.Do()
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if resp.Status != "200 OK" {
		fmt.Println(resp.Status)
		// buf, _ := ioutil.ReadAll(resp.Body)
		// fmt.Printf("%s", string(buf))
		return nil, fmt.Errorf("bad response from nws\n")
	}

	output = new(Station)
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(output); err != nil {
		return nil, err
	}
	return
}

func getCurrentObservation(stationID string) (o *Observation, err error) {
	n := NewRequest(fmt.Sprintf(
		"/stations/%s/observations/current", stationID))
	resp, err := n.Do()
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	o = new(Observation)
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(o); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *Result) String() string {
	t := template.New("results")
	t, err := t.Parse(`Current Weather For {{.Name}}
Observatory: {{.Station}}
Time of Observation: {{.Timestamp}}
Conditions: {{.Conditions}}
Temperature: {{.Temperature}} F
Relative humidity: {{.RelativeHumidity}}%
Heat index: {{.HeatIndex}} F
Barometric pressure: {{.BarometricPressure}} Pa
Wind speed: {{.WindSpeed}} m/s
Wind gust: {{.WindGust}} m/s
Precipitation in the last hour: {{.PrecipitationLastHour}} m
`)
	if err != nil {
		fmt.Println(err.Error())
	}
	buf := new(bytes.Buffer)
	t.Execute(buf, o)
	return buf.String()
}

func toFahrenheit(in float32) string {
	return fmt.Sprintf("%.1f", in*1.8+32)
}

func main() {
	zipMap = readZips()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		zip := r.URL.Query().Get("zip")
		if len(zip) != 5 {
			fmt.Fprintf(w, "%s", zip)
			return
		}

		wthr, err := stationFromZip(zipCode(zip))
		if err != nil {
			fmt.Printf("%s", err.Error())
			return
		}

		o, err := getCurrentObservation(wthr.ID())
		if err != nil {
			fmt.Printf("%s", err.Error())
			return
		}

		fmt.Fprintf(w, "%s", &Result{
			Name:                  zip,
			Station:               wthr.ID(),
			Conditions:            o.TextDescription,
			Timestamp:             o.Timestamp,
			Temperature:           toFahrenheit(o.Temperature.Value),
			BarometricPressure:    o.BarometricPressure.Value,
			WindSpeed:             o.WindSpeed.Value,
			WindGust:              o.WindGust.Value,
			PrecipitationLastHour: o.PrecipitationLastHour.Value,
			HeatIndex:             toFahrenheit(o.HeatIndex.Value),
			RelativeHumidity:      fmt.Sprintf("%.2f", o.RelativeHumidity.Value),
		})
	})

	fmt.Printf("%s", http.ListenAndServe(":8080", nil))
}
