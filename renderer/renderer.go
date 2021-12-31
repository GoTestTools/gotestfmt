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
	return RenderWithSettings(
		prefixes,
		downloadsChannel,
		packagesChannel,
		downloadsTemplate,
		packagesTemplate,
		RenderSettings{},
	)
}

// RenderWithSettings takes the two input channels from the parser and renders them into text output fragments.
func RenderWithSettings(
	prefixes <-chan string,
	downloadsChannel <-chan *parser.Downloads,
	packagesChannel <-chan *parser.Package,
	downloadsTemplate []byte,
	packagesTemplate []byte,
	settings RenderSettings,
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
				Downloads{
					downloads,
					settings,
				},
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
				Package{
					pkg,
					settings,
				},
			)
		}
	}()
	return result
}

// Downloads contains the downloads for rendering.
type Downloads struct {
	*parser.Downloads

	Settings RenderSettings
}

// Package contains a single package for rendering.
type Package struct {
	*parser.Package

	Settings RenderSettings
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

// RenderSettings influence the output.
type RenderSettings struct {
	// HideSuccessfulDownloads hides successful package downloads from the output.
	HideSuccessfulDownloads bool
	// HideSuccessfulPackages hides all packages that have only successful tests from the output.
	HideSuccessfulPackages bool
	// HideEmptyPackages hides the packages from the output that have no test cases.
	HideEmptyPackages bool
	// HideSuccessfulTests hides all tests from the output that are successful.
	HideSuccessfulTests bool
	// ShowTestStatus adds words to indicate the test status next to the icons (PASS, FAIl, SKIP).
	ShowTestStatus bool
}
