package countproducts

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leancensus"
)

// loweringPackages are the packages the measuring may never import, each with every package under it: internal/lower,
// which turns source into IR, internal/javascript, the other backend, and internal/oracle, whose tests key the rows and
// whose recipe a count must not hold. A count's key holds this package's whole import closure (leanRecipeOf), so one
// import of any would put its every edit back in all 16 shards' keys, the cost this package exists to cut (#a4frknq).
var loweringPackages = []string{leancensus.ModulePath + "/internal/lower", leancensus.ModulePath + "/internal/javascript", leancensus.ModulePath + "/internal/oracle"}

// This package, its own files without its tests, imports no package that lowers, directly or through another: go list's
// dependencies of it name none of loweringPackages, nor any package under one, but this one. Its line in
// cloud/fast-gate/compiler-dependencies.json is the compiler it does link, through internal/native. Mutant: a blank
// import of internal/lower in run.go, which this names.
func TestTheMeasuringImportsNothingThatLowers(t *testing.T) {
	t.Parallel()
	var found []string
	for _, dependency := range dependencies(t) {
		if dependency == importPath {
			continue
		}
		for _, lowering := range loweringPackages {
			if dependency == lowering || strings.HasPrefix(dependency, lowering+"/") {
				found = append(found, dependency)
			}
		}
	}
	if len(found) > 0 {
		t.Fatalf("internal/oracle/countproducts imports what lowers, so every edit of it moves every count's key: %s", strings.Join(found, ", "))
	}
}
