package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/haveyoudebuggedit/gotestfmt/v2"
	"github.com/haveyoudebuggedit/gotestfmt/v2/renderer"
)

// ciEnvironments maps environment variables to directories to check for templates.
var ciEnvironments = map[string][]string{
	"GITHUB_WORKFLOW": {
		"./.gotestfmt/github",
		"./.gotestfmt",
	},
	"TEAMCITY_VERSION": {
		"./.gotestfmt/teamcity",
		"./.gotestfmt",
	},
	"GITLAB_CI": {
		"./.gotestfmt/gitlab",
		"./.gotestfmt",
	},
}

type hide string

const (
	hideDownloads     hide = "successful-downloads"
	hidePackages      hide = "successful-packages"
	hideEmptyPackages hide = "empty-packages"
	hideTests         hide = "successful-tests"
	hideAll           hide = "all"
)

var hideDescriptions = map[hide]string{
	hideDownloads:     "Hide successful dependency downloads",
	hidePackages:      "Hide packages with only successful tests",
	hideEmptyPackages: "Hide packages that have no tests",
	hideTests:         "Hide successful tests",
	hideAll:           "Hide all non-error items",
}

func validHideValues() string {
	result := make([]string, len(hideDescriptions))
	i := 0
	for h := range hideDescriptions {
		result[i] = string(h)
		i++
	}
	return strings.Join(result, ", ")
}

func configFromHide(hideText string) (cfg renderer.RenderSettings, err error) {
	if hideText == "" {
		return renderer.RenderSettings{}, nil
	}
	for _, hidePart := range strings.SplitN(hideText, ",", -1) {
		switch p := hide(strings.TrimSpace(hidePart)); p {
		case hideDownloads:
			cfg.HideSuccessfulDownloads = true
		case hidePackages:
			cfg.HideSuccessfulPackages = true
		case hideEmptyPackages:
			cfg.HideEmptyPackages = true
		case hideTests:
			cfg.HideSuccessfulTests = true
		case hideAll:
			cfg.HideSuccessfulDownloads = true
			cfg.HideSuccessfulPackages = true
			cfg.HideEmptyPackages = true
			cfg.HideSuccessfulTests = true
		default:
			return cfg, fmt.Errorf("invalid value for -hide: %s (valid values are: %s)", p, validHideValues())
		}
	}
	return cfg, nil
}

func hideDescription() string {
	description := "Comma-separated list of things to hide from the output. The following options are valid:\n"
	for h, d := range hideDescriptions {
		description = description + fmt.Sprintf("- %s: %s\n", h, d)
	}
	return description
}

func main() {
	dirs := []string{
		"./.gotestfmt",
	}
	ci := ""
	inputFile := "-"
	hide := ""
	flag.StringVar(
		&ci,
		"ci",
		ci,
		"Which subdirectory to use within the .gotestfmt folder. Defaults to detecting the CI from environment variables.",
	)
	flag.StringVar(
		&inputFile,
		"input",
		inputFile,
		"Read build log from file. Defaults to standard input.",
	)
	flag.StringVar(
		&hide,
		"hide",
		hide,
		hideDescription(),
	)
	flag.Parse()

	if ci != "" {
		dirs = []string{
			fmt.Sprintf("./.gotestfmt/%s", filepath.Clean(ci)),
			"./.gotestfmt",
		}
	} else {
		for env, directories := range ciEnvironments {
			if os.Getenv(env) != "" {
				dirs = directories
			}
		}
	}

	cfg, err := configFromHide(hide)
	if err != nil {
		panic(err)
	}

	format, err := gotestfmt.New(
		dirs,
	)
	if err != nil {
		panic(err)
	}

	input := os.Stdin
	if inputFile != "-" {
		fh, err := os.Open(inputFile)
		if err != nil {
			panic(err)
		}
		defer func() {
			_ = fh.Close()
		}()
		input = fh
	}

	format.FormatWithConfig(input, os.Stdout, cfg)
}
