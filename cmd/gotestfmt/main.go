package main

import (
	"os"

	"github.com/haveyoudebuggedit/gotestfmt"
)

func main() {
	fmt, err := gotestfmt.New("./.gotestfmt")
	if err != nil {
		panic(err)
	}
	fmt.Format(os.Stdin, os.Stdout)
}
