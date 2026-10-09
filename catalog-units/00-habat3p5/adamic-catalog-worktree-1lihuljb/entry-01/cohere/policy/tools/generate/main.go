// Command generate writes the Swift sources that compile cohere's policy files into the Swift engine
// (policy.SwiftFiles). Run it from the module root after editing a policy file the Swift engine reads:
//
//	go run ./policy/tools/generate          write the files
//	go run ./policy/tools/generate -check   exit 1 if any is stale, writing nothing
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/system-inc/cohere/policy"
)

func main() {
	check := flag.Bool("check", false, "report stale files and exit 1, writing nothing")
	flag.Parse()

	if err := run(*check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(check bool) error {
	files, err := policy.SwiftFiles()
	if err != nil {
		return err
	}
	var stale []string
	for _, file := range files {
		existing, err := os.ReadFile(file.Path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if bytes.Equal(existing, file.Contents) {
			continue
		}
		stale = append(stale, file.Path)
		if !check {
			if err := os.WriteFile(file.Path, file.Contents, 0o644); err != nil {
				return err
			}
			fmt.Println("wrote", file.Path)
		}
	}
	if check && len(stale) > 0 {
		return fmt.Errorf("stale, so run go run ./policy/tools/generate: %v", stale)
	}
	return nil
}
