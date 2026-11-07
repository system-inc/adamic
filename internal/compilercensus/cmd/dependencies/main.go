// Command dependencies prints the compiler dependency census of the repository it runs in, as the version-1 map of
// cloud/fast-gate/compiler-dependencies.json: {"version": 1, "packages": {directory: [the compiler's packages]}}. The
// gate's readers run it on a tree in place of that hand-kept list (#ng0ksed):
//
//	go run ./internal/compilercensus/cmd/dependencies
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/compilercensus"
)

func main() {
	root, workingDirectoryError := os.Getwd()
	if workingDirectoryError != nil {
		fmt.Fprintln(os.Stderr, "dependencies:", workingDirectoryError)
		os.Exit(1)
	}
	packages, dependenciesError := compilercensus.Dependencies(root)
	if dependenciesError != nil {
		fmt.Fprintln(os.Stderr, "dependencies:", dependenciesError)
		os.Exit(1)
	}
	encoded, encodeError := json.MarshalIndent(map[string]any{"version": 1, "packages": packages}, "", "  ")
	if encodeError != nil {
		fmt.Fprintln(os.Stderr, "dependencies:", encodeError)
		os.Exit(1)
	}
	fmt.Println(string(encoded))
}
