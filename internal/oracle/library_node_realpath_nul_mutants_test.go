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
func realpathNULRuntimeMutant(t *testing.T, file, before, after string) {
	t.Helper()
	path, err := filepath.Abs("testdata/host_realpath/nul.a")
	if err != nil {
		t.Fatal(err)
	}
	cwd := realpathDirectory(t)
	loader, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	truth := realpathRun(t, cwd, "node", "--disable-warning=ExperimentalWarning", loader, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	directory := t.TempDir()
	entries, err := os.ReadDir("../native/runtime")
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join("../native/runtime", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() == file {
			if strings.Count(string(data), before) != 1 {
				t.Fatal("runtime mutation must match exactly once")
			}
			data = []byte(strings.Replace(string(data), before, after, 1))
			changed = true
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if !changed {
		t.Fatal("runtime mutation changed nothing")
	}
	options := native.Options{Sanitize: true}
	archive, err := native.RuntimeLibraryForSource(directory, source, options)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.BuildWithRuntime(source, binary, options, archive); err != nil {
		t.Fatal(err)
	}
	got := realpathRun(t, cwd, binary)
	if truth.exitCode != 0 || got.exitCode != 0 || len(truth.stderr) != 0 || len(got.stderr) != 0 {
		t.Fatalf("must finish sanitizer-clean: Node %+v mutant %+v", truth, got)
	}
	if bytes.Equal(truth.stdout, got.stdout) {
		t.Fatal("mutant survived TestHostRealpathNULMatchesNode's comparison")
	}
	t.Logf("TestHostRealpathNULMatchesNode goes red: Node %q; mutant %q", truth.stdout, got.stdout)
}

func TestHostRealpathNULMutantNonThrowingPath(t *testing.T) {
	t.Parallel()
	realpathNULRuntimeMutant(t, "node_fs_file.c",
		"void adamic_fs_file_realpath_error(const adamic_string *path) {\n    int error = errno;\n    char *name = path_bytes(path, true);",
		"void adamic_fs_file_realpath_error(const adamic_string *path) {\n    int error = errno;\n    char *name = path_bytes(path, false);")
}

func TestHostRealpathNULMutantTruncatedPath(t *testing.T) {
	t.Parallel()
	realpathNULRuntimeMutant(t, "host_runtime.c",
		"\tif (memchr(path->bytes, 0, path->length) != NULL) { errno = EINVAL; return NULL; }\n", "")
}
