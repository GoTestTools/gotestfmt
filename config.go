package testformatter

// Config is the base configuration structure to use for running tests. This configuration structure is typically
// located in the .gotestfmt.yaml file.
type Config struct {
	// Templates contains the various templates used to render the output.
	Templates TemplateConfig `json:"templates" yaml:"templates"`
}

// TemplateConfig contains the various templates used to render the output
type TemplateConfig struct {
	// DownloadStart is the template that is printed out before downloading of packages are printed out.
	DownloadStart string `json:"downloadStart" yaml:"downloadStart"`
	// Downloading is the template used to render a line where a package is downloaded as part of a test suite. The
	// {{ .Package }} and {{ .Version }} tags can be used to print the package name and version.
	Downloading string `json:"downloading" yaml:"downloading"`
	// DownloadEnd is the template that is printed out after the downloading of packages has finished.
	DownloadEnd string `json:"downloadEnd" yaml:"downloadEnd"`
	// Package contains the template for a package. The {{ .Package }} fragment can be used to print the package name.
	// The {{ .Result }} fragment will print PASS, SKIP, or FAIL, depending on the package result. The {{ .Elapsed }}
	// tag contains the duration that this package took to complete tests. The {{ .Content }} fragment can be used to
	// print the tests contained in the package. The {{ .Reason }} tag may contain the failure reason.
	Package string `json:"package" yaml:"package"`
	// Syntax contains the template on rendering a syntax error. The {{ .Output }} contains the error message with the
	// syntax error.
	Syntax string `json:"syntax" yaml:"syntax"`
	// Test is the template used when a single test is printed. The {{ .Output }} fragment contains the output of that
	// test. The {{ .Result }} fragment will print PASS, SKIP, or FAIL, depending on the test result. The {{ .Elapsed }}
	// fragment will print the elapsed time for this test. The {{ .Reason }} tag may contain the failure reason.
	Test string `json:"test" yaml:"test"`
}