package main

import (
	"fmt"
	"strings"
	//"path/filepath"
	"encoding/csv"
	"os"
	//"path"
	"io"

	"github.com/k0kubun/pp"	
	log "github.com/sirupsen/logrus"
	"github.com/karrick/godirwalk"
)

var (
	isDryMode = true
	yoloClass = map[string]string{
		"0": "car",
	}
)

/*
[]string{
  "TRAIN",
  "../../shared/datasets/stanford-cars/cars_train/05089.jpg",
  "(1280, 782)",
  "68",
  "287",
  "1228",
  "667",
  "Spyker C8 Convertible 2009",
  "(0.138671875, 1.2116368286445014, 0.17109375000000002, 83.08823529411765)",
}
*/

func main() {

	fp := "../shared/datasets/stanford-cars/yolo_cars_data.csv"
	file, err := os.Open(fp)
	checkErr(err)

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
		    if perr, ok := err.(*csv.ParseError); ok && perr.Err == csv.ErrFieldCount {
		        continue
		    }
		    checkErr(err)
		}

		// cleanup dimension
		row[2] = strings.Replace(row[2], "(", "", -1)
		row[2] = strings.Replace(row[2], ")", "", -1)
		imgDim := strings.Split(row[2], ",")
		for i, d := range imgDim {
			imgDim[i] = strings.TrimSpace(d)
		}

		// cleanup yolo annotation
		row[8] = strings.Replace(row[8], "(", "", -1)
		row[8] = strings.Replace(row[8], ")", "", -1)
		yoloAnno := strings.Split(row[8], ",")
		for i, y := range yoloAnno {
			yoloAnno[i] = strings.TrimSpace(y)
		}

		pp.Println("imgSrc", row[1])
		pp.Println("imgDim", imgDim)
		pp.Println("yoloAnno", yoloAnno)

		annoTextCnt := "0 " + strings.Join(yoloAnno, " ")
		annoTextFile := strings.Replace(row[1], ".jpg", ".txt", -1)
		annoTextFile = strings.Replace(annoTextFile, "../", "", 1)
		// write file
		ft, err := os.Create(annoTextFile)
		checkErr(err)

		_, err = ft.WriteString(annoTextCnt)
		checkErr(err)
		ft.Sync()
		ft.Close()

	}

}

func walkImages(extension string, dirnames ...string) (err error ){
	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if !de.IsDir() {
					if strings.Contains(osPathname, extension) {
						fmt.Println("found ", osPathname, "extension", extension)
						// dispatch(osPathname)
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
