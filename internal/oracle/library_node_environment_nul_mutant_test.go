package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Change the runtime's lookup itself. A rejected NUL name incorrectly misses
// the explicitly present A; both literal and runtime-built names must catch it.
func TestProcessEnvironmentNulRejectMutantIsCaught(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("testdata/process_env/nul.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	environment := processEnvironmentFixtureEnvironment("nul.a")
	truth := processEnvironmentRun(
		test,
		environment,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		path,
	)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	source := native.C(program)
	directory := test.TempDir()
	entries, readDirectoryError := os.ReadDir("../native/runtime")
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	changed := false
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, readError := os.ReadFile(filepath.Join("../native/runtime", entry.Name()))
		if readError != nil {
			test.Fatal(readError)
		}
		if entry.Name() == "host_runtime.c" {
			const before = "adamic_string *adamic_host_environment(const adamic_string *name) {"
			const after = before + "\n\tif (memchr(name->bytes, 0, name->length) != NULL) { return NULL; }"
			if strings.Count(string(data), before) != 1 {
				test.Fatal("runtime mutation must match exactly once")
			}
			data = []byte(strings.Replace(string(data), before, after, 1))
			changed = true
		}
		if writeError := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if !changed {
		test.Fatal("runtime mutation changed nothing")
	}
	options := native.Options{Sanitize: true}
	archive, runtimeLibraryForSourceError := native.RuntimeLibraryForSource(directory, source, options)
	if runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildWithRuntimeError := native.BuildWithRuntime(
		source,
		binary,
		options,
		archive,
	); buildWithRuntimeError != nil {
		test.Fatal(buildWithRuntimeError)
	}
	got := processEnvironmentRun(test, environment, binary)
	if truth.exitCode != 0 || got.exitCode != 0 || len(truth.stderr) != 0 || len(got.stderr) != 0 {
		test.Fatalf("must finish sanitizer-clean: Node %+v mutant %+v", truth, got)
	}
	if bytes.Equal(truth.stdout, got.stdout) {
		test.Fatal("mutant survived TestProcessEnvironmentNULFixtureMatchesNode/nul")
	}
	test.Logf("TestProcessEnvironmentNULFixtureMatchesNode/nul goes red: Node %q; mutant %q", truth.stdout, got.stdout)
}
