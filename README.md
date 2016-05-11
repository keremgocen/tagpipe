'A simple JSON tag parser'
=====

Install
-----
```
go get github.com/keremgocen/tagparser
```

Usage
-----
Inside the project root folder;

```
./tagparser tag1 tag2 ..
```

Corner Cases
-----
- This program is not covering the case when "data" folder contains nested folders,
it assumes that "data" folder contains files only with text content
- For the case when a "tag" is used as key of a JSON object (as in {"tag":"value"}), count for that particular tag is incremented. After JSON structure
in the file is validated, a regex search is made over the whole content. This may not
be ideal for when tags are not expected to appear as keys.
