package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	pnm "github.com/jbuchbinder/gopnm"
)

// PGM tool
// for editing / changing pgm file into routing pgm
//   for astar-routing in synerex.

var (
	inputPGMFile = flag.String("in-pgm", "projection_edit.pgm", "Input PGM File")
	jsonFname    = flag.String("json", "out.json", "JSON output file name")
	pgmFile      = flag.String("pgm", "out.pgm", "PGM output file name")
	pngFile      = flag.String("png", "", "PNG output file name")

//	width     = flag.Int("width", 1280, "Output PGM file width")
//	store           = flag.Bool("store", false, "store csv data")
)

type Feature struct {
	MinLon, MinLat, MaxLon, MaxLat float64
	DLon, DLat                     float64
	Count                          int
	Scale                          float64
	PGMFile                        string
	PGMWidth, PGMHeight            int
	//	GeoJsonFC                      *geojson.FeatureCollection `json:"-"`
}

func setMinMax(f *Feature, lon, lat float64) {
	//	fmt.Printf("Still max:current %v %f %f\n", *f, lon, lat)
	if lon < f.MinLon {
		f.MinLon = lon
	}
	if lon > f.MaxLon {
		f.MaxLon = lon
	}
	if lat < f.MinLat {
		f.MinLat = lat
	}
	if lat > f.MaxLat {
		f.MaxLat = lat
	}
	f.Count += 1
	//	fmt.Printf("Done %v %f %f\n", *f, lon, lat)
}

func scanImage(img image.Image) *Feature {
	f := &Feature{
		MinLat: math.MaxFloat64,
		MinLon: math.MaxFloat64,
		MaxLat: -math.MaxFloat64,
		MaxLon: -math.MaxFloat64,
	}
	bound := img.Bounds()
	W := bound.Dx()
	H := bound.Dy()
	log.Printf("file loaded with %dx%d", W, H)

	for j := H - 1; j >= 0; j-- {
		for i := 0; i < W; i++ {
			oldPix := img.At(i, j)
			pixel := color.GrayModel.Convert(oldPix)
			pixelU := color.GrayModel.Convert(pixel).(color.Gray).Y
			if float64(pixelU) < 100 { // black area
				setMinMax(f, float64(i), float64(j))
			}
		}
	}
	return f
}

func outputPGM(f *Feature, fname *string, img image.Image) {
	fmt.Printf("Generating Img\n")

	x0 := int(f.MinLon)
	y0 := int(f.MinLat)
	f.PGMWidth = int(f.MaxLon - f.MinLon + 1)
	f.PGMHeight = int(f.MaxLat - f.MinLat + 1)

	newImg := image.NewGray(image.Rect(0, 0, f.PGMWidth, f.PGMHeight))

	for x := 0; x < f.PGMWidth; x++ {
		for y := 0; y < f.PGMHeight; y++ {
			oldPix := img.At(x0+x, y0+y)
			pixel := color.GrayModel.Convert(oldPix)
			pixelU := color.GrayModel.Convert(pixel).(color.Gray).Y
			if float64(pixelU) > 100 { // black area
				newImg.SetGray(x, y, color.Gray{255}) // into white area
			}
		}
	}

	fp, err := os.Create(*fname)
	if err != nil {
		log.Fatal("Can't open file %s", *fname)
	}
	defer fp.Close()
	fmt.Printf("Writing PGM image file: %s\n", *fname)
	pnm.Encode(fp, newImg, pnm.PGM)

	if *pngFile != "" {
		fp2, err2 := os.Create(*pngFile)
		if err2 != nil {
			log.Fatal("Can't open file %s", *pngFile)
		}
		defer fp2.Close()
		fmt.Printf("Writing PNG image file: %s\n", *pngFile)
		png.Encode(fp2, newImg)
	}

}

func loadImage(imgFile string) image.Image {
	file, err := os.Open(imgFile)
	if err != nil {
		return nil
	}
	defer file.Close()

	imData, _, err := image.Decode(file)
	if err != nil {
		return nil
	}
	return imData
}

func main() {
	flag.Parse()
	var feature = &Feature{}
	var img image.Image
	if *inputPGMFile != "" { // load PGM file

		img = loadImage(*inputPGMFile)
		if img == nil {
			log.Fatal("Can't load image: ", *inputPGMFile)
		}
		fmt.Printf("Loaded image: %#v %#v\n", img.Bounds(), img.ColorModel())
		// check img colors

		feature = scanImage(img)

		fmt.Printf("Feature %#v\n", feature)

		outputPGM(feature, pgmFile, img)
		feature.PGMFile = *pgmFile
		feature.Scale = 1.0
	}
	if *jsonFname != "" {
		file, err := os.Create(*jsonFname)
		if err != nil {
			log.Fatal("Can't open ", *jsonFname)
		}
		defer file.Close()

		//		b, _ := json.Marshal(&feature)
		b, _ := json.MarshalIndent(&feature, "", "   ")
		file.Write(b)
	}
}
