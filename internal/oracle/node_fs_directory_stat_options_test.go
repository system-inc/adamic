package oracle

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestNodeFSDirectoryStatConstOptionsMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_directory_stat_options.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: filepath.Dir(path)}
	truth := onNodeWith(t, how, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("Node failed: %+v", truth)
	}
	original := native.C(program)
	source := strings.ReplaceAll(original, "adamic_fs_file_stat(", "mutant_stat(")
	if source == original {
		t.Fatal("mutant changed nothing")
	}
	helper := `static adamic_object *mutant_stat(const adamic_string *path,bool throws) {(void)throws;return adamic_fs_file_stat(path,true);}`
	source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+"\n"+helper, 1)
	shared := sharedDirectory(t)
	binary := filepath.Join(shared, "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeInput(t, how, nodeFSDirectorySanitizerEnvironment(), binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: exit %d stderr %s", got.exitCode, got.stderr)
	}
	if difference := disagreement(truth, got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	t.Log("false option forced true; caught only by Node stdout; sanitizers and leaks clean")
}

func nodeFSDirectorySanitizerEnvironment() []string {
	environment := []string{"UBSAN_OPTIONS=halt_on_error=1"}
	if runtime.GOOS == "linux" {
		environment = append(environment, "ASAN_OPTIONS=detect_leaks=1")
	}
	return environment
}
