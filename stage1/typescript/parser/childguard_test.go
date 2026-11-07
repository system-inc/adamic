package parser

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// Across Go, Node and sanitized native on 348 JSX inputs, the largest output
// gap was 57.56ms unloaded (25.59ms with 2 x NumCPU = ten busy processes).
// Ten seconds leaves over 170x headroom; silent startup uses the shared default.
const parserChildStall = 10 * time.Second

// Capture the large corpus in a file while childguard observes output progress.
func execute(test *testing.T, directory, name string, args ...string) execution {
	test.Helper()
	command := exec.Command(name, args...)
	command.Dir = directory
	output, temporaryFileError := os.CreateTemp(test.TempDir(), "stdout-")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	runError := childguard.Run(command, childguard.Options{Stall: parserChildStall})
	duration := time.Since(started)
	if runError != nil || stderr.Len() != 0 {
		test.Fatalf("%s %v: %v\n%s", name, args, runError, &stderr)
	}
	data, readError := os.ReadFile(output.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	return execution{data, duration}
}

var parserGuardProducts struct {
	sync.Once
	native string
}

// Fetch the declared products before this shard starts its comparison work.
func TestParserChildGuard_000(test *testing.T) {
	test.Parallel()
	parserGuardProducts.Do(func() {
		parserGuardProducts.native = compilerExpressionsNative(test, compilerExpressionsProductDirectory(test))
	})
	oracle := parserGuardOracle(test)
	started := time.Now()
	test.Cleanup(func() { test.Logf("own work %.3fs", time.Since(started).Seconds()) })
	manifest, count := jsxManifest(test)
	directory, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := execute(test, "", oracle, "--manifest", manifest, "--whole")
	for _, got := range []execution{
		wholeNode(test, directory, manifest, false),
		execute(test, "", parserGuardProducts.native, "--manifest", manifest, "--whole"),
	} {
		if diff := difference(got.output, want.output); diff != "" {
			test.Fatal(diff)
		}
	}
	test.Logf(
		"%d JSX inputs, %d identical whole-tree bytes, %d output lines",
		count,
		len(want.output),
		strings.Count(string(want.output), "\n"),
	)
}

var parserOracleProduct struct {
	sync.Once
	binary string
}

func parserGuardOracle(test *testing.T) string {
	test.Helper()
	parserOracleProduct.Do(func() { parserOracleProduct.binary = wholeMutantBuildOracleProduct(test) })
	if parserOracleProduct.binary == "" {
		test.Fatal("parser oracle product preparation failed")
	}
	return parserOracleProduct.binary
}
