package replace

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var errHeld = errors.New("the destination is held open")

// A rename refused while the destination is held succeeds once it is let go, and one refused for any
// other reason fails at once. Windows refuses the way the fake below does; nothing else does.
func TestARenameRefusedWhileHeldIsRetriedUntilItIsLetGo(t *testing.T) {
	t.Parallel()
	refusals := 3
	rename := func(string, string) error {
		if refusals > 0 {
			refusals--
			return errHeld
		}
		return nil
	}
	slept := time.Duration(0)
	sleep := func(pause time.Duration) { slept += pause }
	held := func(err error) bool { return errors.Is(err, errHeld) }

	if err := replacing(rename, held, sleep, "from", "to"); err != nil {
		t.Fatalf("a rename refused three times and then allowed failed: %v", err)
	}
	if refusals != 0 || slept == 0 {
		t.Fatalf("it did not retry: %d refusals left, slept %s", refusals, slept)
	}

	other := errors.New("no such directory")
	calls := 0
	failing := func(string, string) error { calls++; return other }
	if err := replacing(failing, held, sleep, "from", "to"); !errors.Is(err, other) || calls != 1 {
		t.Fatalf("a rename refused for another reason was retried (%d calls) or lost its error: %v", calls, err)
	}
}

// A destination held for good fails after the bound with the refusal itself, rather than waiting forever.
func TestARenameHeldForGoodFailsAfterTheBound(t *testing.T) {
	t.Parallel()
	slept := time.Duration(0)
	err := replacing(func(string, string) error { return errHeld }, func(error) bool { return true },
		func(pause time.Duration) { slept += pause }, "from", "to")
	if !errors.Is(err, errHeld) {
		t.Fatalf("a held destination did not fail with its refusal: %v", err)
	}
	if slept < retryBound || slept > retryBound+200*time.Millisecond {
		t.Fatalf("waited %s, want about %s", slept, retryBound)
	}
}

// File replaces a file, on every platform, including one another handle in this process has open: on
// Windows that is the refusal it retries through, released here partway into the bound.
func TestFileReplacesAFileHeldOpenBriefly(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	destination := filepath.Join(directory, "table.gob")
	source := filepath.Join(directory, ".table-new")
	if err := os.WriteFile(destination, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	reader, err := os.Open(destination)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		reader.Close()
	}()
	if err := File(source, destination); err != nil {
		t.Fatalf("replacing a file held open for 100ms failed: %v", err)
	}
	if contents, _ := os.ReadFile(destination); string(contents) != "new" {
		t.Fatalf("the destination holds %q", contents)
	}
}
