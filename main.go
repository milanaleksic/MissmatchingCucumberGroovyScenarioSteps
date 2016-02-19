package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"io/ioutil"
	"bufio"
	"bytes"
	"log"
	"regexp"
)

var path *string
var visited map[string]bool
var definition *regexp.Regexp

var definedSteps map[string]bool

func init() {
	path = flag.String("path", ".", "Where is the cucumber groovy project?")
	definition = regexp.MustCompile(`(?:(?:^\s*Given)|(?:^\s*When)|(?:^\s*Then)|(?:^\s*And)|(?:^\s*But))[^/']+([/'])`)
	flag.Parse()
}

func main() {
	visited = make(map[string]bool)
	definedSteps = make(map[string]bool)
	err := filepath.Walk(*path, visitor)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func visitor(path string, info os.FileInfo, err error) error {
	if info == nil {
		return err
	}
	if info.IsDir() && info.Name() != "." && !visited[path] {
		visited[path] = true
		filepath.Walk(path, visitor)
	} else if strings.Index(path, ".groovy") == len(path) - 7 {
		byteData, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(bytes.NewReader(byteData))
		for scanner.Scan() {
			text := scanner.Text()
			if definition.MatchString(text) {
				matches := definition.FindStringSubmatchIndex(text)
				stepChar := text[matches[1*2]:matches[1*2+1]]
				matchedDefinition := text[matches[1*2+1]:strings.LastIndex(text, stepChar)]
				definedSteps[matchedDefinition] = true
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}
	return err
}
