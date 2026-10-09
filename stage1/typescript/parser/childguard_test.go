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
func execute(t *testing.T, directory, name string, args ...string) execution {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = directory
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = childguard.Run(command, childguard.Options{Stall: parserChildStall})
	duration := time.Since(started)
	if err != nil || stderr.Len() != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, duration}
}

var parserGuardProducts struct {
	sync.Once
	native string
}

// Fetch the declared products before this shard starts its comparison work.
func TestParserChildGuard_000(t *testing.T) {
	t.Parallel()
	parserGuardProducts.Do(func() {
		parserGuardProducts.native = compilerExpressionsNative(t, compilerExpressionsProductDirectory(t))
	})
	oracle := parserGuardOracle(t)
	started := time.Now()
	t.Cleanup(func() { t.Logf("own work %.3fs", time.Since(started).Seconds()) })
	manifest, count := jsxManifest(t)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	want := execute(t, "", oracle, "--manifest", manifest, "--whole")
	for _, got := range []execution{
		wholeNode(t, directory, manifest, false),
		execute(t, "", parserGuardProducts.native, "--manifest", manifest, "--whole"),
	} {
		if diff := difference(got.output, want.output); diff != "" {
			t.Fatal(diff)
		}
	}
	t.Logf("%d JSX inputs, %d identical whole-tree bytes, %d output lines", count, len(want.output), strings.Count(string(want.output), "\n"))
}

var parserOracleProduct struct {
	sync.Once
	binary string
}

func parserGuardOracle(t *testing.T) string {
	t.Helper()
	parserOracleProduct.Do(func() { parserOracleProduct.binary = wholeMutantBuildOracleProduct(t) })
	if parserOracleProduct.binary == "" {
		t.Fatal("parser oracle product preparation failed")
	}
	return parserOracleProduct.binary
}
