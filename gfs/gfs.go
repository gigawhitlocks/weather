package gfs

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"net/http"
	"time"
)

const numFrames int = 10

type Frame struct {
	Image *image.Paletted
	Num   int
}

const northCent = "nc"
const northEast = "ne"
const northWest = "nw"
const southCent = "sc"
const southEast = "se"
const southWest = "sw"

var validQueries = map[string]string{
	northCent:       northCent,
	northEast:       northEast,
	northWest:       northWest,
	southCent:       southCent,
	southEast:       southEast,
	southWest:       southWest,
	"southwest":     southWest,
	"southeast":     southEast,
	"northwest":     northWest,
	"texas":         southCent,
	"northeast":     northEast,
	"south central": southCent,
}

func region(in string) string {
	if i, ok := validQueries[in]; ok {
		return i
	}
	return ""
}

func Do(input string) *gif.GIF {
	date := fmt.Sprintf("%s", time.Now().Format("20060102"))
	region := region(input)
	frames := make(chan *Frame)
	for i := 1; i < numFrames+1; i++ {
		go func(n int) {
			url := fmt.Sprintf(
				"https://www.tropicaltidbits.com/analysis/models/gfs/%s00/gfs_mslp_pcpn_frzn_%sus_%d.png",
				date, region, n)
			fmt.Printf("%s\n", url)
			r, err := http.Get(url)
			if err != nil {
				return
			}

			var frame image.Image
			frame, err = png.Decode(r.Body)
			if err != nil {
				return
			}

			im := image.NewPaletted(frame.Bounds(), palette.Plan9)
			draw.Draw(im, im.Bounds(), frame, frame.Bounds().Min, draw.Over)
			frames <- &Frame{Image: im, Num: n}
		}(i)
	}

	dumb := map[int]*image.Paletted{}
	for i := 0; i < numFrames; i++ {
		f := <-frames
		dumb[f.Num] = f.Image
	}

	o := &gif.GIF{}
	for i := 1; i < numFrames+1; i++ {
		o.Image = append(o.Image, dumb[i])
		o.Delay = append(o.Delay, 75)
	}

	return o
	// if err = gif.EncodeAll(f, o); err != nil {
	// 	panic(err)
	// }
}
