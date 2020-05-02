package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	darknet "github.com/LdDl/go-darknet"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
	"github.com/k0kubun/pp"
	"github.com/oschwald/geoip2-golang"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/lucmichalski/cars-dataset/pkg/grab"
	// "github.com/lucmichalski/cars-dataset/pkg/middlewares"
	"github.com/lucmichalski/cars-dataset/pkg/models"
)

/*
	TO DO:
	- Make more elegant this dirty hacky code
	- Comment the code as much as possible
*/

var (
	isHelp      bool
	isVerbose   bool
	isDryMode   bool
	gpuId	    int
	weightsFile string
	configFile  string
	geoIpFile   string
	labelmeVer  string
	geoReader   *geoip2.Reader
	m           yoloModel
)

// yoloModel represents the yolllow network model loaded and it read/write mutex
type yoloModel struct {
	n darknet.YOLONetwork
	l sync.Mutex
}

// bboxInfo represents detection info for an enityt detected by yolo
type bboxInfo struct {
	minX    int
	minY    int
	maxX    int
	maxY    int
	width   int
	height  int
	surface int
	class   string
}

func main() {

	// Define cli flag parameters
	pflag.StringVarP(&weightsFile, "weights-file", "w", "./models/yolov4.cfg", "Path to weights file. Example: yolov4.weights")
	pflag.StringVarP(&configFile, "config-file", "c", "./models/yolov4.cfg", "Path to network layer configuration file. Example: cfg/yolov4.cfg")
	pflag.StringVarP(&geoIpFile, "geoip-db", "", "./geoip2/GeoLite2-City.mmdb", "geoip filepath.")
	pflag.StringVarP(&labelmeVer, "labelme-ver", "", "3.6.10", "labelme version JSON format.")
        pflag.IntVarP(&gpuId, "gpu", "", 0, "gpu id (eg 0,1).")
	pflag.BoolVarP(&isVerbose, "verbose", "v", false, "verbose mode.")
	pflag.BoolVarP(&isHelp, "help", "h", false, "help info.")
	pflag.Parse()
	if isHelp {
		pflag.PrintDefaults()
		return
	}

	if configFile == "" || weightsFile == "" {
		pflag.PrintDefaults()
		return
	}

	// Define the yolo network config
	n := darknet.YOLONetwork{
		GPUDeviceIndex:           gpuId,
		NetworkConfigurationFile: configFile,
		WeightsFile:              weightsFile,
		Threshold:                .25,
	}

	// Instanciate the yolo network
	if err := n.Init(); err != nil {
		printError(err)
		return
	}
	defer n.Close()

	// create the global yolo model
	m = yoloModel{
		n: n,
		l: sync.Mutex{},
	}

	// Instanciate geoip2 database
	// geoReader = must(geoip2.Open(geoIpFile)).(*geoip2.Reader)

	// launch the web service
	server()

}

func server() {

	r := gin.Default()

	// globally use middlewares
	//r.Use(
		// middlewares.RealIP(),
		// middlewares.RecoveryWithWriter(os.Stderr),
		// middlewares.Logger(geoReader),
		// middlewares.CORS(),
	//	gin.ErrorLogger(),
	//)

	// The route [/] is a dunny hone page
	//
	// Method: GET
	r.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})

	// The route [/labelme] allows to save the detection in th label me JSON format
	//
	// Method: GET
	//
	// Behaviour:
	// 1. Get an input image issued from a focused crawler (cars, motorcyle)
	// 2. Resize the input image if wider or higher than 700px
	// 3. Process yolov4 detection
	// 4. Detect the largest boudning box (assumption made that the biggest object detection is our focused target)
	// 5. Return a labelme JSON response with the annotation info and the image base64 encoded
	//
	// Parameters:
	// url       = image url to process
	// classes   = filter detections by classes
	// threshold = minium threshold of confidence in the dection
	r.GET("/labelme", func(c *gin.Context) {
		m.l.Lock()
		defer m.l.Unlock()

		var err error
		u := c.Query("url")
		u, err = url.QueryUnescape(u)

		if isVerbose {
			fmt.Println("url:", u)
		}

		classesStr := c.Query("classes")
		classes := strings.Split(classesStr, ",")
		thresholdStr := c.Query("threshold")
		var threshold float64
		if thresholdStr != "" {
			threshold, err = strconv.ParseFloat(thresholdStr, 64)
			if err != nil {
				panic(err.Error())
			}
		}

		if isVerbose {
			fmt.Println("classes", classes)
			fmt.Println("threshold", threshold)
		}

		var size int64
		var file *os.File
		if u != "" && strings.HasPrefix(u, "http") {
			file, size, err = grabFileByURL(u)
			if err != nil {
				panic(err.Error())
			}
		}

		if isVerbose {
			fmt.Println("file", file.Name())
			fmt.Println("size", size)
		}

		bboxInfos := make([]*bboxInfo, 0)
		if size > 0 {
			buf, _ := ioutil.ReadFile(file.Name())
			kind, _ := filetype.Match(buf)

			var src image.Image
			if isVerbose {
				log.Println("kind.MIME.Value:", kind.MIME.Value)
			}

			switch kind.MIME.Value {
			case "image/jpeg":
				src, err = jpeg.Decode(file)
				if err != nil {
					panic(err.Error())
				}
			case "image/png":
				src, err = png.Decode(file)
				if err != nil {
					panic(err.Error())
				}
			default:
				c.String(200, "")
				return
			}

			b := src.Bounds()

			if isVerbose {
				pp.Println("Original Height:", b.Max.Y)
				pp.Println("Original Width:", b.Max.X)
			}

			if b.Max.X > 700 {
				src = imaging.Resize(src, 700, 0, imaging.Lanczos)
			}

			// Ge the new bounds ? shall we reload the object ?
			b = src.Bounds()
			if isVerbose {
				pp.Println("Height post resing:", b.Max.Y)
				pp.Println("Width post resing:", b.Max.X)
			}

			imgDarknet, err := darknet.Image2Float32(src)
			if err != nil {
				panic(err.Error())
			}
			defer imgDarknet.Close()

			dr, err := m.n.Detect(imgDarknet)
			if err != nil {
				printError(err)
				return
			}

			lc := models.Labelme{}
			lc.FillColor = []int{255, 0, 0, 128}

			lc.ImageHeight = b.Max.Y
			lc.ImageWidth = b.Max.X

			// Encode as base64.
			lc.ImageData = base64.StdEncoding.EncodeToString(buf)

			lc.LineColor = []int{0, 255, 0, 128}
			lc.Version = labelmeVer

			if isVerbose {
				log.Println("Network-only time taken:", dr.NetworkOnlyTimeTaken)
				log.Println("Overall time taken:", dr.OverallTimeTaken, len(dr.Detections))
			}
			for _, d := range dr.Detections {
				for i := range d.ClassIDs {
					bBox := d.BoundingBox
					if isVerbose {
						fmt.Printf("%s (%d): %.4f%% | start point: (%d,%d) | end point: (%d, %d)\n",
							d.ClassNames[i], d.ClassIDs[i],
							d.Probabilities[i],
							bBox.StartPoint.X, bBox.StartPoint.Y,
							bBox.EndPoint.X, bBox.EndPoint.Y,
						)
					}
					if (d.ClassNames[i] == "car" || d.ClassNames[i] == "motorbike" || d.ClassNames[i] == "truck") && d.Probabilities[i] >= 70 {
						bboxInfos = append(bboxInfos, &bboxInfo{
							class:   d.ClassNames[i],
							minX:    bBox.StartPoint.X,
							minY:    bBox.StartPoint.Y,
							maxX:    bBox.EndPoint.X,
							maxY:    bBox.EndPoint.Y,
							width:   bBox.EndPoint.X - bBox.StartPoint.X,
							height:  bBox.EndPoint.Y - bBox.StartPoint.Y,
							surface: (bBox.EndPoint.X - bBox.StartPoint.X) * (bBox.EndPoint.Y - bBox.StartPoint.Y),
						})
					}
				}
			}

			sort.Slice(bboxInfos[:], func(i, j int) bool {
				return bboxInfos[i].surface > bboxInfos[j].surface
			})

			if isVerbose {
				pp.Println("bboxInfos:", bboxInfos)
			}

			if len(bboxInfos) > 0 {

				minX, minY := bboxInfos[0].minX, bboxInfos[0].minY
				maxX, maxY := bboxInfos[0].maxX, bboxInfos[0].maxY

				if minX < 0 {
					minX = 0
				}

				if minY < 0 {
					minY = 0
				}

				if maxX > b.Max.X {
					maxX = b.Max.X
				}

				if maxY > b.Max.Y {
					maxY = b.Max.Y
				}

				sc := models.Shape{}
				sc.Label = bboxInfos[0].class
				sc.ShapeType = "rectangle"
				if isVerbose {
					fmt.Println("minX=", minX, "maxX=", maxX, "minY=", minY, "maxY=", maxY)
				}
				x := []int{int(maxX), int(maxY)}
				y := []int{int(minX), int(minY)}

				points := [][]int{x, y}
				sc.Points = append(sc.Points, points...)
				lc.Shapes = append(lc.Shapes, sc)
			} else {
                                lc.Shapes = nil
			}
			c.JSON(200, &lc)
		} else {
			c.String(200, "Nothing")
		}
		log.Println("crop end")

	})

	r.GET("/bbox", func(c *gin.Context) {
		m.l.Lock()
		defer m.l.Unlock()

		var err error
		u := c.Query("url")
		u, err = url.QueryUnescape(u)

		fmt.Println("url:", u)

		classesStr := c.Query("classes")
		classes := strings.Split(classesStr, ",")
		thresholdStr := c.Query("threshold")
		var threshold float64
		//var err error
		if thresholdStr != "" {
			threshold, err = strconv.ParseFloat(thresholdStr, 64)
			if err != nil {
				panic(err.Error())
			}
		}

		fmt.Println("classes", classes)
		fmt.Println("threshold", threshold)

		var file *os.File
		var size int64
		if u != "" && strings.HasPrefix(u, "http") {
			file, size, err = grabFileByURL(u)
			if err != nil {
				panic(err.Error())
			}
		}

		fmt.Println("file", file.Name())
		fmt.Println("size", size)
		bboxInfos := make([]*bboxInfo, 0)

		if size > 0 {
			buf, _ := ioutil.ReadFile(file.Name())
			kind, _ := filetype.Match(buf)

			var src image.Image
			log.Println("kind.MIME.Value:", kind.MIME.Value)

			switch kind.MIME.Value {
			case "image/jpeg":
				src, err = jpeg.Decode(file)
				if err != nil {
					panic(err.Error())
				}
			case "image/png":
				src, err = png.Decode(file)
				if err != nil {
					panic(err.Error())
				}
			default:
				c.String(200, "")
				return
			}

			imgDarknet, err := darknet.Image2Float32(src)
			if err != nil {
				panic(err.Error())
			}
			defer imgDarknet.Close()

			dr, err := m.n.Detect(imgDarknet)
			if err != nil {
				printError(err)
				return
			}

			// Use same size as source image has
			b := src.Bounds()
			m := image.NewRGBA(b)

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
					if (d.ClassNames[i] == "car" || d.ClassNames[i] == "motorbike" || d.ClassNames[i] == "truck") && d.Probabilities[i] >= 70 {
						bboxInfos = append(bboxInfos, &bboxInfo{
							minX:    bBox.StartPoint.X,
							minY:    bBox.StartPoint.Y,
							maxX:    bBox.EndPoint.X,
							maxY:    bBox.EndPoint.Y,
							width:   bBox.EndPoint.X - bBox.StartPoint.X,
							height:  bBox.EndPoint.Y - bBox.StartPoint.Y,
							surface: (bBox.EndPoint.X - bBox.StartPoint.X) * (bBox.EndPoint.Y - bBox.StartPoint.Y),
						})
					}
					minX, minY := float64(bBox.StartPoint.X), float64(bBox.StartPoint.Y)
					maxX, maxY := float64(bBox.EndPoint.X), float64(bBox.EndPoint.Y)

					if minX < 0 {
						minX = 0
					}

					if minY < 0 {
						minY = 0
					}

					if maxX > float64(b.Max.X) {
						maxX = float64(b.Max.X)
					}

					if maxY > float64(b.Max.Y) {
						maxY = float64(b.Max.Y)
					}

					draw.Draw(m, src.Bounds(), src, image.ZP, draw.Src)
					drawBbox(round(minX), round(minY), round(maxX), round(maxY), 1, m)
				}
			}

			sort.Slice(bboxInfos[:], func(i, j int) bool {
				return bboxInfos[i].surface > bboxInfos[j].surface
			})

			// Specify the quality, between 0-100, Higher is better
			opt := jpeg.Options{
				Quality: 100,
			}
			err = jpeg.Encode(c.Writer, m, &opt)
			if err != nil {
				// Handle error
				panic(err.Error())
			}

		} else {
			c.String(200, "Nothing")
		}

		log.Println("crop end")

	})

	r.GET("/crop", func(c *gin.Context) {
		m.l.Lock()
		defer m.l.Unlock()

		log.Println("crop start")

		var err error
		u := c.Query("url")
		u, err = url.QueryUnescape(u)

		log.Println("url:", u)

		classesStr := c.Query("classes")
		classes := strings.Split(classesStr, ",")
		thresholdStr := c.Query("threshold")
		var threshold float64

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
		if u != "" && strings.HasPrefix(u, "http") {
			file, size, err = grabFileByURL(u)
		}

		log.Println("file", file.Name())
		log.Println("size", size)

		bboxInfos := make([]*bboxInfo, 0)

		if size > 0 {

			buf, _ := ioutil.ReadFile(file.Name())
			kind, _ := filetype.Match(buf)

			var src image.Image
			log.Println("kind.MIME.Value:", kind.MIME.Value)

			switch kind.MIME.Value {
			case "image/jpeg":
				src, err = jpeg.Decode(file)
				if err != nil {
					panic(err.Error())
				}
			case "image/png":
				src, err = png.Decode(file)
				if err != nil {
					panic(err.Error())
				}
			default:
				c.String(200, "")
				return
			}

			imgDarknet, err := darknet.Image2Float32(src)
			if err != nil {
				panic(err.Error())
			}
			defer imgDarknet.Close()

			dr, err := m.n.Detect(imgDarknet)
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
					if (d.ClassNames[i] == "car" || d.ClassNames[i] == "motorbike" || d.ClassNames[i] == "truck") && d.Probabilities[i] >= 70 {
						bboxInfos = append(bboxInfos, &bboxInfo{
							minX:    bBox.StartPoint.X,
							minY:    bBox.StartPoint.Y,
							maxX:    bBox.EndPoint.X,
							maxY:    bBox.EndPoint.Y,
							width:   bBox.EndPoint.X - bBox.StartPoint.X,
							height:  bBox.EndPoint.Y - bBox.StartPoint.Y,
							surface: (bBox.EndPoint.X - bBox.StartPoint.X) * (bBox.EndPoint.Y - bBox.StartPoint.Y),
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
				b := src.Bounds()
				imgWidth := b.Max.X
				imgHeight := b.Max.Y
				log.Println("src.Width:", imgWidth, "src.Height:", imgHeight)
				if imgWidth > 700 {
					src = imaging.Resize(src, 700, 0, imaging.Lanczos)
				}
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
		m.l.Lock()
		defer m.l.Unlock()

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

		dr, err := m.n.Detect(imgDarknet)
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
				if (d.ClassNames[i] == "car" || d.ClassNames[i] == "truck") && d.Probabilities[i] >= 70 {
					bbox := image.Rect(bBox.StartPoint.X-20, bBox.StartPoint.Y-20, bBox.EndPoint.X+20, bBox.EndPoint.Y+20)
					src = imaging.Crop(src, bbox)
					err = imaging.Encode(c.Writer, src, imaging.JPEG)
					if err != nil {
						log.Fatalf("failed to encode image: %v", err)
					}
				}
			}
		}
	})

	port := "9003"
	if os.Getenv("DARKNET_PORT") != "" {
		port = os.Getenv("DARKNET_PORT")
	}

	r.Run(fmt.Sprintf(":%s", port))
}

func printError(err error) {
	log.Println("error:", err)
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

func grabFileByURL(rawURL string) (*os.File, int64, error) {
	clientGrab := grab.NewClient()

	req, _ := grab.NewRequest(os.TempDir(), rawURL)
	if req == nil {
		return nil, 0, errors.New("----> could not make request.\n")
	}

	// start download
	log.Printf("----> Downloading %v...\n", req.URL())
	resp := clientGrab.Do(req)
	// pp.Println(resp)
	// fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			log.Printf("---->  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		log.Printf("----> Download failed: %v\n", err)
		return nil, 0, errors.Wrap(err, "Download failed")
	}

	// fmt.Printf("----> Downloaded %v\n", rawURL)
	log.Printf("----> Download saved to %v \n", resp.Filename)
	fi, err := os.Stat(resp.Filename)
	if err != nil {
		return nil, 0, errors.Wrap(err, "os stat failed")
	}
	file, _ := os.Open(resp.Filename)

	return file, fi.Size(), nil
}

// fail fast on initialization
func must(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}

	return i
}
