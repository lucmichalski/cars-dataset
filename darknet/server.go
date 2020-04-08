package main

import (
	"bytes"
	"flag"
	"image"
	"image/jpeg"
	"image/color"
	"image/draw"
	"log"
	"strconv"
	"io"
	"net/url"
	"path/filepath"
	"math"
	"os"
	"net/http"
	"fmt"
	"strings"
	"sort"

	"github.com/k0kubun/pp"	
	"github.com/gin-gonic/gin"
    "github.com/disintegration/imaging"
	darknet "github.com/LdDl/go-darknet"
)

/*
	Refs:
	- https://hackernoon.com/docker-compose-gpu-tensorflow-%EF%B8%8F-a0e2011d36
	- https://github.com/eywalker/nvidia-docker-compose
	- https://github.com/NVIDIA/nvidia-docker
*/

var (
	n darknet.YOLONetwork
	configFile = flag.String("configFile", "", "Path to network layer configuration file. Example: cfg/yolov3.cfg")
	weightsFile = flag.String("weightsFile", "", "Path to weights file. Example: yolov3.weights")
	imageFile = flag.String("imageFile", "", "Path to image file, for detection. Example: image.jpg")
)

type bboxInfo struct {
	minX int
	minY int
	maxX int
	maxY int
	width int
	height int
	surface int
}

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

		// Use same size as source image has
		b := src.Bounds()
		m := image.NewRGBA(b)

		// offset := image.Pt(0, 0)

		// Draw source
		draw.Draw(m, b, src, image.ZP, draw.Src)

		log.Println("Network-only time taken:", dr.NetworkOnlyTimeTaken)
		log.Println("Overall time taken:", dr.OverallTimeTaken, len(dr.Detections))
		for _, d := range dr.Detections {
			for i := range d.ClassIDs {
				bBox := d.BoundingBox
				fmt.Printf("%s (%d): %.4f%% | start point: (%d,%d) | end point: (%d, %d)\n",
					d.ClassNames[i], d.ClassIDs[i],
					d.Probabilities[i],
					bBox.StartPoint.X, bBox.StartPoint.Y,
					bBox.EndPoint.X, bBox.EndPoint.Y,
				)
				minX, minY := float64(bBox.StartPoint.X), float64(bBox.StartPoint.Y)
				maxX, maxY := float64(bBox.EndPoint.X), float64(bBox.EndPoint.Y)
				// rect := image.Rect(round(minX), round(minY), round(maxX), round(maxY))

				// img, err := drawableJPEGImage(f)
				// fmt.Println(rect)

				// Draw watermark
				// draw.Draw(m, watermarkImage.Bounds().Add(offset), watermarkImage, image.ZP, draw.Over)

				drawBbox(round(minX), round(minY), round(maxX), round(maxY), 10, m)
				draw.Draw(m, src.Bounds(), src, image.ZP, draw.Over)

			}
		}
		// Specify the quality, between 0-100
		// Higher is better
		opt := jpeg.Options{
		    Quality: 100,
		}
		err = jpeg.Encode(c.Writer, m, &opt)
		if err != nil {
		    // Handle error
			panic(err.Error())
		}

	})

	r.GET("/crop", func(c *gin.Context) {

		log.Println("crop start")

		url := c.Query("url")
		classesStr := c.Query("classes")
		classes := strings.Split(classesStr, ",")
		thresholdStr := c.Query("threshold")
		var threshold float64
		var err error
		if thresholdStr != "" {
			threshold, err = strconv.ParseFloat(thresholdStr, 64)
			if err != nil {
				panic(err.Error())
			}
		}

		log.Println("classes", classes)
		log.Println("threshold", threshold)

		var file *os.File
		var size int64
		if url != "" && strings.HasPrefix(url, "http") {
			file, size, err = openFileByURL(url)
		}

		log.Println("file", file.Name())
		log.Println("size", size)

		bboxInfos := make([]*bboxInfo, 0)

		if size > 0 {

			src, err := jpeg.Decode(file)
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
					log.Printf("%s (%d): %.4f%% | start point: (%d,%d) | end point: (%d, %d)\n",
						d.ClassNames[i], d.ClassIDs[i],
						d.Probabilities[i],
						bBox.StartPoint.X, bBox.StartPoint.Y,
						bBox.EndPoint.X, bBox.EndPoint.Y,
					)
					if d.ClassNames[i] == "car" && d.Probabilities[i] > 90 {
						bboxInfos = append(bboxInfos, &bboxInfo{
							minX: bBox.StartPoint.X,
							minY: bBox.StartPoint.Y,
							maxX: bBox.EndPoint.X,
							maxY: bBox.EndPoint.Y,
							width: bBox.EndPoint.X-bBox.StartPoint.X,
							height: bBox.EndPoint.Y-bBox.StartPoint.Y,
							surface: (bBox.EndPoint.X-bBox.StartPoint.X)*(bBox.EndPoint.Y-bBox.StartPoint.Y),
						})
					}
				}
			}

			sort.Slice(bboxInfos[:], func(i, j int) bool {
				return bboxInfos[i].surface > bboxInfos[j].surface
			})

			pp.Println("bboxInfos:", bboxInfos)

			if len(bboxInfos) == 0 {
				c.String(200, "Nothing")
			} else {
				bbox := image.Rect(bboxInfos[0].minX-20, bboxInfos[0].minY-20, bboxInfos[0].maxX+20, bboxInfos[0].maxY+20)
			    src = imaging.Crop(src, bbox)
				err = imaging.Encode(c.Writer, src, imaging.JPEG)
			    if err != nil {
			        log.Fatalf("failed to encode image: %v", err)
			    }
			}

		} else {
			c.String(200, "Nothing")
		}

		log.Println("crop end")
	})

	r.POST("/crop", func(c *gin.Context) {

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
				log.Printf("%s (%d): %.4f%% | start point: (%d,%d) | end point: (%d, %d)\n",
					d.ClassNames[i], d.ClassIDs[i],
					d.Probabilities[i],
					bBox.StartPoint.X, bBox.StartPoint.Y,
					bBox.EndPoint.X, bBox.EndPoint.Y,
				)
				if d.ClassNames[i] == "car" && d.Probabilities[i] > 0.90 {
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
				    //src, err := imaging.Decode(f)
				    //if err != nil {
				    //   log.Fatalf("failed to open image: %v", err)
				    //}
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

func drawableJPEGImage(r io.Reader) (draw.Image, error) {
    img, err := jpeg.Decode(r)
    if err != nil {
        return nil, err
    }
    dimg, ok := img.(draw.Image)
    if !ok {
        return nil, fmt.Errorf("%T is not a drawable image type", img)
    }
    return dimg, nil
}

func drawBbox(x1, y1, x2, y2, thickness int, img *image.RGBA) {
    col := color.RGBA{0, 0, 0, 255}
    for t:=0; t<thickness; t++ {
        // draw horizontal lines
        for x := x1; x<= x2; x++ {
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

func openFileByURL(rawURL string) (*os.File, int64, error) {
	if fileURL, err := url.Parse(rawURL); err != nil {
		return nil, 0, err
	} else {
		path := fileURL.Path
		segments := strings.Split(path, "/")
		// extension := filepath.Ext(path)
		fileName := segments[len(segments)-1]

		filePath := filepath.Join(os.TempDir(), fileName)

		if fi, err := os.Stat(filePath); err == nil {
			file, err := os.Open(filePath)
			return file, fi.Size(), err
		}

		file, err := os.Create(filePath)
		if err != nil {
			return file, 0, err
		}

		check := http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}
		resp, err := check.Get(rawURL) // add a filter to check redirect
		if err != nil {
			return file, 0, err
		}
		defer resp.Body.Close()
		fmt.Printf("----> Downloaded %v\n", rawURL)

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return file, 0, err
		}

		fi, err := file.Stat()
		if err != nil {
			return file, 0, err
		}

		return file, fi.Size(), nil
	}
}

