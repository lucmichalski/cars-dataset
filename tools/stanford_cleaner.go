package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/karrick/godirwalk"
	log "github.com/sirupsen/logrus"
)

var (
	isDryMode = false
	counter   = 0
	total     = 0
)

func main() {
	walkImages(".jpg", "../shared/dataset/stanford_train/")
	fmt.Println("purged ", counter, "files over ", total, "files")
}

func walkImages(extension string, dirnames ...string) (err error) {
	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if !de.IsDir() {
					if strings.Contains(osPathname, extension) {
						fmt.Println("found ", osPathname, "extension", extension)
						annotationFile := strings.Replace(osPathname, ".jpg", ".json", -1)
						fmt.Println("checking if annotation file exists, ", annotationFile)
						if _, err := os.Stat(annotationFile); err != nil {
							if !isDryMode {
								os.Remove(osPathname)
							}
							counter++
							fmt.Println("removed ", osPathname)
						}
						total++
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
