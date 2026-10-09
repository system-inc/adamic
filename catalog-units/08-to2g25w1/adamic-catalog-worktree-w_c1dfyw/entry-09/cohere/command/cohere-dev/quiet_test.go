package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// A quiet window holds until it ends and not after, so a holder that never says off cannot switch the trim
// off for good. A marker that cannot be read holds, and a --for past the longest is refused.
func TestAQuietWindowHoldsUntilItEnds(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	opened := time.Date(2026, 10, 6, 4, 0, 0, 0, time.Local)
	if holds, line := quietHolds(directory, opened); holds || line != "" {
		t.Fatalf("no marker held, or was described: %q", line)
	}
	if code := quietOn(directory, []string{"--for", "30m", "a cold measurement"}, opened); code != 0 {
		t.Fatalf("opening a quiet window exited %d", code)
	}
	if holds, line := quietHolds(directory, opened.Add(29*time.Minute)); !holds ||
		!strings.Contains(line, "holds until 04:30:00") || !strings.Contains(line, "a cold measurement") {
		t.Errorf("29 minutes into a 30 minute window it does not hold, or does not say why: %q", line)
	}
	if holds, line := quietHolds(directory, opened.Add(31*time.Minute)); holds || !strings.Contains(line, "expired at 04:30:00") {
		t.Errorf("a minute after a 30 minute window ended it holds, or does not say it expired: %q", line)
	}
	if code := quietOn(directory, []string{"--for"}, opened); code != 2 {
		t.Errorf("a --for with no duration was not refused: exited %d", code)
	}
	if code := quietOn(directory, []string{"--for", "5h"}, opened); code != 2 {
		t.Errorf("a five hour window was not refused: exited %d", code)
	}
	if err := os.WriteFile(quietPath(directory), []byte("tomorrow\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if holds, line := quietHolds(directory, opened); !holds || !strings.Contains(line, "does not start with when it ends") {
		t.Errorf("an unreadable marker does not hold, or does not say what is wrong: %q", line)
	}
}

// quietTrimCase is a cohere-dev built for a test, a pool home of its own, and a Go build cache of ten 1 KB
// entries three hours old under a 6 KB cap, so a trim that runs removes some and one held off removes none.
type quietTrimCase struct {
	wrapper, slots, cache string
	environment           []string
}

func newQuietTrimCase(t *testing.T) quietTrimCase {
	t.Helper()
	wrapper := filepath.Join(t.TempDir(), "cohere-dev")
	if output, err := exec.Command("go", "build", "-o", wrapper, ".").CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, output)
	}
	home := t.TempDir()
	slots := filepath.Join(userCacheDirectoryIn(home), "cohere", "test-slots")
	if err := os.MkdirAll(slots, 0o755); err != nil {
		t.Fatal(err)
	}
	// The cache in the pool home's user cache directory, where the trim guards it (poolGuards).
	cache := filepath.Join(userCacheDirectoryIn(home), "go-build")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
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
	environment := append(outsideThePool(os.Environ()), homeVariable+"="+home, "HOME="+scratchHome(t),
		"GOCACHE="+cache, tokensVariable+"=2", cacheCapVariable+"="+strconv.FormatFloat(6*1024.0/(1<<30), 'g', -1, 64))
	return quietTrimCase{wrapper: wrapper, slots: slots, cache: cache, environment: environment}
}

func (c quietTrimCase) run(t *testing.T, arguments ...string) string {
	t.Helper()
	command := exec.Command(c.wrapper, arguments...)
	command.Env = c.environment
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("cohere-dev %s: %v\n%s", strings.Join(arguments, " "), err, output)
	}
	return string(output)
}

func (c quietTrimCase) size() int64 {
	var total int64
	filepath.WalkDir(c.cache, func(path string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && entry.Name() != "README" {
			information, _ := entry.Info()
			total += information.Size()
		}
		return nil
	})
	return total
}

func (c quietTrimCase) log() string {
	contents, _ := os.ReadFile(filepath.Join(c.slots, "gocache-trim.log"))
	return string(contents)
}

// A quiet window switches the trim off, both ways (#cqfy7cv): with the marker open a trim over its cap
// removes nothing, not even looking, and its log names the window and the marker; once the window is
// released the same trim runs.
func TestAQuietWindowSwitchesTheTrimOff(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("the trim holds tokens, a Unix lock")
	}
	c := newQuietTrimCase(t)

	if output := c.run(t, "quiet", "on", "--for", "10m", "cache's cold runs"); !strings.Contains(output, "quiet window open until") {
		t.Errorf("quiet on does not say the window is open:\n%s", output)
	}
	if output := c.run(t, "status"); !strings.Contains(output, "quiet window: open") || !strings.Contains(output, "cache's cold runs") {
		t.Errorf("cohere-dev status does not show the open window:\n%s", output)
	}
	c.run(t, trimCacheVerb)
	if remaining := c.size(); remaining != 10*1024 {
		t.Errorf("the trim removed entries inside a quiet window: %d bytes left of 10 KB", remaining)
	}
	if log := c.log(); !strings.Contains(log, "not looked at: a quiet window holds until") ||
		!strings.Contains(log, quietPath(c.slots)) || !strings.Contains(log, "cache's cold runs") {
		t.Errorf("the trim's log does not name the window it held off for:\n%s", log)
	}

	if output := c.run(t, "quiet", "off"); !strings.Contains(output, "quiet window released") {
		t.Errorf("quiet off does not say the window was released:\n%s", output)
	}
	c.run(t, trimCacheVerb)
	if remaining := c.size(); remaining > 6*1024/4*3 {
		t.Errorf("with the window released the trim left %d bytes, over three quarters of the 6 KB cap", remaining)
	}
	if log := c.log(); !strings.Contains(log, "trimmed from") {
		t.Errorf("with the window released the trim did not trim:\n%s", log)
	}
}

// A window opened while the trim waits for an empty pool holds too: the trim has already measured, and
// it looks again once it holds every token, before it removes anything.
func TestAQuietWindowOpenedWhileTheTrimWaitsHoldsToo(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("the trim holds tokens, a Unix lock")
	}
	c := newQuietTrimCase(t)

	held, err := takeSlot(c.slots, 2, "a run that holds a token")
	if err != nil || held == nil {
		t.Fatalf("taking a token for the test: %v", err)
	}
	trim := exec.Command(c.wrapper, trimCacheVerb)
	trim.Env = c.environment
	if err := trim.Start(); err != nil {
		held.release()
		t.Fatal(err)
	}
	// Measuring 10 KB takes milliseconds; two seconds puts the trim well into its wait for the pool.
	time.Sleep(2 * time.Second)
	c.run(t, "quiet", "on", "--for", "10m", "lint_rules' pairs")
	held.release()
	if err := trim.Wait(); err != nil {
		t.Fatalf("the trim failed: %v", err)
	}
	if remaining := c.size(); remaining != 10*1024 {
		t.Errorf("a window opened while the trim waited did not hold it: %d bytes left of 10 KB", remaining)
	}
	log := c.log()
	if strings.Contains(log, "not looked at") {
		t.Fatalf("the trim had not reached its wait for the pool in two seconds, so this run tells nothing:\n%s", log)
	}
	if !strings.Contains(log, "cap, not trimmed: a quiet window holds until") || !strings.Contains(log, "lint_rules' pairs") {
		t.Errorf("the trim's log does not say the window stopped it after it held the pool:\n%s", log)
	}
}
