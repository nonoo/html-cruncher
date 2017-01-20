package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// HTMLLoad loads all tags from the given file.
func HTMLLoad(wg *sync.WaitGroup, fileName string) {
	defer wg.Done()

	fmt.Println("html: loading tags from " + fileName)

	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println("html error: can't open " + fileName + ": " + err.Error())
		return
	}
	defer f.Close()

	tokenizer := html.NewTokenizer(bufio.NewReader(f))

	// Iterating through all HTML tokens until EOF.
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				fmt.Println("html: loading tags from " + fileName + " finished")
				return
			}
		}

		token := tokenizer.Token()
		// Iterating through all token attributes.
		for _, attr := range token.Attr {
			switch attr.Key {
			case "id":
				TagAdd(TagTypeID, fileName, attr.Val, false)
			case "class":
				classes := strings.Split(attr.Val, " ")
				for _, class := range classes {
					TagAdd(TagTypeClass, fileName, class, false)
				}
			}
		}
	}
}

// HTMLReplaceTags replaces all known tags to their new names.
func HTMLReplaceTags(wg *sync.WaitGroup, fileName string, outDirName string) {
	defer wg.Done()

	fmt.Println("html: replacing tags in " + fileName)

	fi, err := os.Open(fileName)
	if err != nil {
		fmt.Println("html error: can't open " + fileName + ": " + err.Error())
		return
	}
	defer fi.Close()

	err = os.MkdirAll(filepath.Dir(outDirName+"/"+fileName), 0777)
	if err != nil {
		fmt.Println("html error: can't create output directory " + outDirName +
			": " + err.Error())
		return
	}

	fo, err := os.Create(outDirName + "/" + fileName)
	if err != nil {
		fmt.Println("html error: can't write " + outDirName + "/" +
			fileName + ": " + err.Error())
		return
	}
	defer fo.Close()

	tokenizer := html.NewTokenizer(bufio.NewReader(fi))

	// Iterating through all HTML tokens until EOF.
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
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

				fmt.Printf("html: replacing tags in "+fileName+" finished (%d -> %d, %.1f%%)\n",
					bytesIn, bytesOut, (float64(bytesOut)/float64(bytesIn))*100)
				return
			}
		}

		token := tokenizer.Token()
		// Iterating through all token attributes.
		for nr, attr := range token.Attr {
			switch attr.Key {
			case "id":
				tag, err := TagGet(TagTypeID, attr.Val)
				if err == nil {
					token.Attr[nr].Val = tag.NewName
				}
			case "class":
				classes := strings.Split(attr.Val, " ")
				for _, class := range classes {
					tag, err := TagGet(TagTypeClass, class)
					if err == nil {
						token.Attr[nr].Val = strings.Replace(token.Attr[nr].Val,
							class, tag.NewName, 1)
					}
				}
			}
		}

		fo.WriteString(token.String())
	}
}
