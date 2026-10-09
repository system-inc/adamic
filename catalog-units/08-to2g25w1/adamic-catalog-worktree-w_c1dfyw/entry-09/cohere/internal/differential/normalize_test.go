package differential_test

import (
	"testing"

	"github.com/system-inc/cohere/internal/differential"
)

// A slash means two opposite things depending on who wrote it, and getting it wrong does not look
// like a parse failure.
//
// `NormalizeRuleName` had no test at all, which is how a bug survived the fix written for it. The
// original defect collapsed every config key to its plugin, so the configured set became
// `{nexus, structure, typescript}` and every genuine disagreement was excused as not-configured. The
// fix told the two forms apart by looking for a hyphen in the first segment, and that was right for
// every plugin whose name has no hyphen.
//
// **`better-tailwindcss` has one.** So all eight of its rules collapsed to `better-tailwindcss` and
// the family read as unconfigured, which is the same failure in a narrower form. It survived because
// nothing here asserted the plugin-prefixed direction on a hyphenated plugin name.
//
// The distinguishing shape is the second segment: a message id is camelCase and therefore never
// hyphenated, so a hyphen after the slash means the config's `plugin/rule-name` form.
func TestNormalizeRuleName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		raw   string
		want  string
		notes string
	}{
		{
			name:  "hyphenated plugin prefix",
			raw:   "better-tailwindcss/no-duplicate-classes",
			want:  "no-duplicate-classes",
			notes: "the case that collapsed an entire family to its plugin name",
		},
		{name: "plain plugin prefix", raw: "nexus/consistency-no-enum", want: "consistency-no-enum"},
		{name: "typescript prefix", raw: "typescript/no-unused-vars", want: "no-unused-vars"},
		{
			name:  "rule and message id",
			raw:   "react-component-no-multiple-primary/noMultiplePrimary",
			want:  "react-component-no-multiple-primary",
			notes: "cohere's own findings put the rule before the slash",
		},
		{name: "short rule and message id", raw: "no-enum/enumDeclared", want: "no-enum"},
		{name: "bare name", raw: "no-debugger", want: "no-debugger"},
		{name: "parenthesized plugin", raw: "structure(react-component-no-multiple-primary)", want: "react-component-no-multiple-primary"},
		{name: "surrounding space", raw: "  no-debugger  ", want: "no-debugger"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := differential.NormalizeRuleName(testCase.raw)
			if got != testCase.want {
				message := "normalizing %q gave %q, wanted %q"
				if testCase.notes != "" {
					message += " (" + testCase.notes + ")"
				}
				t.Errorf(message, testCase.raw, got, testCase.want)
			}
		})
	}
}
