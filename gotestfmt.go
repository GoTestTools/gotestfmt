package gotestfmt

import (
    "embed"
    "fmt"
    "io"
    "io/ioutil"
    "path"

    "github.com/gotesttools/gotestfmt/v2/parser"
    "github.com/gotesttools/gotestfmt/v2/renderer"
    "github.com/gotesttools/gotestfmt/v2/tokenizer"
)

//go:embed .gotestfmt/*.gotpl
//go:embed .gotestfmt/*/*.gotpl
var fs embed.FS

func New(
    templateDirs []string,
) (CombinedExitCode, error) {
    downloadsTpl := findTemplate(templateDirs, "downloads.gotpl")

    packageTpl := findTemplate(templateDirs, "package.gotpl")

    return &goTestFmt{
        downloadsTpl: downloadsTpl,
        packageTpl:   packageTpl,
    }, nil
}

func findTemplate(dirs []string, tpl string) []byte {
    var lastError error
    for _, dir := range dirs {
        templateContents, err := ioutil.ReadFile(path.Join(dir, tpl))
        if err == nil {
            return templateContents
        }
        lastError = err
    }
    for _, dir := range dirs {
        templateContents, err := fs.ReadFile(path.Join(dir, tpl))
        if err == nil {
            return templateContents
        }
        lastError = err
    }
    panic(fmt.Errorf("bug: %s not found in binary (%w)", tpl, lastError))
}

// Combined is an interface that combines both the classic GoTestFmt interface and the Formatter interface.
//goland:noinspection GoDeprecation
type Combined interface {
    GoTestFmt
    Formatter
}

// CombinedExitCode contains Combined and adds a function to format with exit code.
type CombinedExitCode interface {
    Combined
    FormatterExitCode
}

// GoTestFmt implements the classic Format instruction. This is no longer in use.
//
// Deprecated: please use the Formatter interface instead.
//goland:noinspection GoDeprecation
type GoTestFmt interface {
    Format(input io.Reader, target io.WriteCloser)
}

// Formatter contains an extended format function to accept render settings.
type Formatter interface {
    FormatWithConfig(input io.Reader, target io.WriteCloser, cfg renderer.RenderSettings)
}

// FormatterExitCode contains an extended format function to accept render settings and returns an exit code
type FormatterExitCode interface {
    FormatWithConfigAndExitCode(input io.Reader, target io.WriteCloser, cfg renderer.RenderSettings) int
}

type goTestFmt struct {
    packageTpl   []byte
    downloadsTpl []byte
}

func (g *goTestFmt) Format(input io.Reader, target io.WriteCloser) {
    g.FormatWithConfigAndExitCode(input, target, renderer.RenderSettings{})
}

func (g *goTestFmt) FormatWithConfig(input io.Reader, target io.WriteCloser, cfg renderer.RenderSettings) {
    _ = g.FormatWithConfigAndExitCode(input, target, cfg)
}

func (g *goTestFmt) FormatWithConfigAndExitCode(input io.Reader, target io.WriteCloser, cfg renderer.RenderSettings) int {
    tokenizerOutput := tokenizer.Tokenize(input)
    prefixes, downloads, packages := parser.Parse(tokenizerOutput)
    result, exitCodeChan := renderer.RenderWithSettingsAndExitCode(
        prefixes,
        downloads,
        packages,
        g.downloadsTpl,
        g.packageTpl,
        cfg,
    )

    for {
        fragment, ok := <-result
        if !ok {
            return <-exitCodeChan
        }
        if _, err := target.Write(fragment); err != nil {
            panic(fmt.Errorf("failed to write to output: %w", err))
        }
    }
}
