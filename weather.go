package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

var zipMap map[zipCode]latLong

type Weather struct {
}

type zipCode string
type latLong [2]string

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
			log.Fatal(err)
		}
		zipMap[zipCode(record[0])] = latLong([2]string{
			record[1],
			record[2]})
	}
	return zipMap
}

func getWeather() (output string, err error) {
	// var resp *http.Response
	// if resp, err = http.Get(""); err != nil {
	// 	return "", err
	// }
	// decoder := json.NewDecoder(resp.Body)
	// if decoder == nil {
	// 	return "", fmt.Errorf("fuck")
	// }
	// w := &Weather{}
	// err = decoder.Decode(w)
	// if w == nil {
	// 	return "", err
	// }
	// return w.String(), nil
	return "", nil
}

func main() {
	zipMap = readZips()
}
