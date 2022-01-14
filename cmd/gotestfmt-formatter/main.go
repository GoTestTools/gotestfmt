package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

var goLogRegexp = regexp.MustCompile(`^\s+([^:]+):([0-9]+): (.*)$`)
var kLogRegexp = regexp.MustCompile(`^([IWE])([0-9]+)\s+([0-9:.]+)\s+([0-9]+)\s+([^:]+):([0-9]+)]\s+(.*)`)

// main is a demo formatter that showcases how to write a formatter for gotestfmt.
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if !first {
			fmt.Println()
		}
		first = false
		if goLogMatch := goLogRegexp.FindSubmatch([]byte(line)); len(goLogMatch) > 0 {
			fmt.Printf("    ⚙ %s:%s: %s", goLogMatch[1], goLogMatch[2], goLogMatch[3])
		} else if kLogMatch := kLogRegexp.FindSubmatch([]byte(line)); len(kLogMatch) > 0 {
			symbol := "⚙"
			switch string(kLogMatch[1]) {
			case "I":
			case "W":
				symbol = "⚠️"
			case "E":
				symbol = "❌"
			}
			fmt.Printf("    %s %s:%s: %s", symbol, kLogMatch[5], kLogMatch[6], kLogMatch[7])
		} else {
			fmt.Printf("    %s", line)
		}
	}
}
