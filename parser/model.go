package parser

import (
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
	// Name is the name of a test case. It may contain slashes (`/`) if this test case is for a subtest.
	Name string
	// Result is the result of this test case.
	Result Result
	// Duration is the time it took to execute this test.
	Duration time.Duration
	// Coverage is the percentage of code coverage in this test case, or a negative number if no coverage data is
	// present.
	Coverage *float64
	// Output is the log output of this test case.
	Output string
}

// Package is the structure for all tests in a package.
type Package struct {
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
	// Reason is a description of why the Result happened. Empty in most cases.
	Reason string
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
	// Packages is a list of packages
	Packages []*Download `json:"packages"`
	// Failed indicates that one or more package downloads failed.
	Failed bool `json:"failed"`
}

// ParseResult is an overall structure for parser results, containing the prefix text, downloads and packages.
type ParseResult struct {
	Prefix    []string  `json:"prefix"`
	Downloads Downloads `json:"downloads"`
	Packages  []Package `json:"packages"`
}
