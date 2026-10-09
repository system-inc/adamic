package oracle

import (
	"bytes"
	"testing"
)

// Ruled source/backend divergence: the TypeScript promise is checked. Node
// reads undefined from an absent merged field; Adamic stops before trusting T.
// Like fixtures.checked, this list pins each backend stop independently of Node.
var mergedFieldRuledDivergences = []struct{ file, sourceOutput, stop string }{
	{"merge_missing.a", "undefined\n", "adamic: panic: field read failed: new Missing().promised is not initialized; expected number, found missing\n"},
	{"merge_missing_computed.a", "undefined\n", "adamic: panic: field read failed: new Missing()['promised'] is not initialized; expected number, found missing\n"},
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{refusalFixtureRoot + "merge_present.a", true, false})
	additionalFixtureCounts = append(additionalFixtureCounts, func(t *testing.T) []string {
		return []string{rulingCount(t, "merge_unused.a", true), rulingCount(t, "merge_missing.a", true), rulingCount(t, "merge_missing_computed.a", true), rulingCount(t, "merge_present.a", true)}
	})
}

func TestMergedFieldAdamicUnusedRefused(t *testing.T) {
	t.Parallel()
	rulingRefusal(t, "merge_unused.a", false, "class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view")
}
func TestMergedFieldAdamicReadRefused(t *testing.T) {
	t.Parallel()
	rulingRefusal(t, "merge_missing.a", false, "class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view")
}
func TestMergedFieldTypeScriptUnusedAccepted(t *testing.T) {
	t.Parallel()
	rulingAgrees(t, "merge_unused.a", true)
}
func TestMergedFieldTypeScriptMissingChecked(t *testing.T) {
	t.Parallel()
	mergedFieldChecked(t, 0)
}

func TestMergedFieldTypeScriptComputedChecked(t *testing.T) {
	t.Parallel()
	mergedFieldChecked(t, 1)
}

func mergedFieldChecked(t *testing.T, index int) {
	t.Helper()
	fixture := mergedFieldRuledDivergences[index]
	node := rulingNode(t, fixture.file)
	if node.exitCode != 0 || string(node.stdout) != fixture.sourceOutput || len(node.stderr) != 0 {
		t.Fatalf("Node outcome: %#v", node)
	}
	program, err := rulingProgram(t, fixture.file, true)
	if err != nil {
		t.Fatal(err)
	}
	actual, _ := natively(t, program)
	for _, result := range []run{actual, onJavaScriptBackend(t, program), released(t, program)} {
		if result.exitCode != 70 || len(result.stdout) != 0 || !bytes.Equal(result.stderr, []byte(fixture.stop)) {
			t.Fatalf("want pinned exit 70: %#v", result)
		}
	}
}
func TestMergedFieldAdamicInitializedAgreesWithNode(t *testing.T) {
	t.Parallel()
	rulingAgrees(t, "merge_present.a", false)
}
func TestMergedFieldTypeScriptInitializedAgreesWithNode(t *testing.T) {
	t.Parallel()
	rulingAgrees(t, "merge_present.a", true)
}

func TestMergedFieldAdamicComputedRefused(t *testing.T) {
	t.Parallel()
	rulingRefusal(t, "merge_missing_computed.a", false, "class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view")
}
