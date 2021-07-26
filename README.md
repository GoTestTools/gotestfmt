# 🚧 WORK IN PROGRESS: Go test output formatter 🚧

Are you tired of scrolling through endless Golang test logs in GitHub Actions or other CI systems? Would you like a test log like this? (Click the test cases.)

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

Then this is the tool for you. Here's how you use it:

```yaml
jobs:
  build:
    name: Build
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2

        # You MUST set up Go before calling gotestfmt.
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.16

        # Running gotestfmt will call your tests
      - name: Run gotestfmt
        uses: haveyoudebuggedit/gotestfmt@v1
```

Tadam, your tests will now show up in a beautifully formatted fashion in GitHub Actions.

## How does it work?

This application runs `go test -v`, then parses the output and reformats it. The default output format is for GitHub Actions, but you can customize it as needed (see below). The best part is: you can download and run it yourself, you don't have to use our action!

## Customizing test run

There are two ways you can customize the runner. The first possibility is to create a file named `.gotestfmt.yaml` in your project that has the following structure:

```yaml
# Add arguments to go test here. These can be overridden on the command
# line, see below.
args: -cover
templates:
  # Add your custom rendering templates here.
  downloading: |
    Add a Go template here to render a line where a package is downloaded
    as part of a test suite. The {{ .Package }} and {{ .Version }} tags can
    be used to print the package name and version.
  package: |
    Add a Go template here to render a package. The {{ .Package }} fragment
    can be used to print the package name. The {{ .Result }} fragment will
    print PASS, SKIP, or FAIL, depending on the package result. The
    {{ .Elapsed }} tag contains the duration that this package took to
    complete tests. The {{ .Content }} fragment can be used to print the
    tests contained in the package. The {{ .Reason }} tag may contain the
    failure reason.
  syntax: |
    Add a Go template here to render a syntax error. The {{ .Output }}
    contains the error message with the syntax error.
  test: |
    Add a Go template here to render the results of a single test. The
    {{ .Output }} fragment contains the output of that test. The
    {{ .Result }} fragment will print PASS, SKIP, or FAIL, depending on the
    test result. The {{ .Elapsed }} fragment will print the elapsed time for
    this test. The {{ .Reason }} tag may contain the failure reason.
```

When using the GitHub actions, you can override the `args` parameter as follows:

```yaml
- name: Run gotestfmt
  uses: haveyoudebuggedit/gotestfmt@v1
  with:
    args: -cover ./...
```

If you are running this tool from the command line, simply pass the options to the tool and they will be passed through to `go test`.
