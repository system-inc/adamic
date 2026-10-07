package nexus

import (
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
	"os"
	"testing"
)

// Go's production loader parses the .a witness as virtual TSX. This is an
// external positive witness for the native parser boundary, not a native port.
func TestAdamicWave14JSXWitness(t *testing.T) {
	source, err := os.ReadFile(os.Getenv("ADAMIC_WAVE14_JSX_WITNESS"))
	if err != nil {
		t.Fatal(err)
	}
	result := rule_testing.RunTyped(t, CorrectnessNoLeakedNumberRender, "/repository/source/Witness.tsx", string(source))
	rule_testing.ExpectFindings(t, result, "leakedNumberRender")
	for _, finding := range result.Diagnostics {
		t.Logf("%d:%d %s %s fixes=%d suggestions=%d", finding.Range.Pos(), finding.Range.End(), finding.Message.Id, finding.Message.Description, len(finding.Fixes), len(finding.Suggestions))
	}
}
