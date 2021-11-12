package tokenizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// Event is an event that occurred during output.
type Event struct {
	// Received is the time this event was seen.
	Received time.Time `json:"-"`
	// Action is the action that happened in this event.
	Action Action `json:"action"`
	// Package, if present, specifies the package being tested.
	Package string `json:"package"`
	// Version contains the downloaded package version for download events.
	Version string `json:"version"`
	// Test, if present, specifies the test, example, or benchmark function that caused the event.
	Test string `json:"test"`
	// Elapsed gives the total elapsed time for a test.
	Elapsed time.Duration `json:"elapsed"`
	// Output contains the output written to the standard output or standard error, depending on the Action field.
	Output []byte `json:"output"`
	// Cached indicates that the results are cached.
	Cached bool `json:"cached"`
	// Coverage shows the code coverage.
	Coverage *float64 `json:"coverage"`
}

func (e Event) Equals(o Event) bool {
	return e.Action == o.Action &&
		e.Package == o.Package &&
		e.Version == o.Version &&
		e.Test == o.Test &&
		e.Elapsed == o.Elapsed &&
		e.Cached == o.Cached &&
		(e.Coverage == o.Coverage ||
			(e.Coverage != nil && o.Coverage != nil && *e.Coverage == *o.Coverage)) &&
		bytes.Equal(e.Output, o.Output)
}

func (e Event) String() string {
	data, err := e.MarshalJSON()
	if err != nil {
		panic(err)
	}
	return string(data)
}

type tmpEvent struct {
	// Action is the action that happened in this event.
	Action Action `json:"action"`
	// Package, if present, specifies the package being tested.
	Package string `json:"package"`
	// Version contains the downloaded package version for download events.
	Version string `json:"version"`
	// Test, if present, specifies the test, example, or benchmark function that caused the event.
	Test string `json:"test"`
	// Elapsed gives the total elapsed time for a test.
	Elapsed string `json:"elapsed"`
	// Output contains the output written to the standard output or standard error, depending on the Action field.
	Output []byte `json:"output"`
	// Cached indicates that the results are cached.
	Cached bool `json:"cached"`
	// Coverage shows the code coverage.
	Coverage *float64 `json:"coverage"`
}

func (e *Event) UnmarshalJSON(data []byte) error {
	tmp := tmpEvent{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	var elapsed time.Duration
	var err error
	if tmp.Elapsed == "" {
		elapsed = 0
	} else {
		elapsed, err = time.ParseDuration(tmp.Elapsed)
		if err != nil {
			return fmt.Errorf("failed to parse elapsed: %s (%w)", tmp.Elapsed, err)
		}
	}
	e.Action = tmp.Action
	e.Package = tmp.Package
	e.Version = tmp.Version
	e.Test = tmp.Test
	e.Elapsed = elapsed
	e.Output = tmp.Output
	e.Cached = tmp.Cached
	e.Coverage = tmp.Coverage
	return nil
}

func (e *Event) MarshalJSON() ([]byte, error) {
	tmp := tmpEvent{
		Action:   e.Action,
		Package:  e.Package,
		Version:  e.Version,
		Test:     e.Test,
		Elapsed:  e.Elapsed.String(),
		Output:   e.Output,
		Cached:   e.Cached,
		Coverage: e.Coverage,
	}
	return json.Marshal(tmp)
}
