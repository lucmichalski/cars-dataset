package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/h2non/filetype"
)

// -imageFile=/mnt/nasha/lucmichalski/cars-dataset/datasets/cars/MERCEDES-BENZ/E 200/2017/e5dc95fcf140e0a079c14998e512362c.jpg,559,424,57,259
// -imageFile=/mnt/nasha/lucmichalski/cars-dataset/datasets/cars/FORD/FOCUS/2018/73e7afb10b5fcf6e7e0721bd645b0742.jpg -bboxData=640,476,0,0

var imageFile = flag.String("imageFile", "", "Path to image file, for detection. Example: image.jpg")
var bboxData = flag.String("bboxData", "", "bounding box data.")

func main() {
	flag.Parse()

	infile, err := os.Open(*imageFile)
	if err != nil {
		panic(err.Error())
	}
	defer infile.Close()

	buf, _ := ioutil.ReadFile(infile.Name())
	kind, _ := filetype.Match(buf)

	var src image.Image
	log.Println("kind.MIME.Value:", kind.MIME.Value)
	switch kind.MIME.Value {
	case "image/jpeg":
		src, err = jpeg.Decode(infile)
		if err != nil {
			panic(err.Error())
		}
	case "image/png":
		src, err = png.Decode(infile)
		if err != nil {
			panic(err.Error())
		}
	default:
		log.Fatal("unkown format image.")
	}

	// Use same size as source image has
	b := src.Bounds()
	m := image.NewRGBA(b)

	// Draw source
	draw.Draw(m, b, src, image.ZP, draw.Src)
	draw.Draw(m, src.Bounds(), src, image.ZP, draw.Src)

	bboxParts := strings.Split(*bboxData, ",")
	if len(bboxParts) != 4 {
		log.Fatal("bbox data are abnormaly formed.")
	}

	maxX, _ := strconv.Atoi(bboxParts[0]) // maxX, maxY, minX, minY
	maxY, _ := strconv.Atoi(bboxParts[1])
	minX, _ := strconv.Atoi(bboxParts[2])
	minY, _ := strconv.Atoi(bboxParts[3])

	drawBbox(minX, minY, maxX, maxY, 1, m)

	basename := filepath.Base(*imageFile)
	dest, err := os.Create("./cropped_" + basename)
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}

	opt := jpeg.Options{
		Quality: 100,
	}
	err = jpeg.Encode(dest, m, &opt)
	if err != nil {
		// Handle error
		panic(err.Error())
	}

	//bboxParts := strings.Split(*bboxData, ",")
	// cropZone(*imageFile, image.Rect(559, 424, 57, 259))
}

func round(v float64) int {
	if v >= 0 {
		return int(math.Floor(v + 0.5))
	}
	return int(math.Ceil(v - 0.5))
}

func drawBbox(x1, y1, x2, y2, thickness int, img *image.RGBA) {
	col := color.RGBA{0, 255, 0, 128}
	for t := 0; t < thickness; t++ {
		// draw horizontal lines
		for x := x1; x <= x2; x++ {
			img.Set(x, y1+t, col)
			img.Set(x, y2-t, col)
		}
		// draw vertical lines
		for y := y1; y <= y2; y++ {
			img.Set(x1+t, y, col)
			img.Set(x2-t, y, col)
		}
	}
}

func printError(err error) {
	log.Println("error:", err)
}

func cropZone(inputFile string, bbox image.Rectangle) error {
	// Open a test image.
	src, err := imaging.Open(inputFile)
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
		return err
	}

	src = imaging.Crop(src, bbox)
	// Save the resulting image as JPEG.
	basename := filepath.Base(inputFile)

	err = imaging.Save(src, "./cropped_"+basename)
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
		return err
	}
	return nil
}
