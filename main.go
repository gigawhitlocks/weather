package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gigawhitlocks/weather/nws"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("zip")

		if strings.HasPrefix(q, "weather") {
			var err error
			var result *nws.Result
			zip := strings.TrimSpace(strings.TrimPrefix(q, "weather"))
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

	})

	fmt.Printf("%s", http.ListenAndServe("0.0.0.0:8111", nil))
}
