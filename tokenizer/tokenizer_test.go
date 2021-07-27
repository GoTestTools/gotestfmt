package tokenizer_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	path2 "path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haveyoudebuggedit/gotestfmt/tokenizer"
)

// TestTokenization reads the *.txt and *.tokenizer.json files from the ../testdata directory, then runs
// the tokenizer.Tokenize function on the text input and compares the output to the events read from the JSON files.
func TestTokenization(t *testing.T) {
	tryDirectories := []string{
		"./testdata",
		"../testdata",
	}

	foundDir := ""
	for _, dir := range tryDirectories {
		if _, err := os.Stat(dir); err == nil {
			foundDir = dir
		}
	}
	if foundDir == "" {
		t.Fatalf("failed to find testdata directory in %v", tryDirectories)
	}

	if e := filepath.Walk(foundDir, func(path string, info fs.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".txt") {
			return nil
		}
		base := strings.Replace(path2.Base(filepath.ToSlash(path)), ".txt", "", 1)
		t.Run(
			base,
			func(t *testing.T) {
				sourceFile := path
				expectedFile := strings.Replace(path, ".txt", ".tokenizer.json", 1)

				var source, err = ioutil.ReadFile(sourceFile)
				if err != nil {
					t.Fatalf("failed to read source file %s (%v)", source, err)
				}
				expectedFh, err := os.Open(expectedFile)
				if err != nil {
					t.Fatalf("failed to read expected file %s (%v)", expectedFile, err)
				}
				decoder := json.NewDecoder(expectedFh)
				var expectedOutput []tokenizer.Event
				if err := decoder.Decode(&expectedOutput); err != nil {
					t.Fatalf("failed to decode expected file %s (%v)", expectedFile, err)
				}
				reader := tokenizer.Tokenize(bytes.NewReader(source))
				remainingOutput := expectedOutput
				for {
					event, ok := <-reader
					if !ok {
						if len(remainingOutput) != 0 {
							t.Fatalf(
								"Reader closed even though there are %d expected items remaining: %v",
								len(remainingOutput),
								remainingOutput,
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
		return nil
	}); e != nil {
		t.Fatal(e)
	}
}
