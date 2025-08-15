package main

import (
	"fmt"
	"os"

	appcmd "github.com/ripplego/ripplego/internal/cmd"
)

func main() {
	rootCmd := appcmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}