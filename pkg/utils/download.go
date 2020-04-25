package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"io/ioutil"

	"github.com/k0kubun/pp"
	"github.com/h2non/filetype"
	// "github.com/cavaliercoder/grab"
	"github.com/pkg/errors"

	"github.com/lucmichalski/cars-dataset/pkg/grab"
)

func OpenFileByURL(rawURL string) (*os.File, int64, string, error) {
	if fileURL, err := url.Parse(rawURL); err != nil {
		return nil, 0, "", err
	} else {
		q := fileURL.Query()
		var segments []string
		if q.Get("url") != "" {
			// if 
			if strings.Contains(strings.ToLower(q.Get("url")), "jpg") {
				segments = append(segments, ".jpg")
			} else if strings.Contains(strings.ToLower(q.Get("url")), "png") {
				segments = append(segments, ".png")
			} else {
				segments = strings.Split(q.Get("url"), "/")
			}
		} else {
			path := fileURL.Path
			segments = strings.Split(path, "/")
		}

		fileName := GetMD5Hash(rawURL) + "-" + segments[len(segments)-1]

		if strings.Contains(fileName, "?") {
                        // clean up query string
                        fileParts := strings.Split(fileName, "?")
                        if len(fileParts) > 0 {
                                fileName = fileParts[0]
                        }
		}
		filePath := filepath.Join(os.TempDir(), fileName)

		file, err := os.Create(filePath)
		if err != nil {
			return file, 0, "", err
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
			return file, 0, "", err
		}
		defer resp.Body.Close()
		fmt.Printf("----> Downloaded %v\n", rawURL)

		fmt.Println("Content-Length:", resp.Header.Get("Content-Length"))

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return file, 0, "", err
		}

		buf, _ := ioutil.ReadFile(file.Name())
		kind, _ := filetype.Match(buf)
		pp.Println("kind: ", kind)

		fi, err := file.Stat()
		if err != nil {
			return file, 0, "", err
		}

		checksum, err := GetMD5File(filePath)
		if err != nil {
			return file, 0, "", err
		}

		return file, fi.Size(), checksum, nil
	}
}

func GrabFileByURL(rawURL string) (*os.File, int64, error) {
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
