package gotestfmt

import (
	"embed"
	"fmt"
	"io"
	"io/ioutil"
	"path"

	"github.com/haveyoudebuggedit/gotestfmt/v2/parser"
	"github.com/haveyoudebuggedit/gotestfmt/v2/renderer"
	"github.com/haveyoudebuggedit/gotestfmt/v2/tokenizer"
)

//go:embed .gotestfmt/*.gotpl
//go:embed .gotestfmt/*/*.gotpl
var fs embed.FS

func New(
	templateDirs []string,
) (Combined, error) {
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

type goTestFmt struct {
	packageTpl   []byte
	downloadsTpl []byte
}

func (g *goTestFmt) Format(input io.Reader, target io.WriteCloser) {
	g.FormatWithConfig(input, target, renderer.RenderSettings{})
}

func (g *goTestFmt) FormatWithConfig(input io.Reader, target io.WriteCloser, cfg renderer.RenderSettings) {
	tokenizerOutput := tokenizer.Tokenize(input)
	prefixes, downloads, packages := parser.Parse(tokenizerOutput)
	result := renderer.RenderWithSettings(prefixes, downloads, packages, g.downloadsTpl, g.packageTpl, cfg)

	for {
		fragment, ok := <-result
		if !ok {
			return
		}
		if _, err := target.Write(fragment); err != nil {
			panic(fmt.Errorf("failed to write to output: %w", err))
		}
	}
}
