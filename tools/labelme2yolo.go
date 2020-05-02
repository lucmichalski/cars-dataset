package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/karrick/godirwalk"
	log "github.com/sirupsen/logrus"
)

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
	walkImages(".json", "../tools/dataset/bus/")
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
	FillColor []int   `json:"fill_color"`
	Label     string  `json:"label"`
	LineColor []int   `json:"line_color"`
	Points    [][]int `json:"points"` // []array
	ShapeType string  `json:"shape_type"`
}

func convert(fp, fi string) error {
	file, err := os.Open(fp)
	if err != nil {
		return err
	}

	byteValue, _ := ioutil.ReadAll(file)
	var labelme Labelme
	if err := json.Unmarshal(byteValue, &labelme); err != nil {
		log.Fatalln("unmarshal error, ", err)
	}
	pp.Println(labelme)

	return nil

}

func walkImages(extension string, dirnames ...string) (err error) {
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
