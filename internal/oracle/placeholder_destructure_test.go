package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A field that may still hold its placeholder (undefined!), taken out by destructuring, is a placeholder use
// as node.parent read into a const is (#jjf9ppe). Main read it with no check, so const { parent } = node and
// then parent.kind dereferenced null natively (UBSan, record.c) where Node throws a TypeError. Now each road
// that destructures a field, a declaration, a parameter and an assignment, checks it where it's read, and
// stops loudly on both backends with the placeholder's words, before the program prints past "before". A
// default over such a field is not yet: Node takes the default for the placeholder's undefined, where the
// check would stop. set_field, the same destructuring of a field that was set, agrees with Node. Mutant:
// placeholderDestructuredRead hands back the read unchecked, and the three loud rows run on unchecked again.
func TestPlaceholderDestructuredBindingIsChecked(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, panic, stop string }{
		{"const_binding.ts", "placeholder 'parent' is unset at destructured binding via parent", ""},
		{"parameter.ts", "placeholder 'parent' is unset at destructured binding via parent", ""},
		{"assignment.ts", "placeholder 'parent' is unset at destructured binding via parent", ""},
		{"defaulted.ts", "", "a destructuring default for field parent, which may still hold its placeholder"},
		{"set_field.ts", "", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/placeholder_destructure", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			program, err := lowered(t, path)
			if probe.stop != "" {
				var pending *lower.NotYet
				if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.stop) || !strings.Contains(err.Error(), probe.name) {
					t.Fatalf("want a not yet with %q, got %v", probe.stop, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			sanitized, binary := nativelyUncached(t, program)
			if probe.panic == "" {
				for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(node, got); difference != "" {
						t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr)
					}
				}
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			if node.exitCode == 0 || string(node.stdout) != "before\n" || !strings.Contains(string(node.stderr), "TypeError") {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			for index, got := range []run{sanitized, releasedUncached(t, program)} {
				if got.exitCode != 70 || string(got.stdout) != "before\n" || string(got.stderr) != "adamic: panic: "+probe.panic+"\n" {
					t.Fatalf("checked use on %s: exit %d, stdout %q, stderr %q", []string{"sanitized native", "release native"}[index], got.exitCode, got.stdout, got.stderr)
				}
			}
			if difference := disagreement(onJavaScriptBackend(t, program), sanitized); difference != "" {
				t.Fatalf("the JavaScript backend against native: %s", difference)
			}
		})
	}
}
