package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"testing"
)

func TestStep21NarrowedMemberUseStops(t *testing.T) {
	t.Parallel()
	step21UseStop(t, "step21_narrow_member_use.ts", "undefined\ncontinued\n", "adamic: panic: stale narrowing use failed: value.detail expected number\n")
}
func TestStep21NarrowedParameterUseStops(t *testing.T) {
	t.Parallel()
	step21UseStop(t, "step21_narrow_parameter_use.ts", "Error\ncontinued\n", "adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n")
}
func step21UseStop(t *testing.T, file, output, message string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", file))
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if difference := disagreement(run{stdout: []byte(output)}, source); difference != "" {
		t.Fatalf("source Node: %s", difference)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"sanitized native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(run{exitCode: 70, stderr: []byte(message)}, got); difference != "" {
			t.Fatalf("%s per-use stop: %s; got exit %d stdout %q stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
		}
	}
	helper := "narrowed_property_use"
	if file == "step21_narrow_parameter_use.ts" {
		helper = "narrowed_union_member"
	}
	removed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name == helper {
			position := 0
			if helper == "narrowed_property_use" {
				position = 2
			}
			if _, check := function.Body[position].(ir.If); check {
				function.Body = append(function.Body[:position], function.Body[position+1:]...)
				removed++
			}
		}
	}
	if removed != 1 {
		t.Fatalf("mutant removed %d checks; want one", removed)
	}
	mutantNative, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"sanitized native": mutantNative, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if disagreement(run{exitCode: 70, stderr: []byte(message)}, got) == "" {
			t.Fatalf("%s per-use-check mutant survived", name)
		}
		t.Logf("%s removed %s check caught: exit %d stdout %q stderr %q", name, helper, got.exitCode, got.stdout, got.stderr)
	}

}
func TestStep21NarrowedFlagsUseAgrees(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/step21_narrow_flags_use.ts"))
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("12\n12\n")}, source); difference != "" {
		t.Fatalf("source Node: %s", difference)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, binary := nativelyUncached(t, program)
	for name, got := range map[string]run{"sanitized native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(source, got); difference != "" {
			t.Fatalf("%s stored flags: %s", name, difference)
		}
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
}

func TestStep21WiderParameterUseAgrees(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/step21_narrow_wider_parameter.ts"))
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("Error\nobject\ncontinued\n")}, source); difference != "" {
		t.Fatal(difference)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, binary := nativelyUncached(t, program)
	for name, got := range map[string]run{"sanitized native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(source, got); difference != "" {
			t.Fatalf("%s wider parameter: %s", name, difference)
		}
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
}

func init() {
	for _, file := range []string{"step21_builtin_narrow_terminal.ts", "step21_narrow_flags_use.ts", "step21_narrow_length_use.ts", "step21_narrow_wider_parameter.ts", "step21_narrow_member_use.ts", "step21_narrow_parameter_use.ts"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + file, true, file == "step21_narrow_member_use.ts" || file == "step21_narrow_parameter_use.ts"})
	}
}

func TestStep21StoredPrimitiveLengthUseAgrees(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/step21_narrow_length_use.ts"))
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("7\n")}, source); difference != "" {
		t.Fatal(difference)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, binary := nativelyUncached(t, program)
	for name, got := range map[string]run{"sanitized native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(source, got); difference != "" {
			t.Fatalf("%s primitive length: %s", name, difference)
		}
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
}
func TestStep21NullPropertyUseStops(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "null-use.ts")
	source := `let value: unknown = new TypeError('first');
 function change(): void { value = undefined; }
 try { if (value instanceof TypeError) { change(); console.log(value.name); } }
 catch { console.log('caught'); }
 console.log('continued');`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("caught\ncontinued\n")}, node); difference != "" {
		t.Fatal(difference)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, _ := nativelyUncached(t, program)
	want := run{exitCode: 70, stderr: []byte("adamic: panic: stale narrowing use failed: value.name expected string\n")}
	for name, got := range map[string]run{"sanitized native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s null stop: %s", name, difference)
		}
	}
	removed := 0
	for i := range program.Functions {
		f := &program.Functions[i]
		if f.Name == "narrowed_property_use" {
			f.Body = f.Body[1:]
			removed++
		}
	}
	if removed != 1 {
		t.Fatalf("null mutant removed %d", removed)
	}
	mutant, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"sanitized native": mutant, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if disagreement(want, got) == "" {
			t.Fatalf("%s null guard mutant survived", name)
		}
		t.Logf("%s null guard mutant caught: exit %d stdout %q stderr %q", name, got.exitCode, got.stdout, got.stderr)
	}
}

func TestStep21LiteralPropertyUseStops(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "literal-use.ts")
	source := `class Typed { readonly detail: 21 = 21; }
 let value: unknown = new Typed();
 function change(): void { value = { detail: 22 }; }
 if (value instanceof Typed) { change(); console.log("" + value.detail); }`

	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("22\n")}, node); difference != "" {
		t.Fatal(difference)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, _ := nativelyUncached(t, program)
	want := run{exitCode: 70, stderr: []byte("adamic: panic: stale narrowing use failed: value.detail expected 21\n")}
	for name, got := range map[string]run{"sanitized native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s literal stop: %s", name, difference)
		}
	}
	removed := 0
	for i := range program.Functions {
		f := &program.Functions[i]
		if f.Name == "narrowed_property_use" {
			if _, check := f.Body[3].(ir.If); check {
				f.Body = append(f.Body[:3], f.Body[4:]...)
				removed++
			}
		}
	}
	if removed != 1 {
		t.Fatalf("literal mutant removed %d", removed)
	}
	mutant, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"sanitized native": mutant, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(node, got); difference != "" {
			t.Fatalf("%s literal mutant did not admit wrong value: %s", name, difference)
		}
		t.Logf("%s literal guard mutant caught: exit %d stdout %q", name, got.exitCode, got.stdout)
	}
}
