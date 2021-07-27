package parser

import (
	"github.com/haveyoudebuggedit/gotestfmt/tokenizer"
)

// Parse creates a new formatter that reads the go test output from the input reader, formats it and writes
// to the output according to the passed configuration.
// The result are two channels: the download channel will receive either zero or one result and then be closed.
// Once the downloads channel is closed the parsed package results will be streamed over the second result.
func Parse(
	evts <-chan tokenizer.Event,
) (<-chan *Downloads, <-chan *Package) {
	downloadsChannel := make(chan *Downloads)
	packagesChannel := make(chan *Package)

	go func() {
		defer close(packagesChannel)
		currentPackage := ""
		currentTest := ""
		lastPackage := ""

		var downloadResultsList []*Download
		downloadResults := map[string]*Download{}
		var testResultsList []*TestCase
		testResults := map[string]*TestCase{}
		packageResults := map[string]*Package{}
		downloadsFinished := false
		for {
			evt, ok := <-evts
			if !ok {
				return
			}
			if evt.Action != tokenizer.ActionDownload && evt.Action != tokenizer.ActionDownloadFailed {
				if len(downloadResultsList) > 0 {
					failed := false
					for _, dl := range downloadResultsList {
						if dl.Failed {
							failed = true
							break
						}
					}
					downloadsChannel <- &Downloads{
						Packages: downloadResultsList,
						Failed:   failed,
					}
					downloadResultsList = nil
					downloadResults = map[string]*Download{}
				}
				if !downloadsFinished {
					close(downloadsChannel)
					downloadsFinished = true
				}
			}
			if evt.Test != "" {
				currentTest = evt.Test
				if _, ok := testResults[currentTest]; !ok {
					testResults[currentTest] = &TestCase{
						Name:     currentTest,
						Result:   "",
						Duration: 0,
						Coverage: -1,
						Output:   nil,
					}
					testResultsList = append(testResultsList, testResults[currentTest])
				}
			}
			if evt.Package != "" && evt.Action != tokenizer.ActionDownload && evt.Action != tokenizer.ActionDownloadFailed {
				packageResults[evt.Package] = &Package{
					Name:      evt.Package,
					Result:    "",
					Duration:  0,
					Coverage:  -1,
					Output:    nil,
					TestCases: nil,
				}
			}
			switch evt.Action {
			case tokenizer.ActionRun:
				lastPackage = ""
			case tokenizer.ActionFail:
				result := ResultFail
				if evt.Test != "" {
					testResults[evt.Test].Result = result
					testResults[evt.Test].Duration = evt.Elapsed
					testResults[evt.Test].Coverage = evt.Coverage
					currentTest = ""
				}
				if evt.Package != "" {
					packageResults[evt.Package].Coverage = evt.Coverage
					packageResults[evt.Package].Duration = evt.Elapsed
					packageResults[evt.Package].Result = result
					packageResults[evt.Package].Reason = evt.Output
					currentPackage = ""
					packagesChannel <- packageResults[evt.Package]
					delete(packageResults, evt.Package)
				}
			case tokenizer.ActionPass:
				result := ResultPass
				if evt.Test != "" {
					testResults[evt.Test].Result = result
					testResults[evt.Test].Duration = evt.Elapsed
					testResults[evt.Test].Coverage = evt.Coverage
					currentTest = ""
				}
				if evt.Package != "" {
					packageResults[evt.Package].Coverage = evt.Coverage
					packageResults[evt.Package].Duration = evt.Elapsed
					packageResults[evt.Package].Result = result
					packageResults[evt.Package].Reason = evt.Output
					currentPackage = ""
					packagesChannel <- packageResults[evt.Package]
					delete(packageResults, evt.Package)
				}
			case tokenizer.ActionSkip:
				result := ResultSkip
				if evt.Test != "" {
					testResults[evt.Test].Result = result
					testResults[evt.Test].Duration = evt.Elapsed
					testResults[evt.Test].Coverage = evt.Coverage
					currentTest = ""
				}
				if evt.Package != "" {
					packageResults[evt.Package].Coverage = evt.Coverage
					packageResults[evt.Package].Duration = evt.Elapsed
					packageResults[evt.Package].Result = result
					packageResults[evt.Package].Reason = evt.Output
					currentPackage = ""
					packagesChannel <- packageResults[evt.Package]
					delete(packageResults, evt.Package)
				}
			case tokenizer.ActionDownload:
				downloadResults[evt.Package] = &Download{
					Package: evt.Package,
					Version: evt.Version,
				}
				downloadResultsList = append(downloadResultsList, downloadResults[evt.Package])
				lastPackage = evt.Package
				currentPackage = ""
				currentTest = ""
			case tokenizer.ActionDownloadFailed:
				downloadResults[evt.Package] = &Download{
					Package: evt.Package,
					Version: evt.Version,
					Failed:  true,
					Reason:  evt.Output,
				}
				lastPackage = evt.Package
				currentPackage = ""
				currentTest = ""
			case tokenizer.ActionPackage:
				packageResults[evt.Package].TestCases = testResultsList
				currentPackage = evt.Package
				testResults = map[string]*TestCase{}
				lastPackage = ""
			case tokenizer.ActionStdout:
				if currentTest != "" {
					testResults[currentTest].Output = append(
						testResults[currentTest].Output,
						evt.Output...,
					)
				} else if currentPackage != "" {
					packageResults[currentPackage].Output = append(
						packageResults[currentPackage].Output,
						evt.Output...,
					)
				} else if lastPackage != "" {
					output := []byte("\n")
					downloadResults[evt.Package].Reason = append(
						downloadResults[evt.Package].Reason,
						append(output, evt.Output...)...,
					)
				}
			}
		}
	}()
	return downloadsChannel, packagesChannel
}
