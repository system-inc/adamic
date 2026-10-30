package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestNodePathBasenameMutants(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, body string }{
		{"suffix", `return adamic_node_path_basename(path,NULL);`},
		{"dirname", `return adamic_node_path_dirname(path);`},
	}
	for _, one := range cases {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_path_basename.a"))
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
			source := strings.ReplaceAll(original, "adamic_node_path_basename(", "mutant_basename(")
			if source == original {
				t.Fatal("mutant changed nothing")
			}
			source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+"\n"+`static adamic_string *mutant_basename(const adamic_string *path,const adamic_string *suffix) {`+one.body+`}`, 1)
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
			t.Log("caught only by Node stdout comparison; sanitizers and leaks clean")
		})
	}
}
