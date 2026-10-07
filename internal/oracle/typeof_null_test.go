package oracle

import (
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
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
func TestTypeOfNullMutant(t *testing.T) {
	t.Parallel()
	for _, fixture := range typeofNullFixtures {
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			changed := restoreNullTypeOf(reflect.ValueOf(program))
			if changed == 0 {
				t.Fatal("mutant changed no typeof observations")
			}
			got, binary := natively(t, program)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: exit %d, stderr %q", got.exitCode, got.stderr)
			}
			if difference := disagreement(want, got); difference != "stdout differs" {
				t.Fatalf("want Node to catch stdout alone, got %q", difference)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatalf("mutant must not leak: %s", report)
			}
			t.Logf("caught %d restored observations: Node %q; mutant %q", changed, want.stdout, got.stdout)
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
func TestTypeOfNullSlotPresenceMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typeof_null_slots.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	source := native.C(program)
	presence := regexp.MustCompile(`(adamic_union_typeof\([^\n]+, )adamic_temporary_[0-9]+ != NULL\)`)
	mutant := presence.ReplaceAllString(source, "${1}true)")
	if mutant == source {
		t.Fatal("mutant changed no slot-presence tests")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	// LeakSanitizer is Linux's: macOS's AddressSanitizer aborts when asked for it.
	var leakCheck []string
	if runtime.GOOS == "linux" {
		leakCheck = []string{"ASAN_OPTIONS=detect_leaks=1"}
	}
	got := executeWith(t, leakCheck, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly without leaks: exit %d, stderr %q", got.exitCode, got.stderr)
	}
	if difference := disagreement(want, got); difference != "stdout differs" {
		t.Fatalf("want Node to catch stdout alone, got %q", difference)
	}
	t.Logf("caught: Node %q; mutant %q", want.stdout, got.stdout)
}
