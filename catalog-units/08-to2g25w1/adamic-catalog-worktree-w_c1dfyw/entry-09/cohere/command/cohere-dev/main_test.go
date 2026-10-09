package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWholeModuleAndCountOnceAreReadFromTheArguments(t *testing.T) {
	t.Parallel()
	if !wholeModule([]string{"-race", "./..."}) || wholeModule([]string{"./internal/edit"}) {
		t.Error("whole-module detection is wrong")
	}
	for _, arguments := range [][]string{{"-count=1", "./..."}, {"-count", "1", "./..."}, {"--count=1", "./..."}} {
		if !countsOnce(arguments) {
			t.Errorf("%v does not read as -count=1", arguments)
		}
	}
	if countsOnce([]string{"-count=2", "./..."}) {
		t.Error("-count=2 reads as -count=1")
	}
}

// A slot held is a slot no one else takes until it is released, and a second slot is free while the first
// is held.
func TestASlotIsHeldUntilReleased(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("slots are a Unix lock")
	}
	directory := t.TempDir()
	first, err := takeSlot(directory, 1, "go test ./...")
	if err != nil || first == nil {
		t.Fatalf("an empty machine gave no slot: %v", err)
	}
	if second, err := takeSlot(directory, 1, "go test ./..."); err != nil || second != nil {
		t.Fatalf("a held slot was taken again (%v)", err)
	}
	if lines := holders(directory, 1); len(lines) != 1 || !strings.Contains(lines[0], "go test ./...") {
		t.Fatalf("the holder is not named: %v", lines)
	}
	if two, err := takeSlot(directory, 2, "go test ./..."); err != nil || two == nil {
		t.Fatalf("with two slots the second was not free (%v)", err)
	} else {
		two.release()
	}
	first.release()
	if again, err := takeSlot(directory, 1, "go test ./..."); err != nil || again == nil {
		t.Fatalf("a released slot was not free (%v)", err)
	} else {
		again.release()
	}
}

// Two whole-module runs on a one-slot machine: the second says it is waiting and whose run holds the slot,
// and starts when the first ends. A -count=1 whole-module run is refused before anything runs.
func TestASecondWholeModuleRunWaitsAndSaysForWhom(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("slots are a Unix lock")
	}
	wrapper := filepath.Join(t.TempDir(), "cohere-dev")
	if output, err := exec.Command("go", "build", "-o", wrapper, ".").CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, output)
	}
	// A go that takes two seconds to test anything, on the front of the path. It leaves a mark as it starts,
	// which only a run holding the slot reaches, so the second run starts once the first holds it. A fixed
	// stagger instead lost the race under the house's load, where the second run's process could start
	// first and take the slot.
	bin := t.TempDir()
	started := filepath.Join(bin, "started")
	script := "#!/bin/sh\ntouch '" + started + "'\nsleep 2\necho tested \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	home := privatePool(t)
	environment := append(outsideThePool(os.Environ()), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		homeVariable+"="+home, tokensVariable+"=1")

	run := func(arguments ...string) (string, error) {
		command := exec.Command(wrapper, append([]string{"test"}, arguments...)...)
		command.Env = environment
		output, err := command.CombinedOutput()
		return string(output), err
	}

	began := time.Now()
	outputs := make([]string, 2)
	var group sync.WaitGroup
	for index := range outputs {
		group.Add(1)
		go func() {
			defer group.Done()
			if index == 1 {
				for _, err := os.Stat(started); err != nil; _, err = os.Stat(started) {
					if time.Since(began) > time.Minute {
						t.Errorf("the first run never started testing")
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
			output, err := run("./...")
			if err != nil {
				t.Errorf("run %d failed: %v\n%s", index, err, output)
			}
			outputs[index] = output
		}()
	}
	group.Wait()
	if elapsed := time.Since(began); elapsed < 4*time.Second {
		t.Errorf("two runs on one slot finished in %s, so they overlapped", elapsed)
	}
	if !strings.Contains(outputs[1], "waiting for a token") || !strings.Contains(outputs[1], "go test -vet=all ./...") ||
		!strings.Contains(outputs[1], "after waiting") {
		t.Errorf("the second run did not say it waited and for whom:\n%s", outputs[1])
	}
	if strings.Contains(outputs[0], "waiting") {
		t.Errorf("the first run waited:\n%s", outputs[0])
	}

	output, err := run("-count=1", "./...")
	if err == nil || !strings.Contains(output, "never passes -count=1") || strings.Contains(output, "tested") {
		t.Errorf("a -count=1 whole-module run was not refused before running:\n%s", output)
	}
	if output, err := run("-count=1", "./internal/edit"); err != nil || !strings.Contains(output, "tested test -vet=all -count=1 ./internal/edit") {
		t.Errorf("a -count=1 run of one package was held back:\n%s (%v)", output, err)
	}
}

// Waiters take the slot in the order they arrived, and one that dies while waiting is passed over rather
// than waited on (#nf1qj58). Five runs on a one-slot machine: the first holds the slot, three more queue
// behind it in a known order, the middle one is killed while it waits, and the rest run in arrival order.
// With the old polling, whichever waiter polled first after a release went next, so this order held by
// chance one time in six.
func TestWaitersTakeTheSlotInArrivalOrder(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("slots are a Unix lock")
	}
	wrapper := filepath.Join(t.TempDir(), "cohere-dev")
	if output, err := exec.Command("go", "build", "-o", wrapper, ".").CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, output)
	}
	bin := t.TempDir()
	order := filepath.Join(bin, "order")
	home := privatePool(t)
	slots := filepath.Join(userCacheDirectoryIn(home), "cohere", "test-slots")
	tickets := func() int {
		entries, _ := os.ReadDir(queueDirectory(slots))
		count := 0
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), ".") {
				count++
			}
		}
		return count
	}
	waitFor := func(what string, done func() bool) {
		t.Helper()
		for deadline := time.Now().Add(time.Minute); !done(); time.Sleep(10 * time.Millisecond) {
			if time.Now().After(deadline) {
				t.Fatalf("waited a minute for %s", what)
			}
		}
	}

	// A go that records which run reached it and takes a second to test. The first run's waits for a release
	// file too, so the line forms behind it before anyone can move.
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte("#!/bin/sh\necho \"$COHERE_DEV_TEST_RUN\" >> '"+order+"'\n"+
		"[ \"$COHERE_DEV_TEST_RUN\" = A ] && while [ ! -e '"+order+".release' ]; do sleep 0.05; done\nsleep 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	runs := map[string]*exec.Cmd{}
	for index, name := range []string{"A", "B", "C", "D", "E"} {
		command := exec.Command(wrapper, "test", "./...")
		command.Env = append(outsideThePool(os.Environ()), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
			homeVariable+"="+home, tokensVariable+"=1", "COHERE_DEV_TEST_RUN="+name)
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		runs[name] = command
		if index == 0 {
			waitFor("the first run to reach go", func() bool { contents, _ := os.ReadFile(order); return string(contents) == "A\n" })
		} else {
			waitFor(name+" to join the line", func() bool { return tickets() == index })
		}
	}
	if err := runs["C"].Process.Kill(); err != nil {
		t.Fatal(err)
	}
	runs["C"].Wait()
	if err := os.WriteFile(order+".release", nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"A", "B", "D", "E"} {
		if err := runs[name].Wait(); err != nil {
			t.Errorf("run %s failed: %v", name, err)
		}
	}
	if contents, _ := os.ReadFile(order); string(contents) != "A\nB\nD\nE\n" {
		t.Errorf("the runs reached go in the order %q, want A, B, D, E: arrival order, passing over C, which died waiting",
			strings.ReplaceAll(strings.TrimSpace(string(contents)), "\n", ", "))
	}
	if left := tickets(); left != 0 {
		t.Errorf("%d tickets were left in line after every run ended", left)
	}
}

func TestOrdinalSaysAPlaceAsAPersonWould(t *testing.T) {
	t.Parallel()
	for place, want := range map[int]string{1: "1st", 2: "2nd", 3: "3rd", 4: "4th", 11: "11th", 12: "12th", 13: "13th", 21: "21st", 112: "112th"} {
		if got := ordinal(place); got != want {
			t.Errorf("ordinal(%d) is %q, want %q", place, got, want)
		}
	}
}

// A run makes itself nice, so the go command and test binaries it starts inherit it and yield the machine
// to a person; --full-priority keeps the priority it was started with, for a timed measurement, and never
// reaches go test (#2qc6j8g). Shown at nice 19 through COHERE_DEV_NICENESS, since the house runs its tests
// at nice 10 already, where a run that failed to lower itself would look the same as one that did.
func TestARunIsNiceUnlessItAsksForFullPriority(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("nice is a Unix priority")
	}
	wrapper := filepath.Join(t.TempDir(), "cohere-dev")
	if output, err := exec.Command("go", "build", "-o", wrapper, ".").CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, output)
	}
	// A go that says how nice it is and what it was asked.
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte("#!/bin/sh\necho nice=$(ps -o nice= -p $$ | tr -d ' ') \"$@\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A pool of its own, so the run never waits in line behind the machine's real ones.
	home := privatePool(t)
	niceOf := func(arguments ...string) (int, string) {
		t.Helper()
		command := exec.Command(wrapper, append([]string{"test"}, arguments...)...)
		command.Env = append(outsideThePool(os.Environ()), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
			homeVariable+"="+home, nicenessVariable+"=19")
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("cohere-dev test %v: %v\n%s", arguments, err, output)
		}
		// The stand-in's own line; the wrapper's notes, such as the corpora a private home leaves uncovered,
		// are around it.
		line := ""
		for _, candidate := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(candidate, "nice=") {
				line = candidate
			}
		}
		fields := strings.Fields(strings.TrimPrefix(line, "nice="))
		if len(fields) == 0 {
			t.Fatalf("the stand-in go printed no nice line: %q", output)
		}
		nice, err := strconv.Atoi(fields[0])
		if err != nil {
			t.Fatalf("the stand-in go printed %q", output)
		}
		return nice, strings.Join(fields[1:], " ")
	}
	output, err := exec.Command("ps", "-o", "nice=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		t.Fatal(err)
	}
	own, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		t.Fatalf("ps printed %q for this test's nice", output)
	}

	if own == 19 {
		t.Skip("this test already runs at nice 19, the most a run can lower itself, so lowering cannot be seen")
	}
	if nice, asked := niceOf("./internal/edit"); nice != 19 || asked != "test -vet=all ./internal/edit" {
		t.Errorf("a default run's go ran at nice %d asked %q, want nice 19 asked %q", nice, asked, "test -vet=all ./internal/edit")
	}
	if nice, asked := niceOf("--full-priority", "./internal/edit"); nice != own || asked != "test -vet=all ./internal/edit" {
		t.Errorf("a --full-priority run's go ran at nice %d asked %q, want this test's own nice %d and the flag left out", nice, asked, own)
	}
}

// outsideThePool is an environment with no token held, for a wrapper a test starts: the test itself may run
// under a gate's token, and a wrapper that inherited it would run inside that share rather than queue.
func outsideThePool(environment []string) []string {
	kept := make([]string, 0, len(environment))
	for _, variable := range environment {
		if !strings.HasPrefix(variable, heldTokenVariable+"=") && !strings.HasPrefix(variable, "GOMAXPROCS=") &&
			!strings.HasPrefix(variable, "GOFLAGS=") && !strings.HasPrefix(variable, homeVariable+"=") {
			kept = append(kept, variable)
		}
	}
	return kept
}

// A token's run starts its commands with Go held to the token's packages and threads, the token named, and
// the cores it is worth for a runner that is not Go,
// replacing any -p and GOMAXPROCS the caller had and keeping its other GOFLAGS, so every go command and test
// binary beneath it stays in its share (#qhg0ntb).
func TestABudgetHoldsGoToTheTokensThreads(t *testing.T) {
	t.Parallel()
	environment := budgetEnvironment([]string{"HOME=/home", "GOMAXPROCS=16", "GOFLAGS=-trimpath -p=16 -buildvcs=false", heldTokenVariable + "=9"},
		2, poolShape{tokens: 4, packages: 3, threads: 5})
	want := []string{"HOME=/home", "GOFLAGS=-trimpath -buildvcs=false -p=3", "GOMAXPROCS=5", heldTokenVariable + "=2", coresVariable + "=15"}
	if strings.Join(environment, "|") != strings.Join(want, "|") {
		t.Errorf("the budget environment is %q, want %q", environment, want)
	}
}

// The pool runs as many commands at once as it has tokens, each held to its threads and named by its
// token, and the next waits; a command started beneath a token's run runs in that share and takes no
// second token, which would wait forever in a pool its parent fills. Shown through every verb.
func TestThePoolRunsAsManyAsItHasTokens(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("tokens are a Unix lock")
	}
	wrapper := filepath.Join(t.TempDir(), "cohere-dev")
	if output, err := exec.Command("go", "build", "-o", wrapper, ".").CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, output)
	}
	// A go, and a tool, that log when they start and end, with their share, and take a second.
	bin := t.TempDir()
	log := filepath.Join(bin, "log")
	stand := "#!/bin/sh\necho \"start $COHERE_DEV_TEST_RUN token=$" + heldTokenVariable + " flags=$GOFLAGS threads=$GOMAXPROCS $*\" >> '" + log + "'\n" +
		"sleep 1\necho \"end $COHERE_DEV_TEST_RUN\" >> '" + log + "'\n"
	for _, name := range []string{"go", "tool"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(stand), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	home := privatePool(t)
	start := func(name string, environment []string, arguments ...string) *exec.Cmd {
		command := exec.Command(wrapper, arguments...)
		command.Env = append(environment, "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
			homeVariable+"="+home, tokensVariable+"=2", packagesVariable+"=4", threadsVariable+"=3",
			"COHERE_DEV_TEST_RUN="+name)
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		return command
	}
	runs := []*exec.Cmd{
		start("test", outsideThePool(os.Environ()), "test", "./internal/edit"),
		start("build", outsideThePool(os.Environ()), "build", "./command/cohere"),
		start("vet", outsideThePool(os.Environ()), "vet", "./..."),
		start("exec", outsideThePool(os.Environ()), "exec", "--", "tool", "--flag"),
	}
	// Already under a token: runs at once, in its parent's share.
	nested := start("nested", append(outsideThePool(os.Environ()), heldTokenVariable+"=7", "GOMAXPROCS=5", "GOFLAGS=-p=6"), "test", "./internal/edit")
	for _, command := range append(runs, nested) {
		if err := command.Wait(); err != nil {
			t.Errorf("a run failed: %v", err)
		}
	}

	contents, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	running, most := 0, 0
	started := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(contents)), "\n") {
		fields := strings.Fields(line)
		if fields[1] == "nested" {
			continue
		}
		if fields[0] == "start" {
			running++
			most = max(most, running)
			started[fields[1]] = strings.Join(fields[2:], " ")
		} else {
			running--
		}
	}
	if most != 2 {
		t.Errorf("at most %d ran at once on a two-token pool, want 2:\n%s", most, contents)
	}
	for name, want := range map[string]string{
		"test":  "flags=-p=4 threads=3 test -vet=all ./internal/edit",
		"build": "flags=-p=4 threads=3 build ./command/cohere",
		"vet":   "flags=-p=4 threads=3 vet ./...",
		"exec":  "flags=-p=4 threads=3 --flag",
	} {
		if got := started[name]; !strings.HasSuffix(got, want) || (!strings.HasPrefix(got, "token=1 ") && !strings.HasPrefix(got, "token=2 ")) {
			t.Errorf("%s ran with %q, want a token of the two and %q", name, got, want)
		}
	}
	if !strings.Contains(string(contents), "start nested token=7 flags=-p=6 threads=5 test -vet=all ./internal/edit") {
		t.Errorf("a run already under a token did not run in its parent's share:\n%s", contents)
	}
}

// The cache trim keeps the builds out rather than guessing which entries they read (#jc6ca7r): over its cap
// it waits for every token, so it does not touch the cache while a run holds one, then trims the least
// recently used entries to three quarters of the cap, sparing what was just written, gives the tokens back,
// and says what it did in cohere-dev status. A cache under the cap costs no token. Looks are rationed to one
// every ten minutes.
func TestTheCacheTrimHoldsThePoolAndSettlesUnderTheCap(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("the trim holds tokens, a Unix lock")
	}
	wrapper := filepath.Join(t.TempDir(), "cohere-dev")
	if output, err := exec.Command("go", "build", "-o", wrapper, ".").CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, output)
	}
	home := t.TempDir()
	slots := filepath.Join(userCacheDirectoryIn(home), "cohere", "test-slots")
	if err := os.MkdirAll(slots, 0o755); err != nil {
		t.Fatal(err)
	}

	// A Go build cache of ten 1 KB entries three hours old, and one written a minute ago.
	// The cache in the pool home's user cache directory, where the trim guards it (poolGuards).
	cache := filepath.Join(userCacheDirectoryIn(home), "go-build")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "README"), []byte("This directory holds cached build artifacts from the Go build system.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	put := func(name string, age time.Duration) string {
		path := filepath.Join(cache, name[:2], name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, make([]byte, 1024), 0o644); err != nil {
			t.Fatal(err)
		}
		when := time.Now().Add(-age)
		if err := os.Chtimes(path, when, when); err != nil {
			t.Fatal(err)
		}
		return path
	}
	old := []string{}
	for index := 0; index < 10; index++ {
		old = append(old, put(fmt.Sprintf("%02xaa-d", index), 3*time.Hour+time.Duration(index)*time.Minute))
	}
	fresh := put("a0aa-a", time.Minute)
	size := func() int64 {
		var total int64
		filepath.WalkDir(cache, func(path string, entry os.DirEntry, err error) error {
			if err == nil && !entry.IsDir() && entry.Name() != "README" {
				information, _ := entry.Info()
				total += information.Size()
			}
			return nil
		})
		return total
	}
	// A cap of 6 KB, so the trim brings 11 KB down to three quarters of it.
	capGigabytes := strconv.FormatFloat(6*1024.0/(1<<30), 'g', -1, 64)
	environment := append(outsideThePool(os.Environ()), homeVariable+"="+home, "HOME="+scratchHome(t),
		"GOCACHE="+cache, tokensVariable+"=2", cacheCapVariable+"="+capGigabytes)

	// A run holds a token, so the trim must wait for it before it touches the cache.
	held, err := takeSlot(slots, 2, "a run that holds a token")
	if err != nil || held == nil || held.number != 1 {
		t.Fatalf("taking a token for the test: %v", err)
	}
	trim := exec.Command(wrapper, trimCacheVerb)
	trim.Env = environment
	if err := trim.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	if remaining := size(); remaining != 11*1024 {
		held.release()
		t.Fatalf("the trim removed entries while a run held a token: %d bytes left", remaining)
	}
	held.release()
	if err := trim.Wait(); err != nil {
		t.Fatalf("the trim failed: %v", err)
	}
	if remaining := size(); remaining > 6*1024/4*3 {
		t.Errorf("the trim left %d bytes, over three quarters of the 6 KB cap", remaining)
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("the entry written a minute ago was trimmed")
	}
	if _, err := os.Stat(old[9]); err == nil {
		t.Error("the oldest entry survived the trim")
	}

	status := exec.Command(wrapper, "status")
	status.Env = environment
	output, err := status.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "trimmed from") || !strings.Contains(string(output), "Go build cache: cap") {
		t.Errorf("cohere-dev status does not report the trim (%v):\n%s", err, output)
	}

	if !claimCacheLook(t.TempDir(), time.Now()) {
		t.Error("a first look was not due")
	}
	rationed := t.TempDir()
	claimCacheLook(rationed, time.Now())
	if claimCacheLook(rationed, time.Now().Add(time.Minute)) {
		t.Error("a second look a minute later was due, inside the ten minutes")
	}
}

// privatePool is a home of its own for the wrappers a test starts, so they queue in a pool of their own, with
// its cache already looked at, so no wrapper starts the background trim: it would run the test's stand-in go
// to find the cache.
func privatePool(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	slots := filepath.Join(userCacheDirectoryIn(home), "cohere", "test-slots")
	if err := os.MkdirAll(slots, 0o755); err != nil {
		t.Fatal(err)
	}
	claimCacheLook(slots, time.Now())
	return home
}

// The trim never keeps a token idle while another runs: with a long job holding token 2, it holds nothing,
// so token 1 stays free for anyone, and it trims once the job ends. Taking tokens one at a time and
// waiting for each, it held token 1 idle for the whole job, half the pool behind a land gate for minutes
// (2026-10-05 15:55, #jc6ca7r).
func TestTheCacheTrimKeepsNoTokenIdleWhileAnotherRuns(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("the trim holds tokens, a Unix lock")
	}
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

	// The long job holds token 2: take token 1 for a moment to reach it, then give token 1 back.
	first, err := takeSlot(slots, 2, "reaching token 2")
	if err != nil || first == nil || first.number != 1 {
		t.Fatalf("taking token 1: %v", err)
	}
	job, err := takeSlot(slots, 2, "a long land gate")
	if err != nil || job == nil || job.number != 2 {
		t.Fatalf("taking token 2: %v", err)
	}
	first.release()

	trim := exec.Command(wrapper, trimCacheVerb)
	trim.Env = append(outsideThePool(os.Environ()), homeVariable+"="+home, "HOME="+scratchHome(t),
		"GOCACHE="+cache, tokensVariable+"=2", cacheCapVariable+"="+strconv.FormatFloat(6*1024.0/(1<<30), 'g', -1, 64))
	if err := trim.Start(); err != nil {
		t.Fatal(err)
	}

	// Through three seconds of the job, token 1 must be free whenever the test looks: anyone may take it.
	// A hold counts once two looks in a row find it: the trim's own look takes token 1 for microseconds
	// before it finds token 2 busy and lets go, and a look landing in that instant is not idle time.
	idle := time.Duration(0)
	const step = 20 * time.Millisecond
	wasHeld := false
	for end := time.Now().Add(3 * time.Second); time.Now().Before(end); time.Sleep(step) {
		lock, err := os.OpenFile(lockPath(slots, 1), os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		taken, _ := tryLockFile(lock)
		if taken {
			unlockFile(lock)
		} else if wasHeld {
			idle += step
		}
		wasHeld = !taken
		lock.Close()
	}
	job.release()
	if err := trim.Wait(); err != nil {
		t.Fatalf("the trim failed: %v", err)
	}
	if idle > 0 {
		t.Errorf("the trim held token 1 idle for about %s while the job held token 2", idle)
	}
	status, _ := os.ReadFile(filepath.Join(slots, "gocache.status"))
	if !strings.Contains(string(status), "trimmed from") {
		t.Errorf("the trim did not trim once the job ended: %q", status)
	}
}
