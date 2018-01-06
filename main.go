package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gigawhitlocks/weather/gfs"
	"github.com/gigawhitlocks/weather/nws"
	"github.com/gigawhitlocks/weather/openweathermap"
	"github.com/gigawhitlocks/weather/wunderground"

	"golang.org/x/sync/syncmap"
)

func main() {

	var imagestore = new(syncmap.Map)
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
			if query == "" {
				help(w)
				return
			}
			var result *wunderground.Weather
			if result, err = wunderground.GetWeather(query); err != nil {
				fmt.Fprintf(w, "%s", err)
			}

			fmt.Fprintf(w, "%s", result.String())
			return
		case strings.HasPrefix(q, "forecast"):
			query := strings.TrimSpace(strings.TrimPrefix(q, "forecast"))
			if query == "" {
				help(w)
				return
			}
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
			return

		case strings.HasSuffix(q, ".gif"):
			i, ok := imagestore.Load(q)
			if !ok {
				fmt.Fprintf(w, "%s", fmt.Errorf("Not ok"))
				return
			}
			w.Header().Set("Content-Type", "image/gif")
			switch i := i.(type) {
			case *gif.GIF:
				if err := gif.EncodeAll(w, i); err != nil {
					fmt.Fprintf(w, "%s", err)
					return
				}
			}
			return

		case strings.HasSuffix(q, ".png"):
			i, ok := imagestore.Load(q)
			if !ok {
				fmt.Fprintf(w, "%s", fmt.Errorf("Not ok"))
				return
			}
			w.Header().Set("Content-Type", "image/png")
			switch i := i.(type) {
			case *image.NRGBA:
				if err := png.Encode(w, i); err != nil {
					fmt.Fprintf(w, "%s", err)
					return
				}
			}
			return

		case strings.HasPrefix(q, "satellite"):
			query := strings.TrimSpace(strings.TrimPrefix(q, "satellite"))
			var err error
			var result *image.NRGBA
			var loc *openweathermap.Location
			if loc, err = openweathermap.GetTileNumbers(query); err != nil {
				fmt.Fprintf(w, "Bad result from OpenWeatherMap API: %s", err)
				return
			}

			if result, err = openweathermap.GetSatellite(loc); err != nil {
				fmt.Fprintf(w, "Bad result from OpenWeatherMap API: %s", err)
				return
			}

			uid := fmt.Sprintf("%s%d", query, time.Now().Nanosecond())
			imagestore.Store(fmt.Sprintf("satellite%s.png", uid), result)

			path := fmt.Sprintf("?zip=satellite%s.png", uid)

			// links to click
			if os.Getenv("DEBUG") == "1" {
				fmt.Fprintf(w, "http://127.0.0.1:8111/%s\n", path)
				return
			}

			path = fmt.Sprintf("weather%s", path)
			fmt.Fprintf(w, "https://shouting.online/%s\n", path)

			return
		case strings.HasPrefix(q, "precip"):
			query := strings.TrimSpace(strings.TrimPrefix(q, "precip"))
			var err error
			var result *image.NRGBA
			if result, err = openweathermap.GetComposite(query); err != nil {
				fmt.Fprintf(w, "Bad result from OpenWeatherMap API: %s", err)
				return
			}

			uid := fmt.Sprintf("%s%d", query, time.Now().Nanosecond())
			imagestore.Store(fmt.Sprintf("composite%s.png", uid), result)

			path := fmt.Sprintf("?zip=composite%s.png", uid)

			// links to click
			if os.Getenv("DEBUG") == "1" {
				fmt.Fprintf(w, "http://127.0.0.1:8111/%s\n", path)
				return
			}

			path = fmt.Sprintf("weather%s", path)
			fmt.Fprintf(w, "https://shouting.online/%s\n", path)

			return

		case strings.HasPrefix(q, "map"):
			query := strings.TrimSpace(strings.TrimPrefix(q, "map"))

			result := gfs.Do(query)
			imagestore.Store(fmt.Sprintf("%sus.gif", query), result)
			path := fmt.Sprintf("?zip=%sus.gif", query)

			// links to click
			if os.Getenv("DEBUG") == "1" {
				fmt.Fprintf(w, "http://127.0.0.1:8111/%s\n", path)
				return
			}

			path = fmt.Sprintf("weather%s", path)
			fmt.Fprintf(w, "https://shouting.online/%s\n", path)

			return

		default:
			help(w)
			return
		}
	})

	fmt.Printf("%s", http.ListenAndServe("0.0.0.0:8111", nil))
}

func help(w http.ResponseWriter) {
	fmt.Fprintf(w, "%s", "*Commands:*\n\n`!nws`: get a weather report from the NWS's public data source. Use a zip code, expect results to be from airports.\n`!weather`: get the current weather and today's forcast. Use `zip` or `city, state` e.g. `!weather 78703` or `!weather san francisco, ca`\n`!forecast`: short-term forecast by zip or city, state\n`!precip` get a precipitation map of the region centered on provided zip\n`!satellite` get a recent (as old as a week) satellite of a region centered on a zip")
}
