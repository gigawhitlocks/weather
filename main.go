package main

import (
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

		if strings.HasPrefix(q, "nws") {
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
		}

		if strings.HasPrefix(q, "weather") {
			query := strings.TrimSpace(strings.TrimPrefix(q, "weather"))
			var result *wunderground.Weather
			if result, err = wunderground.GetWeather(query); err != nil {
				fmt.Fprintf(w, "%s", err)
			}

			fmt.Fprintf(w, "%s", result.String())

		}
	})

	fmt.Printf("%s", http.ListenAndServe("0.0.0.0:8111", nil))
}
