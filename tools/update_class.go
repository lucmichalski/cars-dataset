package main

import (
	"encoding/csv"
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
	walkImages(".txt", "../shared/dataset/stanford_train/")
	fmt.Println("purged ", counter, "files over ", total, "files")
}

func walkImages(extension string, dirnames ...string) (err error) {
	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if !de.IsDir() {
					if strings.Contains(osPathname, extension) {

						file, err := os.Open(osPathname)
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
						file.Close()

						for _, row := range data {
							classId := 1
							centerX := row[1]
							centerY := row[2]
							width := row[3]
							height := row[4]

							ft, err := os.Create(osPathname)
							checkErr(err)

							_, err = ft.WriteString(fmt.Sprintf("%s %s %s %s %s", classId, centerX, centerY, width, height))
							checkErr(err)
							ft.Sync()
							ft.Close()

						}

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
