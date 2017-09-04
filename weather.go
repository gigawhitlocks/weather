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
	Alerts                []Alert
}

type zipCode string
type latLong [2]float64

var zipMap map[zipCode]latLong

const NWSAPI string = "https://api.weather.gov"

type StationListProperties struct {
	StationIdentifier string `json:"stationIdentifier"`
	Name              string `json:"name"`
}

type StationListFeature struct {
	Id         string                `json:"id"`
	Properties StationListProperties `json:"properties"`
}

type StationList struct {
	Features []StationListFeature `json:"features"`
}

func (s *StationList) ID(which int) string {
	if len(s.Features) < which {
		return ""
	}
	return s.Features[which].Properties.StationIdentifier
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
		convFloat := func(record string) float64 {
			record = strings.Trim(record, " ")
			s, _ := strconv.ParseFloat(record, 64)
			return s
		}
		zipMap[zipCode(record[0])] = latLong([2]float64{
			convFloat(record[1]),
			convFloat(record[2])})
	}
	return zipMap
}

func stationsFromZip(z zipCode) (output *StationList, err error) {
	l, err := zipToLatLong(z)
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	for i := 2; i >= 0; i-- {
		points := fmt.Sprintf(fmt.Sprintf("%%.%df,%%.%df", i, i), l[0], l[1])
		n := NewRequest(fmt.Sprintf(
			"points/%s/stations", points))

		resp, err = n.Do()
		defer resp.Body.Close()
		if err != nil {
			continue
		}

		if resp.Status == "200 OK" {
			break
		}
	}
	if resp.Status != "200 OK" {
		return nil, fmt.Errorf("Bad response from NWS")
	}
	output = new(StationList)
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(output); err != nil {
		return nil, err
	}
	return
}

type AlertProperties struct {
	Severity    string `json:"severity"`
	Certainty   string `json:"certainty"`
	Urgency     string `json:"urgency"`
	Event       string `json:"event"`
	Sender      string `json:"sender"`
	Headline    string `json:"headline"`
	Description string `json:"description"`
	Instruction string `json:"instruction"`
}

type Alert struct {
	AlertProperties `json:"properties"`
}

type AlertList struct {
	Alerts []Alert `json:"features"`
}

type StationProperties struct {
	County string `json:"county"`
}

type Station struct {
	StationProperties `json:"properties"`
}

func getCountyCode(stationID string) string {
	n := NewRequest(fmt.Sprintf(
		"/stations/%s", stationID))
	resp, err := n.Do()
	defer resp.Body.Close()
	if err != nil {
		return ""
	}
	s := new(Station)
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(s); err != nil {
		return ""
	}

	parts := strings.Split(s.County, "/")
	return parts[len(parts)-1]
}

func getCurrentAlerts(stationID string) (a *AlertList, err error) {
	c := getCountyCode(stationID)
	n := NewRequest(fmt.Sprintf(
		"/alerts/active/zone/%s", c))
	resp, err := n.Do()
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	a = new(AlertList)
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(a); err != nil {
		return nil, err
	}
	fmt.Printf("%+v", a)
	return a, nil
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
Barometric pressure: {{.BarometricPressure}} in Hg
Wind speed: {{.WindSpeed}} m/s
Wind gust: {{.WindGust}} m/s
Precipitation in the last hour: {{.PrecipitationLastHour}} m
`)
	if err != nil {
		fmt.Println(err.Error())
	}
	buf := new(bytes.Buffer)
	t.Execute(buf, o)

	if len(o.Alerts) <= 0 {
		return buf.String()
	}

	t = template.New("Alerts")
	t, err = t.Parse(`
**{{.Headline}}**
Severity: {{.Severity}}
Certainty: {{.Certainty}}
Urgency: {{.Urgency}}
{{.Description}}
{{.Instruction}}
`)
	for _, a := range o.Alerts {
		t.Execute(buf, a)
	}

	return buf.String()
}

func toFahrenheit(in float32) string {
	return fmt.Sprintf("%.1f", in*1.8+32)
}

func toInchesHg(pascals float32) float32 {
	return pascals / 3386.38866
}

func main() {
	zipMap = readZips()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		zip := r.URL.Query().Get("zip")
		if len(zip) != 5 {
			fmt.Fprintf(w, "%s", zip)
			return
		}

		wthr, err := stationsFromZip(zipCode(zip))
		if err != nil {
			fmt.Printf("%s", err.Error())
			return
		}

		i := 0
		o, err := getCurrentObservation(wthr.ID(i))
		if err != nil {
			fmt.Printf("%s", err.Error())
			return
		}
		for o.Timestamp == "" {
			i++
			if wthr.ID(i) == "" {
				fmt.Fprintln(w, "No forecast found :(")
				return
			}
			o, err = getCurrentObservation(wthr.ID(i))
			if err != nil {
				fmt.Printf("%s", err.Error())
				continue
			}
		}
		stationName := wthr.Features[i].Properties.Name
		a, err := getCurrentAlerts(wthr.ID(i))

		if err != nil {
			fmt.Println(err.Error())
			a = nil
		}

		fmt.Fprintf(w, "%s", &Result{
			Name:                  zip,
			Station:               stationName,
			Conditions:            o.TextDescription,
			Timestamp:             o.Timestamp,
			Temperature:           toFahrenheit(o.Temperature.Value),
			BarometricPressure:    toInchesHg(o.BarometricPressure.Value),
			WindSpeed:             o.WindSpeed.Value,
			WindGust:              o.WindGust.Value,
			PrecipitationLastHour: o.PrecipitationLastHour.Value,
			HeatIndex:             toFahrenheit(o.HeatIndex.Value),
			RelativeHumidity:      fmt.Sprintf("%.2f", o.RelativeHumidity.Value),
			Alerts:                a.Alerts,
		})
	})

	fmt.Printf("%s", http.ListenAndServe(":8111", nil))
}
