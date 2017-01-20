package main

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Tag types.
const (
	TagTypeID = iota
	TagTypeClass
	TagTypeName
)

// TagType describes the type of the tag.
type TagType int

// Tag describes a HTML/CSS/JS tag.
type Tag struct {
	Type               TagType
	Name               string
	NewName            string
	FileNameFirstFound string
	Weight             int
	OnlyExistInOneFile bool
}
type tagsArray []Tag

var tagsMutex = &sync.Mutex{}
var tags tagsArray

func (t TagType) String() string {
	switch t {
	case TagTypeID:
		return "id"
	case TagTypeClass:
		return "class"
	case TagTypeName:
		return "name"
	}
	return "unknown"
}

// TagGetIndexByName searches for the given tag in tags and returns it's array
// position, if found. If not, returns -1.
func tagGetIndexByName(tagType TagType, name string) int {
	for nr, tag := range tags {
		if tag.Type == tagType && tag.Name == name {
			return nr
		}
	}
	return -1
}

// TagGet returns the given tag from the tags array.
func TagGet(tagType TagType, name string) (Tag, error) {
	tagsMutex.Lock()
	defer tagsMutex.Unlock()
	nr := tagGetIndexByName(tagType, name)
	if nr < 0 {
		return Tag{}, errors.New("not found")
	}
	return tags[nr], nil
}

// TagAdd adds the given tag to the tags array.
// onlyIfExists should be set to true if we only want to increase weight of
// existing tags.
func TagAdd(tagType TagType, fileName string, name string, onlyIfExists bool) {
	tagsMutex.Lock()

	nr := tagGetIndexByName(tagType, name)
	if nr == -1 {
		if !onlyIfExists {
			// fmt.Println("tag: adding " + tagType.String() + ": " + name)
			tags = append(tags, Tag{Type: tagType, Name: name,
				FileNameFirstFound: fileName, Weight: len(name),
				OnlyExistInOneFile: true})
		}
	} else {
		// fmt.Println("tag: increasing weight of " + tagType.String() + ": " + name)
		tags[nr].Weight += len(name)
		if tags[nr].FileNameFirstFound != fileName {
			tags[nr].OnlyExistInOneFile = false
		}
	}

	tagsMutex.Unlock()
}

func tagDelete(s tagsArray, i int) tagsArray {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

// TagDropNoCommon checks all tags and drops the ones which don't have
// reference in other files.
func TagDropNoCommon() {
	tagsMutex.Lock()

	for nr, tag := range tags {
		if tag.OnlyExistInOneFile {
			// fmt.Println("tag: no reference in other files, dropping " + tag.Name)
			tags = tagDelete(tags, nr)
			tagsMutex.Unlock()
			TagDropNoCommon()
			return
		}
	}

	tagsMutex.Unlock()
}

func (s tagsArray) Len() int {
	return len(s)
}
func (s tagsArray) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s tagsArray) Less(i, j int) bool {
	return s[i].Weight > s[j].Weight
}

// TagSortByWeight sorts given tags by their weights.
func TagSortByWeight() {
	tagsMutex.Lock()

	sort.Sort(tags)

	tagsMutex.Unlock()
}

// TagPrint prints the contents of the tags array.
func TagPrint() {
	tagsMutex.Lock()

	fmt.Println("tags:")
	for nr, tag := range tags {
		fmt.Printf("#%d (%d): %s -> %s\n", nr, tag.Weight, tag.Name, tag.NewName)
	}

	tagsMutex.Unlock()
}

// TagGiveNewNames associates new short names to tags.
func TagGiveNewNames() {
	tagsMutex.Lock()

	var namePos int
	var newNameFirstChars = "abcdefghijklmnopqrstuvwxyz"
	var newNameOtherChars = newNameFirstChars + "0123456789"
	charsPos := make([]int, 10)

	for nr := range tags {
		// Filling up new name to the current name position.
		for i := 0; i < namePos; i++ {
			if namePos == 0 {
				tags[nr].NewName = string(newNameFirstChars[charsPos[0]])
			} else {
				tags[nr].NewName += string(newNameOtherChars[charsPos[i]])
			}
		}

		if namePos == 0 {
			tags[nr].NewName = string(newNameFirstChars[charsPos[0]])
			charsPos[0]++
			if charsPos[0] == len(newNameFirstChars) {
				charsPos[0] = 0
				namePos++
				if namePos >= len(charsPos) {
					var newCharsPos int
					charsPos = append(charsPos, newCharsPos)
				}
			}
		} else {
			tags[nr].NewName += string(newNameOtherChars[charsPos[namePos]])
			charsPos[namePos]++
			if charsPos[namePos] == len(newNameOtherChars) {
				if namePos == 1 {
					if charsPos[0] == len(newNameFirstChars)-1 {
						namePos++
					} else {
						charsPos[namePos-1]++
					}
				} else {
					if charsPos[namePos-1] == len(newNameOtherChars)-1 {
						namePos++
					} else {
						charsPos[namePos-1]++
					}
				}
				if namePos >= len(charsPos) {
					var newCharsPos int
					charsPos = append(charsPos, newCharsPos)
				}
				charsPos[namePos] = 0
			}
		}
	}

	tagsMutex.Unlock()
}
