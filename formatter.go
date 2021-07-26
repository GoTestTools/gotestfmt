package testformatter

import (
	"io"
	"time"

	"github.com/haveyoudebuggedit/gotestfmt/gotest"
)

// Parse creates a new formatter that reads the go test output from the input reader, formats it and writes
// to the output according to the passed configuration.
func Parse(
	input io.Reader,
) (downloads <-chan Download, packages <-chan Package) {
	evts := gotest.NewEventReader(input)
	currentPackage := ""
	currentTest := ""
	lastPackage := ""

	downloadResults := map[string]*Download{}
	testResults := map[string]*TestCase{}
	packageResults := map[string]*Package{}
	for {
		evt, ok := <- evts
		if !ok {
			return
		}
		if evt.Test != "" {
			currentTest = evt.Test
			if _, ok := testResults[currentTest]; !ok {
				testResults[currentTest] = &TestCase{
					Name:     currentTest,
					Result:   "",
					Duration: 0,
					Coverage: 0,
					Output:   nil,
				}
			}
		}
		if evt.Package != "" {
			packageResults[evt.Package] = &Package{
				Name:     evt.Package,
				Result:   "",
				Duration: 0,
				Coverage: 0,
			}
		}
		switch evt.Action {
		case gotest.ActionRun:
			lastPackage = ""
		case gotest.ActionFail:
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
				currentPackage = ""
			}

		case gotest.ActionPass:
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
				currentPackage = ""
			}
		case gotest.ActionSkip:
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
				currentPackage = ""
			}
		case gotest.ActionDownload:
			downloadResults[evt.Package] = &Download{
				Name:    evt.Package,
				Version: evt.Version,
				Failed:  false,
				Reason:  nil,
			}
			lastPackage = evt.Package
			currentPackage = ""
			currentTest = ""
		case gotest.ActionDownloadFailed:
			downloadResults[evt.Package] = &Download{
				Name:    evt.Package,
				Version: evt.Version,
				Failed:  true,
				Reason: evt.Output,
			}
			lastPackage = evt.Package
			currentPackage = ""
			currentTest = ""
		case gotest.ActionPackage:
			packageResults[evt.Package].TestCases = testResults
			currentPackage = evt.Package
			testResults = map[string]*TestCase{}
			lastPackage = ""
		case gotest.ActionStdout:
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
}

type TestCase struct {
	Name string
	Result Result
	Duration time.Duration
	Coverage float64
	Output []byte
}

type Package struct {
	Name string
	Result Result
	Duration time.Duration
	Coverage float64
	Output []byte
	TestCases map[string]*TestCase
}

type Download struct {
	Name string
	Version string
	Failed bool
	Reason []byte
}
