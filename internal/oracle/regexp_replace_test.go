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
	for _, name := range []string{"callback", "dynamic", "throw", "arguments", "effects", "move_effect"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/regexp_replace/" + name + ".a", true, false})
	}
}

// Each mutant builds cleanly and completes without sanitizer diagnostics. Only the
// source-on-Node output decides whether its callback argument construction is correct.
func TestRegExpReplacementNodeMutants(t *testing.T) {
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
	for _, mutant := range []struct{ name, old, new string }{
		{"offset off by one", "value.number = (double)offset;", "value.number = (double)offset + 1;"},
		{"groups wrong order", "value = match->elements[j];", "value = match->elements[j == 1 ? 2 : j == 2 ? 1 : j];"},
		{"missing named groups argument", "adamic_object *groups = match->properties->slots[2].reference;", "adamic_object *groups = NULL;"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			if bytes.Count(runtime, []byte(mutant.old)) != 1 {
				t.Fatal("mutation site moved")
			}
			changed := strings.Replace(string(runtime), mutant.old, mutant.new, 1)
			changed = strings.ReplaceAll(changed, "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
			source := "#define ADAMIC_REGEXP_REPLACE_CALLBACK 1\n" + changed + "\n" + strings.ReplaceAll(native.C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
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
		})
	}
}

func TestRegExpReplacementTypeGuardMutants(t *testing.T) {
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_replace.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, old, new string }{
		{"argument_guard", "if (!accepted)", "if (!accepted && false)"},
		{"group_guard", "if (field == NULL && !rule->optional)", "if (field == NULL && !rule->optional && false)"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regexp_replace/"+mutant.name+".a"))
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onJavaScriptBackend(t, program)
			if expected.exitCode != 70 || !bytes.Contains(expected.stderr, []byte("does not fit its declared type")) {
				t.Fatalf("guard not exercised: %+v", expected)
			}
			if bytes.Count(runtime, []byte(mutant.old)) != 1 {
				t.Fatal("mutation site moved")
			}
			changed := strings.Replace(string(runtime), mutant.old, mutant.new, 1)
			changed = strings.ReplaceAll(changed, "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
			source := "#define ADAMIC_REGEXP_REPLACE_CALLBACK 1\n" + changed + "\n" + strings.ReplaceAll(native.C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
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
		})
	}
}
