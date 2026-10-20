// adamiclint runs adamic's own static analysis rules as a vet tool: go vet -vettool=$(which adamiclint) ./...
package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/system-inc/adamic/internal/adamiclint/waitdelay"
)

func main() {
	unitchecker.Main(waitdelay.Analyzer)
}
