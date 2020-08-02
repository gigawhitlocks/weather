package geocoding

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

var ApiKey = os.Getenv("GEOCODING_KEY")
var ApiURL = fmt.Sprintf("https://api.opencagedata.com/geocode/v1/json?key=%s", ApiKey)

func init() {
	if ApiKey == "" {
		panic("must provide OpenCageData API key to use this Geocoding package (export GEOCODING_KEY)")
	}
}

type Geocoder interface {
	Latlong() *Coordinates
	ParsedLocation() string
}

var _ Geocoder = &OpenCageData{}

type OpenCageData struct {
	*OpenCageDataGeocodeResponse
}

func NewOpenCageData(location string) (*OpenCageData, error) {
	o := &OpenCageData{}
	err := o.Geocode(location)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (o *OpenCageData) Latlong() *Coordinates {
	return &Coordinates{
		Latitude:  o.Results[0].Geometry.Lat,
		Longitude: o.Results[0].Geometry.Lng,
	}
}

func (o *OpenCageData) Geocode(location string) (err error) {
	o.OpenCageDataGeocodeResponse, err = doGeocode(location)
	return
}

func (o *OpenCageData) ParsedLocation() string {
	result := o.Results[0].Components
	if result.City != "" && result.State != "" {
		return fmt.Sprintf("%s, %s", result.City, result.State)
	}
	if result.City != "" && result.State == "" {
		if result.CountryCode == "US" || result.CountryCode == "" {
			return fmt.Sprintf("%s", result.City)
		}
		if result.CountryCode != "US" {
			return fmt.Sprintf("%s, %s", result.City, result.CountryCode)
		}
	}

	if result.State == "" {
		return result.Country
	}
	if result.Country == "" {
		return result.State
	}
	return fmt.Sprintf("%s, %s", result.State, result.CountryCode)
}

func (o *OpenCageData) Map(location string) string {
	return o.Results[0].Annotations.OSM.URL
}

func buildQuery(query string) string {
	return fmt.Sprintf("%s&q=%s", ApiURL, url.QueryEscape(query))
}

func doGet(url string) (ocdgr *OpenCageDataGeocodeResponse, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch geocode data from Open Cage Data")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	ocdgr = new(OpenCageDataGeocodeResponse)
	err = json.Unmarshal(body, ocdgr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON from response body")
	}
	return
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

func doGeocode(location string) (*OpenCageDataGeocodeResponse, error) {
	response, err := doGet(buildQuery(location))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch coordinates for location %s", location)
	}

	if len(response.Results) == 0 {
		return nil, errors.Errorf("no results found for location '%s'", location)
	}

	return response, nil
}
