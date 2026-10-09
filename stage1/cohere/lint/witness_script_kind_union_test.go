package lint

import (
	"strings"
	"testing"
)

func TestWitnessScriptKind(t *testing.T) {
	t.Parallel()
	witnessScriptKindSetup(t)
	stop := witnessScriptKindDeadline(t)
	defer stop()
	witnessScriptKindUnion(t)
	t.Logf("TestWitnessScriptKind (union): %s", strings.Join(witnessScriptKindState.Keys, ", "))
}
