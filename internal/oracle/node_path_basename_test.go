package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestNodePathBasenameMutants(test *testing.T) {
	test.Parallel()
	cases := []struct{ name, body string }{
		{"suffix", `return adamic_node_path_basename(path,NULL);`},
		{"dirname", `return adamic_node_path_dirname(path);`},
	}
	for _, one := range cases {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/node_path_basename.a"),
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
			source := strings.ReplaceAll(original, "adamic_node_path_basename(", "mutant_basename(")
			if source == original {
				test.Fatal("mutant changed nothing")
			}
			source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+
				"\n"+`static adamic_string *mutant_basename(const adamic_string *path,const adamic_string *suffix) {`+
				one.body+
				`}`, 1)
			shared := sharedDirectory(test)
			binary := filepath.Join(shared, "mutant")
			if buildError := native.Build(source, binary, native.Options{Sanitize: true}); buildError != nil {
				test.Fatal(buildError)
			}
			got := executeInput(test, how, nodeFSDirectorySanitizerEnvironment(), binary)
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
