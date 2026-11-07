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

// Each fixture is named by its file. One that casts by a tag is .ts, since its tag test is a check no proof erases
// (item 8); one that casts only between classes stays .a.
var checkedCastFixtures = []struct {
	file, source, target string
}{
	{"cast_enum_numeric_fails.ts", "A | B", "A"},
	{"cast_enum_string_fails.ts", "A | B", "A"},
	{"cast_enum_const_numeric_fails.ts", "A | B", "A"},
	{"cast_enum_const_string_fails.ts", "A | B", "A"},
	{"cast_subunion_fails.ts", "A | B | C", "A | C"},
	{"cast_subunion_string_fails.ts", "A | B | C", "A | C"},
	{"cast_subunion_const_numeric_fails.ts", "A | B | C", "A | C"},
	{"cast_subunion_const_string_fails.ts", "A | B | C", "A | C"},
	{"cast_class_fails.a", "Animal", "Dog"},
	{"cast_class_union_fails.a", "Bird | Cat | Dog", "Cat | Dog"},
	{"cast_class_generic_fails.a", "Base<number>", "Box<number>"},
}

func TestCheckedCastFlushesOutput(test *testing.T) {
	test.Parallel()
	program, _ := castFixture(test, "cast_enum_string_fails.ts")
	text := strings.Repeat("preceding stdout\n", 20000)
	index := len(program.Strings)
	program.Strings = append(program.Strings, text)
	program.Main = append(
		[]ir.Statement{ir.WriteLine{Stream: ir.Stdout, Value: ir.StringConstant{Index: index}}},
		program.Main...,
	)
	native, _ := natively(test, program)
	for backend, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(test, program)} {
		if result.exitCode != 70 ||
			string(result.stdout) != text+"\ncasting\n" ||
			string(result.stderr) != "adamic: panic: cast failed: this A | B is not a A\n" {
			test.Errorf(
				"%s: panic must flush %d bytes, got %d, exit %d stderr %q",
				backend,
				len(text)+len("\ncasting\n"),
				len(result.stdout),
				result.exitCode,
				result.stderr,
			)
		}
	}
}

// This witness can compile to valid C if tag uniqueness is mistakenly trusted. In that mutant,
// source Node independently exposes the invalid payload interpretation; -Werror must not kill it.
func TestUncheckableCastAdmission(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	source := `enum Kind { A = 1, B = 1 }
type A = { readonly kind: Kind.A; readonly n: number };
type B = { readonly kind: Kind.B; readonly n: string };
function cast(value: A | B): A { return value as A; }
console.log(String(cast({ kind: Kind.B, n: 'wrong'.repeat(2) }).n + 1));
`
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loweredError := lowered(test, path)
	var refused *lower.Refused
	if errors.As(loweredError, &refused) && strings.Contains(loweredError.Error(), "adamic/no-unchecked-cast") {
		return
	}
	if loweredError != nil {
		test.Fatalf("want cast refusal, got %v", loweredError)
	}
	result, _ := natively(test, program)
	node := onNode(test, path)
	test.Fatalf(
		"uncheckable cast admitted: Node exit %d stdout %q; native exit %d stdout %q stderr %q; %s",
		node.exitCode,
		node.stdout,
		result.exitCode,
		result.stdout,
		result.stderr,
		disagreement(node, result),
	)
}

func TestCheckedCastRunsNoCatchOrFinally(test *testing.T) {
	test.Parallel()
	for _, fixture := range []string{"cast_enum_numeric_fails.ts", "cast_class_fails.a"} {
		test.Run(fixture, func(test *testing.T) {
			test.Parallel()
			program, _ := castFixture(test, fixture)
			directory := test.TempDir()
			marker := filepath.Join(directory, "handler-ran")
			code := "import { writeFileSync } from 'node:fs';\n" + javascript.JavaScript(program)
			for _, handler := range []string{"caught", "finally"} {
				original := strconv.Quote(handler)
				if !strings.Contains(code, original) {
					test.Fatalf("missing %s instrumentation site", handler)
				}
				code = strings.ReplaceAll(
					code,
					original,
					"(writeFileSync("+strconv.Quote(marker)+", "+original+"), "+original+")",
				)
			}
			path := filepath.Join(directory, "program.mjs")
			if writeError := os.WriteFile(path, []byte(code), 0644); writeError != nil {
				test.Fatal(writeError)
			}
			result := onNode(test, path)
			if result.exitCode != 70 || string(result.stdout) != "casting\n" {
				test.Fatalf("panic failed: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
			}
			if _, statError := os.Stat(marker); !os.IsNotExist(statError) {
				test.Fatalf("catch or finally ran after cast panic: %v", statError)
			}
		})
	}
}

func init() {
	for _, file := range []string{
		"cast_enum_numeric.ts",
		"cast_enum_string.ts",
		"cast_enum_const.ts",
		"cast_class.a",
		"cast_operand_once.ts",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + file, true, false})
	}
	for _, fixture := range checkedCastFixtures {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + fixture.file, true, true})
	}
}

func castFixture(test *testing.T, file string) (*ir.Program, string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", file))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	return program, path
}

// Pin the externally decided contract as well as backend agreement: two backends can agree on
// the same wrong message, catchable failure, or missing stdout flush.
func TestCheckedCastFailureContract(test *testing.T) {
	test.Parallel()
	for _, fixture := range checkedCastFixtures {
		test.Run(fixture.file, func(test *testing.T) {
			test.Parallel()
			program, path := castFixture(test, fixture.file)
			want := "adamic: panic: cast failed: this " + fixture.source + " is not a " + fixture.target + "\n"
			native, _ := natively(test, program)
			for backend, result := range map[string]run{
				"native":     native,
				"JavaScript": onJavaScriptBackend(test, program),
			} {
				if result.exitCode != 70 || string(result.stdout) != "casting\n" || string(result.stderr) != want {
					test.Errorf(
						"%s: exit %d, stdout %q, stderr %q; want panic/70 and %q",
						backend,
						result.exitCode,
						result.stdout,
						result.stderr,
						want,
					)
				}
			}
			if node := onNode(test, path); node.exitCode != 0 {
				test.Fatalf("source on Node must run on without a check: %d %s", node.exitCode, node.stderr)
			}
		})
	}
}

// Mutations of native input only are held against the unmodified JavaScript backend for failures
// and against original source on Node for successful programs. Every mutant must build valid C.
func TestCheckedCastMutants(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct{ name, fixture string }{
		{"skip numeric", "cast_enum_numeric_fails.ts"},
		{"skip string", "cast_enum_string_fails.ts"},
		{"skip const numeric", "cast_enum_const_numeric_fails.ts"},
		{"skip const string", "cast_enum_const_string_fails.ts"},
		{"skip sub-union", "cast_subunion_fails.ts"},
		{"skip string sub-union", "cast_subunion_string_fails.ts"},
		{"skip const numeric sub-union", "cast_subunion_const_numeric_fails.ts"},
		{"skip const string sub-union", "cast_subunion_const_string_fails.ts"},
		{"skip class", "cast_class_fails.a"},
		{"skip class union", "cast_class_union_fails.a"},
		{"skip generic class", "cast_class_generic_fails.a"},
		{"wrong numeric tag", "cast_enum_numeric.ts"},
		{"wrong string tag", "cast_enum_string.ts"},
		{"wrong class identity", "cast_class_fails.a"},
		{"twice tag operand", "cast_operand_once.ts"},
		{"twice class operand", "cast_operand_once.ts"},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			program, path := castFixture(test, testCase.fixture)
			want := onNode(test, path)
			if strings.HasPrefix(testCase.name, "skip") || testCase.name == "wrong class identity" {
				want = onJavaScriptBackend(test, program)
			}
			changed := false
			mutate := func(value ir.Expression) ir.Expression {
				if changed {
					return value
				}
				switch node := value.(type) {
				case ir.CheckedCast:
					if strings.HasPrefix(testCase.name, "skip") {
						changed = true
						return node.Value
					}
					if testCase.name == "wrong numeric tag" {
						node.Allowed = []ir.Expression{ir.NumberConstant{Value: 99}}
						changed = true
						return node
					}
					if testCase.name == "wrong string tag" {
						node.Allowed = []ir.Expression{ir.StringConstant{Index: len(program.Strings)}}
						program.Strings = append(program.Strings, "wrong")
						changed = true
						return node
					}
					if testCase.name == "twice tag operand" {
						changed = true
						return ir.Conditional{Condition: ir.Binary{
							Operator: ir.Equal,
							Left:     node.Value,
							Right:    ir.Undefined{Of: ir.Object},
						}, WhenTrue: node, WhenNot: node, Of: ir.Object}
					}
				case ir.InstanceOf:
					if strings.HasPrefix(testCase.name, "skip class") || testCase.name == "skip generic class" {
						changed = true
						return ir.BooleanConstant{Value: true}
					}
					if testCase.name == "wrong class identity" {
						for index, class := range program.Classes {
							if class.Name == "Animal" {
								node.Class = index + 1
								changed = true
								return node
							}
						}
					}
				case ir.Call:
					if testCase.name == "twice class operand" &&
						program.Functions[node.Function].Name == "checked_class_cast" {
						changed = true
						return ir.Conditional{Condition: ir.Binary{
							Operator: ir.Equal,
							Left:     node.Arguments[0],
							Right:    ir.Undefined{Of: ir.Object},
						}, WhenTrue: node, WhenNot: node, Of: ir.Object}
					}
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if !changed {
				test.Fatal("mutant changed no cast")
			}
			result, binary := natively(test, program)
			if strings.Contains(string(result.stderr), "Sanitizer") {
				test.Fatalf("mutant must be caught by semantic comparison: %s", result.stderr)
			}
			if difference := disagreement(want, result); difference == "" {
				test.Fatal("mutant not caught")
			} else {
				test.Logf(
					"caught: %s; expected exit %d stdout %q, mutant exit %d stdout %q",
					difference,
					want.exitCode,
					want.stdout,
					result.exitCode,
					result.stdout,
				)
			}
			if result.exitCode == 0 {
				if report := leaks(test, program, binary); report != "" {
					test.Fatalf("successful mutant must be leak-clean: %s", report)
				}
			}
		})
	}
}
