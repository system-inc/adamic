package oracle

import (
	"bytes"
	"errors"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
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
	for _, name := range []string{"callback", "throw", "arguments", "effects", "move_effect", "reentrant"} {
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
	for _, mutant := range []struct{ name, old, new, javascript string }{
		{"offset off by one", "value.number = (double)offset;", "value.number = (double)offset + 1;", "values[values.length - (typeof values[values.length - 1] === 'object' ? 3 : 2)] += 1; "},
		{"groups wrong order", "value = match->elements[j];", "value = match->elements[j == 1 ? 2 : j == 2 ? 1 : j];", "[values[1], values[2]] = [values[2], values[1]]; "},
		{"missing named groups argument", "adamic_object *groups = match->properties->slots[2].reference;", "adamic_object *groups = NULL;", "if (typeof values[values.length - 1] === 'object') values.pop(); "},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			if bytes.Count(runtime, []byte(mutant.old)) != 1 {
				t.Fatal("mutation site moved")
			}
			changed := strings.Replace(string(runtime), mutant.old, mutant.new, 1)
			changed = strings.ReplaceAll(changed, "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
			source := changed + "\n" + strings.ReplaceAll(native.C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant failed before the Node comparison: %+v", actual)
			}
			results := map[string]run{"native": actual}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				results["WASI"] = onWASI(t, source)
			}
			js := javascript.JavaScript(program)
			site := "return adamicCall(callback, values);"
			if strings.Count(js, site) != 1 {
				t.Fatal("JavaScript mutation site moved")
			}
			results["JavaScript"] = replacementJavaScriptMutant(t, strings.Replace(js, site, mutant.javascript+site, 1))
			for backend, result := range results {
				if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(expected, result) != "stdout differs" {
					t.Fatalf("%s: mutant must finish cleanly and differ only from Node: %+v", backend, result)
				}
				t.Logf("%s: Node alone caught stdout differs", backend)
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
			js := javascript.JavaScript(program)
			marker := "panic('RegExp replacement callback argument does not fit its declared type'); "
			end := strings.Index(js, marker)
			if end < 0 {
				t.Fatal("JavaScript guard mutation site moved")
			}
			start := strings.LastIndex(js[:end], "if (!(")
			if start < 0 {
				t.Fatal("JavaScript guard condition missing")
			}
			result := replacementJavaScriptMutant(t, js[:start]+js[end+len(marker):])
			if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(expected, result) == "" {
				t.Fatalf("JavaScript guard mutant escaped comparison: %+v", result)
			}

			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				result := onWASI(t, source)
				if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(expected, result) == "" {
					t.Fatalf("WASI guard mutant escaped comparison: %+v", result)
				}
			}
		})
	}
}

// The closure-convention ruling requires this mutant to fail in clang, rather
// than relying on a later observation of corrupted callback arguments.
func TestRegExpReplacementDropCount(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/closure_convention_regexp_count.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_replace.c"))
	if err != nil {
		t.Fatal(err)
	}
	original := "adamic_closure_call(callback, packed, argument_count)"
	if bytes.Count(runtime, []byte(original)) != 1 {
		t.Fatal("runtime mutation site moved")
	}
	source := strings.ReplaceAll(string(runtime), "adamic_regex_replace_callback", "adamic_regex_replace_callback_abi") + "\n" + strings.ReplaceAll(native.C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_abi")
	binary := filepath.Join(t.TempDir(), "valid")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(onNode(t, path), executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)); difference != "" {
		t.Fatal(difference)
	}
	changed := strings.Replace(source, original, "adamic_closure_call(callback, packed)", 1)
	err = native.Build(changed, filepath.Join(t.TempDir(), "mutant"), native.Options{Sanitize: true})
	if err == nil || !strings.Contains(err.Error(), "too few arguments to function call") || !strings.Contains(err.Error(), "expected 3, have 2") {
		t.Fatalf("drop-count mutant escaped typed arity: %v", err)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		err = native.Build(changed, filepath.Join(t.TempDir(), "mutant.wasm"), native.Options{Target: "wasm32-wasi"})
		if err == nil || !strings.Contains(err.Error(), "too few arguments to function call") || !strings.Contains(err.Error(), "expected 3, have 2") {
			t.Fatalf("WASI drop-count mutant escaped typed arity: %v", err)
		}
	}
	t.Log("runtime count-drop mutant rejected under -Werror: expected 3, have 2")
}

func TestRegExpReplacementCountNodeMutant(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/closure_convention_regexp_count.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_replace.c"))
	if err != nil {
		t.Fatal(err)
	}
	original := "adamic_closure_call(callback, packed, argument_count)"
	if bytes.Count(runtime, []byte(original)) != 1 {
		t.Fatal("runtime mutation site moved")
	}
	changed := strings.Replace(string(runtime), original, "adamic_closure_call(callback, packed, packed_count)", 1)
	changed = strings.ReplaceAll(changed, "adamic_regex_replace_callback", "adamic_regex_replace_callback_count_mutant")
	source := changed + "\n" + strings.ReplaceAll(native.C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_count_mutant")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed outside Node comparison: %+v", actual)
	}
	if bytes.Equal(expected.stdout, actual.stdout) {
		t.Fatal("Node did not catch packed count replacing the logical count")
	}
	results := map[string]run{"native": actual}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		results["WASI"] = onWASI(t, source)
	}
	js := javascript.JavaScript(program)
	site := "return adamicCall(callback, values);"
	if !strings.Contains(js, site) {
		t.Fatal("JavaScript count mutation site moved")
	}
	results["JavaScript"] = replacementJavaScriptMutant(t, strings.ReplaceAll(js, site, "values.length = 1; "+site))
	for backend, result := range results {
		if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(expected, result) != "stdout differs" {
			t.Fatalf("%s: count mutant escaped Node-only comparison: %+v", backend, result)
		}
	}
	t.Log("Node caught physical packed count replacing logical callback count")
}

// The front candidate may precede the library's runtime pattern compiler. Keep
// its exact fixture ready, and report the missing dependency rather than use a
// static pattern as a substitute for the dynamic constructor.
func TestRegExpReplacementDynamicCandidate(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regexp_replace/dynamic.a"))
	program, err := lowered(t, path)
	if err != nil {
		var notYet *lower.NotYet
		if errors.As(err, &notYet) && strings.Contains(err.Error(), "RegExp with a nonconstant pattern") {
			t.Skipf("library runtime compiler dependency absent: %v", err)
		}
		t.Fatal(err)
	}
	expected := onNode(t, path)
	actual, _ := natively(t, program)
	for name, result := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program), "WASI": onWASI(t, native.C(program))} {
		if difference := disagreement(expected, result); difference != "" {
			t.Fatalf("%s: %s", name, difference)
		}
	}
}

// Run the generated backend artifact after changing only its callback dispatch.
func replacementJavaScriptMutant(t *testing.T, source string) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path)
}
