package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/keremgocen/tagpipe"
)

var dataPath string
var configPath string
var cachePtr bool

func init() {
	flag.StringVar(&dataPath, "d", "../data/", "data folder path is not set, use: --d <path-to-files>")
	flag.StringVar(&configPath, "t", "../tags.txt", "tags file path is not set, use: --t <path-to-tags>")
	flag.BoolVar(&cachePtr, "c", true, "-c=false disables cache")
}

func main() {
	defer tagpipe.TimeTrack(time.Now(), "everything")

	var tags []string

	flag.Parse()

	if flag.NFlag() > 0 || flag.NArg() == 0 { // parse arguments from 'configPath'
		log.Println("Fetching tags from the file:", configPath)
		dat, err := ioutil.ReadFile(configPath)
		if err != nil {
			log.Fatalln("ReadFile returned error:", err)
			flag.PrintDefaults()
			return
		}
		tags = strings.Fields(string(dat))
	} else if flag.NArg() > 0 { // flag.NArg() gives number of regular command line arguments excluding flag usage
		log.Println("Fetching tags from command line arguments..")
		tags = flag.Args()
	}

	log.Println("JSON files will be parsed at location:", dataPath)

	// Calculate the MD5 sum of all files under the specified directory,
	// then print the results sorted by path name.
	m, err := tagpipe.DigestAllFiles(dataPath, tags, cachePtr)
	if err != nil {
		log.Println(err)
		return
	}

	// final output, in the format expected
	fmt.Println("Final output:")
	for _, v := range m {
		fmt.Println(v.Tag, v.Count)
	}
}
