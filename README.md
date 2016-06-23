'A simple JSON tag parser'
=====

[![Go Report Card](https://goreportcard.com/badge/github.com/keremgocen/tagpipe)](https://goreportcard.com/report/github.com/keremgocen/tagpipe)
[![GoDoc](https://godoc.org/github.com/keremgocen/tagpipe?status.svg)](https://godoc.org/github.com/keremgocen/tagpipe)
[![Build Status](https://travis-ci.org/keremgocen/tagpipe.svg?branch=master)](https://travis-ci.org/keremgocen/tagpipe)

Install
-----
```
go get github.com/keremgocen/tagpipe
```

Usage
-----

![alt tag](/docs/out.gif)

Tags can be optionally passed as command line arguments. If no tags are found, a local file "tags.txt" will be used.
You can try the example program as shown below, although the 2 usage options can't be mixed;

###### Option 1

```
./run tag1 tag2 ..
```

###### Option 2

Another running option is to provide custom path for `tags.txt` file or data folder
which contains JSON files to be parsed. Using `-c=false` also ignores if a cache is present, overriding the new results.

```
$./run --help
Usage of ./run:
  -c	-c=false overrides cache (default true)
  -d string
    	data folder path is not set, use: --d <path-to-files> (default "../data/")
  -t string
    	tags file path is not set, use: --t <path-to-tags> (default "../tags.txt")
```

Corner Cases
-----
- If a cache file is present after a subsequent run on identical JSON files, changing the tags used will output the same result. Manually removing the cache file works around this.
- Usage options can't be mixed. Tags provided via command line won't be parsed if any of the flags are present, and vice versa.
- For the case when a "tag" is used as key of a JSON object (as in {"tag":"value"}), count for that particular tag is incremented. After JSON structure
in the file is validated, a regex search is made over the whole content. This may not
be ideal for when tags are not expected to appear as keys.
