package gotestfmt

import (
	"embed"
	"fmt"
	"io"
	"io/ioutil"
	"path"

	"github.com/haveyoudebuggedit/gotestfmt/parser"
	"github.com/haveyoudebuggedit/gotestfmt/renderer"
	"github.com/haveyoudebuggedit/gotestfmt/tokenizer"
)

//go:embed .gotestfmt/*.gotpl
var fs embed.FS

func New(
	templateDir string,
) (GoTestFmt, error) {
	downloadsTpl, err := ioutil.ReadFile(path.Join(templateDir, "downloads.gotpl"))
	if err != nil {
		downloadsTpl, err = fs.ReadFile(".gotestfmt/downloads.gotpl")
		if err != nil {
			panic(fmt.Errorf("bug: downloads.gotpl not found in binary (%w)", err))
		}
	}

	packageTpl, err := ioutil.ReadFile(path.Join(templateDir, "package.gotpl"))
	if err != nil {
		packageTpl, err = fs.ReadFile(".gotestfmt/package.gotpl")
		if err != nil {
			panic(fmt.Errorf("bug: package.gotpl not found in binary (%w)", err))
		}
	}

	return &goTestFmt{
		downloadsTpl: downloadsTpl,
		packageTpl:   packageTpl,
	}, nil
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
