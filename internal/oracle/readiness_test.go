package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"testing"
)

func init() {
	for _, name := range []string{"initialized", "uninitialized", "uninitialized_field", "uninitialized_capture", "uninitialized_exception", "uninitialized_loop", "uninitialized_default", "uninitialized_const", "weak", "definite", "definite_local", "definite_field", "static_initialized", "static_uninitialized", "uninitialized_optional", "uninitialized_spread", "literal_return", "literal_assignment", "literal_statement", "uninitialized_iteration", "uninitialized_catch", "uninitialized_interface", "uninitialized_append", "uninitialized_map_entry", "lazy_initialized", "lazy_read", "lazy_field", "lazy_static", "lazy_default"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/non_null_" + name + ".ts", true, name != "initialized" && name != "definite" && name != "static_initialized" && name != "uninitialized_optional"})
	}
}

// These expected outcomes are independent of both emitters, so a shared lowering mistake
// cannot turn a green native-versus-JavaScript comparison into a false proof.
func TestReadinessMutants(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, fixture, stdout, stderr string }{
		{"replace-unset-check-with-zero", "uninitialized", "before\n", "placeholder 'value' is unset at use as T via value"},
		{"erase-without-proof", "uninitialized_loop", "", "placeholder 'value' is unset at use as T via value"},
		{"initialize-to-zero", "uninitialized", "before\n", "placeholder 'value' is unset at use as T via value"},
		{"miss-captured-read", "uninitialized_capture", "before\n", "placeholder 'value' is unset at use as T via value"},
		{"miss-exception-path", "uninitialized_exception", "caught\n", "placeholder 'value' is unset at use as T via value"},
		{"replace-missing-assertion-with-empty", "lazy_read", "", "non-null assertion failed: textInitial! is null or undefined"},
		{"weak-generic-message", "weak", "before\n", "non-null assertion failed: holder.value! is null or undefined"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_"+probe.fixture+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			if probe.fixture == "lazy_read" || probe.fixture == "weak" {
				expression := "textInitial!"
				if probe.fixture == "weak" {
					expression = "holder.value!"
				}
				assertMigratedNonNullCheck(t, path, expression, probe.stdout, true)
				return
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(probe.stdout), stderr: []byte("adamic: panic: " + probe.stderr + "\n"), exitCode: 70}
			baseline, _ := nativelyUncached(t, program)
			if difference := disagreement(want, baseline); difference != "" {
				t.Fatal(difference)
			}
			if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "" {
				t.Fatal(difference)
			}
			changes := 0
			program.Main = mutateReadiness(program.Main, func(node any) any {
				switch value := node.(type) {
				case ir.Coalesce:
					if probe.name == "replace-missing-assertion-with-empty" && value.Panic != nil && value.Value.Type() == ir.String {
						index := len(program.Strings)
						program.Strings = append(program.Strings, "")
						changes++
						value.Panic = nil
						value.Fallback = ir.StringConstant{Index: index}
						return value
					}
					if probe.name != "initialize-to-zero" && probe.name != "weak-generic-message" && value.Panic != nil {
						changes++
						if value.Value.Type() == ir.Union {
							value.Panic = nil
							value.Fallback = ir.Box{Value: ir.NumberConstant{}}
							return value
						}
						if value.Value.Type().IsMaybe() {
							return ir.Unwrap{Value: value.Value, Proven: true}
						}
						return value.Value
					}
				case ir.Read:
					if probe.name != "initialize-to-zero" && probe.name != "weak-generic-message" && value.Readiness != "" {
						value.Readiness = ""
						changes++
						return value
					}
				case ir.Declare:
					if probe.name == "initialize-to-zero" && value.Uninitialized {
						value.Uninitialized = false
						if program.Locals[value.Local].Type == ir.Union {
							value.Value = ir.Box{Value: ir.NumberConstant{}}
						}
						if program.Locals[value.Local].Type.IsMaybe() {
							value.Value = ir.MaybeOf{Value: ir.NumberConstant{}, Of: program.Locals[value.Local].Type}
						}
						changes++
						return value
					}
				case ir.WeakTarget:
					if probe.name == "weak-generic-message" && !value.Present {
						value.Present = true
						changes++
						return value
					}
				}
				return node
			})
			for i := range program.Functions {
				program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(node any) any {
					switch value := node.(type) {
					case ir.Coalesce:
						if probe.name == "replace-missing-assertion-with-empty" && value.Panic != nil && value.Value.Type() == ir.String {
							index := len(program.Strings)
							program.Strings = append(program.Strings, "")
							changes++
							value.Panic = nil
							value.Fallback = ir.StringConstant{Index: index}
							return value
						}
						if probe.name != "initialize-to-zero" && probe.name != "weak-generic-message" && value.Panic != nil {
							changes++
							if value.Value.Type() == ir.Union {
								value.Panic = nil
								value.Fallback = ir.Box{Value: ir.NumberConstant{}}
								return value
							}
							if value.Value.Type().IsMaybe() {
								return ir.Unwrap{Value: value.Value, Proven: true}
							}
							return value.Value
						}
					case ir.Read:
						if probe.name != "initialize-to-zero" && probe.name != "weak-generic-message" && value.Readiness != "" {
							value.Readiness = ""
							changes++
							return value
						}
					case ir.Declare:
						if probe.name == "initialize-to-zero" && value.Uninitialized {
							value.Uninitialized = false
							if program.Locals[value.Local].Type == ir.Union {
								value.Value = ir.Box{Value: ir.NumberConstant{}}
							}
							if program.Locals[value.Local].Type.IsMaybe() {
								value.Value = ir.MaybeOf{Value: ir.NumberConstant{}, Of: program.Locals[value.Local].Type}
							}
							changes++
							return value
						}
					}
					return node
				})
			}
			if changes == 0 {
				t.Fatal("mutant changed nothing")
			}
			mutant, _ := nativelyUncached(t, program)
			if disagreement(want, mutant) == "" {
				t.Fatal("mutant escaped pinned assertion")
			}
			if mutant.exitCode != 0 && mutant.exitCode != 70 {
				t.Fatalf("mutant caught only by a crash: exit %d stderr %s", mutant.exitCode, mutant.stderr)
			}
			t.Logf("caught by pinned output: exit %d stdout %q stderr %q", mutant.exitCode, mutant.stdout, mutant.stderr)
		})
	}
}

func TestUninitializedIsNotNullishMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_initialized.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	baseline, _ := nativelyUncached(t, program)
	if difference := disagreement(want, baseline); difference != "" {
		t.Fatal(difference)
	}
	message := len(program.Strings)
	program.Strings = append(program.Strings, "non-null assertion failed: undefined! is null or undefined")
	changes := 0
	for i := range program.Functions {
		program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(node any) any {
			if declaration, ok := node.(ir.Declare); ok && declaration.Uninitialized && program.Locals[declaration.Local].Placeholder != "" && program.Locals[declaration.Local].Type == ir.Union && changes == 0 {
				declaration.Uninitialized = false
				declaration.Value = fitMutantNumber(ir.Coalesce{Value: ir.MaybeOf{Of: ir.MaybeNumber}, Panic: ir.StringConstant{Index: message}, Of: ir.Number}, program.Locals[declaration.Local].Type)
				changes++
				return declaration
			}
			return node
		})
	}
	if changes != 1 {
		t.Fatal("ordinary-nullish mutant did not replace one initializer")
	}
	mutant, _ := nativelyUncached(t, program)
	if mutant.exitCode != 70 || disagreement(want, mutant) == "" {
		t.Fatalf("ordinary-nullish mutant escaped Node: %#v", mutant)
	}
	t.Logf("ordinary-nullish initializer mutant caught by Node output: %s", mutant.stderr)
}

func mutateReadiness(statements []ir.Statement, mutate func(any) any) []ir.Statement {
	var transform func(reflect.Value) reflect.Value
	transform = func(value reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			mapped := transform(value.Elem())
			result := reflect.New(value.Type()).Elem()
			result.Set(reflect.ValueOf(mutate(mapped.Interface())))
			return result
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(transform(value.Field(i)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(transform(value.Index(i)))
			}
			return result
		}
		return value
	}
	return transform(reflect.ValueOf(statements)).Interface().([]ir.Statement)
}

// Native Weak lifetime is already ruled to differ from Node's tracing collector.
// Pin both observations, including this assertion's expression-specific native diagnostic.
func TestNonNullWeakFreedNamesExpression(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_weak_freed.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	node := run{stdout: []byte("before\nlivelive\n")}
	for _, got := range []run{onNode(t, path), onJavaScriptBackend(t, program)} {
		if difference := disagreement(node, got); difference != "" {
			t.Fatal(difference)
		}
	}
	want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: non-null assertion failed at " + path + ":9:17: holder.value! is null or undefined\n"), exitCode: 70}
	actual, _ := nativelyUncached(t, program)
	if difference := disagreement(want, actual); difference != "" {
		t.Fatal(difference)
	}
	changes := 0
	program.Main = mutateReadiness(program.Main, func(node any) any {
		if target, ok := node.(ir.WeakTarget); ok && !target.Present {
			target.Present = true
			changes++
			return target
		}
		return node
	})
	if changes == 0 {
		t.Fatal("Weak diagnostic mutant changed nothing")
	}
	mutant, _ := nativelyUncached(t, program)
	if mutant.exitCode != 70 || disagreement(want, mutant) == "" {
		t.Fatal("freed Weak diagnostic mutant escaped pinned output")
	}
	t.Logf("freed Weak diagnostic mutant caught: %s", mutant.stderr)
}

// Non-literal assertions are checked at initialization, even if a later write would replace them.
func TestNonliteralInitializerCannotSkipCheck(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_lazy_initialized.ts"))
	if err != nil {
		t.Fatal(err)
	}
	assertMigratedNonNullCheck(t, path, "textInitial!", "", true)
}

func fitMutantNumber(value ir.Expression, of ir.Type) ir.Expression {
	if of == ir.Union {
		return ir.Box{Value: value}
	}
	if of.IsMaybe() {
		return ir.MaybeOf{Value: value, Of: of}
	}
	return value
}
