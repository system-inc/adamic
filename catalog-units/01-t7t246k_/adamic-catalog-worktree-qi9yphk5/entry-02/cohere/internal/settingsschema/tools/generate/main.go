// Command generate writes the settings schemas and their reference (internal/settingsschema) into the
// module's schema/ directory. Run it from the module root:
//
//	go run ./internal/settingsschema/tools/generate          write the files
//	go run ./internal/settingsschema/tools/generate -check   exit 1 if any is stale, writing nothing
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/settingsschema"
)

func main() {
	check := flag.Bool("check", false, "report stale files and exit 1, writing nothing")
	flag.Parse()

	files, err := settingsschema.Build(registry.Names())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	stale := 0
	for _, path := range paths {
		existing, _ := os.ReadFile(path)
		if bytes.Equal(existing, files[path]) {
			continue
		}
		stale++
		if *check {
			fmt.Printf("stale: %s\n", path)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(path, files[path], 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", path)
	}
	if *check && stale > 0 {
		os.Exit(1)
	}
}
