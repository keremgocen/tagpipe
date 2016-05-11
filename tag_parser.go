package tagparser

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// FileCache holds info of each file, including tags observed per file
type FileCache struct {
	md5  string
	name string
	tc   TagCount
}

// Cache is used to cache parsed files, to avoid parsing the same file again
var Cache map[string]FileCache

// TagCount holds tags as key and their count
type TagCount struct {
	Key   string
	Value int
}

// SortedTagCounts used to sort TagCounts by value
type SortedTagCounts []TagCount

func (p SortedTagCounts) Len() int           { return len(p) }
func (p SortedTagCounts) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p SortedTagCounts) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Check exits on error
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// CountTagsInFile counts all given tags inside the file
func CountTagsInFile(file *strings.Reader, tag string) int {

	var telephone = regexp.MustCompile(`[A-Za-z]+`)
	// var telephone = regexp.MustCompile(`\(\d+\)\s\d+-\d+`)

	// do I need buffered channels here?
	tags := make(chan string)
	results := make(chan int)

	// I think we need a wait group, not sure.
	wg := new(sync.WaitGroup)

	// start up some workers that will block and wait?
	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go MatchTags(tags, results, wg, telephone)
	}

	// Go over a file line by line and queue up a ton of work
	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			// Later I want to create a buffer of lines, not just line-by-line here ...
			tags <- scanner.Text()
		}
		close(tags)
	}()

	func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			// Later I want to create a buffer of lines, not just line-by-line here ...
			tags <- scanner.Text()
		}
		close(tags)
	}()

	// Now collect all the results...
	// But first, make sure we close the result channel when everything was processed
	go func() {
		wg.Wait()
		close(results)
	}()

	// Add up the results from the results channel.
	counts := 0
	for v := range results {
		counts += v
	}

	return counts
}

// MatchTags counts tags in the given file
func MatchTags(tags <-chan string, results chan<- int, wg *sync.WaitGroup, telephone *regexp.Regexp) {
	// func matchTags(tags <-chan string, results chan<- int, wg *sync.WaitGroup, telephone *regexp.Regexp) {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer wg.Done()

	// eventually I want to have a []string channel to work on a chunk of lines not just one line of text
	for j := range tags {
		if telephone.MatchString(j) {
			results <- 1
		}
	}
}

// IsValidJSON checks if the given string has a valid JSON format, generalized
func IsValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// TimeTrack utility to measure the elapsed time in ms
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("\n%s took %s\n\n", name, elapsed)
}

func main() {
	defer TimeTrack(time.Now(), "total execution")

	const dataPath = "data/"
	var tags []string
	Cache = make(map[string]FileCache)
	tM := make(map[string]int) // tag map keeping total counts

	files, _ := ioutil.ReadDir(dataPath)

	if len(os.Args[1:]) > 0 { // os.Args[1:] holds the arguments to the program.
		tags = os.Args[1:]
	} else { // parse arguments from 'tags.txt'
		fmt.Println("Missing command-line arguments! Fetching from file `tags.txt`..")
		dat, err := ioutil.ReadFile("tags.txt")
		Check(err)
		tags = strings.Fields(string(dat))
	}

	for _, f := range files { // search for tags in each file asynchronously

		// lookup for file in cache, avoid parsing again
		// fmt.Println("--metadata name:", f.Name(), " mod_time:", f.ModTime(), " size:", f.Size())
		md5Str := f.Name() + f.ModTime().String() + string(f.Size())
		hash := md5.New()
		io.WriteString(hash, md5Str)
		fmt.Printf("\nfile:%s, md5:%x\n", f.Name(), hash.Sum(nil))

		content, err := ioutil.ReadFile(dataPath + f.Name())
		Check(err)

		fileHasValidJSON := IsValidJSON(string(content))

		if fileHasValidJSON {
			// fmt.Printf("contents:%s,  \nisJSON:%v\n---------------------\n", string(content), true)
			for _, tag := range tags {
				r, _ := regexp.Compile(tag)
				c := len(r.FindAllString(string(content), -1))
				if c > 0 {
					tM[tag]++
				}
				fmt.Println("tag:", tag, " - count:", c)
				// if r.Match(content) {
				// 	fmt.Println("found tag:", tag, " in file:", f.Name(), " count:")
				// }
			}
			fmt.Printf("---------------------\n")
		} else {
			fmt.Println("skipping file ", f.Name(), " with invalid JSON")
			fmt.Println("---------------------")
		}

	}

	sT := make(SortedTagCounts, len(tM))
	i := 0
	for k, v := range tM {
		sT[i] = TagCount{k, v}
		i++
	}
	sort.Sort(sort.Reverse(sT))

	fmt.Println("tag map:", tM)
	fmt.Println("sorted tag map:", sT)

}
