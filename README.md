# 🚧 WORK IN PROGRESS: Go test output formatter 🚧

Are you tired of scrolling through endless Golang test logs in GitHub Actions (or other CI systems)? Would you like a test log like this? (Click the test cases.)

<pre>
<details><summary>✅ TestCase1</summary>
<p>
Here are the details of the first test case.
</p>
</details>
<details><summary>✅ TestCase2</summary>
<p>
Here are the details of the second test case.
</p>
</details>
<details><summary>❌ TestCase3</summary>
<p>
Here are the details why the third test case failed.
</p>
</details>
</pre>

Then this is the tool for you. Here's how you use it with GitHub Actions:

```yaml
jobs:
  build:
    name: Test
    runs-on: ubuntu-latest
    steps:
      # Checkout your project with git
      - name: Checkout
        uses: actions/checkout@v2

      # Install Go on the VM running the action.
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.16

      # Install gotestfmt on the VM running the action.
      - name: Set up gotestfmt
        uses: haveyoudebuggedit/gotestfmt-action@v1

      # Run tests with nice formatting. Save the original log in /tmp/gotest.log
      - name: Run tests
        run: go test -v ./... 2>&1 | tee /tmp/gotest.log | gotestfmt

      # Upload the original go test log as an artifact for later review.
      - name: Upload test log
        uses: actions/upload-artifact@v2
        with:
          name: test-log
          path: /tmp/gotest.log
          if-no-files-found: error
```

Tadam, your tests will now show up in a beautifully formatted fashion in GitHub Actions and the original log will be uploaded as an artifact next to the test run. Alternatively, you can grab the binary from [the releases section](https://github.com/haveyoudebuggedit/gotestfmt/releases) and run it in a different CI:

```bash
go test -v ./... 2>&1 | gotestfmt -log /tmp/gotest.log
```

**Note:** Please always save the original log. You will need it if you have to file a bug report.

## How does it work?

You can run your tests as normal. The output is piped to gotestfmt which parses and reformats it.

## Customizing the output

You can, of course, customize the output to match your CI system. This can be done by creating a folder named `.gotestfmt` in your project and adding the files below. You can find the default templates in the [.gotestfmt](.gotestfmt) folder in this repository.

### downloads.tpl

This file contains the output fragment showing the package downloads in the Go template format. It has the following variables available:

| Variable | Type | Description |
|----------|------|-------------|
| `.Failed` | `bool` | Indicates an overall failure. |
| `.Packages` | `[]Package` | A list of packages that have been processed.

The `Package` items have the following format:

| Variable | Type | Description |
|----------|------|-------------|
| `.Package` | `string` | Name of the package. (e.g. `github.com/haveyoudebuggedit/gotestfmt`) |
| `.Version` | `string` | Version of the package. (e.g. `v1.0.0`) |
| `.Failed` | `bool` | If the package download has failed. |
| `.Reason` | `string` | Text explaining the failure. |

## package.tpl

This template is the output format for the results of a single package and the tests in it. If multiple packages are tested, this template is called multiple times in a row. It has the following fields:

| Variable | Type | Description |
|----------|------|-------------|
| `.Name`   | `string` | Name of the package under test.
| `.Result` | `string` | Result of all tests in this package. Can be `PASS`, `FAIL`, or `SKIP`. |
| `.Duration` | `time.Duration` | Duration of all test runs in this package. |
| `.Coverage` | `*float64` | If coverage data was provided, this indicates the code coverage percentage. |
| `.Output` | `string` | Additional output from failures. (e.g. syntax error indications) |
| `.TestCases` | `[]TestCase` | A list of test case results. |
| `.Reason` | `string` | Text explaining the failure. Empty in most cases. |

Test cases have the following format:

| Variable | Type | Description |
|----------|------|-------------|
| `.Name` | `string` | Name of the test case. May contain slashes (`/`) if subtests are run. |
| `.Result` | `string` | Result of the test. Can be `PASS`, `FAIL`, or `SKIP`. |
| `.Duration` | `time.Duration` | Duration of all test runs in this package. |
| `.Coverage` | `float64` | If coverage data was provided, this indicates the code coverage percentage. Contains a negative number if no coverage data is available. |
| `.Output` | `string` | Log output from the test. |

## Architecture

This application has 3 main pieces: the tokenizer, the parser, and the renderer. All of them run in separate goroutines and pipeline data using channels.

The **tokenizer** takes the raw output from `go test` and turns it into a stream of events that can be consumed.

The **parser** takes the tokens from the tokenizer and interprets them, constructing logical units for test cases, packages, and package downloads.

Finally, the **renderer** takes the two streams from the parser and renders them into human-readable text templates, which are then streamed out to the main application for writing.

## Building

If you wish to build `gotestfmt` for yourself you'll need at least Go 1.16. You can then build it by running `go build cmd/gotestfmt`.

## License

This project is licensed under the [Unlicense](LICENSE.md), you are free to do with it as you please. It has no external dependencies.