package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var path *string
var verbose *bool
var excludesPattern *regexp.Regexp

// Version holds the main version string which should be updated externally when building release
var Version = "undefined"
var showVersion *bool

type lineProcessorFunc func(string, string)

func init() {
	showVersion = flag.Bool("version", false, "Get application version")
	path = flag.String("path", ".", "Where is the cucumber groovy project?")
	verbose = flag.Bool("verbose", false, "Show verbose run information")
	excludeUsagePattern := flag.String("excludeUsagePattern", "", "Glob file pattern of which files should be avoided when scanning usages")
	flag.Parse()
	excludesPattern = regexp.MustCompile(*excludeUsagePattern)
}

func main() {
	if *showVersion {
		fmt.Printf("MissmatchingCucumberGroovyScenarioSteps version: %v\n", Version)
		return
	}

	definedSteps := make(map[string]*regexp.Regexp)
	if err := processFiles(isGroovy, createGroovyLineProcessor(definedSteps)); err != nil {
		log.Fatalf("Error while searching definition files: %v", err)
	}

	if *verbose {
		log.Printf("Unique step definitions found: %d", len(definedSteps))
		for step := range definedSteps {
			log.Println(step)
		}
	}

	foundUsages := make(map[string]string)
	if err := processFiles(isFeature, createFeatureLineProcessor(foundUsages)); err != nil {
		log.Fatalf("Error while searching usages: %v", err)
	}
	if *verbose {
		log.Printf("Unique usages found: %v", len(foundUsages))
	}

	failures := listUnmatched(foundUsages, definedSteps)
	if len(failures) > 0 {
		fmt.Println("Non-matched usages:")
		for _, failure := range failures {
			fmt.Println(failure)
		}
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func isGroovy(path string) bool {
	return endsWith(path, ".groovy")
}

func isFeature(path string) bool {
	return endsWith(path, ".feature")
}

func createFeatureLineProcessor(foundUsages map[string]string) lineProcessorFunc {
	usage := regexp.MustCompile(`(?:(?:^\s*Given)|(?:^\s*When)|(?:^\s*Then)|(?:^\s*And)|(?:^\s*But))(.*)$`)
	return func(path string, line string) {
		if usage.MatchString(line) {
			matches := usage.FindStringSubmatchIndex(line)
			matchedUsage := strings.TrimSpace(line[matches[1*2]:matches[1*2+1]])
			if excludesPattern.MatchString(path) {
				if *verbose {
					log.Printf("Excluding detected usage %v because of pattern (source file is %v)", matchedUsage, path)
				}
				return
			}
			foundUsages[matchedUsage] = path
		}
	}
}

func createGroovyLineProcessor(definedSteps map[string]*regexp.Regexp) lineProcessorFunc {
	definition := regexp.MustCompile(`(?:(?:^\s*Given)|(?:^\s*When)|(?:^\s*Then)|(?:^\s*And)|(?:^\s*But))[^/']+([/'])`)
	return func(path string, line string) {
		if definition.MatchString(line) {
			matches := definition.FindStringSubmatchIndex(line)
			stepChar := line[matches[1*2]:matches[1*2+1]]
			matchedDefinition := strings.TrimSpace(line[matches[1*2+1]:strings.LastIndex(line, stepChar)])
			if stepChar == "'" {
				matchedDefinition = strings.Replace(matchedDefinition, `\\`, `\`, -1)
			}
			definedSteps[matchedDefinition] = regexp.MustCompile(matchedDefinition)
		}
	}
}

func listUnmatched(foundUsages map[string]string, definedSteps map[string]*regexp.Regexp) (failures []string) {
	parametrizedUsage := regexp.MustCompile(`<[^>]+>`)
	failures = make([]string, 0)
outer:
	for usage, file := range foundUsages {
		for _, pattern := range definedSteps {
			if pattern.MatchString(usage) {
				continue outer
			}
		}
		if parametrizedUsage.MatchString(usage) {
			if *verbose {
				log.Printf("Excluding detected usage %v because it's parametrized (NYI in this tool)", usage)
			}
		} else {
			failures = append(failures, fmt.Sprintf("%v:%v", file, usage))
		}
	}
	return failures
}

func processFileFunc(filePath string, lineProcessor lineProcessorFunc) error {
	byteData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(byteData))
	for scanner.Scan() {
		text := scanner.Text()
		lineProcessor(filePath, text)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func processFiles(pathValid func(string) bool, lineProcessor lineProcessorFunc) error {
	visitedPath := make(map[string]bool)
	var walkFunc filepath.WalkFunc
	walkFunc = func(filePath string, info os.FileInfo, incomingErr error) error {
		if info == nil || incomingErr != nil {
			return incomingErr
		}
		if info.IsDir() && info.Name() != "." && !visitedPath[filePath] {
			visitedPath[filePath] = true
			if err := filepath.Walk(filePath, walkFunc); err != nil {
				return err
			}
		} else if pathValid(filePath) {
			if err := processFileFunc(filePath, lineProcessor); err != nil {
				return err
			}
		}
		return incomingErr
	}
	return filepath.Walk(*path, walkFunc)
}

func endsWith(s string, suffix string) bool {
	return strings.Index(s, suffix) == len(s)-len(suffix)
}
