package main

import (
	"fmt"
	"os"
	"webCrawler/sitemap"
)

func main() {
	startingPoint := os.Args[1]

	sm := sitemap.NewSiteMap()
	err := sm.ProduceFrom(startingPoint)
	if err != nil {
		fmt.Printf("Could not produce site map: " + err.Error())
	} else {
		sm.Print()
	}
}
