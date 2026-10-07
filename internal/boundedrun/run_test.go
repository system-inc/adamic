package boundedrun

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/boundedrun/testfixture"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Not parallel: these proofs install a process-wide readiness clock.
func TestDeadlineKillsGrandchild(t *testing.T) {
	child, heartbeat := testfixture.Tree(t, "fake-node")
	factory, ready := testfixture.ReadyContext(t, heartbeat)
	installReadyClock(t, factory)
	command, release := Command(200*time.Millisecond, child)
	defer release()
	done := make(chan error, 1)
	go func() { _, err := command.CombinedOutput(); done <- err }()
	start := testfixture.WaitStarted(t, ready)
	err := waitReadyResult(t, done)
	if err == nil {
		t.Fatal("hung child passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), child)
}

func TestExitedLeaderClosesDescendantPipes(t *testing.T) {
	child, heartbeat := testfixture.Tree(t, "fake-clang")
	factory, ready := testfixture.ReadyContext(t, heartbeat)
	installReadyClock(t, factory)
	command, release := Command(3*time.Second, "/bin/sh", "-c", `"$1" & while [ ! -s "$2.ready" ]; do sleep .01; done; exit 0`, "child", child, heartbeat)
	defer release()
	done := make(chan error, 1)
	go func() { _, err := command.CombinedOutput(); done <- err }()
	start := testfixture.WaitStarted(t, ready)
	err := waitReadyResult(t, done)
	if !errors.Is(err, exec.ErrWaitDelay) || !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("pipe deadline missing: %v", err)
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), "/bin/sh")
}

type blockedWriter struct{ release, entered chan struct{} }

func (writer blockedWriter) Write(data []byte) (int, error) {
	close(writer.entered)
	<-writer.release
	return len(data), nil
}

func TestWaitCannotHangInOutputWriter(t *testing.T) {
	writer := blockedWriter{make(chan struct{}), make(chan struct{})}
	factory, ready := testfixture.ContextWhenReady(t, func() bool {
		select {
		case <-writer.entered:
			return true
		default:
			return false
		}
	})
	installReadyClock(t, factory)
	defer close(writer.release)
	command, release := Command(100*time.Millisecond, "/bin/sh", "-c", "echo output")
	defer release()
	command.Stdout = writer
	done := make(chan error, 1)
	go func() { done <- command.Run() }()
	start := testfixture.WaitStarted(t, ready)
	var err error
	select {
	case err = <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("reaping exceeded the parent wait bound")
	}
	if err == nil || !strings.Contains(err.Error(), "deadline") || time.Since(start) > 5*time.Second {
		t.Fatalf("reaping hung or lost deadline: %v", err)
	}
	t.Logf("blocked writer bounded in %.3fs", time.Since(start).Seconds())
}

func installReadyClock(t *testing.T, factory func(context.Context, time.Duration) (context.Context, context.CancelFunc)) {
	t.Helper()
	previous := WithTimeout
	WithTimeout = factory
	t.Cleanup(func() { WithTimeout = previous })
}
func waitReadyResult(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("ready child did not stop")
		return nil
	}
}

func TestReadyDeadlineStartsAfterGrandchild(t *testing.T) {
	child, heartbeat := testfixture.Tree(t, "delayed-grandchild", 650*time.Millisecond)
	factory, ready := testfixture.ReadyContext(t, heartbeat)
	installReadyClock(t, factory)
	start := time.Now()
	command, release := Command(200*time.Millisecond, child)
	defer release()
	done := make(chan error, 1)
	go func() { _, err := command.CombinedOutput(); done <- err }()
	var armed time.Time
	var err error
	select {
	case armed = <-ready:
		err = waitReadyResult(t, done)
	case err = <-done:
		select {
		case armed = <-ready:
		default:
			t.Fatalf("child terminated before grandchild readiness: %v", err)
		}
	case <-time.After(35 * time.Second):
		t.Fatal("child startup did not finish")
	}
	if armed.Before(start.Add(650 * time.Millisecond)) {
		t.Fatal("execution clock armed before delayed grandchild")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("execution deadline cause lost", err)
	}
	testfixture.AssertStopped(t, armed, heartbeat, err.Error(), child)
}

func TestCommandDefaultDeadlineIncludesStartup(t *testing.T) {
	t.Setenv("ADAMIC_CHILD_DEADLINE", "")
	before := time.Now()
	command, release := Command(time.Second, "/bin/sh", "-c", "exit 0")
	defer release()
	after := time.Now()
	deadline, ok := command.context.Deadline()
	if !ok || deadline.Before(before.Add(time.Second)) || deadline.After(after.Add(time.Second)) {
		t.Fatal("production deadline no longer starts at construction", deadline)
	}
}
