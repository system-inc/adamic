package oracle

import (
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

var typeofNullFixtures = []string{
	"internal/oracle/testdata/typeof_null.a",
	"internal/oracle/testdata/typeof_null_compare.a",
	"internal/oracle/testdata/typeof_null_switch.a",
	"internal/oracle/testdata/typeof_null_slots.a",
	"internal/oracle/testdata/typeof_null_roll.a",
}

func init() {
	for _, path := range typeofNullFixtures[1:] {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// Restore the old typeof path only for null: erase its missing-pointer meaning and send a literal
// through the undefined path. All programs must still compile and finish cleanly; Node's stdout
// alone must catch every fixture, including the comparison and switch forms.
func TestTypeOfNullMutant(test *testing.T) {
	test.Parallel()
	for _, fixture := range typeofNullFixtures {
		test.Run(filepath.Base(fixture), func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture))
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
			}
			want := onNode(test, path)
			changed := restoreNullTypeOf(reflect.ValueOf(program))
			if changed == 0 {
				test.Fatal("mutant changed no typeof observations")
			}
			got, binary := natively(test, program)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				test.Fatalf("mutant must finish cleanly: exit %d, stderr %q", got.exitCode, got.stderr)
			}
			if difference := disagreement(want, got); difference != "stdout differs" {
				test.Fatalf("want Node to catch stdout alone, got %q", difference)
			}
			if report := leaks(test, program, binary); report != "" {
				test.Fatalf("mutant must not leak: %s", report)
			}
			test.Logf("caught %d restored observations: Node %q; mutant %q", changed, want.stdout, got.stdout)
		})
	}
}

func restoreNullTypeOf(value reflect.Value) int {
	if !value.IsValid() {
		return 0
	}
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return 0
		}
		if observation, ok := value.Interface().(ir.TypeOf); ok {
			_, literal := observation.Value.(ir.Null)
			if observation.Null || literal {
				observation.Null = false
				if literal {
					observation.Value = ir.Undefined{Of: observation.Value.Type()}
				}
				value.Set(reflect.ValueOf(observation))
				return 1
			}
		}
		copy := reflect.New(value.Elem().Type()).Elem()
		copy.Set(value.Elem())
		changed := restoreNullTypeOf(copy)
		value.Set(copy)
		return changed
	}
	changed := 0
	switch value.Kind() {
	case reflect.Pointer:
		if !value.IsNil() {
			changed += restoreNullTypeOf(value.Elem())
		}
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			changed += restoreNullTypeOf(value.Field(index))
		}
	case reflect.Slice:
		for index := 0; index < value.Len(); index++ {
			changed += restoreNullTypeOf(value.Index(index))
		}
	}
	return changed
}

// Mistaking a missing slot for a present null must be caught separately from mistaking null for
// undefined. Only the slot-presence argument is changed; the value and its ownership stay intact.
func TestTypeOfNullSlotPresenceMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typeof_null_slots.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	want := onNode(test, path)
	source := native.C(program)
	presence := regexp.MustCompile(`(adamic_union_typeof\([^\n]+, )adamic_temporary_[A-Za-z0-9_]+ != NULL\)`)
	mutant := presence.ReplaceAllString(source, "${1}true)")
	if mutant == source {
		test.Fatal("mutant changed no slot-presence tests")
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	if report := leakChecked(test, mutant, binary); report != "" {
		test.Fatalf("mutant must not leak: %s", report)
	}
	got := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("mutant must finish cleanly without leaks: exit %d, stderr %q", got.exitCode, got.stderr)
	}
	if difference := disagreement(want, got); difference != "stdout differs" {
		test.Fatalf("want Node to catch stdout alone, got %q", difference)
	}
	test.Logf("caught: Node %q; mutant %q", want.stdout, got.stdout)
}
