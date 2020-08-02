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

type Geocoder interface {
	Latlong(location string) (*Coordinates, error)
	ParsedLocation(location string) (string, error)
}

var _ Geocoder = &OpenCageData{}

type OpenCageData struct {
	response *OpenCageDataGeocodeResponse
}

func init() {
	if ApiKey == "" {
		panic("must provide OpenCageData API key to use this Geocoding package (export GEOCODING_KEY)")
	}
}

type OpenCageDataGeocodeResponse struct {
	Documentation string `json:"documentation"`
	Licenses      []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"licenses"`
	Rate struct {
		Limit     int `json:"limit"`
		Remaining int `json:"remaining"`
		Reset     int `json:"reset"`
	} `json:"rate"`
	Results []struct {
		Annotations struct {
			DMS struct {
				Lat string `json:"lat"`
				Lng string `json:"lng"`
			} `json:"DMS"`
			FIPS struct {
				County string `json:"county"`
				State  string `json:"state"`
			} `json:"FIPS"`
			MGRS       string `json:"MGRS"`
			Maidenhead string `json:"Maidenhead"`
			Mercator   struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"Mercator"`
			OSM struct {
				EditURL string `json:"edit_url"`
				NoteURL string `json:"note_url"`
				URL     string `json:"url"`
			} `json:"OSM"`
			UNM49 struct {
				Regions struct {
					AMERICAS        string `json:"AMERICAS"`
					NORTHERNAMERICA string `json:"NORTHERN_AMERICA"`
					US              string `json:"US"`
					WORLD           string `json:"WORLD"`
				} `json:"regions"`
				StatisticalGroupings []string `json:"statistical_groupings"`
			} `json:"UN_M49"`
			Callingcode int `json:"callingcode"`
			Currency    struct {
				AlternateSymbols     []string `json:"alternate_symbols"`
				DecimalMark          string   `json:"decimal_mark"`
				DisambiguateSymbol   string   `json:"disambiguate_symbol"`
				HTMLEntity           string   `json:"html_entity"`
				IsoCode              string   `json:"iso_code"`
				IsoNumeric           string   `json:"iso_numeric"`
				Name                 string   `json:"name"`
				SmallestDenomination int      `json:"smallest_denomination"`
				Subunit              string   `json:"subunit"`
				SubunitToUnit        int      `json:"subunit_to_unit"`
				Symbol               string   `json:"symbol"`
				SymbolFirst          int      `json:"symbol_first"`
				ThousandsSeparator   string   `json:"thousands_separator"`
			} `json:"currency"`
			Flag     string  `json:"flag"`
			Geohash  string  `json:"geohash"`
			Qibla    float64 `json:"qibla"`
			Roadinfo struct {
				DriveOn string `json:"drive_on"`
				SpeedIn string `json:"speed_in"`
			} `json:"roadinfo"`
			Sun struct {
				Rise struct {
					Apparent     int `json:"apparent"`
					Astronomical int `json:"astronomical"`
					Civil        int `json:"civil"`
					Nautical     int `json:"nautical"`
				} `json:"rise"`
				Set struct {
					Apparent     int `json:"apparent"`
					Astronomical int `json:"astronomical"`
					Civil        int `json:"civil"`
					Nautical     int `json:"nautical"`
				} `json:"set"`
			} `json:"sun"`
			Timezone struct {
				Name         string `json:"name"`
				NowInDst     int    `json:"now_in_dst"`
				OffsetSec    int    `json:"offset_sec"`
				OffsetString string `json:"offset_string"`
				ShortName    string `json:"short_name"`
			} `json:"timezone"`
			What3Words struct {
				Words string `json:"words"`
			} `json:"what3words"`
			Wikidata string `json:"wikidata"`
		} `json:"annotations"`
		Bounds struct {
			Northeast struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"northeast"`
			Southwest struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"southwest"`
		} `json:"bounds"`
		Components struct {
			ISO31661Alpha2 string `json:"ISO_3166-1_alpha-2"`
			ISO31661Alpha3 string `json:"ISO_3166-1_alpha-3"`
			Category       string `json:"_category"`
			Type           string `json:"_type"`
			City           string `json:"city"`
			Continent      string `json:"continent"`
			Country        string `json:"country"`
			CountryCode    string `json:"country_code"`
			County         string `json:"county"`
			State          string `json:"state"`
			StateCode      string `json:"state_code"`
		} `json:"components"`
		Confidence int    `json:"confidence"`
		Formatted  string `json:"formatted"`
		Geometry   struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
	} `json:"results"`
	Status struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
	StayInformed struct {
		Blog    string `json:"blog"`
		Twitter string `json:"twitter"`
	} `json:"stay_informed"`
	Thanks    string `json:"thanks"`
	Timestamp struct {
		CreatedHTTP string `json:"created_http"`
		CreatedUnix int    `json:"created_unix"`
	} `json:"timestamp"`
	TotalResults int `json:"total_results"`
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

func (o *OpenCageData) Latlong(location string) (*Coordinates, error) {
	var response *OpenCageDataGeocodeResponse
	if o.response == nil {
		response, err := doGeocode(location)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch lat/long for location '%s'", location)
		}
		o.response = response
	} else {
		response = o.response
	}
	return &Coordinates{
		Latitude:  response.Results[0].Geometry.Lat,
		Longitude: response.Results[0].Geometry.Lng,
	}, nil
}

func (o *OpenCageData) Geocode(location string) (*OpenCageDataGeocodeResponse, error) {
	response, err := doGeocode(location)
	if err != nil {
		return nil, err
	}

	o.response = response
	return response, nil
}

func (o *OpenCageData) ParsedLocation(location string) (string, error) {
	var response *OpenCageDataGeocodeResponse
	if o.response == nil {
		response, err := doGeocode(location)
		if err != nil {
			return "", errors.Wrapf(err, "failed to fetch coordinates for location %s", location)
		}
		o.response = response
	} else {
		response = o.response
	}

	result := response.Results[0].Components
	return fmt.Sprintf("%s, %s, %s", result.City, result.State, result.Country), nil
}

func (o *OpenCageData) Map(location string) (string, error) {
	response, err := doGeocode(location)
	if err != nil {
		return "", errors.Wrapf(err, "failed to fetch map for location '%s'", location)
	}

	return response.Results[0].Annotations.OSM.URL, nil
}
