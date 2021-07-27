package tokenizer_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/haveyoudebuggedit/gotestfmt/tokenizer"
)

type testEntry struct {
	input          string
	expectedOutput []tokenizer.Event
}

var testdata = map[string]testEntry{
	"single-package": {
		input: `ok      github.com/haveyoudebuggedit/example   0.019s`,
		expectedOutput: []tokenizer.Event{
			{
				Action:  tokenizer.ActionPass,
				Package: "github.com/haveyoudebuggedit/example",
				Test:    "",
				Elapsed: 19 * time.Millisecond,
				Output:  nil,
			},
		},
	},
	"single-package-verbose": {
		input: `=== RUN   TestNothing
--- PASS: TestNothing (0.00s)
PASS
ok      github.com/haveyoudebuggedit/example   0.019s`,
		expectedOutput: []tokenizer.Event{
			{
				Action:  tokenizer.ActionRun,
				Package: "",
				Test:    "TestNothing",
				Output:  nil,
			},
			{
				Action:  tokenizer.ActionPass,
				Package: "",
				Test:    "TestNothing",
				Elapsed: 0,
				Output:  nil,
			},
			{
				Action:  tokenizer.ActionPass,
				Package: "",
				Test:    "",
				Elapsed: 0,
				Output:  nil,
			},
			{
				Action:  tokenizer.ActionPass,
				Package: "github.com/haveyoudebuggedit/example",
				Test:    "",
				Elapsed: 19 * time.Millisecond,
				Output:  nil,
			},
		},
	},
	"mod-download": {
		input: `go: downloading github.com/stretchr/testify v1.7.0
go: downloading github.com/pmezard/go-difflib v1.0.0
go: downloading github.com/davecgh/go-spew v1.1.0
go: downloading gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
=== RUN   TestNothing
--- PASS: TestNothing (0.00s)
PASS
ok      github.com/haveyoudebuggedit/example    0.027s`,
		expectedOutput: []tokenizer.Event{
			{
				Action:  tokenizer.ActionDownload,
				Package: "github.com/stretchr/testify",
				Version: "v1.7.0",
				Test:    "",
				Elapsed: 0,
				Output:  nil,
			},
			{
				Action:  tokenizer.ActionDownload,
				Package: "github.com/pmezard/go-difflib",
				Version: "v1.0.0",
				Test:    "",
				Elapsed: 0,
				Output:  nil,
			},
			{
				Action:  tokenizer.ActionDownload,
				Package: "github.com/davecgh/go-spew",
				Version: "v1.1.0",
				Test:    "",
				Elapsed: 0,
				Output:  nil,
			},
			{
				Action:  tokenizer.ActionDownload,
				Package: "gopkg.in/yaml.v3",
				Version: "v3.0.0-20200313102051-9f266ea9e77c",
				Test:    "",
				Elapsed: 0,
				Output:  nil,
			},
			{
				Action:  tokenizer.ActionRun,
				Package: "",
				Test:    "TestNothing",
				Output:  nil,
			},
			{
				Action:  tokenizer.ActionPass,
				Package: "",
				Test:    "TestNothing",
				Elapsed: 0,
				Output:  nil,
			},
			{
				Action:  tokenizer.ActionPass,
				Package: "",
				Test:    "",
				Elapsed: 0,
				Output:  nil,
			},
			{
				Action:  tokenizer.ActionPass,
				Package: "github.com/haveyoudebuggedit/example",
				Test:    "",
				Elapsed: 27 * time.Millisecond,
				Output:  nil,
			},
		},
	},
	"parallel": {
		input: `=== RUN   TestParallel1
=== PAUSE TestParallel1
=== RUN   TestParallel2
=== PAUSE TestParallel2
=== CONT  TestParallel1
    parallel_test.go:10: Test message 1
=== CONT  TestParallel2
=== CONT  TestParallel1
    parallel_test.go:12: Test message 2
--- PASS: TestParallel1 (5.01s)
=== CONT  TestParallel2
    parallel_test.go:18: Test message 1
    parallel_test.go:20: Test message 2
--- PASS: TestParallel2 (10.02s)
PASS
ok      github.com/haveyoudebuggedit/example    10.048s`,
		expectedOutput: []tokenizer.Event{
			{
				Action: tokenizer.ActionRun,
				Test:   "TestParallel1",
			},
			{
				Action: tokenizer.ActionPause,
				Test:   "TestParallel1",
			},
			{
				Action: tokenizer.ActionRun,
				Test:   "TestParallel2",
			},
			{
				Action: tokenizer.ActionPause,
				Test:   "TestParallel2",
			},
			{
				Action: tokenizer.ActionCont,
				Test:   "TestParallel1",
			},
			{
				Action: tokenizer.ActionStdout,
				Output: []byte("    parallel_test.go:10: Test message 1"),
			},
			{
				Action: tokenizer.ActionCont,
				Test:   "TestParallel2",
			},
			{
				Action: tokenizer.ActionCont,
				Test:   "TestParallel1",
			},
			{
				Action: tokenizer.ActionStdout,
				Output: []byte("    parallel_test.go:12: Test message 2"),
			},
			{
				Action:  tokenizer.ActionPass,
				Test:    "TestParallel1",
				Elapsed: 5010 * time.Millisecond,
			},
			{
				Action: tokenizer.ActionCont,
				Test:   "TestParallel2",
			},
			{
				Action: tokenizer.ActionStdout,
				Output: []byte("    parallel_test.go:18: Test message 1"),
			},
			{
				Action: tokenizer.ActionStdout,
				Output: []byte("    parallel_test.go:20: Test message 2"),
			},
			{
				Action:  tokenizer.ActionPass,
				Test:    "TestParallel2",
				Elapsed: 10020 * time.Millisecond,
			},
			{
				Action: tokenizer.ActionPass,
			},
			{
				Action:  tokenizer.ActionPass,
				Package: "github.com/haveyoudebuggedit/example",
				Elapsed: 10048 * time.Millisecond,
			},
		},
	},
	"syntax-error": {
		input: `# github.com/haveyoudebuggedit/example
nothing_test.go:7:11: expected '(', found Nothing
FAIL    github.com/haveyoudebuggedit/example [setup failed]
FAIL`,
		expectedOutput: []tokenizer.Event{
			{
				Action:  tokenizer.ActionPackage,
				Package: "github.com/haveyoudebuggedit/example",
			},
			{
				Action: tokenizer.ActionStdout,
				Output: []byte("nothing_test.go:7:11: expected '(', found Nothing"),
			},
			{
				Action:  tokenizer.ActionFail,
				Package: "github.com/haveyoudebuggedit/example",
				Output:  []byte("setup failed"),
			},
			{
				Action: tokenizer.ActionFail,
			},
		},
	},
	"subtest": {
		input: `=== RUN   TestSubtest
=== RUN   TestSubtest/test1
    subtest_test.go:9: Hello world!
=== RUN   TestSubtest/test2
    subtest_test.go:12: Here's an error.
=== RUN   TestSubtest/test3
    subtest_test.go:15: Let's skip this one...
--- FAIL: TestSubtest (0.00s)
    --- PASS: TestSubtest/test1 (0.00s)
    --- FAIL: TestSubtest/test2 (0.00s)
    --- SKIP: TestSubtest/test3 (0.00s)
FAIL
FAIL    github.com/haveyoudebuggedit/example    0.020s
FAIL`,
		expectedOutput: []tokenizer.Event{
			{
				Action: tokenizer.ActionRun,
				Test:   "TestSubtest",
			},
			{
				Action: tokenizer.ActionRun,
				Test:   "TestSubtest/test1",
			},
			{
				Action: tokenizer.ActionStdout,
				Output: []byte("    subtest_test.go:9: Hello world!"),
			},
			{
				Action: tokenizer.ActionRun,
				Test:   "TestSubtest/test2",
			},
			{
				Action: tokenizer.ActionStdout,
				Output: []byte("    subtest_test.go:12: Here's an error."),
			},
			{
				Action: tokenizer.ActionRun,
				Test:   "TestSubtest/test3",
			},
			{
				Action: tokenizer.ActionStdout,
				Output: []byte("    subtest_test.go:15: Let's skip this one..."),
			},
			{
				Action: tokenizer.ActionFail,
				Test:   "TestSubtest",
			},
			{
				Action: tokenizer.ActionPass,
				Test:   "TestSubtest/test1",
			},
			{
				Action: tokenizer.ActionFail,
				Test:   "TestSubtest/test2",
			},
			{
				Action: tokenizer.ActionSkip,
				Test:   "TestSubtest/test3",
			},
			{
				Action: tokenizer.ActionFail,
			},
			{
				Action:  tokenizer.ActionFail,
				Package: "github.com/haveyoudebuggedit/example",
				Elapsed: 20 * time.Millisecond,
			},
			{
				Action: tokenizer.ActionFail,
			},
		},
	},
	"coverage-nostatements": {
		input: `=== RUN   TestNothing
--- PASS: TestNothing (0.00s)
PASS
coverage: [no statements]
ok      github.com/haveyoudebuggedit/example    0.024s  coverage: [no statements]`,
		expectedOutput: []tokenizer.Event{
			{
				Action: tokenizer.ActionRun,
				Test:   "TestNothing",
			},
			{
				Action: tokenizer.ActionPass,
				Test:   "TestNothing",
			},
			{
				Action: tokenizer.ActionPass,
			},
			{
				Action: tokenizer.ActionCoverageNoStatements,
			},
			{
				Action:  tokenizer.ActionPass,
				Package: "github.com/haveyoudebuggedit/example",
				Elapsed: 24 * time.Millisecond,
			},
		},
	},
	"coverage-statements": {
		input: `=== RUN   TestNothing
--- PASS: TestNothing (0.00s)
PASS
coverage: 57.8% of statements
ok      github.com/haveyoudebuggedit/example    (cached)        coverage: 57.8% of statements`,
		expectedOutput: []tokenizer.Event{
			{
				Action: tokenizer.ActionRun,
				Test:   "TestNothing",
			},
			{
				Action: tokenizer.ActionPass,
				Test:   "TestNothing",
			},
			{
				Action: tokenizer.ActionPass,
			},
			{
				Action:   tokenizer.ActionCoverage,
				Coverage: 57.8,
			},
			{
				Action:   tokenizer.ActionPass,
				Package:  "github.com/haveyoudebuggedit/example",
				Cached:   true,
				Coverage: 57.8,
			},
		},
	},
	"gosum": {
		input: "go: github.com/haveyoudebuggedit/nonexistent@v1.0.0: missing go.sum entry; to add it:\n        go mod download github.com/haveyoudebuggedit/nonexistent",
		expectedOutput: []tokenizer.Event{
			{
				Action:  tokenizer.ActionDownloadFailed,
				Package: "github.com/haveyoudebuggedit/nonexistent",
				Version: "v1.0.0",
				Output:  []byte("missing go.sum entry; to add it:"),
			},
			{
				Action: tokenizer.ActionStdout,
				Output: []byte("        go mod download github.com/haveyoudebuggedit/nonexistent"),
			},
		},
	},
	"norevision": {
		input: `go: github.com/haveyoudebuggedit/nonexistent@v1.0.0: reading github.com/haveyoudebuggedit/nonexistent/go.mod at revision v1.0.0: unknown revision v1.0.0`,
		expectedOutput: []tokenizer.Event{
			{
				Action:  tokenizer.ActionDownloadFailed,
				Package: "github.com/haveyoudebuggedit/nonexistent",
				Version: "v1.0.0",
				Output:  []byte("reading github.com/haveyoudebuggedit/nonexistent/go.mod at revision v1.0.0: unknown revision v1.0.0"),
			},
		},
	},
}

func TestParsing(t *testing.T) {
	for name, testEntry := range testdata {
		entry := testEntry
		t.Run(
			name, func(t *testing.T) {
				reader := tokenizer.Tokenize(bytes.NewReader([]byte(entry.input)))
				remainingOutput := entry.expectedOutput
				for {
					event, ok := <-reader
					if !ok {
						if len(remainingOutput) != 0 {
							t.Fatalf(
								"Reader closed even though there are %d expected items remaining",
								len(remainingOutput),
							)
						}
						return
					}
					if len(remainingOutput) == 0 {
						t.Fatalf("Reader returned an event, but there were no more events expected: %v", event)
					}
					expectedEvent := remainingOutput[0]
					remainingOutput = remainingOutput[1:]
					if !expectedEvent.Equals(event) {
						t.Fatalf("The following two events did not match:\n%s", expectedEvent.Diff(event))
					}
				}
			},
		)
	}
}
