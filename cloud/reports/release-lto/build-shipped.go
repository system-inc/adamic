// Command build-shipped builds a pre-emitted C baseline or the shipped native policy.
package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/system-inc/adamic/internal/native"
)

func main() {
	release := flag.Bool("release", false, "use shipped native release flags")
	flag.Parse()
	if flag.NArg() != 2 {
		panic("pass C source and output")
	}
	source, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	options := native.Options{Release: *release}
	library, err := native.RuntimeLibrary("", options)
	if err != nil {
		panic(err)
	}
	if err := native.Build(string(source), flag.Arg(1), options); err != nil {
		panic(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Compile []string
		Link    []string
		Library string
	}{native.Flags(options), native.LinkFlags(options), library}); err != nil {
		panic(err)
	}
}
