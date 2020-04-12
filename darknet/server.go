package main

import (
	"bytes"
	"flag"
	"image"
	"image/png"
	"image/jpeg"
	"image/color"
	"image/draw"
	"log"
	"strconv"
	"io"
	"path/filepath"
	"math"
	"os"
	"net/http"
	"fmt"
	"strings"
	"sort"
	"time"
	"io/ioutil"
	"sync"

	"github.com/h2non/filetype"
	"github.com/cavaliercoder/grab"
	"github.com/pkg/errors"
	"github.com/k0kubun/pp"	
	"github.com/gin-gonic/gin"
    "github.com/disintegration/imaging"
	darknet "github.com/LdDl/go-darknet"
)

/*

	Snippets:
	- gdrivedl
		- !sudo wget -O /usr/sbin/gdrivedl 'https://f.mjh.nz/gdrivedl'
		- !sudo chmod +x /usr/sbin/gdrivedl
		- !gdrivedl https://drive.google.com/open?id=1GL0zdThuAECX6zo1rA_ExKha1CPu1h_h camembert_sentiment.tar.xz
		- !tar xf camembert_sentiment.tar.xz
	- find-object
		- apk add --no-cache qt5-qtbase-dev cmake
		- cmake -DCMAKE_BUILD_TYPE=Release ..
	- nvidia-docker
		- distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
		- curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
		- curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | sudo tee /etc/apt/sources.list.d/nvidia-docker.list
		- sudo apt-get update && sudo apt-get install -y nvidia-container-toolkit
		- sudo systemctl restart docker
		- docker run --gpus all nvidia/cuda:10.0-base nvidia-smi
	- docker-compose gpu
        - sudo apt-get install nvidia-container-runtime
        - ~$ sudo vim /etc/docker/daemon.json
        - then , in this daemon.json file, add this content:
        - {
        - "default-runtime": "nvidia"
        - "runtimes": {
        - "nvidia": {
        - "path": "/usr/bin/nvidia-container-runtime",
        - "runtimeArgs": []
        - }
        - }
        - }
        - ~$ sudo systemctl daemon-reload
        - ~$ sudo systemctl restart docker
    - remove files
    	- find ./ -type f -size 0 -exec rm -f {} \;

	Todo:
	- https://www.lacentrale.fr/robots.txt
	  - https://www.lacentrale.fr/sitemap.php?file=sitemap-index-annonce.xml.gz
	  - https://www.lacentrale.fr/sitemap.php?file=sitemap-index-cotft.xml.gz

	Examples:
	- https://cdn-photos.autosphere.fr/media/FH/FH-662-SQD.jpg (utilitaire)
	- https://cdn-photos.autosphere.fr/media/FH/FH-662-SQC.jpg (utilitaire)
	- https://cdn-photos.autosphere.fr/media/FL/FL-823-GFF.jpg
	- https://cdn-photos.autosphere.fr/media/CY/CY-745-VTC.jpg
	- https://i.pinimg.com/originals/28/1b/ed/281bed127dae148b0e0536ea611e5e67.jpg
	- https://www.lambocars.com/images/lambonews/production_numbers.jpg
	- https://i.pinimg.com/originals/7e/fc/ab/7efcabaff4c082e99955b7b555b8b3da.png

	Refs:
	- https://hackernoon.com/docker-compose-gpu-tensorflow-%EF%B8%8F-a0e2011d36
	- https://github.com/eywalker/nvidia-docker-compose
	- https://github.com/NVIDIA/nvidia-docker
	- https://github.com/dbolya/yolact
	- https://github.com/Jonarod/tensorflow_lite_alpine
	- https://github.com/tinrab/go-tensorflow-image-recognition
	- https://github.com/dereklstinson/coco
	- https://github.com/chtorr/go-tensorflow-realtime-object-detection/blob/master/src/main.go
	- https://github.com/codegangsta/gin
	- https://github.com/shunk031/libtorch-gin-api-server/blob/master/docker/Dockerfile.api
	- https://github.com/tinrab/go-tensorflow-image-recognition/blob/master/main.go
	- https://github.com/x0rzkov/gocv-alpine (runtime,builder)
	- https://stackoverflow.com/questions/15341538/numpy-opencv-2-how-do-i-crop-non-rectangular-region
	- https://www.pyimagesearch.com/2018/11/19/mask-r-cnn-with-opencv/
	- https://note.nkmk.me/en/python-opencv-numpy-alpha-blend-mask/
*/

var (
	// n darknet.YOLONetwork
	m yoloModel
	configFile = flag.String("configFile", "", "Path to network layer configuration file. Example: cfg/yolov3.cfg")
	weightsFile = flag.String("weightsFile", "", "Path to weights file. Example: yolov3.weights")
	imageFile = flag.String("imageFile", "", "Path to image file, for detection. Example: image.jpg")
)

type yoloModel struct {
	n darknet.YOLONetwork
	l sync.RWMutex
}

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
		m.l.RLock()
		defer m.l.RUnlock()

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

		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, f); err != nil {
			panic(err.Error())
		}

		kind, _ := filetype.Match(buf.Bytes())
		var src image.Image

		log.Println("kind.MIME.Value:", kind.MIME.Value)

		switch kind.MIME.Value {
		case "image/jpeg":
			src, err = jpeg.Decode(f)
			if err != nil {
				panic(err.Error())
			}
		case "image/png":
			src, err = png.Decode(f)
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
				// draw.Draw(image3, src.Bounds(), src, image.ZP, draw.Src)
				drawBbox(round(minX), round(minY), round(maxX), round(maxY), 10, m)
				draw.Draw(m, src.Bounds(), src, image.ZP, draw.Src)

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
		m.l.RLock()
		defer m.l.RUnlock()

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
			file, size, err = grabFileByURL(url)
		}

		log.Println("file", file.Name())
		log.Println("size", size)

		bboxInfos := make([]*bboxInfo, 0)

		if size > 0 {

			/*
			buf := bytes.NewBuffer(nil)
			if _, err := io.Copy(buf, file); err != nil {
				panic(err.Error())
			}

			kind, _ := filetype.Match(buf.Bytes())
			*/

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
					if (d.ClassNames[i] == "car" || d.ClassNames[i] == "truck") && d.Probabilities[i] >= 70 {
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
				b := src.Bounds()
				imgWidth := b.Max.X
				imgHeight := b.Max.Y
			    log.Println("src.Width:", imgWidth ,"src.Height:", imgHeight)
			    if imgWidth > 700 {
					src = imaging.Resize(src, 700, 0, imaging.Lanczos)
			    }			  
				err = imaging.Encode(c.Writer, src, imaging.JPEG)
			    if err != nil {
			        log.Fatalf("failed to encode image: %v", err)
			    }
			}

			// remove temporary file
			//err = os.Remove(file.Name())
			//if err != nil {
			//	panic(err)
			//}

		} else {
			c.String(200, "Nothing")
		}

		log.Println("crop end")
	})

	r.POST("/crop", func(c *gin.Context) {
		m.l.RLock()
		defer m.l.RUnlock()
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

	m = yoloModel{
		n: n,
		l: sync.RWMutex{},
	}

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
