package main

import (
	"fmt"
	"icomplie/cmd"
	"os"
)

func main() {
	base := cmd.InitParams()
	if err := cmd.Run(base); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
