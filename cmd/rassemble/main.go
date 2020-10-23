package main

import (
	"fmt"
	"os"

	"github.com/itchyny/rassemble-go"
)

const cmdName = "rassemble"

func main() {
	pat, err := rassemble.Join(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdName, err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", pat)
}
