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
		}{"internal/oracle/testdata/non_null_" + name + ".a", true, name != "initialized" && name != "definite" && name != "static_initialized" && name != "lazy_initialized"})
	}
}

// Historical .a assertion fixtures now pin the refusal instead of a runtime check.
// Checked .ts mutants live in TestCheckedNonNullTypeScript.
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
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_"+probe.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			assertAdamicNonNullRefusal(t, err)
		})
	}
}

func TestUninitializedIsNotNullishMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_initialized.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	assertAdamicNonNullRefusal(t, err)
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

// The historical Weak assertion is refused in .a under the source-mode ruling.
func TestNonNullWeakFreedNamesExpression(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_weak_freed.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	assertAdamicNonNullRefusal(t, err)
}

func TestLazyInitializerIsNotEagerMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_lazy_initialized.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	assertAdamicNonNullRefusal(t, err)
}
