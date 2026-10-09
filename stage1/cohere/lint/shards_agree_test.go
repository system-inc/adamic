package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
	"github.com/system-inc/adamic/stage1/cohere/lint/shards"
)

// Fixed headroom: corpus growth never changes this count or moves another case.
// ADAMIC_TEST_SHARD=i/n selects unit indices congruent to i modulo n; unset runs
// all top-level units. The compiler's 38 root files are pinned by compilerCommit.
const testShardsAgreeShards = 256
const shardsAgreeBudget = 60 * time.Second
const shardsAgreeDeadline = 75 * time.Second

type shardsAgreeCase struct{ Key, Row string }
type shardsAgreePaths struct{ oracle, lowered, native, corpus string }

func shardsAgreeOwner(key string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum64() % testShardsAgreeShards)
}

func shardsAgreePartition(cases []shardsAgreeCase) ([][]shardsAgreeCase, error) {
	units := make([][]shardsAgreeCase, testShardsAgreeShards)
	expected := map[string]string{}
	for _, c := range cases {
		if _, ok := expected[c.Key]; ok {
			return nil, fmt.Errorf("repeated case id %q", c.Key)
		}
		expected[c.Key] = c.Row
		owner := shardsAgreeOwner(c.Key)
		units[owner] = append(units[owner], c)
	}
	seen := map[string]bool{}
	total := 0
	for index, unit := range units {
		for _, c := range unit {
			if seen[c.Key] || expected[c.Key] != c.Row || shardsAgreeOwner(c.Key) != index {
				return nil, fmt.Errorf("invalid union case %q", c.Key)
			}
			seen[c.Key] = true
			total++
		}
	}
	if len(units) != testShardsAgreeShards || total != len(cases) || len(seen) != len(expected) {
		return nil, fmt.Errorf("union has %d cases, want %d", total, len(cases))
	}
	for key := range expected {
		if !seen[key] {
			return nil, fmt.Errorf("missing case %q", key)
		}
	}
	return units, nil
}

func shardsAgreeSelection(text string) (int, int, error) {
	if text == "" {
		return 0, 1, nil
	}
	fields := strings.Split(text, "/")
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("ADAMIC_TEST_SHARD must be i/n")
	}
	i, e1 := strconv.Atoi(fields[0])
	n, e2 := strconv.Atoi(fields[1])
	if e1 != nil || e2 != nil || n < 1 || n > testShardsAgreeShards || i < 0 || i >= n {
		return 0, 0, fmt.Errorf("invalid ADAMIC_TEST_SHARD %q", text)
	}
	return i, n, nil
}

// This is a build-input entry point, not a test: compile with go test -c, then
// run lint.test --adamic-build-shardsagree before measuring any units. Get and
// Product address the same Inputs. All persistent products use internal/buildcache.
func init() {
	if len(os.Args) > 1 && os.Args[1] == "--adamic-build-shardsagree" {
		paths, err := shardsAgreeBuildProducts(nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("oracle %s\nlowered %s\nnative %s\ncorpus %s\n", paths.oracle, paths.lowered, paths.native, paths.corpus)
		os.Exit(0)
	}
}

func shardsAgreeBuildProducts(t *testing.T) (shardsAgreePaths, error) {
	files := []string{"stage1/cohere/lint", "stage1/typescript", "internal", "bridge", "cohere", "go.mod"}
	flags := []string{"package=" + packageDirectory, "sanitize=false", "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_TOOLS=" + os.Getenv("ADAMIC_TOOLS")}
	toolchain := []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}
	product := func(name string, build func(string) error) (string, error) {
		inputs := buildcache.Inputs{Name: name, Files: files, Flags: flags, Toolchain: toolchain}
		if t != nil {
			return buildcache.Product(t, inputs, build), nil
		}
		return buildcache.Get(inputs, build)
	}
	var p shardsAgreePaths
	var err error
	p.oracle, err = product("lint-shardsagree-oracle", func(dir string) error { _, err := goOracleIn(packageDirectory, dir); return err })
	if err != nil {
		return p, err
	}
	p.lowered, err = product("lint-shardsagree-lowered", func(dir string) error {
		if _, err := registry.Generate(packageDirectory); err != nil {
			return err
		}
		program, err := load.Load([]string{filepath.Join(packageDirectory, "main.ts")})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(lowered)), 0644)
	})
	if err != nil {
		return p, err
	}
	p.native, err = product("lint-shardsagree-native", func(dir string) error {
		source, err := os.ReadFile(filepath.Join(p.lowered, "main.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(dir, "scanner"), native.Options{})
	})
	if err != nil {
		return p, err
	}
	p.corpus, err = product("lint-shardsagree-upstream-inputs", func(dir string) error {
		rows, err := captureUpstream(packageDirectory, dir)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return fmt.Errorf("empty upstream corpus")
		}
		fixture := regexp.MustCompile(`(?:^|/)case-[0-9]+/(.*)$`)
		var cases []shardsAgreeCase
		for _, row := range rows {
			if strings.HasSuffix(row, "\tunsupported-recovery") {
				continue
			}
			fields := strings.SplitN(row, "\t", 2)
			mode := ""
			if len(fields) == 2 {
				mode = fields[1]
			}
			name := fixture.FindStringSubmatch(filepath.ToSlash(fields[0]))
			if len(name) != 2 {
				return fmt.Errorf("unknown capture path %s", fields[0])
			}
			source, err := os.ReadFile(fields[0])
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(dir, fields[0])
			if err != nil {
				return err
			}
			key := fmt.Sprintf("cohere/fixtures/%s\t%s\t%x", name[1], mode, sha256.Sum256(source))
			if len(fields) == 2 {
				relative += "\t" + mode
			}
			cases = append(cases, shardsAgreeCase{key, relative})
		}
		data, err := json.Marshal(cases)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "cases.json"), data, 0644)
	})
	return p, err
}

var shardsAgreeState struct {
	sync.Mutex
	paths shardsAgreePaths
	cases []shardsAgreeCase
	units [][]shardsAgreeCase
}

func shardsAgreeProducts(t *testing.T) (string, string) {
	t.Helper()
	shardsAgreeState.Lock()
	defer shardsAgreeState.Unlock()
	if shardsAgreeState.paths.native == "" {
		paths, err := shardsAgreeBuildProducts(t)
		if err != nil {
			t.Fatal(err)
		}
		shardsAgreeState.paths = paths
	}
	return filepath.Join(shardsAgreeState.paths.oracle, "oracle"), filepath.Join(shardsAgreeState.paths.native, "scanner")
}

func shardsAgreeEnumeration(t *testing.T) ([]shardsAgreeCase, [][]shardsAgreeCase, string, string) {
	t.Helper()
	oracle, binary := shardsAgreeProducts(t)
	shardsAgreeState.Lock()
	defer shardsAgreeState.Unlock()
	if shardsAgreeState.units != nil {
		return shardsAgreeState.cases, shardsAgreeState.units, oracle, binary
	}
	var cases []shardsAgreeCase
	rows := generated(t)
	if len(rows) == 0 {
		t.Fatal("empty generated corpus")
	}
	directory, err := os.MkdirTemp(sharedDirectory, "shardsagree-generated-")
	if err != nil {
		t.Fatal(err)
	}
	copied := map[string]string{}
	for _, row := range rows {
		fields := strings.SplitN(row, "\t", 2)
		mode := ""
		if len(fields) == 2 {
			mode = fields[1]
		}
		path, ok := copied[fields[0]]
		if !ok {
			data, err := os.ReadFile(fields[0])
			if err != nil {
				t.Fatal(err)
			}
			path = filepath.Join(directory, filepath.Base(fields[0]))
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			copied[fields[0]] = path
		}
		relativeRow := path
		if len(fields) == 2 {
			relativeRow += "\t" + mode
		}
		cases = append(cases, shardsAgreeCase{"stage1/cohere/lint/lint_test.go/generated/" + filepath.Base(path) + "\t" + mode, relativeRow})
	}
	data, err := os.ReadFile(filepath.Join(shardsAgreeState.paths.corpus, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var upstream []shardsAgreeCase
	if err := json.Unmarshal(data, &upstream); err != nil {
		t.Fatal(err)
	}
	if len(upstream) == 0 {
		t.Fatal("empty upstream corpus")
	}
	for _, c := range upstream {
		fields := strings.SplitN(c.Row, "\t", 2)
		c.Row = filepath.Join(shardsAgreeState.paths.corpus, fields[0])
		if len(fields) == 2 {
			c.Row += "\t" + fields[1]
		}
		cases = append(cases, c)
	}
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("set ADAMIC_TYPESCRIPT_SOURCE to the pinned TypeScript checkout")
	}
	pin := execute(t, source, "git", "rev-parse", "HEAD")
	if strings.TrimSpace(string(pin.output)) != compilerCommit {
		t.Fatal("wrong TypeScript pin")
	}
	matches, err := filepath.Glob(filepath.Join(source, "src/compiler/*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 38 {
		t.Fatalf("pinned compiler corpus has %d files, want 38", len(matches))
	}
	for _, row := range matches {
		cases = append(cases, shardsAgreeCase{"TypeScript/src/compiler/" + filepath.Base(row) + "\tall", row})
	}
	units, err := shardsAgreePartition(cases)
	if err != nil {
		t.Fatal(err)
	}
	shardsAgreeState.cases, shardsAgreeState.units = cases, units
	return cases, units, oracle, binary
}

func shardsAgreeWatchdog(name string) func() {
	timer := time.AfterFunc(shardsAgreeDeadline, func() { fmt.Fprintf(os.Stderr, "COOKED %s: killed at 75s\n", name); os.Exit(125) })
	return func() { timer.Stop() }
}

func shardsAgreeUnion(t *testing.T, merge func([][]byte, int) ([]byte, error)) {
	defer shardsAgreeWatchdog(t.Name())()
	cases, units, _, _ := shardsAgreeEnumeration(t)
	outputs := make([][]byte, len(units))
	var want []byte
	for index, c := range cases {
		block := []byte(fmt.Sprintf("case %d\n", index))
		owner := shardsAgreeOwner(c.Key)
		outputs[owner] = append(outputs[owner], block...)
		want = append(want, block...)
	}
	got, err := merge(outputs, len(cases))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("union differs from live unsplit enumeration")
	}
	if len(units) != testShardsAgreeShards {
		t.Fatal("enumerated shard count differs")
	}
	t.Logf("union: %d live case ids exactly once in %d shards", len(cases), len(units))
}

// Not parallel: initializes the shared generated fixtures and captured corpus state.
func TestShardsAgreeUnion(t *testing.T) { shardsAgreeUnion(t, shards.Merge) }

func shardsAgreeTopLevel(t *testing.T, index int) {
	defer shardsAgreeWatchdog(t.Name())()
	i, n, err := shardsAgreeSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	if index%n != i {
		t.Skip("ADAMIC_TEST_SHARD selects another unit")
	}
	_, units, oracle, binary := shardsAgreeEnumeration(t)
	shardsAgreeUnit(t, oracle, binary, units[index], "", "")
}

// The unit's earlier deadline still wins. Kill a whole child process group,
// including compiler descendants, rather than leaving them after cancellation.
func shardsAgreeChild(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	return cmd
}

func shardsAgreeCommand(ctx context.Context, binary string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	cmd := shardsAgreeChild(ctx, binary, args...)
	output, err := os.CreateTemp(sharedDirectory, "shardsagree-stdout-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(output.Name())
	defer output.Close()
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = output, &stderr
	err = cmd.Run()
	if err != nil || stderr.Len() != 0 {
		return nil, fmt.Errorf("%s %v: %v\n%s", binary, args, err, &stderr)
	}
	return os.ReadFile(output.Name())
}

func shardsAgreeNative(ctx context.Context, binary, path string, count int, countOnly bool, rows int) ([]byte, error) {
	outputs := make([][]byte, count)
	errors := make([]error, count)
	var group sync.WaitGroup
	for index := range outputs {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			args := []string{"--manifest", path, "--shard", fmt.Sprintf("%d/%d", index, count)}
			if countOnly {
				args = append(args, "--count")
			}
			outputs[index], errors[index] = shardsAgreeCommand(ctx, binary, args...)
		}(index)
	}
	group.Wait()
	total := 0
	for index, err := range errors {
		if err != nil {
			return nil, err
		}
		if countOnly {
			value, err := strconv.Atoi(strings.TrimSpace(string(outputs[index])))
			if err != nil {
				return nil, err
			}
			total += value
		}
	}
	if countOnly {
		return []byte(strconv.Itoa(total) + "\n"), nil
	}
	return shards.Merge(outputs, rows)
}

func shardsAgreeCompare(got, want, gotCount, wantCount []byte, count int) error {
	if diff := difference(got, want); diff != "" {
		return fmt.Errorf("%d process shards: %s", count, diff)
	}
	if !bytes.Equal(gotCount, wantCount) {
		return fmt.Errorf("%d shards count %q, want %q", count, gotCount, wantCount)
	}
	return nil
}

func shardsAgreeUnit(t *testing.T, oracle, binary string, cases []shardsAgreeCase, planted, kind string) {
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), shardsAgreeDeadline)
	defer cancel()
	defer func() {
		t.Logf("unit wall %.3fs; cooked=%t; cases=%d", time.Since(started).Seconds(), time.Since(started) > shardsAgreeBudget, len(cases))
	}()
	if len(cases) == 0 {
		return
	}
	rows := make([]string, len(cases))
	for index, c := range cases {
		rows[index] = c.Row
	}
	path := manifest(t, rows)
	flags, err := shardsAgreeCommand(ctx, oracle, "--manifest", path, "--diagnostics")
	if err != nil {
		shardsAgreeError(t, ctx, err)
		return
	}
	values := strings.Fields(string(flags))
	if len(values) != len(rows) {
		t.Fatal("diagnostics row count differs")
	}
	// Identical recovery classification to the unsplit test, inside each deadline.
	for index, row := range rows {
		if values[index] == "1" {
			fields := strings.Split(row, "\t")
			for len(fields) < 7 {
				fields = append(fields, "")
			}
			if fields[6] == "" {
				fields[6] = "recovery"
			}
			rows[index] = strings.Join(fields, "\t")
		}
	}
	path = manifest(t, rows)
	var want, wantCount []byte
	var fullErr, countErr error
	var expected sync.WaitGroup
	expected.Add(2)
	go func() { defer expected.Done(); want, fullErr = shardsAgreeCommand(ctx, binary, "--manifest", path) }()
	go func() {
		defer expected.Done()
		wantCount, countErr = shardsAgreeCommand(ctx, binary, "--manifest", path, "--count")
	}()
	expected.Wait()
	if fullErr != nil {
		shardsAgreeError(t, ctx, fullErr)
		return
	}
	if countErr != nil {
		shardsAgreeError(t, ctx, countErr)
		return
	}
	// Independent comparisons run concurrently on the unit's four CPUs. Every
	// unit still checks both outputs at 1, 2 and NumCPU native process shards.
	counts := []int{1, 2, runtime.NumCPU()}
	errors := make([]error, len(counts))
	var group sync.WaitGroup
	for index, count := range counts {
		group.Add(1)
		go func(index, count int) {
			defer group.Done()
			got, err := shardsAgreeNative(ctx, binary, path, count, false, len(rows))
			if err != nil {
				errors[index] = err
				return
			}
			gotCount, err := shardsAgreeNative(ctx, binary, path, count, true, len(rows))
			if err != nil {
				errors[index] = err
				return
			}
			if count == 1 {
				for local, c := range cases {
					if c.Key == planted {
						if kind == "count" {
							gotCount = []byte("-1\n")
						} else {
							header := []byte(fmt.Sprintf("case %d\n", local))
							got = bytes.Replace(got, header, append(append([]byte(nil), header...), []byte("planted disagreement\n")...), 1)
						}
					}
				}
			}
			errors[index] = shardsAgreeCompare(got, want, gotCount, wantCount, count)
		}(index, count)
	}
	group.Wait()
	for _, err := range errors {
		if err != nil {
			shardsAgreeError(t, ctx, err)
			return
		}
	}
	if time.Since(started) > shardsAgreeBudget {
		t.Errorf("COOKED %s: over 60s budget", t.Name())
	}
}
func shardsAgreeError(t *testing.T, ctx context.Context, err error) {
	t.Helper()
	if ctx.Err() != nil {
		t.Errorf("COOKED %s: killed at 75s", t.Name())
		return
	}
	t.Error(err)
}

func shardsAgreeProofCases(directory string) []shardsAgreeCase {
	var cases []shardsAgreeCase
	for index := 0; index < 8; index++ {
		name := fmt.Sprintf("case-%02d.ts", index)
		cases = append(cases, shardsAgreeCase{"proof/" + name + "\tall", filepath.Join(directory, name)})
	}
	return cases
}

// Not parallel: initializes shared scanner products and fixtures used by proof subprocesses.
func TestShardsAgreeDisagreement(t *testing.T) {
	defer shardsAgreeWatchdog(t.Name())()
	if kind := os.Getenv("ADAMIC_SHARDS_AGREE_PROBE"); kind != "" {
		cases := shardsAgreeProofCases(os.Getenv("ADAMIC_SHARDS_AGREE_PROBE_DIR"))
		units, err := shardsAgreePartition(cases)
		if err != nil {
			t.Fatal(err)
		}
		oracle, binary := shardsAgreeProducts(t)
		for index, unit := range units {
			t.Run(fmt.Sprintf("shard-%03d", index), func(t *testing.T) { t.Parallel(); shardsAgreeUnit(t, oracle, binary, unit, cases[0].Key, kind) })
		}
		return
	}
	directory := t.TempDir()
	cases := shardsAgreeProofCases(directory)
	for _, c := range cases {
		if err := os.WriteFile(c.Row, []byte("debugger; var a;\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	original, err := shardsAgreePartition(cases)
	if err != nil {
		t.Fatal(err)
	}
	grown := append([]shardsAgreeCase{{"proof/new.ts\tall", "new.ts"}}, cases...)
	expanded, err := shardsAgreePartition(grown)
	if err != nil {
		t.Fatal(err)
	}
	for index, unit := range original {
		for _, c := range unit {
			found := false
			for _, other := range expanded[index] {
				if c.Key == other.Key {
					found = true
				}
			}
			if !found {
				t.Fatalf("growth moved %s", c.Key)
			}
		}
	}
	duplicate := append(append([]shardsAgreeCase(nil), cases...), cases[0])
	if _, err := shardsAgreePartition(duplicate); err == nil {
		t.Fatal("duplicate union accepted")
	}
	for _, text := range []string{"0/0", "-1/2", "2/2", "a/2", "0/257"} {
		if _, _, err := shardsAgreeSelection(text); err == nil {
			t.Fatalf("accepted selector %q", text)
		}
	}
	selected := make([]int, testShardsAgreeShards)
	for box := 0; box < 3; box++ {
		i, n, err := shardsAgreeSelection(fmt.Sprintf("%d/3", box))
		if err != nil {
			t.Fatal(err)
		}
		for index := range selected {
			if index%n == i {
				selected[index]++
			}
		}
	}
	for index, count := range selected {
		if count != 1 {
			t.Fatalf("selector enumerated shard %d %d times", index, count)
		}
	}
	target := fmt.Sprintf("shard-%03d", shardsAgreeOwner(cases[0].Key))
	for _, kind := range []string{"output", "count"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), shardsAgreeDeadline)
			defer cancel()
			cmd := shardsAgreeChild(ctx, os.Args[0], "-test.run=^TestShardsAgreeDisagreement$", "-test.v", "-test.timeout=75s")
			for _, value := range os.Environ() {
				if !strings.HasPrefix(value, "ADAMIC_TEST_SHARD=") {
					cmd.Env = append(cmd.Env, value)
				}
			}
			cmd.Env = append(cmd.Env, "ADAMIC_SHARDS_AGREE_PROBE="+kind, "ADAMIC_SHARDS_AGREE_PROBE_DIR="+directory)
			output, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || ctx.Err() != nil {
				t.Fatalf("probe did not report disagreement: %v\n%s", err, output)
			}
			failures := regexp.MustCompile(`(?m)^\s*--- FAIL: TestShardsAgreeDisagreement/(shard-[0-9]{3}) \(`).FindAllStringSubmatch(string(output), -1)
			if len(failures) != 1 || failures[0][1] != target || bytes.Contains(output, []byte("COOKED")) {
				t.Fatalf("want exactly %s, got %v\n%s", target, failures, output)
			}
			t.Logf("planted %s disagreement in %s caught only by %s", kind, cases[0].Key, target)
		})
	}
}
