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
) (GoTestFmt, error) {
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

type GoTestFmt interface {
	Format(input io.Reader, target io.WriteCloser)
}

type goTestFmt struct {
	packageTpl   []byte
	downloadsTpl []byte
}

func (g *goTestFmt) Format(input io.Reader, target io.WriteCloser) {
	tokenizerOutput := tokenizer.Tokenize(input)
	prefixes, downloads, packages := parser.Parse(tokenizerOutput)
	result := renderer.Render(prefixes, downloads, packages, g.downloadsTpl, g.packageTpl)

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
