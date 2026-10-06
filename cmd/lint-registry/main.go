// lint-registry prepares static imports before building the stage 1 lint driver.
package main

import (
	"flag"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
	"os"
)

func main() {
	root := flag.String("root", "stage1/cohere/lint", "lint source directory")
	flag.Parse()
	descriptors, err := registry.Generate(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, d := range descriptors {
		fmt.Println(d.Name)
	}
}
