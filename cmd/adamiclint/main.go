// adamiclint runs adamic's own static analysis rules as a vet tool: go vet -vettool=$(which adamiclint) ./...
package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/system-inc/adamic/internal/adamiclint/parallel"
	"github.com/system-inc/adamic/internal/adamiclint/repowalk"
	"github.com/system-inc/adamic/internal/adamiclint/waitdelay"
)

// main hands go vet the three analyzers; each rule and its tests live in its own package under internal/adamiclint.
func main() {
	unitchecker.Main(waitdelay.Analyzer, repowalk.Analyzer, parallel.Analyzer)
}
