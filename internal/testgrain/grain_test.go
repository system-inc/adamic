package testgrain

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// Embedding TB keeps its private method while recording expected failures.
type recorder struct {
	testing.TB
	mu             sync.Mutex
	cleanups       []func()
	logs, failures []string
}

func (r *recorder) Helper()      {}
func (r *recorder) Name() string { return "recorded" }
func (r *recorder) Cleanup(f func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanups = append(r.cleanups, f)
}
func (r *recorder) Logf(f string, args ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, fmt.Sprintf(f, args...))
}
func (r *recorder) Errorf(f string, args ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failures = append(r.failures, fmt.Sprintf(f, args...))
}
func (r *recorder) Fatalf(f string, args ...any) { r.Errorf(f, args...); panic("expected fatal") }
func (r *recorder) finish() {
	for i := len(r.cleanups) - 1; i >= 0; i-- {
		r.cleanups[i]()
	}
}
func frozen(now *time.Time) clock {
	return clock{func() time.Time { return *now }, func(_ time.Duration, f func()) *time.Timer { return time.AfterFunc(time.Hour, f) }}
}

func TestUnitClock(t *testing.T) {
	t.Parallel()
	now := time.Unix(0, 0)
	c := frozen(&now)
	r := &recorder{TB: t}
	setup(r, t.TempDir(), func() (int, error) { now = now.Add(59 * time.Second); return 1, nil }, c)
	unit(r, c)
	now = now.Add(3 * time.Second)
	r.finish()
	if len(r.failures) != 0 || r.logs[len(r.logs)-1] != "grain recorded: 3.000 s cooked=false" {
		t.Fatalf("setup leaked onto clock: logs=%v failures=%v", r.logs, r.failures)
	}
}
func TestUnitBudget(t *testing.T) {
	t.Parallel()
	for _, elapsed := range []time.Duration{Budget, Budget + time.Second} {
		now := time.Unix(0, 0)
		r := &recorder{TB: t}
		unit(r, frozen(&now))
		now = now.Add(elapsed)
		r.finish()
		if (len(r.failures) > 0) != (elapsed > Budget) {
			t.Fatalf("elapsed %v: %v", elapsed, r.failures)
		}
		if elapsed > Budget && r.failures[0] != "cooked: recorded took 61.0 s, over the 60 s budget; split smaller" {
			t.Fatal(r.failures)
		}
		want := fmt.Sprintf("grain recorded: %.3f s cooked=%t", elapsed.Seconds(), elapsed > Budget)
		if !reflect.DeepEqual(r.logs, []string{want}) {
			t.Fatal(r.logs)
		}
	}
}
func TestSetupOnce(t *testing.T) {
	t.Parallel()
	for _, bad := range []bool{false, true} {
		var calls atomic.Int32
		start := make(chan struct{})
		entered := make(chan struct{}, 16)
		release := make(chan struct{})
		var wg sync.WaitGroup
		key := fmt.Sprintf("%s/%t/%s", t.Name(), bad, t.TempDir())
		shared := new(int)
		sentinel := errors.New("shared error")
		for i := 0; i < 16; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				entered <- struct{}{}
				r := &recorder{TB: t}
				defer func() {
					if v := recover(); v != nil && v != "expected fatal" {
						panic(v)
					}
					if bad && (len(r.failures) != 1 || !strings.Contains(r.failures[0], key+": shared error")) {
						t.Errorf("unshared error: %v", r.failures)
					}
				}()
				value := Setup(r, key, func() (*int, error) {
					calls.Add(1)
					<-release
					if bad {
						return nil, sentinel
					}
					return shared, nil
				})
				if value != shared {
					t.Error("unshared value")
				}
			}()
		}
		close(start)
		for i := 0; i < 16; i++ {
			<-entered
		}
		close(release)
		wg.Wait()
		if calls.Load() != 1 {
			t.Fatalf("prepare called %d times", calls.Load())
		}
	}
}
func TestAssign(t *testing.T) {
	t.Parallel()
	identities := []string{"a", "b", "c", "a", "delta"}
	a := Assign(identities, 7)
	if !reflect.DeepEqual(a, Assign(identities, 7)) {
		t.Fatal("not deterministic")
	}
	owners := func(ids []string, assignments [][]int) map[string]int {
		out := map[string]int{}
		for shard, indices := range assignments {
			last := -1
			for _, index := range indices {
				if index <= last {
					t.Fatal("indices not ascending")
				}
				last = index
				out[ids[index]] = shard
			}
		}
		return out
	}
	reversed := []string{"delta", "a", "c", "b", "a"}
	if !reflect.DeepEqual(owners(identities, a), owners(reversed, Assign(reversed, 7))) {
		t.Fatal("order dependent")
	}
	// Known FNV-64a vectors, rather than a second copy of the algorithm.
	if got := Assign([]string{"a", "b"}, 16); !reflect.DeepEqual(got[12], []int{0}) || !reflect.DeepEqual(got[5], []int{1}) {
		t.Fatal(got)
	}
}
func TestDeclared_000(t *testing.T) { t.Parallel() }
func TestDeclared_001(t *testing.T) { t.Parallel() }
func TestUnion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		prefix      string
		shards      int
		assignments [][]int
		total       int
		want        string
	}{
		{"TestDeclared", 2, [][]int{{0}, {1}}, 2, ""},
		{"TestDeclared", 2, [][]int{{0}, {}}, 2, "index 1 covered 0 times"},
		{"TestDeclared", 2, [][]int{{0}, {0, 1}}, 2, "index 0 covered 2 times"},
		{"TestDeclared", 3, [][]int{{0}, {1}, {}}, 2, "missing top-level test TestDeclared_002"},
		{"TestDeclared", 1, [][]int{{0, 1}}, 2, "extra top-level test TestDeclared_001"},
		{"TestDeclared", 2, [][]int{{-1}, {2}}, 2, "outside [0,2)"},
		{"TestDeclared", 2, [][]int{{0, 1}}, 2, "invalid shard enumeration"},
	}
	for _, tc := range cases {
		r := &recorder{TB: t}
		Union(r, tc.prefix, tc.shards, tc.assignments, tc.total)
		r.finish()
		got := strings.Join(r.failures, "\n")
		if tc.want == "" && got != "" || tc.want != "" && !strings.Contains(got, tc.want) {
			t.Fatalf("want %q, got %q", tc.want, got)
		}
	}
}
func TestCaughtByExactly(t *testing.T) {
	t.Parallel()
	for _, caught := range []map[int]bool{{2: true}, {2: true, 3: false}, {}, {3: true}, {2: true, 3: true}} {
		r := &recorder{TB: t}
		CaughtByExactly(r, caught, 2)
		valid := caught[2] && !caught[3]
		if (len(r.failures) == 0) != valid {
			t.Fatal(caught, r.failures)
		}
	}
}

func TestKill(t *testing.T) {
	t.Parallel()
	if path := os.Getenv("TESTGRAIN_KILL_PID"); path != "" {
		c := realClock
		c.after = func(_ time.Duration, f func()) *time.Timer { return time.AfterFunc(500*time.Millisecond, f) }
		unit(t, c)
		// Shell and its child share a process group. The grandchild records its PID.
		cmd := Command(t, "sh", "-c", `sh -c 'echo $$ > "$1"; exec sleep 300' child "$1" & wait`, "parent", path)
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		_ = cmd.Wait()
		time.Sleep(time.Second)
		return
	}
	path := filepath.Join(t.TempDir(), "pid")
	cmd := exec.Command(os.Args[0], "-test.run=^TestKill$", "-test.timeout=5s")
	cmd.Env = append(os.Environ(), "TESTGRAIN_KILL_PID="+path)
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "cooked: TestKill still running at 90 s; split smaller") {
		t.Fatalf("kill did not panic: %v\n%s", err, output)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		t.Fatal(err)
	}
	defer syscall.Kill(pid, syscall.SIGKILL)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !alive(pid) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("Command grandchild %d survived grain kill", pid)
}
func alive(pid int) bool {
	if err := syscall.Kill(pid, 0); err == syscall.ESRCH {
		return false
	}
	// Linux may retain an orphan zombie until the container's init reaps it.
	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid)); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 2 && fields[2] == "Z" {
			return false
		}
	}
	return true
}

func TestSetupBudget(t *testing.T) {
	t.Parallel()
	now := time.Unix(0, 0)
	r := &recorder{TB: t}
	key := t.Name() + "/" + t.TempDir()
	value := setup(r, key, func() (int, error) { now = now.Add(61 * time.Second); return 42, nil }, frozen(&now))
	if value != 42 || !reflect.DeepEqual(r.logs, []string{"grain setup " + key + ": 61.000 s cooked=true"}) || len(r.failures) != 1 {
		t.Fatal(value, r.logs, r.failures)
	}
	later := &recorder{TB: t}
	if Setup(later, key, func() (int, error) { t.Error("prepared twice"); return 0, nil }) != 42 || len(later.failures) != 0 {
		t.Fatal("budget charged to waiter")
	}
}
func TestCommandCleanup(t *testing.T) {
	t.Parallel()
	r := &recorder{TB: t}
	now := time.Now()
	unit(r, frozen(&now))
	cmd := Command(r, "sleep", "300")
	if cmd.SysProcAttr == nil || !cmd.SysProcAttr.Setpgid {
		t.Fatal("missing process group")
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	r.finish()
	if err := cmd.Wait(); err == nil {
		t.Fatal("cleanup did not kill child")
	}
}

func TestSetupNilInterface(t *testing.T) {
	t.Parallel()
	if got := Setup[any](t, t.TempDir(), func() (any, error) { return nil, nil }); got != nil {
		t.Fatal(got)
	}
}

func TestEnclosingGrainOwnsSetupCommands(t *testing.T) {
	t.Parallel()
	r := &recorder{TB: t}
	now := time.Now()
	unit(r, frozen(&now))
	active.Lock()
	outer := active.grains[r][0]
	active.Unlock()
	Setup(r, t.TempDir(), func() (int, error) {
		cmd := Command(r, "sleep", "300")
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		defer cmd.Process.Kill()
		outer.kill()
		if err := cmd.Wait(); err == nil {
			t.Error("enclosing grain did not own inner command")
		}
		return 0, nil
	})
	r.finish()
}
