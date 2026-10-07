// Package testfixture supplies a child tree for deadline proofs.
package testfixture

import (
	"context"
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
func Tree(t *testing.T, name string, startupDelay ...time.Duration) (string, string) {
	t.Helper()
	directory := t.TempDir()
	heartbeat := filepath.Join(directory, "heartbeat")
	child := filepath.Join(directory, name)
	pidfile := filepath.Join(directory, "pid")
	groupfile := filepath.Join(directory, "group")
	t.Cleanup(func() {
		data, _ := os.ReadFile(pidfile)
		pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		groupBytes, _ := os.ReadFile(groupfile)
		group, _ := strconv.Atoi(strings.TrimSpace(string(groupBytes)))
		if group > 0 && group != syscall.Getpgrp() {
			_ = syscall.Kill(-group, syscall.SIGKILL)
		} else if pid > 0 {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	})
	delay := time.Duration(0)
	if len(startupDelay) > 0 {
		delay = startupDelay[0]
	}
	script := fmt.Sprintf("#!/bin/sh\necho $$ > %q\n/bin/ps -o pgid= -p $$ > %q\ntrap '' TERM\n(sh -c 'trap \"\" TERM; sleep \"$2\"; while :; do echo tick >> \"$1\"; sleep 0.01; done' child %q %q) &\nwhile [ ! -s %q ]; do sleep 0.01; done\necho ready > %q\nwait\n", pidfile, groupfile, heartbeat, fmt.Sprintf("%.3f", delay.Seconds()), heartbeat, heartbeat+".ready")
	if err := os.WriteFile(child, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	return child, heartbeat
}

func AssertStopped(t *testing.T, started time.Time, heartbeat, diagnostic, child string) {
	t.Helper()
	if time.Since(started) > 5*time.Second {
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

// ReadyContext separates the bounded startup phase from the requested execution
// deadline. The parent's signal is written only after its grandchild heartbeat.
// Install this clock only in nonparallel tests, restoring it after the run.
func ReadyContext(t *testing.T, heartbeat string) (func(context.Context, time.Duration) (context.Context, context.CancelFunc), <-chan time.Time) {
	t.Helper()
	return ContextWhenReady(t, func() bool {
		heartbeatBytes, err := os.ReadFile(heartbeat)
		if err != nil || len(heartbeatBytes) == 0 {
			return false
		}
		ready, err := os.ReadFile(heartbeat + ".ready")
		return err == nil && len(ready) > 0
	})
}

func ContextWhenReady(t *testing.T, ready func() bool) (func(context.Context, time.Duration) (context.Context, context.CancelFunc), <-chan time.Time) {
	t.Helper()
	started := make(chan time.Time, 1)
	factory := func(parent context.Context, limit time.Duration) (context.Context, context.CancelFunc) {
		startup, stopStartup := context.WithTimeout(parent, 30*time.Second+limit)
		ctx, cancel := context.WithCancelCause(startup)
		stop := func() { cancel(context.Canceled); stopStartup() }
		t.Cleanup(stop)
		go func() {
			ticker := time.NewTicker(5 * time.Millisecond)
			defer ticker.Stop()
			startupTimer := time.NewTimer(30 * time.Second)
			defer startupTimer.Stop()
			for !ready() {
				select {
				case <-ctx.Done():
					return
				case <-startupTimer.C:
					cancel(context.DeadlineExceeded)
					return
				case <-ticker.C:
				}
			}
			at := time.Now()
			started <- at
			timer := time.NewTimer(limit)
			defer timer.Stop()
			select {
			case <-ctx.Done():
			case <-timer.C:
				cancel(context.DeadlineExceeded)
			}
		}()
		return ctx, stop
	}
	return factory, started
}

func WaitStarted(t *testing.T, started <-chan time.Time) time.Time {
	t.Helper()
	select {
	case at := <-started:
		return at
	case <-time.After(35 * time.Second):
		t.Fatal("child did not signal readiness within bounded startup")
		return time.Time{}
	}
}

// WaitTree covers callers whose execution timer begins in a later operation,
// such as the persistent compiler worker's first request.
func WaitTree(t *testing.T, heartbeat string) time.Time {
	t.Helper()
	factory, started := ReadyContext(t, heartbeat)
	_, cancel := factory(context.Background(), time.Hour)
	defer cancel()
	return WaitStarted(t, started)
}
