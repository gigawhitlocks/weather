package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gigawhitlocks/weather/nws"
	"github.com/gigawhitlocks/weather/wunderground"
)

func main() {

	if len(wunderground.APIKey) == 0 {
		fmt.Println("Set WUNDERGROUND_API_KEY to your Wunderground API key")
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var err error
		q := r.URL.Query().Get("zip")
		q = strings.TrimPrefix(q, "!")

		switch {
		case strings.HasPrefix(q, "nws"):
			var result *nws.Result
			zip := strings.TrimSpace(strings.TrimPrefix(q, "nws"))
			if len(zip) != 5 {
				return
			}

			if result, err = nws.GetWeather(zip); err != nil {
				fmt.Fprintf(w, "%s", err)
				return
			}
			fmt.Fprintf(w, "%s", result)
			return
		case strings.HasPrefix(q, "weather"):
			query := strings.TrimSpace(strings.TrimPrefix(q, "weather"))
			var result *wunderground.Weather
			if result, err = wunderground.GetWeather(query); err != nil {
				fmt.Fprintf(w, "%s", err)
			}

			fmt.Fprintf(w, "%s", result.String())
			return
		case strings.HasPrefix(q, "forecast"):
			query := strings.TrimSpace(strings.TrimPrefix(q, "forecast"))
			var url string
			if wunderground.CityStatePattern.MatchString(query) {
				location := wunderground.CleanCityState(query)
				url = fmt.Sprintf(
					"https://api.wunderground.com/api/%s/features/forecast/q/%s/%s.json",
					wunderground.APIKey,
					location[1],
					location[0])
			} else if wunderground.ZipPattern.MatchString(query) {
				url = fmt.Sprintf(
					"https://api.wunderground.com/api/%s/features/forecast/q/%s.json",
					wunderground.APIKey, query)
			} else {
				fmt.Println("Invalid query string")
				return
			}
			var resp *http.Response
			if resp, err = http.Get(url); err != nil {
				fmt.Printf("Couldn't fetch from forecast API because %s", err)
				return
			}
			decoder := json.NewDecoder(resp.Body)

			f := new(wunderground.Forecast)
			if err := decoder.Decode(f); err != nil {
				fmt.Printf("Couldn't decode JSON because %s", err)
			}
			var forecasts []string
			for _, day := range f.TxtForecast.ForecastDay {
				forecasts = append(forecasts, fmt.Sprintf("*%s*: %s", day.Title, day.Fcttext))
			}

			fmt.Fprintf(w, "%s", strings.Join(forecasts, "\n"))

		default:
			fmt.Fprintf(w, "%s", "Invalid command")

		}
	})

	fmt.Printf("%s", http.ListenAndServe("0.0.0.0:8111", nil))
}
