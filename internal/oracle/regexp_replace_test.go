package oracle

import (
	"bytes"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"argument_guard", "group_guard"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/regexp_replace/" + name + ".a", true, true})
	}
	for _, name := range []string{"callback", "throw", "arguments"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/regexp_replace/" + name + ".a", true, false})
	}
}

// Exercise the same runtime-before-program layout without applying a mutation.
func replacementRuntimeControl(t *testing.T, source string, expected run) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "control")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
	if diff := disagreement(expected, actual); diff != "" {
		t.Fatalf("unmutated replacement runtime: %s: %+v", diff, actual)
	}
	t.Log("unmutated runtime agrees with its oracle with sanitizers and leaks")
}

// Each mutant builds cleanly and completes without sanitizer diagnostics. Only the
// source-on-Node output decides whether its callback argument construction is correct.
func TestRegExpReplacementNodeMutantsOffset(t *testing.T) {
	t.Parallel()
	regExpReplacementNodeMutant(t, "value.number = (double)offset;", "value.number = (double)offset + 1;")
}

func TestRegExpReplacementNodeMutantsGroupOrder(t *testing.T) {
	t.Parallel()
	regExpReplacementNodeMutant(t, "value = match->elements[j];", "value = match->elements[j == 1 ? 2 : j == 2 ? 1 : j];")
}

func TestRegExpReplacementNodeMutantsNamedGroups(t *testing.T) {
	t.Parallel()
	regExpReplacementNodeMutant(t, "adamic_object *groups = match->properties->slots[2].reference;", "adamic_object *groups = NULL;")
}

func TestRegExpReplacementTypeGuardMutantsArgument(t *testing.T) {
	t.Parallel()
	regExpReplacementTypeGuardMutant(t, "argument_guard", "if (!accepted)", "if (!accepted && false)")
}

func TestRegExpReplacementTypeGuardMutantsGroup(t *testing.T) {
	t.Parallel()
	regExpReplacementTypeGuardMutant(t, "group_guard", "if (field == NULL && !rule->optional)", "if (field == NULL && !rule->optional && false)")
}

func regExpReplacementNodeMutant(t *testing.T, old, new string) {
	t.Helper()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regexp_replace/arguments.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_replace.c"))
	if err != nil {
		t.Fatal(err)
	}
	programSource := native.C(program)
	if bytes.Count(runtime, []byte(old)) != 1 {
		t.Fatal("mutation site moved")
	}
	replacementRuntimeControl(t, strings.ReplaceAll(string(runtime)+"\n"+native.C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant"), expected)
	changed := strings.Replace(string(runtime), old, new, 1)
	changed = strings.ReplaceAll(changed, "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	source := changed + "\n" + strings.ReplaceAll(programSource, "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed before the Node comparison: %+v", actual)
	}
	if diff := disagreement(expected, actual); diff == "" {
		t.Fatal("Node did not catch mutant")
	} else {
		t.Log(diff)
	}
}

func regExpReplacementTypeGuardMutant(t *testing.T, name, old, new string) {
	t.Helper()
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_replace.c"))
	if err != nil {
		t.Fatal(err)
	}
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regexp_replace/"+name+".a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onJavaScriptBackend(t, program)
	if expected.exitCode != 70 || !bytes.Contains(expected.stderr, []byte("does not fit its declared type")) {
		t.Fatalf("guard not exercised: %+v", expected)
	}
	if bytes.Count(runtime, []byte(old)) != 1 {
		t.Fatal("mutation site moved")
	}
	replacementRuntimeControl(t, strings.ReplaceAll(string(runtime)+"\n"+native.C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant"), expected)
	changed := strings.Replace(string(runtime), old, new, 1)
	changed = strings.ReplaceAll(changed, "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	source := changed + "\n" + strings.ReplaceAll(native.C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed before comparison: %+v", actual)
	}
	if diff := disagreement(expected, actual); diff == "" {
		t.Fatal("backend comparison did not catch mutant")
	} else {
		t.Log(diff)
	}
}
