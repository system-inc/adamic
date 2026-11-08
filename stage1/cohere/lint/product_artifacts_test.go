package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Owner rows check their outputs; finders leave executable probes to the owner.
func requireLintArtifacts(test *testing.T, directory string, names ...string) {
	test.Helper()
	artifacts := make([]buildcache.Artifact, 0, len(names))
	for _, name := range names {
		artifacts = append(artifacts, buildcache.Artifact{Name: name})
	}
	buildcache.RequireArtifacts(test, directory, artifacts...)
}

// An empty manifest exercises startup and argument handling without walking a type graph.
func requireLintExecutable(test *testing.T, binary string) {
	test.Helper()
	manifest := filepath.Join(test.TempDir(), "empty-manifest")
	if writeError := os.WriteFile(manifest, nil, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	buildcache.RequireArtifacts(test, filepath.Dir(binary), buildcache.Artifact{
		Name: filepath.Base(binary), Executable: true, Arguments: []string{"--manifest", manifest},
	})
}

func requireLintReferences(test *testing.T, directory string) {
	test.Helper()
	// Check both persisted answers for every populated bucket and heavy unit.
	products := testShardsAgreeRead(test, testShardsAgreePreparedRoot)
	var names []string
	for shard, rows := range products.Buckets {
		if len(rows) > 0 {
			names = append(names, fmt.Sprintf("bucket-%03d.output", shard), fmt.Sprintf("bucket-%03d.count", shard))
		}
	}
	for _, entry := range testShardsAgreeHeavy {
		names = append(names, "heavy-"+entry.Name+".output", "heavy-"+entry.Name+".count")
	}
	if len(names) == 0 {
		test.Fatal("no reference artifacts")
	}
	requireLintArtifacts(test, directory, names...)
}

func requireLintJsxBundle(test *testing.T, directory string) {
	test.Helper()
	requireLintArtifacts(test, directory, "bundle.json")
	requireLintExecutable(test, filepath.Join(directory, "native"))
	requireLintExecutable(test, filepath.Join(directory, "parser-oracle"))
}

// Keep each generated declaration's literal rule slug visible to the union census.
func requireMutantRuleProduct(test *testing.T, slug string) {
	test.Helper()
	requireLintArtifacts(test, mutantRuleProduct(test, slug), "program.mjs", "port/main.ts")
}
