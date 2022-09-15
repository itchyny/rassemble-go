# rassemble-go
[![CI Status](https://github.com/itchyny/rassemble-go/workflows/CI/badge.svg)](https://github.com/itchyny/rassemble-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/itchyny/rassemble-go)](https://goreportcard.com/report/github.com/itchyny/rassemble-go)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/itchyny/rassemble-go/blob/main/LICENSE)
[![release](https://img.shields.io/github/release/itchyny/rassemble-go/all.svg)](https://github.com/itchyny/rassemble-go/releases)
[![pkg.go.dev](https://pkg.go.dev/badge/github.com/itchyny/rassemble-go)](https://pkg.go.dev/github.com/itchyny/rassemble-go)

### Go implementation of [Regexp::Assemble](https://metacpan.org/pod/Regexp::Assemble)
This is a Go language package for joining regular expressions into a regular expression which matches all strings that any of the arguments matches.

```go
package main

import (
	"fmt"
	"log"

	"github.com/itchyny/rassemble-go"
)

func main() {
	xs, err := rassemble.Join([]string{"abc", "ab", "acbd", "abe"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(xs) // a(?:b[ce]?|cbd)
}
```

A command line tool is also provided.
```sh
 % go install github.com/itchyny/rassemble-go/cmd/rassemble@latest
 % rassemble abcd abd acd ad
ab?c?d
 % rassemble $(head -n30 /usr/share/dict/words)
A(?:a(?:ni|r(?:on(?:i(?:c(?:al)?|t(?:e|ic)))?|u))|b(?:ab(?:deh|ua))?)?|a(?:a(?:l(?:ii)?|m|rd(?:vark|wolf))?|ba(?:c(?:a(?:te|y)?|i(?:nat(?:e|ion)|s(?:cus|t))|k|tinal(?:ly)?)?)?)?
```

## Bug Tracker
Report bug at [Issues・itchyny/rassemble-go - GitHub](https://github.com/itchyny/rassemble-go/issues).

## Author
itchyny (https://github.com/itchyny)

## License
This software is released under the MIT License, see LICENSE.
