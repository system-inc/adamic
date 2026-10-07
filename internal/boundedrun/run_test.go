package boundedrun

import (
	"errors"
	"github.com/system-inc/adamic/internal/boundedrun/testfixture"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestDeadlineKillsGrandchild(t *testing.T) {
	t.Parallel()
	child, heartbeat := testfixture.Tree(t, "fake-node")
	command, release := Command(200*time.Millisecond, child)
	defer release()
	start := time.Now()
	_, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("hung child passed")
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), child)
}

func TestExitedLeaderClosesDescendantPipes(t *testing.T) {
	t.Parallel()
	child, heartbeat := testfixture.Tree(t, "fake-clang")
	command, release := Command(3*time.Second, "/bin/sh", "-c", `"$1" & sleep .1; exit 0`, "child", child)
	defer release()
	start := time.Now()
	_, err := command.CombinedOutput()
	if !errors.Is(err, exec.ErrWaitDelay) || !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("pipe deadline missing: %v", err)
	}
	testfixture.AssertStopped(t, start, heartbeat, err.Error(), "/bin/sh")
}

type blockedWriter struct{ release chan struct{} }

func (writer blockedWriter) Write(data []byte) (int, error) { <-writer.release; return len(data), nil }

func TestWaitCannotHangInOutputWriter(t *testing.T) {
	t.Parallel()
	writer := blockedWriter{make(chan struct{})}
	defer close(writer.release)
	command, release := Command(100*time.Millisecond, "/bin/sh", "-c", "echo output")
	defer release()
	command.Stdout = writer
	start := time.Now()
	done := make(chan error, 1)
	go func() { done <- command.Run() }()
	var err error
	select {
	case err = <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("reaping exceeded the parent wait bound")
	}
	if err == nil || !strings.Contains(err.Error(), "deadline") || time.Since(start) > 3*time.Second {
		t.Fatalf("reaping hung or lost deadline: %v", err)
	}
	t.Logf("blocked writer bounded in %.3fs", time.Since(start).Seconds())
}
