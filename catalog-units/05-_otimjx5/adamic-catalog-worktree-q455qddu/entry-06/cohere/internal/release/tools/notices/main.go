// Command notices writes NOTICE and THIRD_PARTY_NOTICES.md at the module root, from cohere's build graph
// and the curated upstreams in internal/release/notices.
//
//	go run ./internal/release/tools/notices
//
// Run it from the module root after a dependency, the compiler pin, the Swift engine's packages or the
// curated list changes. A test in internal/release/notices fails, naming this command, when the committed
// files differ from what it would write.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/system-inc/cohere/internal/release/notices"
)

func main() {
	files, err := notices.Generate(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "notices: %v\n", err)
		os.Exit(1)
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(".", name), files[name], 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "notices: writing %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s (%d bytes)\n", name, len(files[name]))
	}
}
