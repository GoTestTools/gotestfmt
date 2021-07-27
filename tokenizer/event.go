package tokenizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Event is an event that occurred during output
type Event struct {
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
	Output []byte `json:"stdout"`
	// Cached indicates that the results are cached.
	Cached bool `json:"cached"`
	// Coverage shows the code coverage.
	Coverage float64 `json:"coverage"`
}

func (e Event) Equals(o Event) bool {
	return e.Action == o.Action &&
		e.Package == o.Package &&
		e.Version == o.Version &&
		e.Test == o.Test &&
		e.Elapsed == o.Elapsed &&
		bytes.Equal(e.Output, o.Output)
}

func (e Event) Diff(o Event) string {
	if e.Equals(o) {
		return ""
	}

	var diff strings.Builder
	diff.WriteString(`--- expected
+++ actual
@@ -1,8 +1,8 @@
{
`)
	e.diff("action", e.Action, o.Action, &diff, ",")
	e.diff("test", e.Test, o.Test, &diff, ",")
	e.diff("elapsed", e.Elapsed, o.Elapsed, &diff, ",")
	e.diff("package", e.Package, o.Package, &diff, ",")
	e.diff("version", e.Version, o.Version, &diff, ",")
	e.diff("cached", e.Cached, e.Cached, &diff, ",")
	e.diff("coverage", e.Coverage, o.Coverage, &diff, ",")
	e.diff("output", string(e.Output), string(o.Output), &diff, "")

	diff.WriteString("}\n")
	return diff.String()
}

func (e Event) diff(name string, expected interface{}, actual interface{}, diff *strings.Builder, suffix string) {
	expectedStr, err := json.Marshal(expected)
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s: %v (%w)", name, expected, err))
	}
	actualStr, err := json.Marshal(actual)
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s: %v (%w)", name, actual, err))
	}
	if bytes.Equal(expectedStr, actualStr) {
		diff.WriteString(fmt.Sprintf("   \"%s\": %s%s\n", name, expectedStr, suffix))
	} else {
		diff.WriteString(fmt.Sprintf("-  \"%s\": %s%s\n", name, expectedStr, suffix))
		diff.WriteString(fmt.Sprintf("+  \"%s\": %s%s\n", name, actualStr, suffix))
	}
}
