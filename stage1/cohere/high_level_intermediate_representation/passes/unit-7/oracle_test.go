//go:build lintoracle

// Overlay beside the pinned Go HIR package. This exports input, not a pass result.
package high_level_intermediate_representation

import (
	"fmt"
	"os"
	"testing"
)

func TestUnit7ScopeTerminalInput(t *testing.T) {
	t.Parallel()
	terminal := &Scope{Scope: 1, Block: 2, Fallthrough: 3}
	line := fmt.Sprintf("%d Scope %s\n", TerminalOrder(terminal), oraclePayload(terminal))
	if err := os.WriteFile(os.Getenv("UNIT7_SCOPE_TERMINAL_OUTPUT"), []byte(line), 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("Go input: %s", line)
}
