package suppression

import (
	"strings"
	"testing"
)

// TestRuleReferencesNameEveryRuleEveryDirectiveWrote pins what the unknown-rule check reads: every
// name a disable or enable comment wrote, with the whole comment as its range.
//
// Enables are the half most likely to be dropped, because the index does not otherwise keep them,
// and ESLint 10.8.1 reports an unknown name in `eslint-enable` exactly as it does in a disable.
func TestRuleReferencesNameEveryRuleEveryDirectiveWrote(t *testing.T) {
	t.Parallel()
	lines := []string{
		"// eslint-disable-next-line no-unused-var -- why",
		"const a = 1; // eslint-disable-line no-consol, no-console",
		"/* eslint-disable no-such-rule */",
		"/* eslint-enable no-such-rule */",
		"// eslint-disable-next-line",
		"const text = '// eslint-disable-next-line inside-a-string';",
		// Not directives: a file-scope disable or an enable counts only as a block comment, ESLint's
		// grammar (directives.isLineComment). The phi web sentence is verbatim.
		"// eslint-disable + generated banner keep the linter and future readers out.",
		"// eslint-enable no-such-rule",
	}
	source := strings.Join(lines, "\n")

	got := []string{}
	for _, reference := range Build(source).RuleReferences() {
		got = append(got, reference.Name+"="+source[reference.Pos:reference.End])
	}
	want := []string{
		"no-unused-var=" + lines[0],
		"no-consol=// eslint-disable-line no-consol, no-console",
		"no-console=// eslint-disable-line no-consol, no-console",
		"no-such-rule=" + lines[2],
		"no-such-rule=" + lines[3],
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("rule references\n got: %q\nwant: %q", got, want)
	}
}
