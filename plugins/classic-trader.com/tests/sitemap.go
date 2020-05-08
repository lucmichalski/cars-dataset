package main

import (
        "fmt"
	// "os"
        "path/filepath"
	"strings"

	// log "github.com/sirupsen/logrus"
	// "github.com/k0kubun/pp"
	"github.com/beevik/etree"
)

func main() {

         pattern := "./*.xml"
         matches, err := filepath.Glob(pattern)
         if err != nil {
                 fmt.Println(err)
         }
         // pp.Println(matches)

	// os.Exit(1)

	var urls []string
	for _, match := range matches {
		doc := etree.NewDocument()
		if err := doc.ReadFromFile(match); err != nil {
		    panic(err)
		}
		urlset := doc.SelectElement("urlset")
		if urlset != nil {
			entries := urlset.SelectElements("url")
			for _, entry := range entries {
				loc := entry.SelectElement("loc")
				l := loc.Text()
				l = strings.TrimLeftFunc(l, func(c rune) bool {
					return c == '\r' || c == '\n' || c == '\t'
				})
				l = strings.TrimSpace(l)
				fmt.Println(l)
				urls = append(urls, l)
			}
		}
	}
	// pp.Println(urls)
}

