package gotest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"
)

// NewEventReader starts a reader of Event in the background that reads until the input is closed. This method starts a
// goroutine in the background and should be stopped by closing the input reader.
func NewEventReader(input io.Reader) <-chan Event {
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
		regexp.MustCompile(`^\s*--- FAIL:\s+(?P<Test>[^\s]+) \((?P<Elapsed>.*)\)$`),
		stateRun,
		ActionFail,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\s*--- FAIL:\s+(?P<Test>[^\s]+) \((?P<Elapsed>.*)\)$`),
		stateBetweenTests,
		ActionFail,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\s*--- PASS:\s+(?P<Test>[^\s]+) \((?P<Elapsed>.*)\)$`),
		stateRun,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\s*--- PASS:\s+(?P<Test>[^\s]+) \((?P<Elapsed>.*)\)$`),
		stateBetweenTests,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\s*--- SKIP:\s+(?P<Test>[^\s]+) \((?P<Elapsed>.*)\)$`),
		stateRun,
		ActionSkip,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^\s*--- SKIP:\s+(?P<Test>[^\s]+) \((?P<Elapsed>.*)\)$`),
		stateBetweenTests,
		ActionSkip,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^ok\s+(?P<Package>[^\s]+)\s+(?P<Elapsed>.*)$`),
		stateInit,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^ok\s+(?P<Package>[^\s]+)\s+(?P<Elapsed>.*)$`),
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
		regexp.MustCompile(`^FAIL\s+(?P<Package>[^\s]+)\s+\((?P<Elapsed>.*)\)$`),
		stateInit,
		ActionFail,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^FAIL\s+(?P<Package>[^\s]+)\s+(?P<Elapsed>.*)$`),
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
		regexp.MustCompile(`^PASS\s+(?P<Package>[^\s]+)\s+\((?P<Elapsed>.*)\)$`),
		stateInit,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^PASS\s+(?P<Package>[^\s]+)\s+\((?P<Elapsed>.*)\)$`),
		stateBetweenTests,
		ActionPass,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^SKIP\s+(?P<Package>[^\s]+)\s+\((?P<Elapsed>.*)\)$`),
		stateInit,
		ActionSkip,
		stateBetweenTests,
	},
	{
		regexp.MustCompile(`^SKIP\s+(?P<Package>[^\s]+)\s+\((?P<Elapsed>.*)\)$`),
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
	close(output)
}

func parseLine(currentState state, line []byte, output chan<- Event) state {
	for _, stateTransition := range stateMachine {
		if stateTransition.inputState != currentState {
			continue
		}

		if match := stateTransition.regexp.FindSubmatch(line); len(match) != 0 {
			elapsed, err := getTimeElapsed(stateTransition.regexp, match, "Elapsed")
			if err == nil {
				evt := Event{
					Action:  stateTransition.action,
					Package: string(extract(stateTransition.regexp, match, "Package")),
					Version: string(extract(stateTransition.regexp, match, "Version")),
					Test:    string(extract(stateTransition.regexp, match, "Test")),
					Elapsed: elapsed,
					Output:  extract(stateTransition.regexp, match, "Output"),
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
