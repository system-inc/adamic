package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// This fixture has its own differential gate so counts.md stays unchanged. The ordinary
// oracle still checks every existing allocation count with the new numeric runtime.
func TestIntegerFastPathsAgreeWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/integer_fast_paths.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	oracle := onNode(t, path)
	result, sanitized := natively(t, program)
	for name, got := range map[string]run{"sanitized": result, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(oracle, got); difference != "" {
			t.Fatalf("%s: %s, exit %d, stderr %s", name, difference, got.exitCode, got.stderr)
		}
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatal(report)
	}
}

// Clone the actual inline runtime into the generated translation unit, then change just one
// rule. Every mutant must compile, exit zero and remain sanitizer clean: Node alone kills it.
func TestIntegerFastPathMutants(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/integer_fast_paths.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	header, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/integer.h"))
	if err != nil {
		t.Fatal(err)
	}
	mutants := []struct{ name, before, after string }{
		{"remainder_zero_sign", "copysign(0.0, left)", "0.0"},
		{"remainder_divisor_sign", ": (double)remainder;", ": copysign(fabs((double)remainder), right);"},
		{"int32_boundary_accepted_as_itself", "bits >= 0x80000000u", "bits > 0x80000000u"},
		{"unsigned_shift_signed", "return (double)(adamic_to_uint32(left) >> adamic_shift_count(right));", "return adamic_signed_bits(adamic_to_uint32(left) >> adamic_shift_count(right));"},
		{"remainder_fraction_accepted", "(double)dividend == left && (double)divisor == right", "(double)divisor == right"},
	}
	oracle := onNode(t, path)
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			if !strings.Contains(string(header), mutant.before) {
				t.Fatal("mutant target absent")
			}
			changed := strings.Replace(string(header), mutant.before, mutant.after, 1)
			changed = strings.ReplaceAll(changed, "adamic_", "mutant_")
			changed = strings.ReplaceAll(changed, "mutant_to_uint32_slow", "adamic_to_uint32_slow")
			changed = strings.ReplaceAll(changed, "ADAMIC_INTEGER_H", "MUTANT_INTEGER_H")
			source := native.C(program)
			// The pure >>> emitter uses unsigned bits directly; this mutant must change its
			// final number conversion too, so both fused and runtime shift paths are tested.
			if mutant.name == "unsigned_shift_signed" {
				source = strings.ReplaceAll(source, "((double)((uint32_t)", "(mutant_signed_bits((uint32_t)")
			}
			for _, name := range []string{"remainder", "signed_bits", "to_uint32", "bitwise_and", "bitwise_or", "bitwise_xor", "bitwise_not", "shift_left", "shift_right_unsigned", "shift_right_bits", "shift_right"} {
				source = strings.ReplaceAll(source, "adamic_"+name+"(", "mutant_"+name+"(")
			}
			source = strings.Replace(source, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+changed, 1)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := execute(t, binary)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: exit %d, stderr %s", got.exitCode, got.stderr)
			}
			if difference := disagreement(oracle, got); difference != "stdout differs" {
				t.Fatalf("Node must catch mutant, got %q", difference)
			}
			t.Log("Node caught stdout differs; exit 0 and sanitizer clean")
		})
	}
}
