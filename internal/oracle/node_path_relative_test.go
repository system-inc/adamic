package oracle

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// This runtime/emitter check does not substitute for compiling the upstream
// acceptance fixture. It can run while fs_file's declaration hook is pending.
func TestNodePathRelativeRuntime(test *testing.T) {
	test.Parallel()
	paths := []string{"", ".", "..", "/", "/a", "/a/b", "/aa", "/a/b/", "a/../b", "é/中", "\x00", "a//b"}
	program := &ir.Program{Source: "node_path_relative.a", Strings: paths}
	for from := range paths {
		for to := range paths {
			call := ir.NodeHostCall{Module: "node:path", Member: "relative", Returns: ir.String, Throws: true,
				Arguments: []ir.Expression{ir.StringConstant{Index: from}, ir.StringConstant{Index: to}}}
			program.Main = append(program.Main, ir.WriteLine{Stream: ir.Stdout, Value: call})
		}
	}
	fixture, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/node_path_relative.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	how := inputRun{directory: root}
	truth := onNodeWith(test, how, fixture)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		test.Fatalf("Node failed: %+v", truth)
	}
	shared := sharedDirectory(test)
	backend := inputBackend(test, how, program, shared)
	got, binary := inputNatively(test, how, program, shared)
	for name, result := range map[string]run{"native": got, "JavaScript": backend} {
		if difference := disagreement(truth, result); difference != "" {
			test.Errorf("%s: %s", name, difference)
		}
	}
	if leaked := inputLeaks(test, func() inputRun { return how }, program, binary); leaked != "" {
		test.Fatal(leaked)
	}
	source := strings.ReplaceAll(native.C(program), "adamic_node_path_relative(", "mutant_relative(")
	helper := `static adamic_string *mutant_relative(const adamic_string *from, const adamic_string *to) { return adamic_node_path_join(2, (adamic_string *const[]){(adamic_string *)from, (adamic_string *)to}); }`
	source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+"\n"+helper, 1)
	mutant := filepath.Join(shared, "mutant")
	if buildError := native.Build(source, mutant, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	// LeakSanitizer is Linux's: macOS's AddressSanitizer aborts when asked for it.
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=1"}
	}
	result := executeInput(test, how, environment, mutant)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		test.Fatalf("mutant failed outside comparison: exit %d stderr %s", result.exitCode, result.stderr)
	}
	if difference := disagreement(truth, result); difference != "stdout differs" {
		test.Fatalf("mutant caught by %q", difference)
	}
	test.Log("relative runtime agrees on both backends; join mutant caught only by Node stdout comparison")
}
