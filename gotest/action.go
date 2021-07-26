package gotest

type Action string

const (
	// ActionRun declares that a test has been started.
	ActionRun Action = "run"
	// ActionPause declares that a test was paused.
	ActionPause Action = "pause"
	// ActionCont declares the continued test for upcoming output
	ActionCont Action = "cont"
	// ActionPass declares a test or package-level pass for tests.
	ActionPass Action = "pass"
	// ActionFail declares a test or package-level test failure.
	ActionFail Action = "fail"
	// ActionStdout is an event when text was written to the standard output. The contents are in the output field.
	ActionStdout Action = "stdout"
	// ActionStderr is an event when text was written to the standard error. The contents are in the output field.
	ActionStderr Action = "stderr"
	// ActionPackage declares the package for an upcoming error message.
	ActionPackage Action = "package"
	// ActionSkip is an action when a test was skipped.
	ActionSkip Action = "skip"
	// ActionDownload is an event when a package is downloaded via Go modules.
	ActionDownload Action = "download"
	// ActionDownloadFailed indicates that the download of a package failed.
	ActionDownloadFailed Action = "download_failed"
	// ActionCoverage is an event showing code coverage-related statements.
	ActionCoverage Action = "coverage"
	// ActionCoverageNoStatements indicates that there were no code statements to cover.
	ActionCoverageNoStatements Action = "coverage_nostatements"
)
