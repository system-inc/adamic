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

func nodeCountOracleProduct(test *testing.T) string {
	test.Helper()
	nodeCountOracle.Do(func() { nodeCountOracle.path = wholeMutantBuildOracleProduct(test) })
	return nodeCountOracle.path
}
func nodeCountNativeProduct(test *testing.T) string {
	test.Helper()
	nodeCountNative.Do(func() {
		nodeCountNative.path = wholeMutantBuildPortProduct(test, compilerExpressionsProductDirectory(test), true)
	})
	return nodeCountNative.path
}
func TestProduct_ParserFindsNodeCountOracle(test *testing.T) {
	test.Parallel()
	nodeCountOracleProduct(test)
}
func TestProduct_NodeCountLower(test *testing.T) {
	test.Parallel()
	wholeMutantLowerProduct(test, compilerExpressionsProductDirectory(test))
}
func TestProduct_NodeCountNative(test *testing.T) { test.Parallel(); nodeCountNativeProduct(test) }

// Count mode takes countTree's separate traversal rather than printTree's count.
// Compare its emitted summary to the unmodified Go parser's traversal.
func TestWholeNodeCountAgrees_000(test *testing.T) {
	test.Parallel()
	setupStarted := time.Now()
	oracle := nodeCountOracleProduct(test)
	binary := nodeCountNativeProduct(test)
	test.Logf("setup wall %.3fs", time.Since(setupStarted).Seconds())
	started := time.Now()
	defer func() { test.Logf("own work wall %.3fs", time.Since(started).Seconds()) }()
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	manifest := wholeManifest(test, []string{`const x = 1;`})
	args := []string{"--manifest", manifest, "--whole", "--count"}
	want := expressionExecute(test, ctx, "", oracle, args...).output
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, side := range []struct {
		name   string
		output []byte
	}{
		{"Node", expressionExecute(
			test,
			ctx,
			"",
			"node",
			append([]string{"--disable-warning=ExperimentalWarning", runner, source}, args...)...,
		).output},
		{"native", expressionExecute(test, ctx, "", binary, args...).output},
	} {
		if diff := difference(side.output, want); diff != "" {
			test.Fatalf("const x = 1; node count %s: %s", side.name, diff)
		}
	}
	test.Logf("const x = 1; node count agrees with Go: %s", want)
}
