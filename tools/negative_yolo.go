package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/k0kubun/pp"	
	"github.com/karrick/godirwalk"
	log "github.com/sirupsen/logrus"
)

var (
	isDryMode      = true
	isDeleteMode   = false
	isDispatchMode = true
	yoloClass      = map[string]string{
		"0": "person",
		"1": "car",
		"2": "motorcycle",
		"3": "bus",
		"4": "truck",
	}
	countDelete = 0
)

const (
	datasetAbsPath = `/home/ubuntu/cars-dataset/shared/datasets/5classes/train2017_set/`
)

var (
	datasetNoCars = filepath.Join("dataset", "train2017_nocars")
)

func main() {

	err := ensureDir(datasetNoCars)
	checkErr(err)

	walkImages(".txt", datasetAbsPath)
	if isDeleteMode {
		fmt.Println("countDelete=", countDelete)
	}
}

func dispatch(fp string) error {
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

	if isDispatchMode {
		pp.Println(data)
		hasCar := false
		for _, row := range data {
			classId := row[0]
			className := yoloClass[classId]
			if className == "car" {
				hasCar = true
			}
		}
		if !hasCar {
			srcFileNameIMG := strings.Replace(fp, ".txt", ".jpg", -1)
			destFileNameTXT := filepath.Join(datasetNoCars, path.Base(fp))
			destFileNameIMG := filepath.Join(datasetNoCars, strings.Replace(path.Base(fp), ".txt", ".jpg", -1))
			copy(fp, destFileNameTXT)
			copy(srcFileNameIMG, destFileNameIMG)			
		}

	}

	if isDeleteMode {
		var isDelete bool
		for _, row := range data {
			if row[0] == "0" {
				isDelete = true
				break
			}
		}
		if isDelete {
			countDelete++
			if !isDryMode {
				err := os.Remove(fp)
				checkErr(err)
				err = os.Remove(strings.Replace(fp, ".txt", ".jpg", -1))
				checkErr(err)
				err = os.Remove(strings.Replace(fp, ".txt", ".json", -1))
				checkErr(err)
			} else {
				fmt.Println("should remove file=", fp)
			}
		}
	}

	return nil

}

func walkImages(extension string, dirnames ...string) (err error) {
	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if !de.IsDir() {
					if strings.Contains(osPathname, extension) {
						fmt.Println("found ", osPathname, "extension", extension)
						dispatch(osPathname)
					}
				}
				return nil
			},
			Unsorted: true,
		})
	}
	return
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func ensureDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		os.MkdirAll(path, os.FileMode(0755))
	} else {
		return err
	}
	d.Close()
	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
