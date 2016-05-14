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
Tags can be optionally passed as command line arguments. If no tags are found, a local file "tags.txt" will be used.
You can try the example program as shown below;

```
./example tag1 tag2 ..
```

Corner Cases
-----
- This program is not covering the case when "data" folder contains nested folders,
it assumes that "data" folder contains files only with text content
- For the case when a "tag" is used as key of a JSON object (as in {"tag":"value"}), count for that particular tag is incremented. After JSON structure
in the file is validated, a regex search is made over the whole content. This may not
be ideal for when tags are not expected to appear as keys.
