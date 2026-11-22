package childguard

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// reapWatch signals as killGroup does, and notes for each signal whether the leader was already reaped then. It asks
// the kernel rather than the group: waitid on a reaped child fails with `ECHILD`. A signal after the reap is noted and
// not sent, since the id may already lead another group.
type reapWatch struct {
	sync.Mutex
	afterReap []bool
}

func (w *reapWatch) signal(leader int) {
	reaped := errors.Is(waitid(leader, syscall.WEXITED|syscall.WNOWAIT|syscall.WNOHANG), syscall.ECHILD)
	w.Lock()
	w.afterReap = append(w.afterReap, reaped)
	w.Unlock()
	if !reaped {
		killGroup(leader)
	}
}

// kills returns each signal so far, true where it came after the reap.
func (w *reapWatch) kills() []bool {
	w.Lock()
	defer w.Unlock()
	return append([]bool(nil), w.afterReap...)
}

// The group is killed once the leader exits, before it is reaped, and never after: the kill still reaches the
// descendant the leader left, and none lands once the leader's pid is free for another process to take.
func TestCaptureSignalsTheGroupOnlyBeforeTheReap(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	var watch reapWatch
	var stdout strings.Builder
	if err := captureTo(ctx, child("descendant"), &stdout, nil, watch.signal); err != nil || !strings.HasPrefix(stdout.String(), "parent ") {
		t.Fatalf("err %v, stdout %q", err, stdout.String())
	}
	if kills := watch.kills(); len(kills) != 1 || kills[0] {
		t.Fatalf("signals, true where after the reap: %v, want one before it", kills)
	}
}

// ctx's kill and the reap share one lock, so ctx ending while the leader runs, as it exits, or after it is reaped never
// signals the group after the reap.
func TestCaptureCancelSignalsTheGroupOnlyBeforeTheReap(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	var watch reapWatch
	err := captureTo(ctx, child("silent"), nil, nil, watch.signal)
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("a silent leader at ctx's end: err %v, want its kill's exit", err)
	}
	// ctx's kill while the leader runs, then the kill of what it left.
	if kills := watch.kills(); len(kills) != 2 || kills[0] || kills[1] {
		t.Fatalf("signals, true where after the reap: %v, want two before it", kills)
	}
	// ctx ending around the leader's exit, at steps across a short child's life, races its kill with the reap.
	for step := range 20 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(step)*10*time.Millisecond)
		var watch reapWatch
		_ = captureTo(ctx, child("finish"), nil, nil, watch.signal)
		cancel()
		// A kill started as stop() ran may still be waiting on the lock; it must find the leader reaped and do nothing.
		time.Sleep(10 * time.Millisecond)
		for _, afterReap := range watch.kills() {
			if afterReap {
				t.Fatalf("ctx ending after %dms: signalled the group after the reap: %v", step*10, watch.kills())
			}
		}
	}
}
