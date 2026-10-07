package oracle

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

var jsonTypesDigest string

// Build once before any tests, including uncached gates. There is no persistent helper cache:
// Go's build cache owns source, dependency, toolchain and target invalidation. A new test process
// always asks Go to build the current source, so a previously built helper cannot go stale.
func prepareJSONTypes() (string, error) {
	if err := os.Unsetenv("ADAMIC_ORACLE_JSON_TYPES"); err != nil {
		return "", err
	}
	// The JSON source helper lands independently of this developer-tools harness. Before it
	// arrives there is nothing to build; Node still refuses JSON sources without its binary.
	if _, err := os.Stat(filepath.Join(repository, "oracle", "json_types.go")); os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	directory, err := os.MkdirTemp("", "adamic-json-types-")
	if err != nil {
		return "", err
	}
	failed := true
	defer func() {
		if failed {
			os.RemoveAll(directory)
		}
	}()
	// Permission probes may run Node as nobody; the binary must remain traversable.
	if err := os.Chmod(directory, 0755); err != nil {
		return "", err
	}
	binary := filepath.Join(directory, "json_types")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "build", "-trimpath", "-buildvcs=false", "-o", binary, "./oracle/json_types.go")
	command.Dir = repository
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = 5 * time.Second
	if output, err := command.CombinedOutput(); err != nil {
		return "", fmt.Errorf("building oracle/json_types.go before fixtures: %w\n%s", err, output)
	}
	if err := os.Chmod(binary, 0755); err != nil {
		return "", err
	}
	contents, err := os.ReadFile(binary)
	if err != nil {
		return "", err
	}
	// The result cache includes the actual helper, covering its transitive Go dependencies too.
	// Exclude its random temporary path from the environment identity; the bytes decide the key.
	jsonTypesDigest = fmt.Sprintf("%x", sha256.Sum256(contents))
	if err := os.Setenv("ADAMIC_ORACLE_JSON_TYPES", binary); err != nil {
		return "", err
	}
	failed = false
	return directory, nil
}

func jsonTypesContext(context, digest string) string {
	return cacheKey(context, digest)
}

func TestJSONTypesObservationKey(t *testing.T) {
	t.Parallel()
	cache := &resultCache{directory: t.TempDir()}
	for _, digest := range []string{"old helper", "changed helper"} {
		key := nodeResultKey("same fixture", "", "same node", jsonTypesContext("same harness", digest))
		got := reusableResult(t, cache, nodeResults, key, func() string { return digest })
		if got != digest {
			t.Fatalf("changed helper served stale observation %q, want %q", got, digest)
		}
	}
}

func TestJSONTypesMissingBinary(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "source.a")
	if err := os.WriteFile(path, []byte("console.log('decodeJson');\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ name, binary string }{{"unset", ""}, {"missing", filepath.Join(t.TempDir(), "missing-json-types")}} {
		t.Run(test.name, func(t *testing.T) {
			binary := test.binary
			t.Parallel()
			got := executeWith(t, []string{"ADAMIC_ORACLE_JSON_TYPES=" + binary}, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
			want := "missing ADAMIC_ORACLE_JSON_TYPES"
			if binary != "" {
				want = "missing or non-executable json_types binary " + binary
			}
			if got.exitCode != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), want) {
				t.Fatalf("missing helper must name %q: exit=%d stdout=%q stderr=%q", want, got.exitCode, got.stdout, got.stderr)
			}
		})
	}
}
