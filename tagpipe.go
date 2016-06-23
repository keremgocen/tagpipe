package tagpipe

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"
)

// Result is returned from digesters containing the necessary information
// about digestion result and tag count map of the file it corresponds to
type Result struct {
	Path string
	Sum  string
	E    error
	T    map[string]int
}

// cacheMap is used to cache parsed files, to avoid parsing the same file again
var cacheMap map[string]Result

// Becomes false when user disables cache via command line flags
var uc bool // use cache

// T holds tag, count pairs
type T struct {
	Tag   string
	Count int
}

// TList sorts tag, count pairs using the implemented functions below
type TList []T

func (t TList) Len() int           { return len(t) }
func (t TList) Less(i, j int) bool { return t[i].Count < t[j].Count }
func (t TList) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

// sortByTagCount sorts a tag map by counts
func sortByTagCount(tagFrequencies map[string]int) TList {
	tl := make(TList, len(tagFrequencies))
	i := 0
	for k, v := range tagFrequencies {
		tl[i] = T{k, v}
		i++
	}
	sort.Sort(sort.Reverse(tl))
	return tl
}

// walkFiles starts a goroutine to walk the directory tree at root and send the
// path of each regular file on the string channel.  It sends the result of the
// walk on the error channel.  If done is closed, walkFiles abandons its work.
func walkFiles(done <-chan struct{}, root string) (<-chan string, <-chan error, int) {
	paths := make(chan string)
	errc := make(chan error, 1)

	// used to optimize digester count
	files, _ := ioutil.ReadDir(root)

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
	return paths, errc, len(files)
}

// digester reads path names from paths and sends digests of the corresponding
// files on c until either paths or done is closed.
func digester(done <-chan struct{}, paths <-chan string, c chan<- Result, tags []string) {

	for path := range paths { // HLpaths
		data, err := ioutil.ReadFile(path)

		tM := make(map[string]int) // tag map keeping total counts

		// TODO before parsing tags, check if the file was processed before
		// skip if file is found in cache
		bytesMD5 := md5.Sum(data)
		sumMD5 := hex.EncodeToString(bytesMD5[:])
		savedResult, ok := cacheMap[sumMD5]

		if uc && ok {
			log.Println("identical file found in cache", path)
			c <- savedResult
			continue
		}

		// discard files containing invalid JSON
		fileHasValidJSON := IsValidJSON(string(data))

		if fileHasValidJSON {

			for _, tag := range tags {
				r, _ := regexp.Compile("\"" + tag + "\"")
				c := len(r.FindAllStringIndex(string(data), -1))
				if c > 0 {
					tM[tag] += c
				}
			}

			if uc {
				if cacheMap == nil {
					cacheMap = map[string]Result{}
				}
				cacheMap[sumMD5] = Result{path, sumMD5, err, tM}
			}

		} else {
			log.Println("skipping file ", path, " with invalid JSON")
			continue
		}

		select {
		case <-done:
			return
		default:
			// Path string
			// Sum  string
			// E  error
			// T map[string]int
			c <- Result{Path: path, Sum: sumMD5, E: err, T: tM}
		}
	}
}

// DigestAllFiles reads all the files in the file tree rooted at root and returns a map
// from file path to the MD5 sum of the file's contents.  If the directory walk
// fails or any read operation fails, DigestAllFiles returns an error.  In that case,
// DigestAllFiles does not wait for inflight read operations to complete.
func DigestAllFiles(root string, tags []string, useCache bool) (TList, error) {
	defer TimeTrack(time.Now(), "DigestAllFiles")

	// DigestAllFiles closes the done channel when it returns; it may do so before
	// receiving all the values from c and errc.
	done := make(chan struct{})
	defer close(done)

	// prepare cache
	uc = useCache
	if uc {
		cacheMap = LoadCache()
	}

	paths, errc, fc := walkFiles(done, root)

	// Start a fixed number of goroutines to read and digest files.
	c := make(chan Result) // HLc
	var wg sync.WaitGroup

	// create as many digesters as the number of files in root path, with an upper limit 20
	numDigesters := fc
	if numDigesters > 20 {
		numDigesters = 20
	}

	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			digester(done, paths, c, tags) // HLc
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	m := make(map[string]int)
	for r := range c {
		if r.E != nil {
			return nil, r.E
		}

		if len(r.T) == 0 {
			log.Println("no tags found in", r.Path)
			continue
		}

		log.Println("received tags for", r.Path)

		for t, i := range r.T {

			fmt.Println(t, i)

			m[t] += r.T[t]
		}
	}

	// override cache
	log.Println("saving cache..", SaveCache(cacheMap))

	// Check whether the Walk failed.
	if err := <-errc; err != nil { // HLerrc
		return nil, err
	}

	return sortByTagCount(m), nil
}

// LoadCache tries to parse a previously saved cache file
func LoadCache() map[string]Result {
	defer TimeTrack(time.Now(), "LoadCache")

	dat, e1 := ioutil.ReadFile("cache")
	if e1 != nil {
		log.Println(e1, ", creating new cache")
		return nil
	}

	if e2 := json.Unmarshal(dat, &cacheMap); e2 != nil {
		if cacheMap == nil {
			cacheMap = make(map[string]Result)
		}
		return nil
	}

	return cacheMap
}

// SaveCache will save parsing results of all files in a file named "cache"
func SaveCache(cache map[string]Result) bool {
	defer TimeTrack(time.Now(), "SaveCache")

	// marshall cache into JSON array
	cacheJSON, errj := json.Marshal(cache)
	if errj != nil {
		log.Println(errj)
		return false
	}

	err := ioutil.WriteFile("cache", cacheJSON, 0644)
	if err != nil {
		return false
	}
	return true
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
