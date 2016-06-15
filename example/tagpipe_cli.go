package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/keremgocen/tagpipe"
)

// TODO - use flags
const dataPath = "../data/"

func main() {
	var tags []string
	// tagpipe.Cache = make(map[string]tagpipe.FileCache)
	// tM := make(map[string]int) // tag map keeping total counts
	// files, _ := ioutil.ReadDir(dataPath)

	// determine tags to be parsed - this part is blocking
	if len(os.Args[1:]) > 0 { // os.Args[1:] holds the arguments to the program.
		tags = os.Args[1:]
	} else { // parse arguments from 'tags.txt'
		fmt.Println("Missing command-line arguments! Fetching from the file `tags.txt`..")
		dat, err := ioutil.ReadFile("../tags.txt")
		if err != nil {
			fmt.Println("ReadFile returned error:", err)
		}
		tags = strings.Fields(string(dat))
	}

	fmt.Println(tags)

	// Calculate the MD5 sum of all files under the specified directory,
	// then print the results sorted by path name.
	m, err := tagpipe.DigestAllFiles(dataPath, tags)
	if err != nil {
		fmt.Println(err)
		return
	}
	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	// sort.Strings(paths)
	// for _, path := range paths {
	// 	fmt.Printf("%x  %s\n", m[path], path)
	// }

	// sT := make(tagpipe.SortedTagCounts, len(tM))
	// i := 0
	// for k, v := range tM {
	// 	sT[i] = tagpipe.TagCount{k, v}
	// 	i++
	// }
	// sort.Sort(sort.Reverse(sT))
	//
	// fmt.Println("tag map:", tM)
	// fmt.Println("sorted tag map:", sT)
}
