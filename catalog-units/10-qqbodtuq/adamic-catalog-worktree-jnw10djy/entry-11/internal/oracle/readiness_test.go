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
		}{nonNullRuntimeFixture(name), true, name != "definite"})
	}
}

func nonNullRuntimeFixture(name string) string {
	return "internal/oracle/testdata/non_null_" + name + ".ts"
}

// The unchanged programs now run as checked TypeScript. Initializer assertions
// stop eagerly; these pins keep their runtime checks covered.
func TestReadinessMutants(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, fixture, stdout, stderr string }{
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
			path, err := filepath.Abs(filepath.Join(repository, nonNullRuntimeFixture(probe.fixture)))
			if err != nil {
				t.Fatal(err)
			}
			expression, stdout := "undefined!", ""
			if probe.fixture == "uninitialized_loop" {
				expression = "null!"
			}
			if probe.fixture == "lazy_read" {
				expression = "textInitial!"
			}
			if probe.fixture == "weak" {
				expression, stdout = "holder.value!", "before\n"
			}
			assertMigratedNonNullCheck(t, path, expression, stdout, true)
		})
	}
}

func TestUninitializedIsNotNullishMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_initialized.ts"))
	if err != nil {
		t.Fatal(err)
	}
	assertMigratedNonNullCheck(t, path, "undefined!", "", true)
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

// Native Weak lifetime differs from Node tracing; keep both original observations.
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

func TestLazyInitializerIsNotEagerMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_lazy_initialized.ts"))
	if err != nil {
		t.Fatal(err)
	}
	assertMigratedNonNullCheck(t, path, "textInitial!", "", true)
}
