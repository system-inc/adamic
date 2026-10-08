package main

import (
	"flag"
	"fmt"
	"github.com/system-inc/adamic/stage3/census/latent/refusalrewrite"
	"os"
)

func main() {
	input := flag.String("input", "", "compiler refusals.go")
	output := flag.String("output", "", "scratch overlay output")
	flag.Parse()
	source, err := os.ReadFile(*input)
	if err == nil {
		var rewritten []byte
		rewritten, err = refusalrewrite.Rewrite(source)
		if err == nil {
			err = os.WriteFile(*output, rewritten, 0600)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
