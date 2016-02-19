package main

import (
	"regexp"
	"fmt"
"strings"
)

func main() {
	var usage = "transfer is awaiting to be picked up by organization ibm"
	var definitionPattern = `^transfer is awaiting to be picked up(?: by organization with id (\\w+))?(?: by organization (\\w+))??(?: by known ONP organization (\\w+))?$`
	definitionPattern = strings.Replace(definitionPattern, `\\`, `\`, -1)
	var definition = regexp.MustCompile(definitionPattern)
	fmt.Printf("matches: %v", definition.MatchString(usage))
}