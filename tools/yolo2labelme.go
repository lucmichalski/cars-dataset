package main

import (
	"fmt"
	"strings"
	"path/filepath"
	"encoding/csv"
	"encoding/json"
    "strconv"
    "os"
    "image"
    _ "image/jpeg"
    _ "image/png"

	log "github.com/sirupsen/logrus"
	"github.com/k0kubun/pp"
	"github.com/karrick/godirwalk"
)

const datasetAbsPath = `/home/ubuntu/cars-dataset/shared/datasets/train2017_set/`

var (
	yoloClass = map[string]string{
		"0": "person",
		"1": "car",
		"2": "motorcycle",
		"3": "bus",
		"4": "truck",
	}
)


func main() {
	walkImages(".jpg", "../shared/datasets/stanford-cars/cars_test/")
	// walkImages(".jpg", "/home/ubuntu/cars-dataset/shared/datasets/train2017_set/")
}

type Labelme struct {
	FillColor   []int       `json:"fillColor"`
	Flags       Flags       `json:"flags"`
	ImageData   interface{} `json:"imageData"`
	ImageHeight int         `json:"imageHeight"`
	ImagePath   string      `json:"imagePath"`
	ImageWidth  int         `json:"imageWidth"`
	LineColor   []int       `json:"lineColor"`
	Shapes      []Shape     `json:"shapes"`
	Version     string      `json:"version"`
}

type Flags struct {
}

type Shape struct {
	FillColor []int 	  `json:"fill_color"`
	Label     string      `json:"label"`
	LineColor []int 	  `json:"line_color"`
	Points    [][]int     `json:"points"` // []array
	ShapeType string      `json:"shape_type"`
}

func convert(fp, fi string) error {
    file, err := os.Open(fp)
    if err != nil {
        return err
    }

	reader := csv.NewReader(file)
	reader.Comma = ' '
	reader.LazyQuotes = true
	data, err := reader.ReadAll()
	if err != nil {
		return err
	}
	width, height := getImageDimension(fi)
	extension := filepath.Ext(fp)
	labelmeFile := strings.Replace(fp, extension, ".json", -1)	

	lc := Labelme{}
	lc.FillColor = []int{255, 0, 0, 128}
	lc.ImageHeight = height
	lc.ImageWidth = width
	lc.ImagePath = filepath.Base(fi)

	lc.LineColor = []int{0, 255, 0, 128}
	lc.Version = "3.6.10"

	for _, row := range data {
		classId := row[0]
		className := yoloClass[classId]
		centerX, err := strconv.ParseFloat(row[1], 64)
		checkErr(err)
		centerY, err := strconv.ParseFloat(row[2], 64)
		checkErr(err)
		widthPercent, err := strconv.ParseFloat(row[3], 64)
		checkErr(err)
		heightPercent, err := strconv.ParseFloat(row[4], 64)
		checkErr(err)
		fmt.Println("classId=",classId,", className=",className,", centerX=",centerX,", centerY=",centerY,", widthPercent=",widthPercent,", heightPercent=",heightPercent,", width=",width,", height=",height)

		sc := Shape{}
		sc.Label = className
		//if className == "car" {
		//	m.Set(ImagePath, true)
		//}
		sc.ShapeType = "rectangle"

		// get the coordinates
		centerXpx := centerX * float64(width)
		centerYpx := centerY * float64(height)
		widthXpx := widthPercent * float64(width)
		heightYpx := heightPercent * float64(height)

		fmt.Println("centerXpx=",centerXpx,"centerYpx=",centerYpx)
		fmt.Println("widthXpx=",centerXpx,"heightYpx=",centerYpx)

		minX := centerXpx - (widthXpx/2)
		maxX := centerXpx + (widthXpx/2)
		minY := centerYpx - (heightYpx/2)
		maxY := centerYpx + (heightYpx/2)

		fmt.Println("minX=",minX,"maxX=",maxX,"minY=",minY,"maxY=",maxY)

		x := []int{int(maxX),int(maxY)}
		y := []int{int(minX),int(minY)}

		points := [][]int{x, y}
		sc.Points = append(sc.Points, points...)
		lc.Shapes = append(lc.Shapes, sc)
	}

	pp.Println(lc)
	slcB, err := json.MarshalIndent(&lc, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("labelmeFile=%s\n", labelmeFile)

	// write file
	f, err := os.Create(labelmeFile)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err = f.WriteString(string(slcB))
	if err != nil {
		log.Fatal(err)
	}
	f.Sync()

	// dump image car dataset
	//m.IterCb(func(url string, v interface{}) {
	//})

	return nil

}

func getImageDimension(imagePath string) (int, int) {
    file, err := os.Open(imagePath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
    }
    defer file.Close()
    image, _, err := image.DecodeConfig(file)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
    }
    return image.Width, image.Height
}

func walkImages(extension string, dirnames ...string) (err error ){
	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if !de.IsDir() {
					if strings.Contains(osPathname, extension) {
						fmt.Println("found ", osPathname, "extension", extension)
						extension := filepath.Ext(osPathname)
						annoPathname := strings.Replace(osPathname, extension, ".txt", -1)
						convert(annoPathname, osPathname)
					}
				}
				return nil
			},
			Unsorted: true,
		})
	}
	return 
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
