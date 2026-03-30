package main

import (
	"fmt"
	"os"
)

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}
