package parser_test

import (
    "encoding/json"
    "io/fs"
    "os"
    path2 "path"
    "path/filepath"
    "strings"
    "testing"

    "github.com/gotesttools/gotestfmt/v2/parser"
    "github.com/gotesttools/gotestfmt/v2/testutil"
    "github.com/gotesttools/gotestfmt/v2/tokenizer"
)

// TestParse takes the *.tokenizer.json and *.parser.json files in ../testdata, runs the tokenizer files as input
// through the parser and compares the result with the parser files.
func TestParse(t *testing.T) {
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

        if !strings.HasSuffix(path, ".tokenizer.json") {
            return nil
        }
        base := strings.Replace(path2.Base(filepath.ToSlash(path)), ".tokenizer.json", "", 1)
        t.Run(
            base,
            func(t *testing.T) {
                sourceFile := path
                expectedFile := strings.Replace(path, ".tokenizer.json", ".parser.json", 1)

                t.Logf("Parsing %s and comparing with %s...", sourceFile, expectedFile)

                var input []tokenizer.Event
                inputFh, err := os.Open(sourceFile)
                if err != nil {
                    t.Fatalf("Failed to open test input: %s (%v)", sourceFile, err)
                }
                defer func() {
                    _ = inputFh.Close()
                }()
                inputDecoder := json.NewDecoder(inputFh)
                if err := inputDecoder.Decode(&input); err != nil {
                    t.Skipf("Failed to decode test input: %s (%v)", sourceFile, err)
                }

                parserInput := make(chan tokenizer.Event)
                parserResult := parser.ParseResult{}
                prefixes, downloads, packages := parser.Parse(parserInput)
                readerDone := make(chan struct{})
                go func() {
                    for {
                        prefix, ok := <-prefixes
                        if !ok {
                            break
                        }
                        parserResult.Prefix = append(parserResult.Prefix, prefix)
                    }
                    for {
                        download, ok := <-downloads
                        if !ok {
                            break
                        }
                        parserResult.Downloads = *download
                    }
                    for {
                        pkg, ok := <-packages
                        if !ok {
                            break
                        }
                        parserResult.Packages = append(parserResult.Packages, *pkg)
                    }
                    close(readerDone)
                }()
                for _, evt := range input {
                    parserInput <- evt
                }
                close(parserInput)
                <-readerDone

                var expectedOutput parser.ParseResult
                expectedFh, err := os.Open(expectedFile)
                if err != nil {
                    t.Logf("Failed to open test expectation, dumping actual result in %s...", expectedFile+".actual")
                    expectedFh, err = os.Create(expectedFile + ".actual")
                    if err != nil {
                        t.Fatalf("Failed to open %s for writing (%v).", expectedFile+".actual", err)
                    }
                    encoder := json.NewEncoder(expectedFh)
                    if err := encoder.Encode(parserResult); err != nil {
                        t.Fatalf("Failed to encode actual test result to %s (%v)", expectedFile+".actual", err)
                    }
                    if err := expectedFh.Close(); err != nil {
                        t.Fatalf("Failed to close actual test result file %s (%v)", expectedFile+".actual", err)
                    }
                    t.Skipf("Failed to open test expectation: %s (%v)", expectedFile, err)
                }
                defer func() {
                    _ = expectedFh.Close()
                }()
                expectedDecoder := json.NewDecoder(expectedFh)
                if err := expectedDecoder.Decode(&expectedOutput); err != nil {
                    t.Skipf("Failed to decode test expectation: %s (%v)", expectedFile, err)
                }

                diff := testutil.Diff(expectedOutput, parserResult)
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
