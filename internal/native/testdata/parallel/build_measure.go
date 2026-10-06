// Build a runtime C measurement harness against the compiler in the current checkout.
package main

import (
	"flag"
	"github.com/system-inc/adamic/internal/native"
	"os"
)

func main() {
	source := flag.String("source", "", "C source")
	output := flag.String("o", "", "output binary")
	flag.Parse()
	data, err := os.ReadFile(*source)
	if err != nil {
		panic(err)
	}
	if err := native.Build(string(data), *output, native.Options{}); err != nil {
		panic(err)
	}
}
