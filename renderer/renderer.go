package renderer

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/gotesttools/gotestfmt/v2/parser"
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

// RenderWithSettingsAndExitCode takes the two input channels from the parser and renders them into text output
// fragments as well as an exit code.
func RenderWithSettingsAndExitCode(
	prefixes <-chan string,
	downloadsChannel <-chan *parser.Downloads,
	packagesChannel <-chan *parser.Package,
	downloadsTemplate []byte,
	packagesTemplate []byte,
	settings RenderSettings,
) (<-chan []byte, <-chan int) {
	result := make(chan []byte)
	exitCodeChan := make(chan int)
	go func() {
		exitCode := 0
		defer func() {
			close(result)
			exitCodeChan <- exitCode
			close(exitCodeChan)
		}()
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
			if downloads.Failed {
				exitCode = 1
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
			if pkg.Result == parser.ResultFail {
				exitCode = 1
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
	return result, exitCodeChan
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

func formatTestOutput(testOutput string, cfg RenderSettings) string {
	if cfg.Formatter == "" {
		return testOutput
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var shell []string
	if runtime.GOOS == "windows" {
		shell = []string{
			"cmd.exe",
			"/C",
			cfg.Formatter,
		}
	} else {
		shell = []string{
			"/bin/bash",
			"-c",
			cfg.Formatter,
		}
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	run := exec.CommandContext(ctx, shell[0], shell[1:]...)
	run.Stdin = bytes.NewReader([]byte(testOutput))
	run.Stdout = stdout
	run.Stderr = stderr
	if err := run.Run(); err != nil {
		panic(fmt.Errorf(
			"failed to run test output formatter '%s', stderr was: %s (%w)",
			strings.Join(shell, " "),
			stderr.String(),
			err,
		))
	}
	return stdout.String()
}

func renderTemplate(templateName string, templateText []byte, data interface{}) []byte {
	result := bytes.Buffer{}
	tpl := template.New(templateName)
	tpl.Funcs(map[string]interface{}{
		"formatTestOutput": formatTestOutput,
	})
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
	// Formatter is the path to an external program that is executed for each test output for format it.
	Formatter string
}
