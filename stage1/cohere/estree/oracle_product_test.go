package estree

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Use an alternate module identity beneath cohere so the existing independent
// oracle source can import cohere's internal formatter without an overlay or a
// copied driver. This file is passed at the repository root, not built as a
// standalone module. GoBuild hashes this module's go.mod/go.sum and all deps.
func estreeOracleProduct(t *testing.T) string {
	t.Helper()
	module := filepath.Join(root(t), "stage1/cohere/estree/oraclebuild/go.mod")
	return buildcache.GoBuild(t, "estree-oracle", "./stage1/cohere/estree/testdata", []string{"-trimpath", "-buildvcs=false"}, "GOWORK=off", "GOFLAGS=-modfile="+module)
}

// In particular, a source-file Go build has no main-module metadata and a
// custom .mod basename can lose module inputs. Guard the package build's graph.
func TestEstreeOracleBuildInputs(t *testing.T) {
	module := filepath.Join(root(t), "stage1/cohere/estree/oraclebuild/go.mod")
	inputs, err := buildcache.GoInputs("estree-oracle", "./stage1/cohere/estree/testdata", []string{"-trimpath", "-buildvcs=false"}, []string{"GOWORK=off", "GOFLAGS=-modfile=" + module})
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string]bool)
	for _, file := range inputs.Files {
		files[file] = true
	}
	for _, file := range []string{"stage1/cohere/estree/oraclebuild/go.mod", "stage1/cohere/estree/oraclebuild/go.sum", "stage1/cohere/estree/testdata/oracle.go"} {
		if !files[file] {
			t.Fatalf("oracle build key omitted %s", file)
		}
	}
}
