package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func init() {
	for _, entry := range []struct {
		path    string
		checked bool
	}{{"caught_sites", false}, {"caught_string_return", true}, {"caught_string_field", true}} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + entry.path + "/main.ts", true, entry.checked})
	}
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/catch_values.a", true, false}, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/caught_members/main.ts", true, false}, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/caught_string_check/main.ts", true, true})
}

// The project permits the expression, but a definite string use must not silently
// trust a missing message. Pin the named check independently of counts.
func TestCaughtMessageDefiniteStringCheck(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ fixture, stdout string }{{"caught_string_check", "[undefined]\nafter\n"}, {"caught_string_field", "undefined\nafter\n"}, {"caught_string_return", "undefined\nafter\n"}} {
		t.Run(probe.fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", probe.fixture, "main.ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			checks := ir.InsertedChecks(program)
			if len(checks) != 1 || checks[0].Kind != "caught-type" || !strings.Contains(checks[0].Where, "main.ts:") {
				t.Fatalf("definite catch use must report its guard site: %+v", checks)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != probe.stdout {
				t.Fatalf("Node observation: %+v", node)
			}
			native, _ := natively(t, program)
			backend := onJavaScriptBackend(t, program)
			for name, result := range map[string]run{"native": native, "javascript": backend} {
				if result.exitCode != 70 || string(result.stdout) != "" || string(result.stderr) != "adamic: panic: adamic/catch-type: caught value does not match its typed use\n" {
					t.Errorf("%s: exit %d stdout %q stderr %q", name, result.exitCode, result.stdout, result.stderr)
				}
			}
		})
	}
}

func TestCaughtValuesMutants(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"every-error", "definite-read", "unchecked-string-use"} {
		t.Run(kind, func(t *testing.T) {
			fixture := "caught_members"
			if kind == "unchecked-string-use" {
				fixture = "caught_string_check"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture, "main.ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := 0
			mutate := func(value ir.Expression) ir.Expression {
				switch node := value.(type) {
				case ir.InstanceOf:
					if kind == "every-error" && node.Class == ir.ErrorClass {
						changed++
						return ir.BooleanConstant{Value: true}
					}
				case ir.ObjectCall:
					if kind == "definite-read" && node.Method == "catchProperty" {
						changed++
						// Keep the carrier ABI unchanged. Only claim the read is definite.
						return ir.Box{Value: ir.Narrow{Value: node, To: ir.String, Checked: true}}
					}
				case ir.Narrow:
					if kind == "unchecked-string-use" && node.Checked {
						changed++
						node.Checked = false
						return node
					}
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if changed == 0 {
				t.Fatal("mutant changed nothing")
			}
			want := onNode(t, path)
			if kind == "unchecked-string-use" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: adamic/catch-type: caught value does not match its typed use\n")}
			}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if disagreement(want, got) == "" {
					t.Fatalf("%s mutant survived in %s", kind, name)
				}
				t.Logf("%s caught %s: exit %d stdout %q stderr %q", kind, name, got.exitCode, got.stdout, got.stderr)
			}
		})
	}
}

func TestCaughtMemberMayThrowMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/caught_sites/main.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for index := range program.Functions {
		if program.Functions[index].Name == "maybeMessage" {
			if !program.Functions[index].MayThrow {
				t.Fatal("nullish catch read must propagate MayThrow")
			}
			program.Functions[index].MayThrow = false
			changed = true
		}
	}
	if !changed {
		t.Fatal("mutant changed no function")
	}
	want := onNode(t, path)
	got, _ := natively(t, program)
	if disagreement(want, got) == "" {
		t.Fatal("uncatchable member failure mutant survived")
	}
	t.Logf("MayThrow under-approximation caught: exit %d stdout %q stderr %q", got.exitCode, got.stdout, got.stderr)
}

func TestCaughtNullTagMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/caught_members/main.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "&adamic_null", "NULL")
	if mutant == source {
		t.Fatal("mutant removed no null tag")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("null tag mutant must finish cleanly: exit %d stderr %q", got.exitCode, got.stderr)
	}
	if disagreement(onNode(t, path), got) != "stdout differs" {
		t.Fatal("dropping null's tag escaped the Node output comparison")
	}
	t.Logf("null tag mutant caught by Node stdout: %q", got.stdout)
}
