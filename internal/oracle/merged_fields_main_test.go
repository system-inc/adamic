package oracle

import (
	"bytes"
	"testing"
)

// Ruled source/backend divergence: the TypeScript promise is checked. Node
// reads undefined from an absent merged field; Adamic stops before trusting T.
// Like fixtures.checked, this list pins each backend stop independently of Node.
var mergedFieldRuledDivergences = []struct{ file, sourceOutput, stop string }{
	{
		"merge_missing.a",
		"undefined\n",
		"adamic: panic: field read failed: new Missing().promised is not initialized; expected number, found missing\n",
	},
	{
		"merge_missing_computed.a",
		"undefined\n",
		"adamic: panic: field read failed: new Missing()['promised'] is not initialized; expected number, found missing\n",
	},
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{refusalFixtureRoot + "merge_present.a", true, false})
	// merge_present.a's bare .a record counts it; its TypeScript-mode agree and leak tests still run it.
	for _, file := range []string{"merge_unused.a", "merge_missing.a", "merge_missing_computed.a"} {
		additionalFixtureCounts = append(
			additionalFixtureCounts,
			countCase{refusalFixtureRoot + file, func(test *testing.T) (string, func() string) {
				return rulingCount(test, file, true)
			}},
		)
	}
}

func TestMergedFieldAdamicUnusedRefused(test *testing.T) {
	test.Parallel()
	rulingRefusal(
		test,
		"merge_unused.a",
		false,
		"class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view",
	)
}
func TestMergedFieldAdamicReadRefused(test *testing.T) {
	test.Parallel()
	rulingRefusal(
		test,
		"merge_missing.a",
		false,
		"class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view",
	)
}
func TestMergedFieldTypeScriptUnusedAccepted(test *testing.T) {
	test.Parallel()
	rulingAgrees(test, "merge_unused.a", true)
}
func TestMergedFieldTypeScriptMissingChecked(test *testing.T) {
	test.Parallel()
	mergedFieldChecked(test, 0)
}

func TestMergedFieldTypeScriptComputedChecked(test *testing.T) {
	test.Parallel()
	mergedFieldChecked(test, 1)
}

func mergedFieldChecked(test *testing.T, index int) {
	test.Helper()
	fixture := mergedFieldRuledDivergences[index]
	node := rulingNode(test, fixture.file)
	if node.exitCode != 0 || string(node.stdout) != fixture.sourceOutput || len(node.stderr) != 0 {
		test.Fatalf("Node outcome: %#v", node)
	}
	program, err := rulingProgram(test, fixture.file, true)
	if err != nil {
		test.Fatal(err)
	}
	actual, _ := natively(test, program)
	for _, result := range []run{actual, onJavaScriptBackend(test, program), released(test, program)} {
		if result.exitCode != 70 || len(result.stdout) != 0 || !bytes.Equal(result.stderr, []byte(fixture.stop)) {
			test.Fatalf("want pinned exit 70: %#v", result)
		}
	}
}
func TestMergedFieldAdamicInitializedAgreesWithNode(test *testing.T) {
	test.Parallel()
	rulingAgrees(test, "merge_present.a", false)
}
func TestMergedFieldTypeScriptInitializedAgreesWithNode(test *testing.T) {
	test.Parallel()
	rulingAgrees(test, "merge_present.a", true)
}

func TestMergedFieldAdamicComputedRefused(test *testing.T) {
	test.Parallel()
	rulingRefusal(
		test,
		"merge_missing_computed.a",
		false,
		"class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view",
	)
}
