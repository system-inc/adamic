package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Mutants preserve valid execution and ownership; only Node stdout catches them.
func TestNodeFileSystemDirectorySymlinkMutants(test *testing.T) {
	test.Parallel()
	cases := []struct{ name, body string }{
		{
			"target",
			`static adamic_string missing=ADAMIC_STRING("missing"); return adamic_node_fs_symlink(&missing,path);`,
		},
		{
			"code",
			`double result=adamic_node_fs_symlink(target,path); if(adamic_thrown!=NULL){static adamic_string wrong=ADAMIC_STRING("WRONG");adamic_release(adamic_thrown->slots[2].reference);adamic_thrown->slots[2].reference=&wrong;}return result;`,
		},
		{
			"message",
			`double result=adamic_node_fs_symlink(target,path); if(adamic_thrown!=NULL){static adamic_string wrong=ADAMIC_STRING("wrong");adamic_release(adamic_thrown->slots[1].reference);adamic_thrown->slots[1].reference=&wrong;}return result;`,
		},
	}
	for _, one := range cases {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/node_fs_directory_symlink.a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			shared := sharedDirectory(test)
			prepare := func(name string) inputRun {
				how := inputRun{directory: filepath.Dir(path), arguments: []string{writable(test, shared, name)}}
				if os.Geteuid() == 0 {
					how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
				}
				return how
			}
			truth := onNodeWith(test, prepare("node"), path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				test.Fatalf("Node failed: %+v", truth)
			}
			if one.name == "target" {
				backend := inputBackend(test, prepare("javascript"), program, shared)
				nativeRun := prepare("native")
				got, binary := inputNatively(test, nativeRun, program, shared)
				for name, observation := range map[string]run{"native": got, "JavaScript": backend} {
					if difference := disagreement(truth, observation); difference != "" {
						test.Fatalf(
							"%s: %s Node=%q got=%q stderr=%q",
							name,
							difference,
							truth.stdout,
							observation.stdout,
							observation.stderr,
						)
					}
				}
				leakRun := 0
				prepareLeaks := func() inputRun {
					leakRun++
					return prepare(fmt.Sprintf("leaks-%d", leakRun))
				}
				if leaked := inputLeaks(test, prepareLeaks, program, binary); leaked != "" {
					test.Fatal(leaked)
				}
			}
			original := native.C(program)
			source := strings.ReplaceAll(original, "adamic_node_fs_symlink(", "mutant_symlink(")
			if source == original {
				test.Fatal("mutant changed nothing")
			}
			helper := `static double mutant_symlink(const adamic_string *target,const adamic_string *path) {` +
				one.body +
				`}`
			source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+"\n"+helper, 1)
			binary := filepath.Join(shared, "mutant")
			if buildError := native.Build(source, binary, native.Options{Sanitize: true}); buildError != nil {
				test.Fatal(buildError)
			}
			how := prepare("mutant-input")
			got := executeInput(test, how, nodeFileSystemDirectorySanitizerEnvironment(), binary, how.arguments...)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				test.Fatalf("mutant failed outside comparison: exit %d stderr %s", got.exitCode, got.stderr)
			}
			if difference := disagreement(truth, got); difference != "stdout differs" {
				test.Fatalf("mutant caught by %q", difference)
			}
			test.Log("caught only by Node stdout comparison; sanitizers and leaks clean")
		})
	}
}
