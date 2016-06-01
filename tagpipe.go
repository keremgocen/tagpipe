package tagpipe

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// TODO - add doc
type result struct {
	path string
	sum  [md5.Size]byte
	err  error
	tmap *map[string]int
}

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

// walkFiles starts a goroutine to walk the directory tree at root and send the
// path of each regular file on the string channel.  It sends the result of the
// walk on the error channel.  If done is closed, walkFiles abandons its work.
func walkFiles(done <-chan struct{}, root string) (<-chan string, <-chan error) {
	paths := make(chan string)
	errc := make(chan error, 1)
	go func() { // HL
		// Close the paths channel after Walk returns.
		defer close(paths) // HL
		// No select needed for this send, since errc is buffered.
		errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error { // HL
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			select {
			case paths <- path: // HL
			case <-done: // HL
				return errors.New("walk canceled")
			}
			return nil
		})
	}()
	return paths, errc
}

// digester reads path names from paths and sends digests of the corresponding
// files on c until either paths or done is closed.
func digester(done <-chan struct{}, paths <-chan string, c chan<- result, tags []string) {

	for path := range paths { // HLpaths
		data, err := ioutil.ReadFile(path)

		// discard files containing invalid JSON
		fileHasValidJSON := IsValidJSON(string(data))

		tM := make(map[string]int) // tag map keeping total counts

		if fileHasValidJSON {

			// debug
			var y map[string]interface{}
			json.Unmarshal(data, &y)
			fmt.Println("\npath:", path, " content:", y)

			// fmt.Printf("contents:%s,  \nisJSON:%v\n---------------------\n", string(content), true)
			for _, tag := range tags {
				r, _ := regexp.Compile("\"" + tag + "\"")
				c := len(r.FindAllStringIndex(string(data), -1))
				if c > 0 {
					tM[tag] += c
				}
				fmt.Println("tag:", tag, " - count:", c)
				// if r.Match(data) {
				// 	fmt.Println("found tag:", tag, " in file:", path, " count:", len(r.FindAllStringSubmatchIndex(string(data), -1)))
				// }
			}

		} else {
			fmt.Println("\nskipping file ", path, " with invalid JSON")
			return
		}

		// debug
		if len(tM) > 0 {
			fmt.Println("\nfound tags:", tM, " in file:", path)
		}

		select {
		case <-done:
			return
		default:
			c <- result{path, md5.Sum(data), err, &tM}
		}
	}
}

// MD5All reads all the files in the file tree rooted at root and returns a map
// from file path to the MD5 sum of the file's contents.  If the directory walk
// fails or any read operation fails, MD5All returns an error.  In that case,
// MD5All does not wait for inflight read operations to complete.
func MD5All(root string, tags []string) (map[string][md5.Size]byte, error) {
	// MD5All closes the done channel when it returns; it may do so before
	// receiving all the values from c and errc.
	done := make(chan struct{})
	defer close(done)

	paths, errc := walkFiles(done, root)

	totaltags := make(map[string]int) // tag map keeping total counts

	// Start a fixed number of goroutines to read and digest files.
	c := make(chan result) // HLc
	var wg sync.WaitGroup
	const numDigesters = 20
	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			digester(done, paths, c, tags) // HLc
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(c) // HLc
	}()
	// End of pipeline. OMIT

	m := make(map[string][md5.Size]byte)
	for r := range c {
		if r.err != nil {
			return nil, r.err
		}
		m[r.path] = r.sum
		fmt.Println("received tags:")
		for t, i := range *r.tmap {
			fmt.Println(t, i)
			totaltags[t] += i
		}
	}

	fmt.Println(totaltags)
	// Check whether the Walk failed.
	if err := <-errc; err != nil { // HLerrc
		return nil, err
	}
	return m, nil
}

// // CountTagsInFile counts all given tags inside the file
// func CountTagsInFile(file *strings.Reader, tag string) int {
//
// 	var telephone = regexp.MustCompile(`[A-Za-z]+`)
// 	// var telephone = regexp.MustCompile(`\(\d+\)\s\d+-\d+`)
//
// 	// do I need buffered channels here?
// 	tags := make(chan string)
// 	results := make(chan int)
//
// 	// I think we need a wait group, not sure.
// 	wg := new(sync.WaitGroup)
//
// 	// start up some workers that will block and wait?
// 	for w := 1; w <= 3; w++ {
// 		wg.Add(1)
// 		go MatchTags(tags, results, wg, telephone)
// 	}
//
// 	// Go over a file line by line and queue up a ton of work
// 	go func() {
// 		scanner := bufio.NewScanner(file)
// 		for scanner.Scan() {
// 			// Later I want to create a buffer of lines, not just line-by-line here ...
// 			tags <- scanner.Text()
// 		}
// 		close(tags)
// 	}()
//
// 	func() {
// 		scanner := bufio.NewScanner(file)
// 		for scanner.Scan() {
// 			// Later I want to create a buffer of lines, not just line-by-line here ...
// 			tags <- scanner.Text()
// 		}
// 		close(tags)
// 	}()
//
// 	// Now collect all the results...
// 	// But first, make sure we close the result channel when everything was processed
// 	go func() {
// 		wg.Wait()
// 		close(results)
// 	}()
//
// 	// Add up the results from the results channel.
// 	counts := 0
// 	for v := range results {
// 		counts += v
// 	}
//
// 	return counts
// }
//
// // MatchTags counts tags in the given file
// func MatchTags(tags <-chan string, results chan<- int, wg *sync.WaitGroup, telephone *regexp.Regexp) {
// 	// func matchTags(tags <-chan string, results chan<- int, wg *sync.WaitGroup, telephone *regexp.Regexp) {
// 	// Decreasing internal counter for wait-group as soon as goroutine finishes
// 	defer wg.Done()
//
// 	// eventually I want to have a []string channel to work on a chunk of lines not just one line of text
// 	for j := range tags {
// 		if telephone.MatchString(j) {
// 			results <- 1
// 		}
// 	}
// }

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

// func main() {
// 	defer TimeTrack(time.Now(), "total execution")
//
//
//
// }
