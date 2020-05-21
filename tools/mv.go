package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/karrick/godirwalk"
	log "github.com/sirupsen/logrus"
)

var (
	countMoved = 0
)

const (
	shellToUse      = "bash"
	sourcePath      = `/home/ubuntu/cars-dataset/public/system/vehicle_images/`
	destinationPath = `/mnt/nasha/lucmichalski/cars-dataset/public/system/vehicle_images/`
)

func main() {
	walkImages(sourcePath)
}

func walkImages(dirnames ...string) (err error) {
	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if de.IsDir() {
					// fmt.Println("osPathname:", osPathname)
					osPathnameParts := strings.Split(osPathname, "/")
					osPathnameID := osPathnameParts[len(osPathnameParts)-1]
					newLocation := destinationPath + osPathnameID
					fmt.Println("osPathname:", osPathname, "newLocation:", newLocation)
					// err := os.Rename(oldLocation, newLocation)
					// if err != nil {
					//	log.Fatal(err)
					// }
					countMoved++
				}
				return nil
			},
			Unsorted: true,
		})
	}
	return
}

func shellout(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(shellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
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
