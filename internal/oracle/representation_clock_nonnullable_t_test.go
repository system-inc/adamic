package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

const clockNonNullableFixture = "internal/oracle/testdata/representation_clock_nonnullable_t.a"
const clockNonNullableConcreteFixture = "internal/oracle/testdata/representation_clock_nonnullable_t_concrete.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{clockNonNullableFixture, false, false}, struct {
		path            string
		lowers, checked bool
	}{clockNonNullableConcreteFixture, true, false})
}

func TestClockNonNullableSource(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, clockNonNullableFixture))
	if err != nil {
		t.Fatal(err)
	}
	got := onNode(t, path)
	if got.exitCode != 0 || string(got.stdout) != "text\ntrue\n" || len(got.stderr) != 0 {
		t.Fatalf("Node: %+v", got)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "a function returning NonNullable<T>") {
		t.Fatalf("want isolated generic return stop, got %v", err)
	}
}

func TestClockNonNullableMutant(t *testing.T) {
	t.Run("clock-nonnullable-t-drop-object-return", func(t *testing.T) {
		path, err := filepath.Abs(filepath.Join(repository, clockNonNullableConcreteFixture))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		changed := 0
		for i := range program.Functions {
			function := &program.Functions[i]
			if !strings.HasPrefix(function.Name, "keep") || function.Returns != ir.Object {
				continue
			}
			for j, statement := range function.Body {
				if result, ok := statement.(ir.Return); ok && result.Value != nil && result.Value.Type() == ir.Object {
					function.Body[j] = ir.Return{Value: ir.Undefined{Of: ir.Object}}
					changed++
				}
			}
		}
		if changed != 1 {
			t.Fatalf("want only object specialization return, changed %d", changed)
		}
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "LSAN_OPTIONS=use_stacks=0:use_registers=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
		if got.exitCode != 0 || len(got.stderr) != 0 || string(got.stdout) != "text\nfalse\n" {
			t.Fatalf("mutant must finish cleanly and print false: %+v", got)
		}
		if difference := disagreement(onNode(t, path), got); difference != "stdout differs" {
			t.Fatalf("want stdout to catch mutant, got %q", difference)
		}
		t.Logf("Node text/true; sanitized mutant text/false, no leaks")
	})
}
