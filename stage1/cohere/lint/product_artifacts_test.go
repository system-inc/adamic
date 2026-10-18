package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Owner rows check their outputs; finders leave executable probes to the owner.
func requireLintArtifacts(t *testing.T, directory string, names ...string) {
	t.Helper()
	artifacts := make([]buildcache.Artifact, 0, len(names))
	for _, name := range names {
		artifacts = append(artifacts, buildcache.Artifact{Name: name})
	}
	buildcache.RequireArtifacts(t, directory, artifacts...)
}

// An empty manifest exercises startup and argument handling without walking a type graph.
func requireLintExecutable(t *testing.T, binary string) {
	t.Helper()
	manifest := filepath.Join(t.TempDir(), "empty-manifest")
	if err := os.WriteFile(manifest, nil, 0600); err != nil {
		t.Fatal(err)
	}
	buildcache.RequireArtifacts(t, filepath.Dir(binary), buildcache.Artifact{
		Name: filepath.Base(binary), Executable: true, Arguments: []string{"--manifest", manifest},
	})
}

func requireLintReferences(t *testing.T, directory string) {
	t.Helper()
	// Check both persisted answers for every populated bucket and heavy unit.
	products := testShardsAgreeRead(t, testShardsAgreePreparedRoot)
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
		t.Fatal("no reference artifacts")
	}
	requireLintArtifacts(t, directory, names...)
}

func requireLintJSXBundle(t *testing.T, directory string) {
	t.Helper()
	requireLintArtifacts(t, directory, "bundle.json")
	requireLintExecutable(t, filepath.Join(directory, "native"))
	requireLintExecutable(t, filepath.Join(directory, "parser-oracle"))
}

// Keep each generated declaration's literal rule slug visible to the union census.
func requireMutantRuleProduct(t *testing.T, slug string) {
	t.Helper()
	requireLintArtifacts(t, mutantRuleProduct(t, slug), "program.mjs", "port/main.ts")
}
