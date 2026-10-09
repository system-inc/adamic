package registry

import (
	"strings"
	"testing"
)

// Shapes the reverse sweep (#pd2chkx) found ESLint refusing and cohere's decoders accepting, each
// through the real config layer: refused now, by upstream's schema, naming the element and the path.
// The controls beside them are the nearest shape ESLint accepts, so a refusal is about the shape and
// not about the rule refusing everything.
func TestTheConfigLayerRefusesWhatUpstreamsSchemaRefuses(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		rule, refused, accepted, mustName string
	}{
		// An enum value with its case flipped fell back to the default.
		{"@typescript-eslint/method-signature-style", `["error", "Method"]`, `["error", "method"]`, "element 1"},
		// A specifier missing `from` decoded and was then dropped from allow.
		{"@typescript-eslint/no-deprecated", `["error", {"allow": [{"name": "x"}]}]`, `["error", {"allow": [{"from": "file", "name": "x"}]}]`, "element 1 at allow[0]"},
		// A negative maximum loaded as -1.
		{"complexity", `["error", -1]`, `["error", 0]`, "element 1"},
		// A duplicate stayed a duplicate.
		{"no-unreachable-loop", `["error", {"ignore": ["WhileStatement", "WhileStatement"]}]`, `["error", {"ignore": ["WhileStatement"]}]`, "element 1 at ignore"},
		// An empty list where upstream says minItems 1.
		{"no-warning-comments", `["error", {"decoration": []}]`, `["error", {"decoration": ["*"]}]`, "element 1 at decoration"},
	} {
		_, err := optionsThroughTheConfigLayer(t, testCase.rule, testCase.refused)
		if err == nil {
			t.Errorf("%s %s loaded, and ESLint refuses it", testCase.rule, testCase.refused)
		} else if !strings.Contains(err.Error(), testCase.mustName) || !strings.Contains(err.Error(), "upstream's schema") {
			t.Errorf("%s: the refusal does not name %q and upstream's schema: %v", testCase.rule, testCase.mustName, err)
		}
		if _, err := optionsThroughTheConfigLayer(t, testCase.rule, testCase.accepted); err != nil {
			t.Errorf("%s refused %s, which ESLint accepts: %v", testCase.rule, testCase.accepted, err)
		}
	}
}
