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

func TestJSONReplacerDuplicatesRecordMutant(t *testing.T) {
	t.Parallel()
	jsonReplacerDedupMutant(t, "records_json_replacer_duplicates")
}

func TestJSONReplacerDuplicatesObjectMutant(t *testing.T) {
	t.Parallel()
	jsonReplacerDedupMutant(t, "json_replacer_duplicates_object")
}

func jsonReplacerDedupMutant(t *testing.T, fixture string) {
	t.Helper()
	path, err := filepath.Abs("../oracle/testdata/" + fixture + ".a")
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(t.Context(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	source := C(program)
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	want := jsonReplacerRun(t, "node", "--disable-warning=ExperimentalWarning", runner, path)
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
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
				t.Fatal("dedup mutant needs exactly one site")
			}
			file.contents = []byte(strings.Replace(content, before, "if (false && adamic_string_equal(key, w->keys[previous]))", 1))
			changed = true
		}
		kept = append(kept, file)
	}
	if !changed {
		t.Fatal("mutant did not change runtime")
	}
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	version, err := exec.CommandContext(t.Context(), compiler, "--version").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	flags := sourceFlags(source, Options{Sanitize: true})
	library, err := cachedRuntime(kept, flags, compiler, string(version), filepath.Join(cache, "adamic", "json-replacer-mutants"))
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	main := filepath.Join(directory, "main.c")
	binary := filepath.Join(directory, "mutant")
	if err := os.WriteFile(main, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	arguments := append(append([]string{}, flags...), "-I", filepath.Dir(library), main, "-o", binary)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	if output, err := exec.CommandContext(t.Context(), compiler, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("mutant compile: %v\n%s", err, output)
	}
	got := jsonReplacerRun(t, binary)
	if bytes.Equal(got, want) {
		t.Fatalf("dedup mutant survived: %q", got)
	}
	if !bytes.Contains(got, []byte(`"a":"oneone","a":"oneone"`)) {
		t.Fatalf("unexpected mutant output: %q", got)
	}
	t.Logf("skipped dedup caught by Node stdout; clean ASan/UBSan/leaks; mutant %q; Node %q", got, want)
}

// Use the existing record harness observation path, avoiding a dependency on
// c-portability's test-only portabilityRun helper.
func jsonReplacerRun(t *testing.T, name string, arguments ...string) []byte {
	t.Helper()
	stdout, stderr, err := recordRun(name, arguments...)
	if err != nil || stderr != "" {
		t.Fatalf("%s: %v\n%s", name, err, stderr)
	}
	return []byte(stdout)
}
