package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func scanDir(scanDirName string, htmlFileNamesArray *[]string, cssFileNamesArray *[]string,
	jsFileNamesArray *[]string) {

	if scanDirName == "" {
		return
	}

	files, err := ioutil.ReadDir(scanDirName)
	if err != nil {
		fmt.Print(err)
	}

	for _, f := range files {
		if f.IsDir() {
			scanDir(scanDirName+"/"+f.Name(), htmlFileNamesArray, cssFileNamesArray,
				jsFileNamesArray)
		} else {
			switch filepath.Ext(f.Name()) {
			case ".html":
				*htmlFileNamesArray = append(*htmlFileNamesArray, scanDirName+"/"+f.Name())
			case ".css":
				*cssFileNamesArray = append(*cssFileNamesArray, scanDirName+"/"+f.Name())
			case ".js":
				*jsFileNamesArray = append(*jsFileNamesArray, scanDirName+"/"+f.Name())
			}
		}
	}
}

func main() {
	var htmlFileNames string
	var cssFileNames string
	var jsFileNames string
	var outDirName string
	var scanDirName string
	var renameOnlyCommonTags bool

	flag.StringVar(&htmlFileNames, "html", "", "HTML files, separated by commas")
	flag.StringVar(&cssFileNames, "css", "", "CSS files, separated by commas")
	flag.StringVar(&jsFileNames, "js", "", "JS files, separated by commas")
	flag.StringVar(&outDirName, "outdir", "out", "Output directory")
	flag.StringVar(&scanDirName, "scandir", "", "Recursive scan directory for HTML/CSS/JS files")
	flag.BoolVar(&renameOnlyCommonTags, "rc", true, "Rename only tags which are present both in HTML and other files")
	flag.Parse()

	var htmlFileNamesArray []string
	var cssFileNamesArray []string
	var jsFileNamesArray []string

	if htmlFileNames != "" {
		htmlFileNamesArray = append(htmlFileNamesArray, strings.Split(htmlFileNames, ",")...)
	}
	if cssFileNames != "" {
		cssFileNamesArray = append(cssFileNamesArray, strings.Split(cssFileNames, ",")...)
	}
	if jsFileNames != "" {
		jsFileNamesArray = append(jsFileNamesArray, strings.Split(jsFileNames, ",")...)
	}

	scanDir(scanDirName, &htmlFileNamesArray, &cssFileNamesArray, &jsFileNamesArray)

	if len(htmlFileNamesArray) == 0 {
		fmt.Println("error: no html files given")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	for _, fileName := range htmlFileNamesArray {
		wg.Add(1)
		go HTMLLoad(&wg, fileName)
	}
	// Need to wait here because CSS and JS tags are only added if they have
	// been previously added by the HTML loader.
	wg.Wait()
	for _, fileName := range cssFileNamesArray {
		wg.Add(1)
		go CSSLoad(&wg, fileName)
	}
	for _, fileName := range jsFileNamesArray {
		wg.Add(1)
		go JSLoad(&wg, fileName)
	}
	wg.Wait()

	if renameOnlyCommonTags {
		TagDropNoCommon()
	}
	TagSortByWeight()
	TagGiveNewNames()
	TagPrint()

	for _, fileName := range htmlFileNamesArray {
		wg.Add(1)
		go HTMLReplaceTags(&wg, fileName, outDirName)
	}
	for _, fileName := range cssFileNamesArray {
		wg.Add(1)
		go CSSReplaceTags(&wg, fileName, outDirName)
	}
	for _, fileName := range jsFileNamesArray {
		wg.Add(1)
		go JSReplaceTags(&wg, fileName, outDirName)
	}
	wg.Wait()

	var bytesIn int64
	var bytesOut int64

	for _, fileName := range htmlFileNamesArray {
		fi, e := os.Stat(fileName)
		if e == nil {
			bytesIn += fi.Size()
		}
		fi, e = os.Stat(outDirName + "/" + fileName)
		if e == nil {
			bytesOut += fi.Size()
		}
	}
	for _, fileName := range cssFileNamesArray {
		fi, e := os.Stat(fileName)
		if e == nil {
			bytesIn += fi.Size()
		}
		fi, e = os.Stat(outDirName + "/" + fileName)
		if e == nil {
			bytesOut += fi.Size()
		}
	}
	for _, fileName := range jsFileNamesArray {
		fi, e := os.Stat(fileName)
		if e == nil {
			bytesIn += fi.Size()
		}
		fi, e = os.Stat(outDirName + "/" + fileName)
		if e == nil {
			bytesOut += fi.Size()
		}
	}

	fmt.Printf("html-cruncher: total bytes in: %d out: %d saved: %d (%.1f%%)\n", bytesIn,
		bytesOut, bytesIn-bytesOut, (float64(bytesOut)/float64(bytesIn))*100)
}
