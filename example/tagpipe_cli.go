package main

import (
	"fmt"
	"sort"

	"github.com/keremgocen/tagpipe"
)

// TODO - use flags
const dataPath = "../data/"

func main() {
	// var tags []string
	// tagpipe.Cache = make(map[string]tagpipe.FileCache)
	// tM := make(map[string]int) // tag map keeping total counts
	// files, _ := ioutil.ReadDir(dataPath)

	// Calculate the MD5 sum of all files under the specified directory,
	// then print the results sorted by path name.
	m, err := tagpipe.MD5All(dataPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Printf("%x  %s\n", m[path], path)
	}
	//
	// if len(os.Args[1:]) > 0 { // os.Args[1:] holds the arguments to the program.
	// 	tags = os.Args[1:]
	// } else { // parse arguments from 'tags.txt'
	// 	fmt.Println("Missing command-line arguments! Fetching from the file `tags.txt`..")
	// 	dat, err := ioutil.ReadFile("../tags.txt")
	// 	tagpipe.Check(err)
	// 	tags = strings.Fields(string(dat))
	// }
	//
	// for _, f := range files { // search for tags in each file asynchronously
	//
	// 	// lookup for file in cache, avoid parsing again
	// 	// fmt.Println("--metadata name:", f.Name(), " mod_time:", f.ModTime(), " size:", f.Size())
	// 	md5Str := f.Name() + f.ModTime().String() + string(f.Size())
	// 	hash := md5.New()
	// 	io.WriteString(hash, md5Str)
	// 	fmt.Printf("\nfile:%s, md5:%x\n", f.Name(), hash.Sum(nil))
	//
	// 	content, err := ioutil.ReadFile(dataPath + f.Name())
	// 	tagpipe.Check(err)
	//
	// 	fileHasValidJSON := tagpipe.IsValidJSON(string(content))
	//
	// 	if fileHasValidJSON {
	// 		// fmt.Printf("contents:%s,  \nisJSON:%v\n---------------------\n", string(content), true)
	// 		for _, tag := range tags {
	// 			r, _ := regexp.Compile(tag)
	// 			c := len(r.FindAllString(string(content), -1))
	// 			if c > 0 {
	// 				tM[tag]++
	// 			}
	// 			fmt.Println("tag:", tag, " - count:", c)
	// 			// if r.Match(content) {
	// 			// 	fmt.Println("found tag:", tag, " in file:", f.Name(), " count:")
	// 			// }
	// 		}
	// 		fmt.Printf("---------------------\n")
	// 	} else {
	// 		fmt.Println("skipping file ", f.Name(), " with invalid JSON")
	// 		fmt.Println("---------------------")
	// 	}
	//
	// }
	//
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
