package main

import (
	"bytes"
	"flag"
	"image"
	"image/jpeg"
	// "image/color"
	"log"
	"strconv"
	"path/filepath"
	"math"
	"os"
	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"
	// "github.com/llgcode/draw2d/draw2dimg"
    "github.com/disintegration/imaging"
	darknet "github.com/LdDl/go-darknet"
)

var (
	n darknet.YOLONetwork
	configFile = flag.String("configFile", "", "Path to network layer configuration file. Example: cfg/yolov3.cfg")
	weightsFile = flag.String("weightsFile", "", "Path to weights file. Example: yolov3.weights")
	imageFile = flag.String("imageFile", "", "Path to image file, for detection. Example: image.jpg")
)

func server() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})


	r.POST("/bbox", func(c *gin.Context) {

		// Source
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}

		f, err := file.Open()
		if err != nil {
			panic(err.Error())
		}

		src, err := jpeg.Decode(f)
		if err != nil {
			panic(err.Error())
		}

		imgDarknet, err := darknet.Image2Float32(src)
		if err != nil {
			panic(err.Error())
		}
		defer imgDarknet.Close()

		dr, err := n.Detect(imgDarknet)
		if err != nil {
			printError(err)
			return
		}

		log.Println("Network-only time taken:", dr.NetworkOnlyTimeTaken)
		log.Println("Overall time taken:", dr.OverallTimeTaken, len(dr.Detections))
		for _, d := range dr.Detections {
			for i := range d.ClassIDs {
				bBox := d.BoundingBox
				fmt.Printf("[%s] %s (%d): %.4f%% | start point: (%d,%d) | end point: (%d, %d)\n",
					imageFile,
					d.ClassNames[i], d.ClassIDs[i],
					d.Probabilities[i],
					bBox.StartPoint.X, bBox.StartPoint.Y,
					bBox.EndPoint.X, bBox.EndPoint.Y,
				)
				minX, minY := float64(bBox.StartPoint.X), float64(bBox.StartPoint.Y)
				maxX, maxY := float64(bBox.EndPoint.X), float64(bBox.EndPoint.Y)
				rect := image.Rect(round(minX), round(minY), round(maxX), round(maxY))

				fmt.Println(rect)
				/*
			    src, err := imaging.Decode(f)
			    if err != nil {
			        log.Fatalf("failed to open image: %v", err)
			    }
			    */

			}
		}

	})

	r.POST("/upload", func(c *gin.Context) {

		// classes := c.PostForm("classes")
		// threshold := c.PostForm("threshold")

		// Source
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}

		f, err := file.Open()
		if err != nil {
			panic(err.Error())
		}

		src, err := jpeg.Decode(f)
		if err != nil {
			panic(err.Error())
		}

		imgDarknet, err := darknet.Image2Float32(src)
		if err != nil {
			panic(err.Error())
		}
		defer imgDarknet.Close()

		dr, err := n.Detect(imgDarknet)
		if err != nil {
			printError(err)
			return
		}

		log.Println("Network-only time taken:", dr.NetworkOnlyTimeTaken)
		log.Println("Overall time taken:", dr.OverallTimeTaken, len(dr.Detections))
		for _, d := range dr.Detections {
			for i := range d.ClassIDs {
				bBox := d.BoundingBox
				log.Printf("[%s] %s (%d): %.4f%% | start point: (%d,%d) | end point: (%d, %d)\n",
					imageFile,
					d.ClassNames[i], d.ClassIDs[i],
					d.Probabilities[i],
					bBox.StartPoint.X, bBox.StartPoint.Y,
					bBox.EndPoint.X, bBox.EndPoint.Y,
				)
				if d.ClassNames[i] == "car" {
					// save bouding boxes
					// cropZone(imageFile, i, d.ClassNames[i], image.Rect(bBox.StartPoint.X-20, bBox.StartPoint.Y-20, bBox.EndPoint.X+20, bBox.EndPoint.Y+20))
					// check image size if not acceptable size
					// Uncomment code below if you want save cropped objects to files
					// minX, minY := float64(bBox.StartPoint.X), float64(bBox.StartPoint.Y)
					// maxX, maxY := float64(bBox.EndPoint.X), float64(bBox.EndPoint.Y)
					// rect := image.Rect(round(minX), round(minY), round(maxX), round(maxY))
					// err := saveToFile(src, rect, fmt.Sprintf("crop_%d.jpeg", i))
					// if err != nil {
					// 	fmt.Println(err)
					// 	return
					// }
				    // Open a test image.
					bbox := image.Rect(bBox.StartPoint.X-20, bBox.StartPoint.Y-20, bBox.EndPoint.X+20, bBox.EndPoint.Y+20)
				    src, err := imaging.Decode(f)
				    if err != nil {
				        log.Fatalf("failed to open image: %v", err)
				    }
				    src = imaging.Crop(src, bbox) // image.Rect(42, 51, 772, 485))
					err = imaging.Encode(c.Writer, src, imaging.JPEG)
				    if err != nil {
				        log.Fatalf("failed to encode image: %v", err)
				    }
				}
			}
		}
	})

	r.Run(":9003")
}

func printError(err error) {
	log.Println("error:", err)
}

func main() {
	flag.Parse()

	if *configFile == "" || *weightsFile == "" {
		flag.Usage()
		return
	}

	n = darknet.YOLONetwork{
		GPUDeviceIndex:           0,
		NetworkConfigurationFile: *configFile,
		WeightsFile:              *weightsFile,
		Threshold:                .25,
	}
	if err := n.Init(); err != nil {
		printError(err)
		return
	}
	defer n.Close()

	server()
}

func round(v float64) int {
	if v >= 0 {
		return int(math.Floor(v + 0.5))
	}
	return int(math.Ceil(v - 0.5))
}

func saveToFile(imgSrc image.Image, bbox image.Rectangle, fname string) error {
	rectcropimg := imaging.Crop(imgSrc, bbox)
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	err = jpeg.Encode(f, rectcropimg, nil)
	if err != nil {
		return err
	}
	return nil
}

func imageToBytes(img image.Image) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, nil)
	return buf.Bytes(), err
}

func cropZone(inputFile string, idx int, className string, bbox image.Rectangle) error {
    // Open a test image.
    src, err := imaging.Open(inputFile)
    if err != nil {
        log.Fatalf("failed to open image: %v", err)
        return err
    }

    src = imaging.Crop(src, bbox) // image.Rect(42, 51, 772, 485))

    // Save the resulting image as JPEG.
	basename := filepath.Base(inputFile)

    err = imaging.Save(src, "/darknet/cropped_"+className+"-"+strconv.Itoa(idx)+"-"+basename)
    if err != nil {
        log.Fatalf("failed to save image: %v", err)
        return err
    }
    return nil
}


