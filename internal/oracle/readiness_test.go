package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"testing"
)

func init() {
	for _, probe := range []struct {
		name    string
		checked bool
	}{
		{"deinitialize_entries", true},
		{"deinitialize_static", true},
		{"deinitialize_shadowed", false},
		{"deinitialize_initialized", false},
		{"deinitialize_read", true},
		{"deinitialize_field", true},
		{"deinitialize_loop", false},
		{"deinitialize_capture", true},
		{"deinitialize_exception", true},
		{"deinitialize_alias", true},
		{"deinitialize_parameter", true},
		{"deinitialize_values", true},
		{"deinitialize_assign", false},
		{"deinitialize_assign_read", true},
		{"deinitialize_field_initialized", false},
		{"deinitialize_store", true},
		{"deinitialize_eager", true},
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/non_null_" + probe.name + ".ts", true, probe.checked})
	}
	for _, name := range []string{"initialized", "uninitialized", "uninitialized_field", "uninitialized_capture", "uninitialized_exception", "uninitialized_loop", "uninitialized_default", "uninitialized_const", "weak", "definite", "definite_local", "definite_field", "static_initialized", "static_uninitialized", "uninitialized_optional", "uninitialized_spread", "literal_return", "literal_assignment", "literal_statement", "uninitialized_iteration", "uninitialized_catch", "uninitialized_interface", "uninitialized_append", "uninitialized_map_entry", "lazy_initialized", "lazy_read", "lazy_field", "lazy_static", "lazy_default", "scanner_initialized", "scanner_var", "scanner_var_uninitialized"} {
		extension := ".ts"
		if name == "scanner_var" || name == "scanner_var_uninitialized" {
			extension = ".a"
		}
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/non_null_" + name + extension, true, name != "initialized" && name != "definite" && name != "static_initialized" && name != "lazy_initialized" && name != "scanner_initialized" && name != "scanner_var"})
	}
}

// These expected outcomes are independent of both emitters, so a shared lowering mistake
// cannot turn a green native-versus-JavaScript comparison into a false proof.
func TestReadinessMutants(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, fixture, stdout, stderr string }{
		{"scanner-var-capture", "scanner_var_uninitialized", "before\n", "read before assignment: variable 'tokenValue' in tokenValue"},
		{"deinitialization-values", "deinitialize_values", "before\n", "read before assignment: field 'value' in Object.values(box)"},
		{"deinitialization-entries", "deinitialize_entries", "before\n", "read before assignment: field 'value' in Object.entries(box)"},
		{"deinitialization-assign-source", "deinitialize_assign_read", "before\n", "read before assignment: field 'value' in Object.assign(target, source)"},
		{"deinitialization-field-alias", "deinitialize_alias", "before\n", "read before assignment: field 'value' in box.value"},
		{"deinitialization-field-method", "deinitialize_field", "before\n", "read before assignment: field 'value' in this.value"},
		{"keep-slot-proven-after-deinitialization", "deinitialize_read", "before\n", "read before assignment: variable 'value' in value"},
		{"deinitialization-through-capture", "deinitialize_capture", "before\n", "read before assignment: variable 'value' in value"},
		{"deinitialization-exception-path", "deinitialize_exception", "caught\n", "read before assignment: variable 'value' in value"},
		{"drop-check", "uninitialized", "before\n", "read before assignment: variable 'value' in value"},
		{"erase-without-proof", "uninitialized_loop", "", "read before assignment: variable 'value' in value"},
		{"initialize-to-zero", "uninitialized", "before\n", "read before assignment: variable 'value' in value"},
		{"miss-captured-read", "uninitialized_capture", "before\n", "read before assignment: variable 'value' in value"},
		{"miss-exception-path", "uninitialized_exception", "caught\n", "read before assignment: variable 'value' in value"},
		{"lazy-read", "lazy_read", "before\n", "read before assignment: variable 'text' in textInitial!"},
		{"weak-generic-message", "weak", "before\n", "non-null assertion failed: holder.value! is null or undefined"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			extension := ".ts"
			if probe.fixture == "scanner_var_uninitialized" {
				extension = ".a"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_"+probe.fixture+extension))
			if err != nil {
				t.Fatal(err)
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
				case ir.Read:
					if probe.name != "initialize-to-zero" && probe.name != "weak-generic-message" && value.Readiness != "" {
						value.Readiness = ""
						changes++
						return value
					}
				case ir.ObjectCall:
					if value.Readiness != "" {
						value.Readiness = ""
						changes++
						return value
					}
				case ir.Property:
					if probe.name != "initialize-to-zero" && probe.name != "weak-generic-message" && value.Readiness != "" {
						value.Readiness = ""
						changes++
						return value
					}
				case ir.Declare:
					if probe.name == "initialize-to-zero" && value.Uninitialized {
						value.Uninitialized = false
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
					case ir.Read:
						if probe.name != "initialize-to-zero" && probe.name != "weak-generic-message" && value.Readiness != "" {
							value.Readiness = ""
							changes++
							return value
						}
					case ir.ObjectCall:
						if value.Readiness != "" {
							value.Readiness = ""
							changes++
							return value
						}
					case ir.Property:
						if probe.name != "initialize-to-zero" && probe.name != "weak-generic-message" && value.Readiness != "" {
							value.Readiness = ""
							changes++
							return value
						}
					case ir.Declare:
						if probe.name == "initialize-to-zero" && value.Uninitialized {
							value.Uninitialized = false
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
			if declaration, ok := node.(ir.Declare); ok && declaration.Uninitialized && program.Locals[declaration.Local].Type == ir.Number && changes == 0 {
				declaration.Uninitialized = false
				declaration.Value = ir.Coalesce{Value: ir.MaybeOf{Of: ir.MaybeNumber}, Panic: ir.StringConstant{Index: message}, Of: ir.Number}
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
	want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: non-null assertion failed: holder.value! is null or undefined\n"), exitCode: 70}
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

func TestLazyInitializerIsNotEagerMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_lazy_initialized.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	baseline, _ := nativelyUncached(t, program)
	if difference := disagreement(node, baseline); difference != "" {
		t.Fatal(difference)
	}
	changed := false
	for fi := range program.Functions {
		body := program.Functions[fi].Body
		for i := 1; i < len(body); i++ {
			declaration, ok := body[i].(ir.Declare)
			if !ok || !declaration.Uninitialized || program.Locals[declaration.Local].InitializerExpression != "textInitial!" {
				continue
			}
			operand, ok := body[i-1].(ir.Declare)
			if !ok {
				t.Fatal("initializer operand was not held once")
			}
			index := len(program.Strings)
			program.Strings = append(program.Strings, "non-null assertion failed: textInitial! is null or undefined")
			declaration.Uninitialized = false
			declaration.Value = ir.Coalesce{Value: ir.Read{Local: operand.Local, Of: program.Locals[operand.Local].Type}, Panic: ir.StringConstant{Index: index}, Of: program.Locals[declaration.Local].Type}
			body[i] = declaration
			changed = true
		}
	}
	if !changed {
		t.Fatal("eager initializer mutant changed nothing")
	}
	mutant, _ := nativelyUncached(t, program)
	if mutant.exitCode != 70 {
		t.Fatalf("mutant failed without a runtime panic: %#v", mutant)
	}
	if disagreement(node, mutant) == "" {
		t.Fatal("eager initializer survived Node comparison")
	}
	t.Logf("eager initializer caught by Node output: %s", mutant.stderr)
}

func TestDeinitializationIsNotOrdinaryStoreMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(checkedViewFixturePath(filepath.Join(repository, "internal/oracle/testdata/non_null_deinitialize_store.ts")))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: read before assignment: variable 'text' in text\n"), exitCode: 70}
	baseline, _ := nativelyUncached(t, program)
	if difference := disagreement(want, baseline); difference != "" {
		t.Fatal(difference)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	changes := 0
	program.Main = mutateReadiness(program.Main, func(node any) any {
		if assign, ok := node.(ir.Assign); ok && assign.Uninitialized {
			assign.Uninitialized = false
			changes++
			return assign
		}
		return node
	})
	if changes != 1 {
		t.Fatalf("ordinary store mutant changed %d assignments", changes)
	}
	mutant := onJavaScriptBackend(t, program)
	node := onNode(t, path)
	if mutant.exitCode != 0 || disagreement(node, mutant) != "" || disagreement(want, mutant) == "" {
		t.Fatalf("ordinary undefined store survived pinned output: %#v", mutant)
	}
	t.Logf("ordinary undefined store caught by pinned output: exit %d stdout %q", mutant.exitCode, mutant.stdout)
}
