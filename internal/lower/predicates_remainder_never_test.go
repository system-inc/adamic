package lower

import (
	"embed"
	"strings"
	"testing"
)

//go:embed testdata/ruling12a/*.ts
var ruling12APrograms embed.FS

func ruling12ASource(t *testing.T, name string) string {
	t.Helper()
	content, err := ruling12APrograms.ReadFile("testdata/ruling12a/" + name + ".ts")
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestClosedNeverRemainderAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, stdout string }{
		{"a1", "a\n#b\n"},
		{"a3", "a\n#b\n1\n"},
		{"a4", "a\n#b\nlast\n"},
		{"a6", "a\n#b\n"},
		{"narrower-constant", "a\nfallback\n"},
		{"a-chain", "a\n#b\nother\n"},
		{"a-property", "a\n#b\n"},
		{"a-loop", "a\n#b\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, ruling12ASource(t, test.name), ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != test.stdout || stderr != "" {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want exit 0 and %q", code, stdout, stderr, test.stdout)
			}
			if test.name == "a6" {
				for _, site := range program.PredicateChecks.Sites {
					if site.Function == "isPrivateIdentifier" {
						for _, direction := range site.Directions {
							if direction.Direction == "false" && direction.Status == "unobservable" {
								return
							}
						}
					}
				}
				t.Fatal("the second predicate's false side must be unobservable")
			}
		})
	}
}

func TestClosedNeverRemainderWrongFalseStops(t *testing.T) {
	t.Parallel()
	overloadCheckStops(t, ruling12ASource(t, "a2"), "", "unreachable read of name at ", ":10:24, never after the check isPrivateIdentifier(name) at ")
}

func TestClosedNeverRemainderRefusesAccessorTag(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, ruling12ASource(t, "a5"), ".ts")
	if err == nil || !strings.Contains(err.Error(), "checked predicate membership over an open tag domain") {
		t.Fatalf("want the open tag domain refusal, got %v", err)
	}
}

// Ruling 12 A, Language's reading adopted at 20:44:38Z on #bbssdcy:
// fallback(name) is the false branch's first read of name and its stop.
// Node returns normally here, so this is a held stop.
func TestClosedNeverRemainderNarrowerPredicateStops(t *testing.T) {
	t.Parallel()
	source := ruling12ASource(t, "narrower")
	node := runAgreementNode(t, writeAgreementSource(t, "main.ts", source))
	if node.code != 0 || string(node.stdout) != "a\nfallback\n" || len(node.stderr) != 0 {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q; want exit 0, a then fallback and empty stderr", node.code, node.stdout, node.stderr)
	}
	overloadCheckStops(t, source, "a\n", "unreachable read of name at ", ":10:21, never after the check isPrivateIdentifier(name) at ", ":9:31")
}
