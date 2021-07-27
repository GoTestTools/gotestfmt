package tokenizer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"
)

// Tokenize starts a reader of Event in the background that reads until the input is closed. This method starts a
// goroutine in the background and should be stopped by closing the input reader.
func Tokenize(input io.Reader) <-chan Event {
	output := make(chan Event)
	go decode(input, output)
	return output
}

type state string

const (
	stateInit         state = "init"
	stateRun          state = "run"
	stateBetweenTests state = "between_tests"
)

type stateChange struct {
	regexp     *regexp.Regexp
	inputState state
	action     Action
	newState   state
}

var stateMachine = []stateChange{
	{
		regexp.MustCompile(`^go: downloading (?P<Package>[^\s]+) (?P<Version>.*)$`),
		stateInit,
		ActionDownload,
		stateInit,
	},
	{
		regexp.MustCompile(`^go: (?P<Package>[^@]+)@(?P<Version>[^:]+): (?P<Output>.*)`),
		stateInit,
		ActionDownloadFailed,
		stateInit,
	},
	{
		regexp.MustCompile(`^# (?P<Package>.*)$`),
		stateInit,
		ActionPackage,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^=== RUN\s+(?P<Test>.*)$`),
		stateInit,
		ActionRun,
		stateRun,
	},
	{
		regexp.MustCompile(`^=== RUN\s+(?P<Test>.*)$`),
		stateBetweenTests,
		ActionRun,
		stateRun,
	},
	{
		regexp.MustCompile(`^=== RUN\s+(?P<Test>.*)$`),
		stateRun,
		ActionRun,
		stateRun,
	},
	{
		regexp.MustCompile(`^=== PAUSE\s+(?P<Test>.*)$`),
		stateRun,
		ActionPause,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^=== PAUSE\s+(?P<Test>.*)$`),
		stateBetweenTests,
		ActionPause,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^=== CONT\s+(?P<Test>.*)$`),
		stateBetweenTests,
		ActionCont,
		stateRun,
	},
	{
		regexp.MustCompile(`^=== CONT\s+(?P<Test>.*)$`),
		stateRun,
		ActionCont,
		stateRun,
	},
	{
		regexp.MustCompile(`^\s*--- FAIL:\s+(?P<Test>[^\s]+) \(((?P<Cached>cached)|(?P<Elapsed>[^\s]*))\)$`),
		stateRun,
		ActionFail,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\s*--- FAIL:\s+(?P<Test>[^\s]+) \(((?P<Cached>cached)|(?P<Elapsed>[^\s]*))\)$`),
		stateBetweenTests,
		ActionFail,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\s*--- PASS:\s+(?P<Test>[^\s]+) \(((?P<Cached>cached)|(?P<Elapsed>[^\s]*))\)$`),
		stateRun,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\s*--- PASS:\s+(?P<Test>[^\s]+) \(((?P<Cached>cached)|(?P<Elapsed>[^\s]*))\)$`),
		stateBetweenTests,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\s*--- SKIP:\s+(?P<Test>[^\s]+) \(((?P<Cached>cached)|(?P<Elapsed>[^\s]*))\)$`),
		stateRun,
		ActionSkip,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\s*--- SKIP:\s+(?P<Test>[^\s]+) \(((?P<Cached>cached)|(?P<Elapsed>[^\s]*))\)$`),
		stateBetweenTests,
		ActionSkip,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^ok\s+(?P<Package>[^\s]+)\s+(\((?P<Cached>cached)\)|(?P<Elapsed>[^\s]*))(|([\s]+)coverage: ((?P<Coverage>.*)% of statements|\[no statements]))$`),
		stateInit,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^ok\s+(?P<Package>[^\s]+)\s+(\((?P<Cached>cached)\)|(?P<Elapsed>[^\s]*))(|\s+coverage: ((?P<Coverage>.*)% of statements|\[no statements]))$`),
		stateBetweenTests,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\?\s+(?P<Package>[^\s]+)\s+(?P<Output>.*)$`),
		stateInit,
		ActionSkip,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\?\s+(?P<Package>[^\s]+)\s+(?P<Output>.*)$`),
		stateBetweenTests,
		ActionSkip,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^FAIL\s+(?P<Package>[^\s]+)\s+\((?P<Elapsed>[^\s]*)\)$`),
		stateInit,
		ActionFail,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^FAIL\s+(?P<Package>[^\s]+)\s+(?P<Elapsed>[^\s]*)$`),
		stateBetweenTests,
		ActionFail,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^FAIL\s+(?P<Package>[^\s]+)\s+\[(?P<Output>.*)\]$`),
		stateBetweenTests,
		ActionFail,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^PASS\s+(?P<Package>[^\s]+)\s+\(((?P<Elapsed>[0-9.smh]+)|(?P<Cached>cached))\)$`),
		stateInit,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^PASS\s+(?P<Package>[^\s]+)\s+\(((?P<Elapsed>[^\s]*)|(?P<Cached>cached))\)$`),
		stateBetweenTests,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^SKIP\s+(?P<Package>[^\s]+)\s+\(((?P<Elapsed>[^\s]*)|(?P<Cached>cached))\)$`),
		stateInit,
		ActionSkip,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^SKIP\s+(?P<Package>[^\s]+)\s+\(((?P<Elapsed>[^\s]*)|(?P<Cached>cached))\)$`),
		stateBetweenTests,
		ActionSkip,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^FAIL$`),
		stateInit,
		ActionFail,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^FAIL$`),
		stateBetweenTests,
		ActionFail,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^PASS$`),
		stateInit,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^PASS$`),
		stateBetweenTests,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^coverage: (?P<Coverage>.*)% of statements$`),
		stateBetweenTests,
		ActionCoverage,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^coverage: (?P<Coverage>.*)% of statements$`),
		stateRun,
		ActionCoverage,
		stateRun,
	},
	{
		regexp.MustCompile(`^coverage: \[no statements]$`),
		stateBetweenTests,
		ActionCoverageNoStatements,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^coverage: \[no statements]$`),
		stateRun,
		ActionCoverageNoStatements,
		stateRun,
	},
	{
		regexp.MustCompile(`^(?P<Output>.*)$`),
		stateInit,
		ActionStdout,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^(?P<Output>.*)$`),
		stateBetweenTests,
		ActionStdout,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^(?P<Output>.*)$`),
		stateRun,
		ActionStdout,
		stateRun,
	},
}

func decode(input io.Reader, output chan<- Event) {
	defer close(output)
	var lastBuffer []byte
	buffer := make([]byte, 4096)
	currentState := stateInit
	for {
		n, err := input.Read(buffer)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				panic(fmt.Errorf("failed to read from input (%w)", err))
			}
			break
		}
		if n == 0 {
			break
		}

		lines := bytes.Split(append(lastBuffer, buffer[:n]...), []byte("\n"))
		lastBuffer = lines[len(lines)-1]
		lines = lines[:len(lines)-1]
		for _, line := range lines {
			currentState = parseLine(currentState, line, output)
		}
	}
	_ = parseLine(currentState, lastBuffer, output)
}

func parseLine(currentState state, line []byte, output chan<- Event) state {
	for _, stateTransition := range stateMachine {
		if stateTransition.inputState != currentState {
			continue
		}

		if match := stateTransition.regexp.FindSubmatch(line); len(match) != 0 {
			elapsed, err := getTimeElapsed(stateTransition.regexp, match, "Elapsed")
			if err == nil {
				coverageString := string(extract(stateTransition.regexp, match, "Coverage"))
				coverage := 0.00
				if coverageString != "" {
					coverage, err = strconv.ParseFloat(coverageString, 64)
					if err != nil {
						continue
					}
				}

				evt := Event{
					Action:   stateTransition.action,
					Package:  string(extract(stateTransition.regexp, match, "Package")),
					Version:  string(extract(stateTransition.regexp, match, "Version")),
					Test:     string(extract(stateTransition.regexp, match, "Test")),
					Cached:   string(extract(stateTransition.regexp, match, "Cached")) == "cached",
					Coverage: coverage,
					Elapsed:  elapsed,
					Output:   extract(stateTransition.regexp, match, "Output"),
				}

				output <- evt
				return stateTransition.newState
			}
		}
	}
	if len(line) != 0 {
		panic(fmt.Errorf("failed to match line: %v", line))
	}
	return currentState
}

func getTimeElapsed(r *regexp.Regexp, match [][]byte, name string) (time.Duration, error) {
	val := extract(r, match, name)
	if val == nil {
		return 0, nil
	}
	return time.ParseDuration(string(val))
}

func extract(r *regexp.Regexp, match [][]byte, name string) []byte {
	idx := r.SubexpIndex(name)
	if idx < 0 {
		return nil
	}
	return match[idx]
}
