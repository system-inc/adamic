package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
)

var checkedCastFixtures = []struct {
	name, source, target string
}{
	{"cast_enum_numeric_fails", "A | B", "A"},
	{"cast_enum_string_fails", "A | B", "A"},
	{"cast_enum_const_numeric_fails", "A | B", "A"},
	{"cast_enum_const_string_fails", "A | B", "A"},
	{"cast_subunion_fails", "A | B | C", "A | C"},
	{"cast_subunion_string_fails", "A | B | C", "A | C"},
	{"cast_subunion_const_numeric_fails", "A | B | C", "A | C"},
	{"cast_subunion_const_string_fails", "A | B | C", "A | C"},
	{"cast_class_fails", "Animal", "Dog"},
	{"cast_class_union_fails", "Bird | Cat | Dog", "Cat | Dog"},
	{"cast_class_generic_fails", "Base<number>", "Box<number>"},
}

func TestCheckedCastFlushesOutput(t *testing.T) {
	t.Parallel()
	program, _ := castFixture(t, "cast_enum_string_fails")
	text := strings.Repeat("preceding stdout\n", 20000)
	index := len(program.Strings)
	program.Strings = append(program.Strings, text)
	program.Main = append([]ir.Statement{ir.WriteLine{Stream: ir.Stdout, Value: ir.StringConstant{Index: index}}}, program.Main...)
	native, _ := natively(t, program)
	for backend, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if result.exitCode != 70 || string(result.stdout) != text+"\ncasting\n" || string(result.stderr) != "adamic: panic: cast failed: this A | B is not a A\n" {
			t.Errorf("%s: panic must flush %d bytes, got %d, exit %d stderr %q", backend, len(text)+len("\ncasting\n"), len(result.stdout), result.exitCode, result.stderr)
		}
	}
}

// This witness can compile to valid C if tag uniqueness is mistakenly trusted. In that mutant,
// source Node independently exposes the invalid payload interpretation; -Werror must not kill it.
func TestUncheckableCastAdmission(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	source := `enum Kind { A = 1, B = 1 }
type A = { readonly kind: Kind.A; readonly n: number };
type B = { readonly kind: Kind.B; readonly n: string };
function cast(value: A | B): A { return value as A; }
console.log(String(cast({ kind: Kind.B, n: 'wrong'.repeat(2) }).n + 1));
`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	var refused *lower.Refused
	if errors.As(err, &refused) && strings.Contains(err.Error(), "adamic/no-unchecked-cast") {
		return
	}
	if err != nil {
		t.Fatalf("want cast refusal, got %v", err)
	}
	result, _ := natively(t, program)
	node := onNode(t, path)
	t.Fatalf("uncheckable cast admitted: Node exit %d stdout %q; native exit %d stdout %q stderr %q; %s", node.exitCode, node.stdout, result.exitCode, result.stdout, result.stderr, disagreement(node, result))
}

func TestCheckedCastRunsNoCatchOrFinally(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"cast_enum_numeric_fails", "cast_class_fails"} {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			program, _ := castFixture(t, fixture)
			directory := t.TempDir()
			marker := filepath.Join(directory, "handler-ran")
			code := "import { writeFileSync } from 'node:fs';\n" + javascript.JavaScript(program)
			for _, handler := range []string{"caught", "finally"} {
				original := strconv.Quote(handler)
				if !strings.Contains(code, original) {
					t.Fatalf("missing %s instrumentation site", handler)
				}
				code = strings.ReplaceAll(code, original, "(writeFileSync("+strconv.Quote(marker)+", "+original+"), "+original+")")
			}
			path := filepath.Join(directory, "program.mjs")
			if err := os.WriteFile(path, []byte(code), 0644); err != nil {
				t.Fatal(err)
			}
			result := onNode(t, path)
			if result.exitCode != 70 || string(result.stdout) != "casting\n" {
				t.Fatalf("panic failed: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("catch or finally ran after cast panic: %v", err)
			}
		})
	}
}

func init() {
	for _, name := range []string{"cast_enum_numeric", "cast_enum_string", "cast_enum_const", "cast_class", "cast_operand_once"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name + ".a", true, false})
	}
	for _, fixture := range checkedCastFixtures {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + fixture.name + ".a", true, true})
	}
}

func castFixture(t *testing.T, name string) (*ir.Program, string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return program, path
}

// Pin the externally decided contract as well as backend agreement: two backends can agree on
// the same wrong message, catchable failure, or missing stdout flush.
func TestCheckedCastFailureContract(t *testing.T) {
	t.Parallel()
	for _, fixture := range checkedCastFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			program, path := castFixture(t, fixture.name)
			want := "adamic: panic: cast failed: this " + fixture.source + " is not a " + fixture.target + "\n"
			native, _ := natively(t, program)
			for backend, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
				if result.exitCode != 70 || string(result.stdout) != "casting\n" || string(result.stderr) != want {
					t.Errorf("%s: exit %d, stdout %q, stderr %q; want panic/70 and %q", backend, result.exitCode, result.stdout, result.stderr, want)
				}
			}
			if node := onNode(t, path); node.exitCode != 0 {
				t.Fatalf("source on Node must run on without a check: %d %s", node.exitCode, node.stderr)
			}
		})
	}
}

// Mutations of native input only are held against the unmodified JavaScript backend for failures
// and against original source on Node for successful programs. Every mutant must build valid C.
func TestCheckedCastMutants(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, fixture string }{
		{"skip numeric", "cast_enum_numeric_fails"},
		{"skip string", "cast_enum_string_fails"},
		{"skip const numeric", "cast_enum_const_numeric_fails"},
		{"skip const string", "cast_enum_const_string_fails"},
		{"skip sub-union", "cast_subunion_fails"},
		{"skip string sub-union", "cast_subunion_string_fails"},
		{"skip const numeric sub-union", "cast_subunion_const_numeric_fails"},
		{"skip const string sub-union", "cast_subunion_const_string_fails"},
		{"skip class", "cast_class_fails"},
		{"skip class union", "cast_class_union_fails"},
		{"skip generic class", "cast_class_generic_fails"},
		{"wrong numeric tag", "cast_enum_numeric"},
		{"wrong string tag", "cast_enum_string"},
		{"wrong class identity", "cast_class_fails"},
		{"twice tag operand", "cast_operand_once"},
		{"twice class operand", "cast_operand_once"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			program, path := castFixture(t, test.fixture)
			want := onNode(t, path)
			if strings.HasPrefix(test.name, "skip") || test.name == "wrong class identity" {
				want = onJavaScriptBackend(t, program)
			}
			changed := false
			mutate := func(value ir.Expression) ir.Expression {
				if changed {
					return value
				}
				switch node := value.(type) {
				case ir.CheckedCast:
					if strings.HasPrefix(test.name, "skip") {
						changed = true
						return node.Value
					}
					if test.name == "wrong numeric tag" {
						node.Allowed = []ir.Expression{ir.NumberConstant{Value: 99}}
						changed = true
						return node
					}
					if test.name == "wrong string tag" {
						node.Allowed = []ir.Expression{ir.StringConstant{Index: len(program.Strings)}}
						program.Strings = append(program.Strings, "wrong")
						changed = true
						return node
					}
					if test.name == "twice tag operand" {
						changed = true
						return ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: node.Value, Right: ir.Undefined{Of: ir.Object}}, WhenTrue: node, WhenNot: node, Of: ir.Object}
					}
				case ir.InstanceOf:
					if strings.HasPrefix(test.name, "skip class") || test.name == "skip generic class" {
						changed = true
						return ir.BooleanConstant{Value: true}
					}
					if test.name == "wrong class identity" {
						for index, class := range program.Classes {
							if class.Name == "Animal" {
								node.Class = index + 1
								changed = true
								return node
							}
						}
					}
				case ir.Call:
					if test.name == "twice class operand" && program.Functions[node.Function].Name == "checked_class_cast" {
						changed = true
						return ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: node.Arguments[0], Right: ir.Undefined{Of: ir.Object}}, WhenTrue: node, WhenNot: node, Of: ir.Object}
					}
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if !changed {
				t.Fatal("mutant changed no cast")
			}
			result, binary := natively(t, program)
			if strings.Contains(string(result.stderr), "Sanitizer") {
				t.Fatalf("mutant must be caught by semantic comparison: %s", result.stderr)
			}
			if difference := disagreement(want, result); difference == "" {
				t.Fatal("mutant not caught")
			} else {
				t.Logf("caught: %s; expected exit %d stdout %q, mutant exit %d stdout %q", difference, want.exitCode, want.stdout, result.exitCode, result.stdout)
			}
			if result.exitCode == 0 {
				if report := leaks(t, program, binary); report != "" {
					t.Fatalf("successful mutant must be leak-clean: %s", report)
				}
			}
		})
	}
}
