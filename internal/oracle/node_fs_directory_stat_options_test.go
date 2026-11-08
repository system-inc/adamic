package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestNodeFileSystemDirectoryStatConstOptionsMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/node_fs_directory_stat_options.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	how := inputRun{directory: filepath.Dir(path)}
	truth := onNodeWith(test, how, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		test.Fatalf("Node failed: %+v", truth)
	}
	original := native.C(program)
	source := strings.ReplaceAll(original, "adamic_fs_file_stat(", "mutant_stat(")
	if source == original {
		test.Fatal("mutant changed nothing")
	}
	helper := `static adamic_object *mutant_stat(const adamic_string *path,bool throws) {(void)throws;return adamic_fs_file_stat(path,true);}`
	source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+"\n"+helper, 1)
	shared := sharedDirectory(test)
	binary := filepath.Join(shared, "mutant")
	if buildError := native.Build(source, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	got := executeInput(test, how, nodeFileSystemDirectorySanitizerEnvironment(), binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("mutant failed outside comparison: exit %d stderr %s", got.exitCode, got.stderr)
	}
	if difference := disagreement(truth, got); difference != "stdout differs" {
		test.Fatalf("mutant caught by %q", difference)
	}
	test.Log("false option forced true; caught only by Node stdout; sanitizers and leaks clean")
}

func nodeFileSystemDirectorySanitizerEnvironment() []string {
	environment := []string{"UBSAN_OPTIONS=halt_on_error=1"}
	if runtime.GOOS == "linux" {
		environment = append(environment, "ASAN_OPTIONS=detect_leaks=1")
	}
	return environment
}

// Dropping an encoding argument reproduces the old string-only closure bug.
// Latin1 and UTF8 remain valid strings, so only Node's bytes catch the mutant.
func TestNodeFileSystemRealpathClosureOptionsMutant(test *testing.T) {
	test.Parallel()
	input, absolutePathError := filepath.Abs("testdata/node_fs_directory_inputs/é.txt")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source := "import {realpathSync} from 'node:fs'; const resolve=realpathSync.native; console.log(resolve(" +
		strconv.Quote(input) +
		",'latin1'));"
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	_, loweredError := lowered(test, path)
	var missing *lower.NotYet
	if !errors.As(loweredError, &missing) ||
		!strings.Contains(missing.What, "is called with zero or multiple arguments") {
		test.Fatalf("want explicit closure overload refusal, got %v", loweredError)
	}
	truth := onNode(test, path)
	mutantPath := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(
		mutantPath,
		[]byte(strings.Replace(source, ",'latin1'", "", 1)),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	program, loweredError := lowered(test, mutantPath)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := native.Build(native.C(program), binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	got := executeWith(test, nodeFileSystemDirectorySanitizerEnvironment(), binary)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("mutant failed outside Node comparison: Node %+v mutant %+v", truth, got)
	}
	if disagreement(truth, got) != "stdout differs" {
		test.Fatal("Node didn't catch discarded realpath encoding")
	}
	test.Log("discarded latin1 encoding caught only by Node stdout; exit 0, clean sanitizer and leak checks")
}
