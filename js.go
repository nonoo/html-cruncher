package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// JSLoad loads all tags from the given file.
func JSLoad(wg *sync.WaitGroup, fileName string) {
	defer wg.Done()

	fmt.Println("js: loading tags from " + fileName)

	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("js error: can't open " + fileName + ": " + err.Error())
		return
	}

	r := regexp.MustCompile("(getElementById|getElementsByClassName|hasClass|addClass|removeClass|querySelector|\\$)\\((\"[^\"]*\"|'[^']*')\\)")
	matches := r.FindAllStringSubmatch(string(fileData), -1)

	for _, s := range matches {
		trimmedMatch := strings.Trim(s[2], "'\"")

		switch s[1] {
		case "getElementById":
			TagAdd(TagTypeID, fileName, trimmedMatch, true)
		case "getElementsByClassName":
			TagAdd(TagTypeClass, fileName, trimmedMatch, true)
		case "hasClass":
			TagAdd(TagTypeClass, fileName, trimmedMatch, true)
		case "addClass":
			TagAdd(TagTypeClass, fileName, trimmedMatch, true)
		case "removeClass":
			TagAdd(TagTypeClass, fileName, trimmedMatch, true)
		case "querySelector":
			CSSProcess(fileName, trimmedMatch)
		case "$":
			CSSProcess(fileName, trimmedMatch)
		}
	}
}

// JSReplaceTags replaces all known tags to their new names.
func JSReplaceTags(wg *sync.WaitGroup, fileName string, outDirName string) {
	defer wg.Done()

	fmt.Println("js: replacing tags in " + fileName)

	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("js error: can't open " + fileName + ": " + err.Error())
		return
	}

	err = os.MkdirAll(filepath.Dir(outDirName+"/"+fileName), 0777)
	if err != nil {
		fmt.Println("js error: can't create output directory " + outDirName +
			": " + err.Error())
		return
	}

	fo, err := os.Create(outDirName + "/" + fileName)
	if err != nil {
		fmt.Println("js error: can't write " + outDirName + "/" +
			fileName + ": " + err.Error())
		return
	}
	defer fo.Close()

	r := regexp.MustCompile("(getElementById|getElementsByClassName|hasClass|addClass|removeClass|querySelector|\\$)\\((\"[^\"]*\"|'[^']*')\\)")
	fo.WriteString(r.ReplaceAllStringFunc(string(fileData), func(str string) string {
		match := r.FindStringSubmatch(str)

		trimmedMatch := strings.Trim(match[2], "'\"")

		switch match[1] {
		case "getElementById":
			tag, err := TagGet(TagTypeID, trimmedMatch)
			if err == nil {
				return strings.Replace(str, trimmedMatch, tag.NewName, 1)
			}
		case "getElementsByClassName":
			tag, err := TagGet(TagTypeClass, trimmedMatch)
			if err == nil {
				return strings.Replace(str, trimmedMatch, tag.NewName, 1)
			}
		case "hasClass":
			tag, err := TagGet(TagTypeClass, trimmedMatch)
			if err == nil {
				return strings.Replace(str, trimmedMatch, tag.NewName, 1)
			}
		case "addClass":
			tag, err := TagGet(TagTypeClass, trimmedMatch)
			if err == nil {
				return strings.Replace(str, trimmedMatch, tag.NewName, 1)
			}
		case "removeClass":
			tag, err := TagGet(TagTypeClass, trimmedMatch)
			if err == nil {
				return strings.Replace(str, trimmedMatch, tag.NewName, 1)
			}
		case "querySelector":
			var buf bytes.Buffer
			err := CSSReplaceFromString(trimmedMatch, &buf, fileName)
			if err == nil {
				return strings.Replace(str, trimmedMatch, string(buf.Bytes()), 1)
			}
			fmt.Print(err.Error())
			return str
		case "$":
			var buf bytes.Buffer
			err := CSSReplaceFromString(trimmedMatch, &buf, fileName)
			if err == nil {
				return strings.Replace(str, trimmedMatch, string(buf.Bytes()), 1)
			}
			fmt.Print(err.Error())
			return str
		}
		return str
	}))

	var bytesIn int64
	var bytesOut int64

	fi, e := os.Stat(fileName)
	if e == nil {
		bytesIn = fi.Size()
	}

	fi, e = fo.Stat()
	if e == nil {
		bytesOut = fi.Size()
	}

	fmt.Printf("js: replacing tags in "+fileName+" finished (%d -> %d, %.1f%%)\n",
		bytesIn, bytesOut, (float64(bytesOut)/float64(bytesIn))*100)
}
