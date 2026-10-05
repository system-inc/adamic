// sanitized builds an Adamic program as the oracle does, under ASan and UBSan, so a review probe runs
// the same binary the oracle would.
//
//	go run ./review/r2/sanitized <file.a> <out>
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: sanitized <file.a> <out>")
		os.Exit(2)
	}
	program, err := load.Load([]string{os.Args[1]})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := native.Build(native.C(lowered), os.Args[2], native.Options{Sanitize: true}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
