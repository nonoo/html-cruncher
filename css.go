package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/css/scanner"
)

// CSSProcess loads all tags from the given CSS string.
func CSSProcess(fileName string, css string) {
	s := scanner.New(css)

	// Iterating through all CSS tokens until EOF.
	var prevTokenType string
	var prevTokenValue string
	for {
		token := s.Next()

		switch token.Type {
		case scanner.TokenEOF:
			return
		case scanner.TokenError:
			fmt.Printf("css error in %s: %s\n", fileName, token.String())
			return
		case scanner.TokenHash:
			TagAdd(TagTypeID, fileName, token.Value[1:], true)
		case scanner.TokenIdent:
			if prevTokenType == "CHAR" && prevTokenValue == "." {
				TagAdd(TagTypeClass, fileName, token.Value, true)
			}
		}

		prevTokenType = token.Type.String()
		prevTokenValue = token.Value
	}
}

// CSSLoad loads all tags from the given file.
func CSSLoad(wg *sync.WaitGroup, fileName string) {
	defer wg.Done()

	fmt.Println("css: loading tags from " + fileName)

	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("css error: can't open " + fileName + ": " + err.Error())
		return
	}

	CSSProcess(fileName, string(fileData))
	fmt.Println("css: loading tags from " + fileName + " finished")
}

// CSSReplaceFromString processes CSS input and writes data with replaced
// tags to outFile.
func CSSReplaceFromString(css string, out *bytes.Buffer, fileName string) error {
	s := scanner.New(css)

	// Iterating through all CSS tokens until EOF.
	var prevTokenType string
	var prevTokenValue string
	for {
		token := s.Next()

		switch token.Type {
		case scanner.TokenEOF:
			return nil
		case scanner.TokenError:
			return fmt.Errorf("css error in %s: %s\n",
				fileName, token.String())
		case scanner.TokenHash:
			tag, err := TagGet(TagTypeID, token.Value[1:])
			if err == nil {
				token.Value = "#" + tag.NewName
			}
		case scanner.TokenIdent:
			if prevTokenType == "CHAR" && prevTokenValue == "." {
				tag, err := TagGet(TagTypeClass, token.Value)
				if err == nil {
					token.Value = tag.NewName
				}
			}
		}

		out.WriteString(token.Value)

		prevTokenType = token.Type.String()
		prevTokenValue = token.Value
	}
}

// CSSReplaceTags replaces all known tags to their new names.
func CSSReplaceTags(wg *sync.WaitGroup, fileName string, outDirName string) {
	defer wg.Done()

	fmt.Println("css: replacing tags in " + fileName)

	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("css error: can't open " + fileName + ": " + err.Error())
		return
	}

	err = os.MkdirAll(filepath.Dir(outDirName+"/"+fileName), 0777)
	if err != nil {
		fmt.Println("css error: can't create output directory " + outDirName +
			": " + err.Error())
		return
	}

	fo, err := os.Create(outDirName + "/" + fileName)
	if err != nil {
		fmt.Println("css error: can't write " + outDirName + "/" +
			fileName + ": " + err.Error())
		return
	}
	defer fo.Close()

	var buf bytes.Buffer
	err = CSSReplaceFromString(string(fileData), &buf, fileName)
	if err == nil {
		buf.WriteTo(fo)
	} else {
		fmt.Print(err.Error())
		fo.Write(fileData)
	}

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

	fmt.Printf("css: replacing tags in "+fileName+" finished (%d -> %d, %.1f%%)\n",
		bytesIn, bytesOut, (float64(bytesOut)/float64(bytesIn))*100)
	return
}
