package testformatter

import (
	"time"
)

type Result string

const (
	ResultPass Result = "PASS"
	ResultFail Result = "FAIL"
	ResultSkip Result = "SKIP"
)

// DownloadingVars is the context for the downloading template.
type DownloadingVars struct {
	Package string
	Version string
}

// PackageVars is the context for the package rendering template.
type PackageVars struct {
	// Package is the name for this package.
	Package string
	// Result is the test result for this package.
	Result Result
	// Elapsed is the time it took for this package to complete.
	Elapsed time.Duration
	// Reason may contain the failure reason for this package.
	Reason string
	// Content is the already rendered content of the tests contained within this package.
	Content string
}

// SyntaxVars is the context for the syntax error rendering template.
type SyntaxVars struct {
	// Output is the error message for the syntax error.
	Output string
}