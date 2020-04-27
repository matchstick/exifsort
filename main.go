package main

import (
	"flag"
	"fmt"
	"github.com/matchstick/exifSort/lib"
)

func main() {
	mediaDir := flag.String("media", "bobo", "Directory with Media")
	flag.Parse()
	fmt.Println("Hello. Photo Media found here ", *mediaDir)
	fmt.Println("tail:", flag.Args())
	exifSort.PhotoSorting()
}

