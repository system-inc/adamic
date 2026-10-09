package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// A waiter names each new holder, not only the first it saw (#vz3g7cn). On a one-token machine the test
// holds the token, two runs queue behind it in a known order, and the test lets go: the first waiter takes
// the token, and the second, still waiting, names that run as the holder now. Before, it went on naming the
// holder it first saw, after that holder was gone.
func TestAWaiterNamesEachNewHolder(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("tokens are a Unix lock")
	}
	wrapper := filepath.Join(t.TempDir(), "cohere-dev")
	if output, err := exec.Command("go", "build", "-o", wrapper, ".").CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, output)
	}
	home := privatePool(t)
	slots := filepath.Join(userCacheDirectoryIn(home), "cohere", "test-slots")
	environment := append(outsideThePool(os.Environ()), homeVariable+"="+home, tokensVariable+"=1")

	first, err := takeSlot(slots, 1, "the first holder, which the test lets go")
	if err != nil || first == nil {
		t.Fatalf("taking the token: %v", err)
	}
	tickets := func() int {
		entries, _ := os.ReadDir(queueDirectory(slots))
		live := 0
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), ".") {
				live++
			}
		}
		return live
	}
	type waiter struct {
		command *exec.Cmd
		output  *strings.Builder
	}
	queue := func(marker string, want int) waiter {
		var output strings.Builder
		command := exec.Command(wrapper, "exec", "--", "sh", "-c", "sleep 2 # "+marker)
		command.Env, command.Stdout, command.Stderr = environment, &output, &output
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		for deadline := time.Now().Add(time.Minute); tickets() < want; time.Sleep(20 * time.Millisecond) {
			if time.Now().After(deadline) {
				t.Fatalf("%s never joined the line", marker)
			}
		}
		return waiter{command, &output}
	}
	second := queue("the second run", 1)
	third := queue("the third run", 2)
	first.release()
	for _, run := range []waiter{second, third} {
		if err := run.command.Wait(); err != nil {
			t.Fatalf("a queued run failed: %v\n%s", err, run.output)
		}
	}
	if output := third.output.String(); !strings.Contains(output, "the first holder") ||
		!strings.Contains(output, "the token holders changed; now:") || !strings.Contains(output, "the second run") {
		t.Errorf("the third run did not name the second as the holder once it took the token:\n%s", output)
	}
}

// cohere-dev status shows land's lock beside the tokens: who holds it, or free, and who waits for it.
func TestStatusShowsTheLandLock(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("the land lock is a Unix lock")
	}
	wrapper := filepath.Join(t.TempDir(), "cohere-dev")
	if output, err := exec.Command("go", "build", "-o", wrapper, ".").CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, output)
	}
	home := privatePool(t)
	land := filepath.Join(userCacheDirectoryIn(home), "cohere", "test-slots", "land")
	if err := os.MkdirAll(land, 0o755); err != nil {
		t.Fatal(err)
	}
	status := func() string {
		t.Helper()
		command := exec.Command(wrapper, "status")
		command.Env = append(outsideThePool(os.Environ()), homeVariable+"="+home)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("cohere-dev status: %v\n%s", err, output)
		}
		return string(output)
	}
	if output := status(); !strings.Contains(output, "land lock: free") || !strings.Contains(output, "land line: 0 waiting") {
		t.Errorf("status does not show a free land lock:\n%s", output)
	}
	held, err := takeSlot(land, 1, "landing 1234abcd from a worktree")
	if err != nil || held == nil {
		t.Fatalf("taking the land lock: %v", err)
	}
	defer held.release()
	ticket, err := joinQueue(land)
	if err != nil {
		t.Fatal(err)
	}
	defer ticket.leave()
	if output := status(); !strings.Contains(output, "land lock: pid") || !strings.Contains(output, "landing 1234abcd") ||
		!strings.Contains(output, "land line: 1 waiting") {
		t.Errorf("status does not show the land lock's holder and its line:\n%s", output)
	}
}
