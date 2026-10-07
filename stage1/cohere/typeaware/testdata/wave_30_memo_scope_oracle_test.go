// Overlay-only test. It invokes the production validator's check method.
package high_level_intermediate_representation

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestWave30MemoScopeOracle(t *testing.T) {
	var output strings.Builder
	for kind := 0; kind < 2; kind++ {
		for flags := 0; flags < 64; flags++ {
			scopes := &ReactiveScopes{byIdentifier: map[IdentifierId]ScopeId{}}
			if flags&1 != 0 {
				scopes.byIdentifier[1] = 7
			}
			function := &Function{}
			if flags&32 != 0 {
				function.Name = "Component"
				function.Params = []Place{{Identifier: 1}}
			}
			validator := manualMemoValidator{function: function, scopes: scopes, liveScopes: map[ScopeId]bool{7: flags&2 != 0}, prunedScopes: map[ScopeId]bool{7: flags&4 != 0}, walkedScopes: map[ScopeId]bool{7: flags&8 != 0}, prunedByChain: map[ScopeId]bool{7: flags&16 != 0}}
			validator.check(1, 42, PreserveManualMemoizationKind(kind))
			fmt.Fprintf(&output, "%d,%d\t", kind, flags)
			for _, finding := range validator.findings {
				fmt.Fprintf(&output, "%d,%d,%d,%d", finding.Identifier, finding.Scope, finding.Order, finding.Kind)
			}
			output.WriteByte('\n')
		}
	}
	if err := os.WriteFile(os.Getenv("ADAMIC_WAVE30_COMPONENT_ORACLE")+"/memo-go.stdout", []byte(output.String()), 0600); err != nil {
		t.Fatal(err)
	}
}
