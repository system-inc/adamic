package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// A run with `HOME` set to a scratch directory still uses the machine's pool, the one under the account's own
// home, so it waits its turn with everyone else's and starts no trim of its own (#d0x1fhp). test's closing
// gate set `HOME` to a scratch directory with the go cache passed through, and ran in a pool of its own.
// Status reads the same poolDirectory every token is taken from, and names it.
func TestARunUnderAScratchHomeUsesTheMachinesPool(t *testing.T) {
	t.Parallel()
	wrapper := filepath.Join(t.TempDir(), "cohere-dev")
	if output, err := exec.Command("go", "build", "-o", wrapper, ".").CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, output)
	}
	account, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	scratch := scratchHome(t)
	status := exec.Command(wrapper, "status")
	status.Env = append(outsideThePool(os.Environ()), "HOME="+scratch)
	output, err := status.CombinedOutput()
	if err != nil {
		t.Fatalf("cohere-dev status: %v\n%s", err, output)
	}
	machine := filepath.Join(userCacheDirectoryIn(account.HomeDir), "cohere", "test-slots")
	if !strings.Contains(string(output), "("+machine+")") {
		t.Errorf("under a scratch home the pool is not the machine's, %s:\n%s", machine, output)
	}
	if _, err := os.Stat(filepath.Join(userCacheDirectoryIn(scratch), "cohere")); !os.IsNotExist(err) {
		t.Errorf("a scratch home grew a pool of its own (%v)", err)
	}
}

// A trim refuses a cache outside the user cache directory its pool lives in, since holding that pool keeps
// out none of the builds reading it: it removes nothing, and says so in its pool's log (#d0x1fhp). The setup
// is test's from 04:44: a scratch `HOME`, and a go cache passed in from elsewhere. With the cache beside the
// pool, the trim tests trim under the same scratch `HOME`, so this is the refusal's half of both ways.
func TestATrimRefusesACacheItsPoolDoesNotGuard(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("the trim holds tokens, a Unix lock")
	}
	wrapper := filepath.Join(t.TempDir(), "cohere-dev")
	if output, err := exec.Command("go", "build", "-o", wrapper, ".").CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, output)
	}
	home := t.TempDir()
	slots := filepath.Join(userCacheDirectoryIn(home), "cohere", "test-slots")
	if err := os.MkdirAll(slots, 0o755); err != nil {
		t.Fatal(err)
	}
	// Elsewhere: not under the pool home's user cache directory.
	cache := t.TempDir()
	if err := os.WriteFile(filepath.Join(cache, "README"), []byte("This directory holds cached build artifacts from the Go build system.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 10; index++ {
		path := filepath.Join(cache, fmt.Sprintf("%02x", index), fmt.Sprintf("%02xaa-d", index))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, make([]byte, 1024), 0o644); err != nil {
			t.Fatal(err)
		}
		old := time.Now().Add(-3 * time.Hour)
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}
	scratch := scratchHome(t)
	trim := exec.Command(wrapper, trimCacheVerb)
	trim.Env = append(outsideThePool(os.Environ()), homeVariable+"="+home, "HOME="+scratch, "GOCACHE="+cache,
		tokensVariable+"=2", cacheCapVariable+"="+strconv.FormatFloat(6*1024.0/(1<<30), 'g', -1, 64))
	output, err := trim.CombinedOutput()
	if exited, ok := err.(*exec.ExitError); !ok || exited.ExitCode() != 2 {
		t.Errorf("the trim of a cache its pool does not guard did not exit 2 (%v):\n%s", err, output)
	}
	remaining := 0
	filepath.WalkDir(cache, func(path string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && entry.Name() != "README" {
			remaining++
		}
		return nil
	})
	if remaining != 10 {
		t.Errorf("the trim removed entries from a cache its pool does not guard: %d of 10 left", remaining)
	}
	log, _ := os.ReadFile(filepath.Join(slots, "gocache-trim.log"))
	if !strings.Contains(string(log), "not trimmed: the Go build cache "+cache+" is outside "+userCacheDirectoryIn(home)) {
		t.Errorf("the pool's log does not say why the cache was not trimmed:\n%s", log)
	}
	if _, err := os.Stat(filepath.Join(userCacheDirectoryIn(scratch), "cohere")); !os.IsNotExist(err) {
		t.Errorf("the trim made a pool under the scratch home (%v)", err)
	}
}

// scratchHome is a home of nothing, for `HOME` in a run that must not follow it. Go's telemetry is off in it:
// a go command run beneath it counts into its home, and the background writer raced the test's cleanup
// ("directory not empty", one run in three).
func scratchHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	for _, configuration := range []string{filepath.Join(home, "Library", "Application Support"), filepath.Join(home, ".config")} {
		telemetry := filepath.Join(configuration, "go", "telemetry")
		if err := os.MkdirAll(telemetry, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(telemetry, "mode"), []byte("off"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return home
}
