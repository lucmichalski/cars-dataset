package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"
)

var (
	isDryMode = false
)

func main() {
	fp := "./negative.csv"
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
		pp.Println("found link: ", row[0])
		f, _, err := openFileByURL(row[0], "./test_negative")
		checkErr(err)

		// create empty bbox file for yolo
		annotationFile := strings.Replace(f.Name(), filepath.Ext(f.Name()), ".txt", -1)

		pp.Println("annotationFile: ", annotationFile)

		if !isDryMode {
			ft, err := os.Create(annotationFile)
			checkErr(err)
			_, err = ft.WriteString("")
			checkErr(err)
			ft.Sync()
			ft.Close()
		}

	}
}

func openFileByURL(rawURL, destPath string) (*os.File, int64, error) {
	if fileURL, err := url.Parse(rawURL); err != nil {
		return nil, 0, err
	} else {
		var segments []string
		path := fileURL.Path
		segments = strings.Split(path, "/")
		// fileName := segments[len(segments)-1]
		// http://35.179.44.166:9008/system/vehicle_images/1790564/file.jpg
		extension := filepath.Ext(path)
		fileName := segments[3]
		// pp.Println("segments", segments)
		filePath := filepath.Join(destPath, fileName+extension)

		pp.Println("download to local path: ", filePath)

		// return nil, 0, err

		file, err := os.Create(filePath)
		if err != nil {
			return file, 0, err
		}

		check := http.Client{
			// Timeout: 10 * time.Second,
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

		fmt.Println("Content-Length:", resp.Header.Get("Content-Length"))

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
