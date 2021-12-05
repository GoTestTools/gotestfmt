package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/haveyoudebuggedit/gotestfmt/v2"
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

func main() {
	dirs := []string{
		"./.gotestfmt",
	}
	ci := ""
	inputFile := "-"
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

	format.Format(input, os.Stdout)
}
