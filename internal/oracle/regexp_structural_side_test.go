package oracle

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

var regexpStructuralSideRows = []struct{ name, stdout, diagnostic string }{
	{"copy", "TypeError: Method RegExp.prototype.exec called on incompatible receiver #<Object>\n", ":3:7: error TS2375: Type '{ test: ((text: string) => boolean) | undefined; }' is not assignable to type 'Maybe' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'test' are incompatible.\n    Type '((text: string) => boolean) | undefined' is not assignable to type '(text: string) => boolean'.\n      Type 'undefined' is not assignable to type '(text: string) => boolean'."},
	{"call", "true\n", ":3:20: stage 0 can't lower a call through ?. (an optional call) yet"},
	{"bind", "true\n", ":3:20: stage 0 can't lower a call through ?. (an optional call) yet"},
	{"apply", "true\n", ":3:20: stage 0 can't lower a call through ?. (an optional call) yet"},
	{"typeof", "function\n", ":3:20: stage 0 can't lower a presence test of optional member test whose receiver may hold a value the program didn't build as a plain object literal (no reaching allocation established; untracked source ir.RegExpNew) yet"},
	{"identity", "true\n", ""},
	{"optional_last_index", "true\nfalse\ntrue\n", ""},
	{"required_last_index", "true\nfalse\ntrue\n", ""},
	{"array", "true\n", ":3:13: stage 0 can't lower an array of boolean | undefined yet"},
	{"generic", "true\n", ""},
	{"narrowed", "true\n", ""},
	{"impostor", "undefined\nno test\n", ":2:20: error TS2353: Object literal may only specify known properties, and 'source' does not exist in type 'Maybe'."},
	{"copy_narrowed", "TypeError: Method RegExp.prototype.exec called on incompatible receiver #<Object>\n", ""},
	{"impostor_undefined", "undefined\nno test\n", ":3:18: Adamic 0.1 refuses a value of type { source: string; flags: string; lastIndex: number; test: undefined; } seen as Maybe, which can write ((text: string) => boolean) | undefined where undefined is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"},
	{"array_string", "true\n", ""},
	{"impostor_readonly", "undefined\nno test\n", ""},
	{"required_bound", "true\n", ""},
}

func init() {
	for _, row := range regexpStructuralSideRows {
		if row.diagnostic == "" {
			fixtures = append(fixtures, struct {
				path            string
				lowers, checked bool
			}{regexpStructuralSidePath(row.name), true, false})
		}
	}
}

func regexpStructuralSidePath(name string) string {
	return "internal/oracle/testdata/regexp_structural_side/" + name + ".a"
}

// Each road is held to Node, or to its exact checker or lowering stop. The raw source runner
// also observes checker-rejected programs: the ordinary cached Node runner checks their imports.
// Removing the copied-function receiver arm reds copy_narrowed; a fresh wrapper on every read
// reds identity, and omitting the optional-read wrapper reds the existing detached witness.
func TestRegExpStructuralSideRoads(t *testing.T) {
	t.Parallel()
	for _, row := range regexpStructuralSideRows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, regexpStructuralSidePath(row.name)))
			if err != nil {
				t.Fatal(err)
			}
			truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			if difference := disagreement(run{stdout: []byte(row.stdout)}, truth); difference != "" {
				t.Fatalf("Node: %s: %+v", difference, truth)
			}
			checked, err := load.Load([]string{path})
			if row.diagnostic != "" {
				if err == nil {
					_, err = lower.Lower(context.Background(), checked)
				}
				if err == nil || err.Error() != path+row.diagnostic {
					t.Fatalf("want exact stop %q, got %v", row.diagnostic, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			native, binary := nativelyUncached(t, program)
			for name, got := range map[string]run{"sanitized": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: %s: exit %d stdout %q stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Error(report)
			}
		})
	}
}

func TestRegExpClosureEntryStaysDirect(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile(filepath.Join(repository, "internal/javascript/javascript.go"))
	if err != nil {
		t.Fatal(err)
	}
	const direct = "const adamicCall = (closure, values) => closure.code(closure, values);"
	if !strings.Contains(string(source), direct) {
		t.Fatal("the ordinary closure entry must not inspect RegExp.prototype.test")
	}
}
