package parser

import (
	"strings"
	"time"
)

// Result is the type for indicating the result for a test case or package.
type Result string

const (
	ResultPass Result = "PASS"
	ResultFail Result = "FAIL"
	ResultSkip Result = "SKIP"
)

// TestCase is the representation for a single test case.
type TestCase struct {
	// StartTime marks the earliest time this test case was seen in the log output.
	StartTime *time.Time `json:"-"`
	// Name is the name of a test case. It may contain slashes (`/`) if this test case is for a subtest.
	Name string
	// Result is the result of this test case.
	Result Result
	// Duration is the time it took to execute this test.
	Duration time.Duration
	// Coverage is the percentage of code coverage in this test case, or a negative number if no coverage data is
	// present.
	//
	// Deprecated: Coverage is not reported per testcase and should not be used.
	Coverage *float64
	// Output is the log output of this test case.
	Output string
	// Cached indicates that the test results are cached and the tests have not actually been run.
	Cached bool
}

// ID returns the Name of the test case without slashes
func (t *TestCase) ID() string {
	return strings.Replace(t.Name, "/", "_", -1)
}

// EndTime returns the calculated end time of the test case
func (t *TestCase) EndTime() *time.Time {
	if t.StartTime == nil {
		return nil
	}
	endTime := (*t.StartTime).Add(t.Duration)
	return &endTime
}

// Package is the structure for all tests in a package.
type Package struct {
	// StartTime marks the earliest time this package was seen in the log output.
	StartTime *time.Time `json:"-"`
	// Name contains the name of the package under test.
	Name string
	// Result is the result of the sum of all tests in this package.
	Result Result
	// Duration is the time it took to execute the tests in this package.
	Duration time.Duration
	// Coverage is the percentage of code coverage in this package, or a negative number if no coverage data is
	// present.
	Coverage *float64
	// Output is the text output of a generic failure (e.g. a syntax error)
	Output string
	// TestCases is a list of test cases run in this package. Subtests are included as separate test cases.
	TestCases []*TestCase
	// TestCasesByName holds the test cases mapped by name.
	TestCasesByName map[string]*TestCase
	// Reason is a description of why the Result happened. Empty in most cases.
	Reason string
	// Cached indicates that the results came from the go test cache.
	Cached bool
}

func (p *Package) EndTime() *time.Time {
	if p.StartTime == nil {
		return nil
	}
	t := p.StartTime.Add(p.Duration)
	return &t
}

// ID returns the Name of the package without dots and slashes
func (p *Package) ID() string {
	return strings.Replace(
		strings.Replace(p.Name, ".", "_", -1),
		"/", "_", -1,
	)
}

type Download struct {
	// Package is the name of the package being downloaded.
	Package string `json:"package"`
	// Version is the version of the package being downloaded
	Version string `json:"version"`
	// Failed indicates that the download failed.
	Failed bool `json:"failed"`
	// Reason is the reason text of the download failure.
	Reason string `json:"reason"`
}

// Downloads is the context for TemplatePackageDownloads.
type Downloads struct {
	// Packages is a list of packagesByName
	Packages []*Download `json:"packages"`
	// Failed indicates that one or more package downloads failed.
	Failed bool `json:"failed"`
	// StartTime indicates the time when the downloads started.
	StartTime *time.Time `json:"-"`
	// EndTime indicates when the downloads finished.
	EndTime *time.Time `json:"-"`
}

// ParseResult is an overall structure for parser results, containing the prefix text, downloads and packagesByName.
type ParseResult struct {
	Prefix    []string  `json:"prefix"`
	Downloads Downloads `json:"downloads"`
	Packages  []Package `json:"packages"`
}
