package main

import (
	"fmt"
	"strings"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/karrick/godirwalk"
)

func main() {
	walkImages(".jpg", "data/obj")
}

func walkImages(extension string, dirnames ...string) (err error ){

	// write file
	ft, err := os.Create("data/train.txt")
	checkErr(err)
	defer ft.Close()

	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if !de.IsDir() {
					if strings.Contains(osPathname, extension) {
						fmt.Println("found ", osPathname, "extension", extension)
						_, err = ft.WriteString(osPathname+"\n")
						checkErr(err)
						ft.Sync()
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

