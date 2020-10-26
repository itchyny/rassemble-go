package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/itchyny/rassemble-go"
)

const cmdName = "rassemble"

func main() {
	xs := os.Args[1:]
	if len(xs) == 0 {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			xs = append(xs, s.Text())
		}
	}
	pat, err := rassemble.Join(xs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdName, err)
		os.Exit(1)
	}
	fmt.Println(pat)
}
