package geocoding

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type Geocoder interface {
	Latlong() *Coordinates
	ParsedLocation() string
}

var _ Geocoder = &OpenCageData{}

type OpenCageData struct {
	ApiURL string
	*OpenCageDataGeocodeResponse
}

func NewOpenCageData(location, apiKey string) (*OpenCageData, error) {
	o := &OpenCageData{ApiURL: fmt.Sprintf("https://api.opencagedata.com/geocode/v1/json?key=%s", apiKey)}
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
	o.OpenCageDataGeocodeResponse, err = o.doGeocode(location)
	return
}

func (o *OpenCageData) ParsedLocation() string {
	result := o.Results[0].Components
	if result.City != "" && result.State != "" {
		if strings.ToUpper(result.CountryCode) == "US" {
			return fmt.Sprintf("%s, %s", result.City, result.State)
		} else {
			return fmt.Sprintf("%s, %s, %s", result.City, result.State, result.Country)
		}
	}

	if result.City != "" && result.State == "" {
		if result.CountryCode == "us" || result.CountryCode == "" {
			return fmt.Sprintf("%s", result.City)
		}
		if result.CountryCode != "us" {
			return fmt.Sprintf("%s, %s", result.City, result.Country)
		}
	}

	if result.State == "" {
		return result.Country
	}
	if result.Country == "" {
		return result.State
	}

	if result.CountryCode != "us" {
		return fmt.Sprintf("%s, %s", result.State, result.Country)
	} else {
		return fmt.Sprintf("%s", result.State)
	}
}

func (o *OpenCageData) Map(location string) string {
	return o.Results[0].Annotations.OSM.URL
}

func (o *OpenCageData) buildQuery(query string) string {
	return fmt.Sprintf("%s&q=%s", o.ApiURL, url.QueryEscape(query))
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

func (o *OpenCageData) doGeocode(location string) (*OpenCageDataGeocodeResponse, error) {
	response, err := doGet(o.buildQuery(location))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch coordinates for location %s", location)
	}

	if len(response.Results) == 0 {
		return nil, errors.Errorf("no results found for location '%s'", location)
	}

	return response, nil
}
