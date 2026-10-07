package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// This runtime/emitter check does not substitute for compiling the upstream
// acceptance fixture. It can run while fs_file's declaration hook is pending.
func TestNodePathRelativeRuntime(t *testing.T) {
	t.Parallel()
	paths := []string{"", ".", "..", "/", "/a", "/a/b", "/aa", "/a/b/", "a/../b", "é/中", "\x00", "a//b"}
	program := &ir.Program{Source: "node_path_relative.a", Strings: paths}
	for from := range paths {
		for to := range paths {
			call := ir.NodeHostCall{Module: "node:path", Member: "relative", Returns: ir.String, Throws: true,
				Arguments: []ir.Expression{ir.StringConstant{Index: from}, ir.StringConstant{Index: to}}}
			program.Main = append(program.Main, ir.WriteLine{Stream: ir.Stdout, Value: call})
		}
	}
	fixture, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_path_relative.a"))
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: root}
	truth := onNodeWith(t, how, fixture)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("Node failed: %+v", truth)
	}
	shared := sharedDirectory(t)
	backend := inputBackend(t, how, program, shared)
	got, binary := inputNatively(t, how, program, shared)
	for name, result := range map[string]run{"native": got, "JavaScript": backend} {
		if difference := disagreement(truth, result); difference != "" {
			t.Errorf("%s: %s", name, difference)
		}
	}
	if leaked := inputLeaks(t, how, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
	source := strings.ReplaceAll(native.C(program), "adamic_node_path_relative(", "mutant_relative(")
	helper := `static adamic_string *mutant_relative(const adamic_string *from, const adamic_string *to) { return adamic_node_path_join(2, (adamic_string *const[]){(adamic_string *)from, (adamic_string *)to}); }`
	source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+"\n"+helper, 1)
	mutant := filepath.Join(shared, "mutant")
	if err := native.Build(source, mutant, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeInput(t, how, []string{"ASAN_OPTIONS=detect_leaks=1"}, mutant)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: exit %d stderr %s", result.exitCode, result.stderr)
	}
	if difference := disagreement(truth, result); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	t.Log("relative runtime agrees on both backends; join mutant caught only by Node stdout comparison")
}
