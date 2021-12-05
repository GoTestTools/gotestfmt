package renderer

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/haveyoudebuggedit/gotestfmt/v2/parser"
)

// Render takes the two input channels from the parser and renders them into text output fragments.
func Render(
	prefixes <-chan string,
	downloadsChannel <-chan *parser.Downloads,
	packagesChannel <-chan *parser.Package,
	downloadsTemplate []byte,
	packagesTemplate []byte,
) <-chan []byte {
	result := make(chan []byte)
	go func() {
		defer close(result)
		for {
			prefix, ok := <-prefixes
			if !ok {
				break
			}
			result <- []byte(fmt.Sprintf("%s\n", prefix))
		}

		for {
			downloads, ok := <-downloadsChannel
			if !ok {
				break
			}
			result <- renderTemplate(
				"downloads.gotpl",
				downloadsTemplate,
				downloads,
			)
		}

		for {
			pkg, ok := <-packagesChannel
			if !ok {
				break
			}
			result <- renderTemplate(
				"package.gotpl",
				packagesTemplate,
				pkg,
			)
		}
	}()
	return result
}

func renderTemplate(templateName string, templateText []byte, data interface{}) []byte {
	result := bytes.Buffer{}
	tpl := template.New(templateName)
	tpl, err := tpl.Parse(string(templateText))
	if err != nil {
		panic(fmt.Errorf("failed to parse template (%w)", err))
	}
	if err := tpl.Execute(&result, data); err != nil {
		panic(fmt.Errorf("failed to render template (%w)", err))
	}
	return result.Bytes()
}
