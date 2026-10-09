package parser

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

var nodeCountOracle struct {
	sync.Once
	path string
}
var nodeCountNative struct {
	sync.Once
	path string
}

func nodeCountOracleProduct(t *testing.T) string {
	t.Helper()
	nodeCountOracle.Do(func() { nodeCountOracle.path = wholeMutantBuildOracleProduct(t) })
	return nodeCountOracle.path
}
func nodeCountNativeProduct(t *testing.T) string {
	t.Helper()
	nodeCountNative.Do(func() {
		nodeCountNative.path = wholeMutantBuildPortProduct(t, compilerExpressionsProductDirectory(t), true)
	})
	return nodeCountNative.path
}
func TestProduct_NodeCountOracle(t *testing.T) { t.Parallel(); nodeCountOracleProduct(t) }
func TestProduct_NodeCountLower(t *testing.T) {
	t.Parallel()
	wholeMutantLowerProduct(t, compilerExpressionsProductDirectory(t))
}
func TestProduct_NodeCountNative(t *testing.T) { t.Parallel(); nodeCountNativeProduct(t) }

// Count mode takes countTree's separate traversal rather than printTree's count.
// Compare its emitted summary to the unmodified Go parser's traversal.
func TestWholeNodeCountAgrees_000(t *testing.T) {
	t.Parallel()
	setupStarted := time.Now()
	oracle := nodeCountOracleProduct(t)
	binary := nodeCountNativeProduct(t)
	t.Logf("setup wall %.3fs", time.Since(setupStarted).Seconds())
	started := time.Now()
	defer func() { t.Logf("own work wall %.3fs", time.Since(started).Seconds()) }()
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	manifest := wholeManifest(t, []string{`const x = 1;`})
	args := []string{"--manifest", manifest, "--whole", "--count"}
	want := expressionExecute(t, ctx, "", oracle, args...).output
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name   string
		output []byte
	}{
		{"Node", expressionExecute(t, ctx, "", "node", append([]string{"--disable-warning=ExperimentalWarning", runner, source}, args...)...).output},
		{"native", expressionExecute(t, ctx, "", binary, args...).output},
	} {
		if diff := difference(side.output, want); diff != "" {
			t.Fatalf("const x = 1; node count %s: %s", side.name, diff)
		}
	}
	t.Logf("const x = 1; node count agrees with Go: %s", want)
}
