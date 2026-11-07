package lint

import (
	"strings"
	"testing"
)

func TestWitnessScriptKind(test *testing.T) {
	test.Parallel()
	witnessScriptKindSetup(test)
	stop := witnessScriptKindDeadline(test)
	defer stop()
	witnessScriptKindUnion(test)
	test.Logf("TestWitnessScriptKind (union): %s", strings.Join(witnessScriptKindState.Keys, ", "))
}
