package main

import (
	"bufio"
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var path *string
var definition *regexp.Regexp
var usage *regexp.Regexp
var excludesPattern *regexp.Regexp

var definedSteps map[string]*regexp.Regexp
var foundUsages map[string]string

func init() {
	path = flag.String("path", ".", "Where is the cucumber groovy project?")
	excludeUsagePattern := flag.String("excludeUsagePattern", "", "Glob file pattern of which files should be avoided when scanning usages")
	definition = regexp.MustCompile(`(?:(?:^\s*Given)|(?:^\s*When)|(?:^\s*Then)|(?:^\s*And)|(?:^\s*But))[^/']+([/'])`)
	usage = regexp.MustCompile(`(?:(?:^\s*Given)|(?:^\s*When)|(?:^\s*Then)|(?:^\s*And)|(?:^\s*But))(.*)$`)
	flag.Parse()
	excludesPattern = regexp.MustCompile(*excludeUsagePattern)
}

func main() {

	definedSteps = make(map[string]*regexp.Regexp)

	visitedDefinitionFile = make(map[string]bool)
	err := filepath.Walk(*path, visitDefinitionFiles)
	if err != nil {
		log.Fatalf("Error while searching definition files: %v", err)
	}
	log.Printf("Unique step definitions found: %d", len(definedSteps))
	for step := range definedSteps {
		log.Println(step)
	}

	visitedUsageFile = make(map[string]bool)
	foundUsages = make(map[string]string)
	err = filepath.Walk(*path, visitUsages)
	if err != nil {
		log.Fatalf("Error while searching usages: %v", err)
	}
	log.Printf("Unique usages found: %v", len(foundUsages))

	log.Println("Non-matched usages:")
	outer:
	for usage, file := range foundUsages {
		for _, pattern := range definedSteps {
			if pattern.MatchString(usage) {
				continue outer
			}
		}
		log.Printf("%v:%v", file, usage)
	}
}

func endsWith(s string, suffix string) bool {
	return strings.Index(s, suffix) == len(s) - len(suffix)
}

var visitedDefinitionFile map[string]bool

func visitDefinitionFiles(path string, info os.FileInfo, err error) error {
	if info == nil {
		return err
	}
	if info.IsDir() && info.Name() != "." && !visitedDefinitionFile[path] {
		visitedDefinitionFile[path] = true
		filepath.Walk(path, visitDefinitionFiles)
	} else if endsWith(path, ".groovy") {
		byteData, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(bytes.NewReader(byteData))
		for scanner.Scan() {
			text := scanner.Text()
			if definition.MatchString(text) {
				matches := definition.FindStringSubmatchIndex(text)
				stepChar := text[matches[1 * 2]:matches[1 * 2 + 1]]
				matchedDefinition := strings.TrimSpace(text[matches[1 * 2 + 1]:strings.LastIndex(text, stepChar)])
				definedSteps[matchedDefinition] = regexp.MustCompile(matchedDefinition)
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}
	return err
}

var visitedUsageFile map[string]bool

func visitUsages(path string, info os.FileInfo, err error) error {
	//TODO: use closure here to simplify
	//TODO: map can be created here as well instead of being global
	if info == nil {
		return err
	}
	if info.IsDir() && info.Name() != "." && !visitedUsageFile[path] {
		visitedUsageFile[path] = true
		filepath.Walk(path, visitUsages)
	} else if endsWith(path, ".feature") {
		byteData, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(bytes.NewReader(byteData))
		for scanner.Scan() {
			text := scanner.Text()
			if usage.MatchString(text) {
				matches := usage.FindStringSubmatchIndex(text)
				matchedUsage := strings.TrimSpace(text[matches[1 * 2]:matches[1 * 2 + 1]])
				if excludesPattern.String() == "" || !excludesPattern.MatchString(path) {
					foundUsages[matchedUsage] = path
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}
	return err
}