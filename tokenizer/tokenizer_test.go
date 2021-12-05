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

	"github.com/haveyoudebuggedit/gotestfmt/v2/testutil"
	"github.com/haveyoudebuggedit/gotestfmt/v2/tokenizer"
)

// TestTokenization reads the *.txt and *.tokenizer.json files from the ../testdata directory, then runs
// the tokenizer.Tokenize function on the text input and compares the output to the events read from the JSON files.
func TestTokenization(t *testing.T) {
	t.Logf("Locating testdata directory...")
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
	t.Logf("Testdata directory is located at %s.", foundDir)

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
				t.Parallel()
				sourceFile := path
				expectedFile := strings.Replace(path, ".txt", ".tokenizer.json", 1)

				t.Logf("Tokenizing %s and comparing with %s...", sourceFile, expectedFile)

				var source, err = ioutil.ReadFile(sourceFile)
				if err != nil {
					t.Fatalf("failed to read source file %s (%v)", source, err)
				}
				expectedFh, err := os.Open(expectedFile)
				if err != nil {
					t.Fatalf("failed to read expected file %s (%v)", expectedFile, err)
				}
				defer func() {
					_ = expectedFh.Close()
				}()

				decoder := json.NewDecoder(expectedFh)
				var expectedOutput []tokenizer.Event
				if err := decoder.Decode(&expectedOutput); err != nil {
					t.Fatalf("failed to decode expected file %s (%v)", expectedFile, err)
				}
				reader := tokenizer.Tokenize(bytes.NewReader(source))
				var events []tokenizer.Event
				for {
					event, ok := <-reader
					if !ok {
						break
					}
					events = append(events, event)
				}
				diff := testutil.Diff(expectedOutput, events)
				if diff != "" {
					t.Fatalf("The expected output did not match the real output:\n%v", diff)
				}
				t.Logf("No difference, test successful.")
			},
		)
		return nil
	}); e != nil {
		t.Fatal(e)
	}
}
