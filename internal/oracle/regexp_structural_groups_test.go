package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var regexpStructuralGroups = []string{"8e67677_structural_method", "8e67677_structural_optional", "903f25b_groups_proto", "903f25b_groups_proto2", "caught", "structural_exec"}

const regexpStructuralGroupsDir = "internal/oracle/testdata/regexp_structural_groups/"

func init() {
	for _, name := range regexpStructuralGroups {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{regexpStructuralGroupsDir + name + ".a", true, false})
	}
}

func TestRegExpStructuralGroups(t *testing.T) {
	t.Parallel()
	for _, name := range regexpStructuralGroups {
		t.Run(name, func(t *testing.T) {
			path, _ := filepath.Abs(filepath.Join(repository, regexpStructuralGroupsDir+name+".a"))
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			actual, _ := natively(t, program)
			if diff := disagreement(expected, actual); diff != "" {
				t.Fatal("native: " + diff)
			}
			if diff := disagreement(expected, onJavaScriptBackend(t, program)); diff != "" {
				t.Fatal("JavaScript: " + diff)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") != "" {
				if diff := disagreement(expected, onWASI(t, native.C(program))); diff != "" {
					t.Fatal("WASI: " + diff)
				}
			}
		})
	}
}

func TestRegExpStructuralGroupsNodeMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct{ name, fixture, old, new string }{
		{"wrong-exec-start", "structural_exec", ".reference = adamic_regex_exec(self, arguments[0].reference)", ".reference = (self->slots[1].number = 2, adamic_regex_exec(self, arguments[0].reference))"},
		{"wrong-test-result", "8e67677_structural_method", ".boolean = adamic_regex_test(self, arguments[0].reference)", ".boolean = !adamic_regex_test(self, arguments[0].reference)"},
		{"missing-test-presence", "8e67677_structural_optional", "{\"test\", \"exec\"}", "{\"absent\", \"exec\"}"},
		{"ordinary-group-prototype", "caught", "groups->null_prototype = true;", "groups->null_prototype = false;"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			path, _ := filepath.Abs(filepath.Join(repository, regexpStructuralGroupsDir+mutant.fixture+".a"))
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp.c"))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(data), mutant.old) != 1 {
				t.Fatal("mutation site moved")
			}
			// Runtime code lives in the archive. Give every exported matcher
			// function a private mutant name so the program really calls this copy.
			names := regexp.MustCompile(`(?m)^(?:[A-Za-z_][A-Za-z_0-9 ]* \*?)(adamic_regex_[A-Za-z_0-9]+)\(`).FindAllStringSubmatch(string(data), -1)
			prefix := "#define ADAMIC_REGEXP_RUNTIME_OWNER 1\n"
			for _, name := range names {
				prefix += "#define " + name[1] + " mutant_" + name[1] + "\n"
			}
			source := prefix + strings.Replace(string(data), mutant.old, mutant.new, 1) + "\n" + native.C(program)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("failed outside Node: %+v", actual)
			}
			if disagreement(expected, actual) == "" {
				t.Fatal("Node did not catch mutant")
			}
			t.Log("Node caught " + mutant.name)
		})
	}
}

// Diagnostic mutants finish normally: only the external Node comparison rejects
// their incorrect catchable error messages, on every backend.
func TestRegExpGroupDiagnosticNodeMutants(t *testing.T) {
	t.Parallel()
	for _, message := range []string{"groups.hasOwnProperty is not a function", "match.groups.toString is not a function"} {
		t.Run(message, func(t *testing.T) {
			path, _ := filepath.Abs(filepath.Join(repository, regexpStructuralGroupsDir+"caught.a"))
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			found := false
			for i, value := range program.Strings {
				if value == message {
					program.Strings[i] = "wrong diagnostic"
					found = true
				}
			}
			if !found {
				t.Fatal("mutation site moved")
			}
			actual, _ := natively(t, program)
			results := []run{actual, onJavaScriptBackend(t, program)}
			if os.Getenv("ADAMIC_ORACLE_WASI") != "" {
				results = append(results, onWASI(t, native.C(program)))
			}
			for _, result := range results {
				if result.exitCode != 0 || len(result.stderr) != 0 {
					t.Fatalf("failed outside Node: %+v", result)
				}
				if disagreement(expected, result) == "" {
					t.Fatal("Node did not catch diagnostic mutant")
				}
			}
			t.Log("Node caught wrong diagnostic on all enabled backends")
		})
	}
}
