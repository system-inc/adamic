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
	{"call", "true\n", ":3:20: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"bind", "true\n", ":3:20: stage 0 can't lower a call through ?. (an optional call) yet"},
	{"apply", "true\n", ":3:20: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"typeof", "function\n", ":3:20: stage 0 can't lower a presence test of optional member test whose receiver may hold a value the program didn't build as a plain object literal (no reaching allocation established; untracked source ir.RegExpNew) yet"},
	{"identity", "true\n", ":3:20: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"optional_last_index", "true\nfalse\ntrue\n", ""},
	{"required_last_index", "true\nfalse\ntrue\n", ""},
	{"array", "true\n", ":3:13: stage 0 can't lower an array of boolean | undefined yet"},
	{"generic", "true\n", ""},
	{"narrowed", "true\n", ""},
	{"impostor", "undefined\nno test\n", ":2:20: error TS2353: Object literal may only specify known properties, and 'source' does not exist in type 'Maybe'."},
	{"copy_narrowed", "TypeError: Method RegExp.prototype.exec called on incompatible receiver #<Object>\n", ":3:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"impostor_undefined", "undefined\nno test\n", ":3:18: Adamic 0.1 refuses a value of type { source: string; flags: string; lastIndex: number; test: undefined; } seen as Maybe, which can write ((text: string) => boolean) | undefined where undefined is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"},
	{"array_string", "true\n", ""},
	{"impostor_readonly", "undefined\nno test\n", ""},
	{"required_bound", "true\n", ":3:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"subclass_required", "true\n", ":2:9: stage 0 can't lower a computed class base; name the base class directly yet"},
	{"subclass_optional", "has test\ntrue\n", ":2:9: stage 0 can't lower a computed class base; name the base class directly yet"},
	{"override_required", "false\n", ":2:9: stage 0 can't lower a computed class base; name the base class directly yet"},
	{"override_optional", "has test\nfalse\n", ":5:5: Adamic 0.1 refuses a value of type R seen as Maybe, whose method test becomes a writable field: a write of ((text: string) => boolean) | undefined would replace a method of type (_text: string) => boolean; make test readonly in the type it's seen as (readonly test: ...), or declare it there as a method, which isn't written (adamic/invariant-mutable)"},
	{"task_passed_maybe", "true,true\n", ":4:25: Adamic 0.1 refuses parallelMap items are not shareable: Maybe.test is a mutable field; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)"},
	{"task_captured_maybe", "true,true\n", ":5:46: Adamic 0.1 refuses task capture 'v' is not shareable: v.test is a mutable field; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)"},
	{"task_readonly_maybe", "true,true\n", ":5:46: Adamic 0.1 refuses task capture 'v' is not shareable: v.test is a function whose captures are not known; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)"},
	{"task_passed_tester", "true,true\n", ":4:45: Adamic 0.1 refuses parallelMap work writes state another task can reach: work calls a method whose complete dispatch hierarchy isn't proven; keep work and everything it calls pure; write only task locals and fresh objects"},
	{"task_captured_tester", "true,true\n", ":5:46: Adamic 0.1 refuses parallelMap work writes state another task can reach: work calls a method whose complete dispatch hierarchy isn't proven; keep work and everything it calls pure; write only task locals and fresh objects"},
	{"task_readonly_tester", "true,true\n", ":5:46: Adamic 0.1 refuses task capture 'v' is not shareable: v.test is a function whose captures are not known; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)"},
	{"task_actual", "true,true\n", ":6:44: Adamic 0.1 refuses task capture 'r' is not shareable: r has a type the compiler cannot see the whole of; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)"},
	{"task_passed_maybe_global", "true,true\n", ":4:25: Adamic 0.1 refuses parallelMap items are not shareable: Maybe.test is a mutable field; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)"},
	{"task_captured_maybe_global", "true,false\n", ":5:46: Adamic 0.1 refuses task capture 'v' is not shareable: v.test is a mutable field; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)"},
	{"task_readonly_maybe_global", "true,false\n", ":5:46: Adamic 0.1 refuses task capture 'v' is not shareable: v.test is a function whose captures are not known; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)"},
	{"task_passed_tester_global", "true,true\n", ":4:45: Adamic 0.1 refuses parallelMap work writes state another task can reach: work calls a method whose complete dispatch hierarchy isn't proven; keep work and everything it calls pure; write only task locals and fresh objects"},
	{"task_captured_tester_global", "true,false\n", ":5:46: Adamic 0.1 refuses parallelMap work writes state another task can reach: work calls a method whose complete dispatch hierarchy isn't proven; keep work and everything it calls pure; write only task locals and fresh objects"},
	{"task_readonly_tester_global", "true,false\n", ":5:46: Adamic 0.1 refuses task capture 'v' is not shareable: v.test is a function whose captures are not known; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)"},
	{"task_actual_global", "true,false\n", ":6:44: Adamic 0.1 refuses task capture 'r' is not shareable: r has a type the compiler cannot see the whole of; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)"},
	{"task_identity", "true,true\ntrue\ntrue\n", ""},
	{"task_identity_global", "true,true\ntrue\nfalse\n", ""},
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
// Subclass allocation and stateful task calls keep their existing named stops. Identity-only
// tasks carry the opaque value without reading lastIndex; main calls test after map completion.
// Copied, bound and identity reads now keep the served-view value stop; calls and task
// identity comparisons keep their measured behavior without handing out a prototype function.
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
