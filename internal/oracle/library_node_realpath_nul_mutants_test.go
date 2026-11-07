package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Mutate the runtime itself, not its caller. Both mutants must build and run
// sanitizer-clean; only the NUL fixture's comparison with Node may kill them.
func realpathNULRuntimeMutant(test *testing.T, file, before, after string) {
	test.Helper()
	path, absolutePathError := filepath.Abs("testdata/host_realpath/nul.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	cwd := realpathDirectory(test)
	loader, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	truth := realpathRun(test, cwd, "node", "--disable-warning=ExperimentalWarning", loader, path)
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
		if entry.Name() == file {
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
	got := realpathRun(test, cwd, binary)
	if truth.exitCode != 0 || got.exitCode != 0 || len(truth.stderr) != 0 || len(got.stderr) != 0 {
		test.Fatalf("must finish sanitizer-clean: Node %+v mutant %+v", truth, got)
	}
	if bytes.Equal(truth.stdout, got.stdout) {
		test.Fatal("mutant survived TestHostRealpathNULMatchesNode's comparison")
	}
	test.Logf("TestHostRealpathNULMatchesNode goes red: Node %q; mutant %q", truth.stdout, got.stdout)
}

func TestHostRealpathNULMutantNonThrowingPath(test *testing.T) {
	test.Parallel()
	realpathNULRuntimeMutant(test, "node_fs_file.c",
		"void adamic_fs_file_realpath_error(const adamic_string *path) {\n    int error = errno;\n    char *name = path_bytes(path, true);",
		"void adamic_fs_file_realpath_error(const adamic_string *path) {\n    int error = errno;\n    char *name = path_bytes(path, false);")
}

func TestHostRealpathNULMutantTruncatedPath(test *testing.T) {
	test.Parallel()
	realpathNULRuntimeMutant(test, "host_runtime.c",
		"\tif (memchr(path->bytes, 0, path->length) != NULL) { errno = EINVAL; return NULL; }\n", "")
}
