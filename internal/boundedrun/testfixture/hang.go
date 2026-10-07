// Package testfixture supplies a child tree for deadline proofs.
package testfixture

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Tree starts a forever-running leader and a grandchild heartbeat. Stopping only
// the leader leaves the heartbeat alive, which AssertStopped rejects.
func Tree(t *testing.T, name string) (string, string) {
	t.Helper()
	directory := t.TempDir()
	heartbeat := filepath.Join(directory, "heartbeat")
	child := filepath.Join(directory, name)
	pidfile := filepath.Join(directory, "pid")
	t.Cleanup(func() {
		data, _ := os.ReadFile(pidfile)
		pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		if pid > 0 {
			_ = syscall.Kill(-pid, syscall.SIGKILL)
		}
	})
	script := fmt.Sprintf("#!/bin/sh\necho $$ > %q\ntrap '' TERM\n(sh -c 'trap \"\" TERM; while :; do echo tick >> \"$1\"; sleep 0.01; done' child %q) &\nwait\n", pidfile, heartbeat)
	if err := os.WriteFile(child, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	return child, heartbeat
}

func AssertStopped(t *testing.T, started time.Time, heartbeat, diagnostic, child string) {
	t.Helper()
	if time.Since(started) > 3*time.Second {
		t.Fatalf("child %s did not return within deadline", child)
	}
	if !strings.Contains(diagnostic, child) || !strings.Contains(diagnostic, "deadline") {
		t.Fatalf("child missing from timeout report: %q", diagnostic)
	}
	before, err := os.ReadFile(heartbeat)
	if err != nil || len(before) == 0 {
		t.Fatalf("grandchild did not start: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	after, err := os.ReadFile(heartbeat)
	if err != nil || string(before) != string(after) {
		t.Fatalf("grandchild survived group kill: %v", err)
	}
	t.Logf("child=%s killed_seconds=%.3f diagnostic=%s", child, time.Since(started).Seconds(), diagnostic)
}
