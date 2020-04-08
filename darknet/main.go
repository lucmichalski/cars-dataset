package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"strconv"
	"path/filepath"

    "github.com/disintegration/imaging"
	darknet "github.com/LdDl/go-darknet"
)

var configFile = flag.String("configFile", "", "Path to network layer configuration file. Example: cfg/yolov3.cfg")
var weightsFile = flag.String("weightsFile", "", "Path to weights file. Example: yolov3.weights")
var imageFile = flag.String("imageFile", "", "Path to image file, for detection. Example: image.jpg")

func printError(err error) {
	log.Println("error:", err)
}

func main() {
	flag.Parse()

	if *configFile == "" || *weightsFile == "" {
		flag.Usage()
		return
	}

	n := darknet.YOLONetwork{
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

	imageFiles := []string{"FL-823-GFB.jpg", "FL-823-GFA.jpg", "FL-823-GFC.jpg", "FL-823-GFD.jpg", "FL-823-GFE.jpg", "FL-823-GFF.jpg", "FL-823-GFG.jpg", "FL-823-GFI.jpg"}

	for _, imageFile := range imageFiles {

		infile, err := os.Open(imageFile)
		if err != nil {
			panic(err.Error())
		}
		defer infile.Close()
		src, err := jpeg.Decode(infile)
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
				if d.ClassNames[i] == "car" && d.Probabilities[i] > 90 {
					// save bouding boxes
					cropZone(imageFile, i, d.ClassNames[i], image.Rect(bBox.StartPoint.X-20, bBox.StartPoint.Y-20, bBox.EndPoint.X+20, bBox.EndPoint.Y+20))
					// check image size if not acceptable size
				}
			}
		}
	}
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

