// Package testgrain enforces the test unit budget. Import it only from tests.
package testgrain

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Budget is the maximum elapsed time for a healthy grain.
const Budget = 60 * time.Second

// Kill is the hard elapsed deadline for a grain.
const Kill = 90 * time.Second

// clock is per-grain so tests can shorten deadlines without changing other grains.
type clock struct {
	now   func() time.Time
	after func(time.Duration, func()) *time.Timer
}

var realClock = clock{time.Now, time.AfterFunc}

type grain struct {
	mu       sync.Mutex
	commands []*exec.Cmd
	stopped  bool
}

var active = struct {
	sync.Mutex
	grains map[testing.TB][]*grain
}{grains: make(map[testing.TB][]*grain)}

func (g *grain) kill() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.stopped = true
	for _, cmd := range g.commands {
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
	}
}

// track owns command cleanup independently of deadline policy.
func track(t testing.TB) (*grain, func()) {
	g := new(grain)
	active.Lock()
	active.grains[t] = append(active.grains[t], g)
	active.Unlock()
	return g, func() {
		g.kill()
		active.Lock()
		stack := active.grains[t]
		for i, item := range stack {
			if item == g {
				stack = append(stack[:i], stack[i+1:]...)
				break
			}
		}
		if len(stack) == 0 {
			delete(active.grains, t)
		} else {
			active.grains[t] = stack
		}
		active.Unlock()
	}
}

func begin(t testing.TB, label string, c clock) func() {
	started := c.now()
	g, finish := track(t)
	timer := c.after(Kill, func() { g.kill(); panic(fmt.Sprintf("cooked: %s still running at 90 s; split smaller", label)) })
	return func() {
		timer.Stop()
		finish()
		elapsed := c.now().Sub(started)
		t.Logf("grain %s: %.3f s cooked=%t", label, elapsed.Seconds(), elapsed > Budget)
		if elapsed > Budget {
			t.Errorf("cooked: %s took %.1f s, over the 60 s budget; split smaller", label, elapsed.Seconds())
		}
	}
}

func beginSetup(t testing.TB, key string, c clock) func() {
	started := c.now()
	_, finish := track(t)
	return func() {
		finish()
		t.Logf("grain setup %s: %.3f s", key, c.now().Sub(started).Seconds())
	}
}

// Unit starts the case clock. Call it after shared preparation.
func Unit(t testing.TB)          { t.Helper(); unit(t, realClock) }
func unit(t testing.TB, c clock) { t.Helper(); t.Cleanup(begin(t, t.Name(), c)) }

type prepared[T any] struct{ value T }

type preparation struct {
	done  chan struct{}
	value any
	err   error
}

var preparations = struct {
	sync.Mutex
	values map[string]*preparation
}{values: make(map[string]*preparation)}

// Setup publishes one process-wide result per key, including errors. Keys must
// have a consistent result type and preparation must not depend on leaf cleanup.
// Setup has no test-side budget or deadline; Loom bounds the whole unit.
// Commands launched during preparation are killed when preparation returns.
func Setup[T any](t testing.TB, key string, prepare func() (T, error)) T {
	t.Helper()
	return setup(t, key, prepare, realClock)
}

func setup[T any](t testing.TB, key string, prepare func() (T, error), c clock) T {
	t.Helper()
	preparations.Lock()
	p, exists := preparations.values[key]
	if !exists {
		p = &preparation{done: make(chan struct{})}
		preparations.values[key] = p
	}
	preparations.Unlock()
	if !exists {
		func() {
			defer close(p.done)
			finish := beginSetup(t, key, c)
			defer finish()
			// A Fatal/Goexit inside legacy preparation must not publish a successful nil value.
			p.err = fmt.Errorf("preparation did not complete")
			value, err := prepare()
			p.value, p.err = prepared[T]{value}, err
		}()
	} else {
		<-p.done
	}
	if p.err != nil {
		t.Fatalf("grain setup %s: %v", key, p.err)
	}
	result, ok := p.value.(prepared[T])
	if !ok {
		t.Fatalf("grain setup %s: incompatible value type", key)
	}
	return result.value
}

// Command creates a command in a process group tracked by the active grains.
func Command(t testing.TB, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	return registerCommand(t, exec.Command(name, arguments...))
}

// CommandContext adds context cancellation that kills the entire process group.
// It uses the same setup/unit tracking and cleanup as Command.
func CommandContext(t testing.TB, ctx context.Context, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.CommandContext(ctx, name, arguments...)
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = time.Second
	return registerCommand(t, cmd)
}

func registerCommand(t testing.TB, cmd *exec.Cmd) *exec.Cmd {
	t.Helper()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	active.Lock()
	stack := append([]*grain(nil), active.grains[t]...)
	if len(stack) == 0 {
		active.Unlock()
		t.Fatalf("grain Command for %s requires Unit or Setup", t.Name())
		return cmd
	}
	active.Unlock()
	// Enclosing units must also own commands launched by inner preparation.
	for _, g := range stack {
		g.mu.Lock()
		g.commands = append(g.commands, cmd)
		stopped := g.stopped
		g.mu.Unlock()
		if stopped {
			t.Fatalf("grain Command for %s started after cleanup", t.Name())
		}
	}
	t.Cleanup(func() {
		if cmd.Process != nil && cmd.ProcessState == nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
	})
	return cmd
}

// Assign hashes stable identities into shards and preserves ascending indexes.
func Assign(identities []string, shards int) [][]int {
	if shards <= 0 {
		panic("grain Assign: shards must be positive")
	}
	assignments := make([][]int, shards)
	for index, identity := range identities {
		h := fnv.New64a()
		_, _ = h.Write([]byte(identity))
		shard := int(h.Sum64() % uint64(shards))
		assignments[shard] = append(assignments[shard], index)
	}
	return assignments
}

// Union checks corpus coverage and the binary's declared top-level shard tests.
func Union(t testing.TB, prefix string, shards int, assignments [][]int, total int) {
	t.Helper()
	if shards <= 0 || total < 0 || len(assignments) != shards {
		t.Errorf("grain union %s: invalid shard enumeration", prefix)
		return
	}
	seen := make([]int, total)
	for _, indices := range assignments {
		for _, index := range indices {
			if index < 0 || index >= total {
				t.Errorf("grain union %s: index %d outside [0,%d)", prefix, index, total)
				continue
			}
			seen[index]++
		}
	}
	for index, count := range seen {
		if count != 1 {
			t.Errorf("grain union %s: index %d covered %d times", prefix, index, count)
		}
	}
	unit(t, realClock)
	cmd := Command(t, os.Args[0], "-test.list", "^"+regexp.QuoteMeta(prefix)+"_[0-9]+$")
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("grain union %s: list tests: %v", prefix, err)
		return
	}
	names := make(map[string]bool)
	for _, name := range strings.Fields(string(output)) {
		if strings.HasPrefix(name, prefix+"_") {
			names[name] = true
		}
	}
	for shard := 0; shard < shards; shard++ {
		name := fmt.Sprintf("%s_%03d", prefix, shard)
		if !names[name] {
			t.Errorf("grain union: missing top-level test %s", name)
		}
		delete(names, name)
	}
	for name := range names {
		t.Errorf("grain union: extra top-level test %s", name)
	}
}

// CaughtByExactly checks the sole owner of a planted failure.
func CaughtByExactly(t testing.TB, caught map[int]bool, want int) {
	t.Helper()
	count := 0
	for _, yes := range caught {
		if yes {
			count++
		}
	}
	if count != 1 || !caught[want] {
		t.Errorf("grain planted failure: caught by %v, want exactly shard %d", caught, want)
	}
}
