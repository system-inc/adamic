package native

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestJSONReplacerDuplicatesRecordMutant(test *testing.T) {
	test.Parallel()
	jsonReplacerDedupMutant(test, "records_json_replacer_duplicates")
}

func TestJSONReplacerDuplicatesObjectMutant(test *testing.T) {
	test.Parallel()
	jsonReplacerDedupMutant(test, "json_replacer_duplicates_object")
}

func jsonReplacerDedupMutant(test *testing.T, fixture string) {
	test.Helper()
	path, absolutePathError := filepath.Abs("../oracle/testdata/" + fixture + ".a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(test.Context(), loaded)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	source := C(program)
	runner, absolutePathError := filepath.Abs("../../oracle/node.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := jsonReplacerRun(test, "node", "--disable-warning=ExperimentalWarning", runner, path)
	files, readRuntimeError := readRuntime(runtime, "runtime")
	if readRuntimeError != nil {
		test.Fatal(readRuntimeError)
	}
	before := "if (adamic_string_equal(key, w->keys[previous]))"
	changed := false
	kept := files[:0]
	for _, file := range files {
		if file.name == "node_host.c" && !strings.Contains(source, "#define ADAMIC_NODE_HOST 1\n") {
			continue
		}
		if file.name == "regexp_replace.c" && !strings.Contains(source, "#define ADAMIC_REGEXP_REPLACE_CALLBACK 1\n") {
			continue
		}
		if file.name == "json_stringify.c" {
			content := string(file.contents)
			if strings.Count(content, before) != 1 {
				test.Fatal("dedup mutant needs exactly one site")
			}
			file.contents = []byte(
				strings.Replace(content, before, "if (false && adamic_string_equal(key, w->keys[previous]))", 1),
			)
			changed = true
		}
		kept = append(kept, file)
	}
	if !changed {
		test.Fatal("mutant did not change runtime")
	}
	compiler, lookPathError := exec.LookPath("clang")
	if lookPathError != nil {
		test.Fatal(lookPathError)
	}
	version, combinedOutputError := exec.CommandContext(test.Context(), compiler, "--version").CombinedOutput()
	if combinedOutputError != nil {
		test.Fatal(combinedOutputError)
	}
	cache, userCacheDirectoryError := os.UserCacheDir()
	if userCacheDirectoryError != nil {
		test.Fatal(userCacheDirectoryError)
	}
	flags := sourceFlags(source, Options{Sanitize: true})
	library, cachedRuntimeError := cachedRuntime(
		kept,
		flags,
		compiler,
		string(version),
		filepath.Join(cache, "adamic", "json-replacer-mutants"),
	)
	if cachedRuntimeError != nil {
		test.Fatal(cachedRuntimeError)
	}
	directory := test.TempDir()
	main := filepath.Join(directory, "main.c")
	binary := filepath.Join(directory, "mutant")
	if writeError := os.WriteFile(main, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	arguments := append(append([]string{}, flags...), "-I", filepath.Dir(library), main, "-o", binary)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	if output, combinedOutputError := exec.CommandContext(
		test.Context(),
		compiler,
		arguments...,
	).CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("mutant compile: %v\n%s", combinedOutputError, output)
	}
	got := jsonReplacerRun(test, binary)
	if bytes.Equal(got, want) {
		test.Fatalf("dedup mutant survived: %q", got)
	}
	if !bytes.Contains(got, []byte(`"a":"oneone","a":"oneone"`)) {
		test.Fatalf("unexpected mutant output: %q", got)
	}
	test.Logf("skipped dedup caught by Node stdout; clean ASan/UBSan/leaks; mutant %q; Node %q", got, want)
}

// Use the existing record harness observation path, avoiding a dependency on
// c-portability's test-only portabilityRun helper.
func jsonReplacerRun(test *testing.T, name string, arguments ...string) []byte {
	test.Helper()
	stdout, stderr, recordRunError := recordRun(name, arguments...)
	if recordRunError != nil || stderr != "" {
		test.Fatalf("%s: %v\n%s", name, recordRunError, stderr)
	}
	return []byte(stdout)
}
