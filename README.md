# Golang test output formatter

Are you tired of scrolling through endless Golang test logs in GitHub Actions? Would you like a test log like this?

<details><summary>✅ TestCase1</summary>
<p>
Here are the details of the first test case/
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

Then this is the tool for you. Here's how you use it:

```yaml
jobs:
  build:
    name: Build
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.16

      - name: Set up go-test-output-formatter
        uses: janoszen/go-test-output-formatter@v1

      - name: Run tests
        run: go test -v ./... | ./test-output-formatter
```

Tadam, your tests will now show up in a beautifully formatted fashion in GitHub Actions.