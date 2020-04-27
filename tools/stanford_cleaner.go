package main

import (
	"fmt"
	"strings"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/karrick/godirwalk"
)

var (
	isDryMode = true
	counter = 0
	total = 0
)

func main() {
	walkImages(".jpg", "../shared/datasets/stanford-cars/cars_train/")
	fmt.Println("purged ", counter, "files over ", total, "files")
}

func walkImages(extension string, dirnames ...string) (err error ){
	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if !de.IsDir() {
					if strings.Contains(osPathname, extension) {
						fmt.Println("found ", osPathname, "extension", extension)
						annotationFile := strings.Replace(osPathname, ".jpg", ".txt", -1)
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
