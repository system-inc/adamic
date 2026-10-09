package suppression

import (
	"strings"
	"testing"
)

// A directive may name a rule by the id ESLint runs it under in a house repo, the Nexus twin of a core
// rule, and cohere honors it as naming the core rule it pairs with (#004vj72).
//
// ESLint matches a directive by exact id, and where Nexus' twin replaces a core rule only the twin's id
// silences it, so a house directive has to be twin-spelled. cohere registers the core id, and the
// plugin-qualified suffix match in matchesRuleName reads `nexus/no-use-before-define` as naming it, so
// one spelling works in both engines with no table in cohere. The core spelling keeps working too: cohere
// ships to users with no Nexus, whose core-spelled directives must not change meaning. Flagging a
// core-spelled directive in a house repo is the parity tool's job, which knows the twins are loaded.
func TestTwinSpelledAndCoreSpelledDirectivesBothSilenceTheCoreRule(t *testing.T) {
	t.Parallel()

	for _, named := range []string{"nexus/no-use-before-define", "no-use-before-define"} {
		t.Run(named, func(t *testing.T) {
			t.Parallel()

			source := "export function run() {\n    // eslint-disable-next-line " + named + " -- the cycle cannot run early\n    return later();\n}\nfunction later() {\n    return 1;\n}\n"
			index := Build(source)
			if !index.Suppresses("no-use-before-define", strings.Index(source, "later();")) {
				t.Fatalf("a directive naming %s did not silence no-use-before-define on the line below", named)
			}
			if index.AppliedCount(0) != 1 || len(index.Unused()) != 0 {
				t.Fatalf("the directive naming %s applied %d times and %d directives read unused, want 1 and 0",
					named, index.AppliedCount(0), len(index.Unused()))
			}
		})
	}

	// The control: the twin spelling silences by naming the rule, not by being a directive at all.
	source := "export function run() {\n    // eslint-disable-next-line nexus/no-use-before-define -- the cycle cannot run early\n    return later();\n}\nfunction later() {\n    return 1;\n}\n"
	index := Build(source)
	if index.Suppresses("require-await", strings.Index(source, "later();")) {
		t.Fatal("a directive naming nexus/no-use-before-define silenced require-await")
	}
	if len(index.Unused()) != 1 {
		t.Fatalf("a directive that silenced nothing read as used: %d unused", len(index.Unused()))
	}
}
