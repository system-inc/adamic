package tsprinter

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func statementProduct(t *testing.T, inputs buildcache.Inputs, build func(dir string) error) string {
	t.Helper()
	return buildcache.Product(t, inputs, func(dir string) error {
		start, cpu := time.Now(), statementCPU()
		if err := build(dir); err != nil {
			return err
		}
		t.Logf("build %s: %.3fs, %.3f CPU s", inputs.Name, time.Since(start).Seconds(), statementCPU()-cpu)
		return nil
	})
}

func statementBytesHash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

func statementFileHash(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return statementBytesHash(data)
}

// External npm inputs cannot appear in repository-relative Inputs.Files.
// WalkDir visits names lexically; hash every name and byte, refusing symlinks.
func statementDirectoryHash(t *testing.T, directory string) string {
	t.Helper()
	hash := sha256.New()
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("oracle input symlink: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		name, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%d:%s:%d:", len(name), name, len(data))
		hash.Write(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Leak checks consume cached executable products, so a cache hit never needs
// an in-memory lowered program or a build inside a leaf.
func statementLeaks(t *testing.T, sanitized, release string, arguments ...string) string {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		report := statementExecute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	case "darwin":
		report := statementExecute(t, nil, "leaks", append([]string{"--atExit", "--", release}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

// Linux accounting includes all reaped descendants (oracle, Node and clang).
func statementCPU() float64 {
	var self, children syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &self); err != nil {
		panic(err)
	}
	if err := syscall.Getrusage(syscall.RUSAGE_CHILDREN, &children); err != nil {
		panic(err)
	}
	return float64(self.Utime.Sec+self.Stime.Sec+children.Utime.Sec+children.Stime.Sec) + float64(self.Utime.Usec+self.Stime.Usec+children.Utime.Usec+children.Stime.Usec)/1e6
}

// Indices, rather than labels (which repeat in synthetic cases), are the case IDs.
type statementShard struct {
	indices           []int
	labels            []string
	text, specs, want string
}

// Keys use a file's relative path, node mode and ordinal within that file/mode.
// Repository growth cannot change another file's assignment. Generated cases
// use their generator file and mode, with an ordinal local to that mode.
func statementCaseKeys(items []printerCase, root, upstream string) ([]string, error) {
	keys := make([]string, len(items))
	counts := map[string]int{}
	for i, item := range items {
		base := "stage1/cohere/tsprinter/testdata/statements_side_test.go:" + item.Label
		end := strings.LastIndex(item.Label, ":")
		if end >= 0 {
			start := strings.LastIndex(item.Label[:end], ":")
			if start < 0 {
				return nil, fmt.Errorf("invalid statement label %q", item.Label)
			}
			path, err := filepath.Abs(item.Label[:start])
			if err != nil {
				return nil, err
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return nil, err
			}
			if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				relative, err = filepath.Rel(upstream, path)
				if err != nil {
					return nil, err
				}
				if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
					return nil, fmt.Errorf("statement outside corpus roots: %s", path)
				}
				relative = "typescript@050880ce59e30b356b686bd3144efe24f875ebc8/" + filepath.ToSlash(relative)
			}
			base = filepath.ToSlash(relative) + ":" + item.Label[end+1:]
		}
		keys[i] = fmt.Sprintf("%s:%d", base, counts[base])
		counts[base]++
	}
	return keys, nil
}

func statementPartition(keys []string) ([][]int, error) {
	shards := make([][]int, testStatementsAgainstGoAndPrettierShards)
	seen := map[string]bool{}
	for id, key := range keys {
		if seen[key] {
			return nil, fmt.Errorf("repeated stable case key %q", key)
		}
		seen[key] = true
		digest := sha256.Sum256([]byte(key))
		var value uint64
		for _, b := range digest[:8] {
			value = value<<8 | uint64(b)
		}
		number := int(value % uint64(testStatementsAgainstGoAndPrettierShards))
		shards[number] = append(shards[number], id)
	}
	if err := statementUnion(len(keys), shards); err != nil {
		return nil, err
	}
	return shards, nil
}

func statementProofKeys(count int) []string {
	keys := make([]string, count)
	for i := range keys {
		keys[i] = fmt.Sprintf("stage1/cohere/tsprinter/statements_shards_test.go:proof:%d", i)
	}
	return keys
}

func TestStatementsShardAssignmentStable(t *testing.T) {
	t.Parallel()
	root, upstream := "/repository", "/typescript"
	items := []printerCase{
		{Label: root + "/stage1/a.ts:0:SourceFile"},
		{Label: root + "/stage1/a.ts:3:VariableStatement"},
		{Label: root + "/stage1/a.ts:9:VariableStatement"},
		{Label: upstream + "/src/compiler/b.ts:0:SourceFile"},
		{Label: "program-sequence"}, {Label: "program-sequence"},
	}
	before, err := statementCaseKeys(items, root, upstream)
	if err != nil {
		t.Fatal(err)
	}
	grown := append([]printerCase{{Label: root + "/bench/new.ts:0:SourceFile"}}, items...)
	grown = append(grown, printerCase{Label: root + "/stage1/last.ts:0:SourceFile"})
	after, err := statementCaseKeys(grown, root, upstream)
	if err != nil {
		t.Fatal(err)
	}
	for i, key := range before {
		if key != after[i+1] {
			t.Fatalf("added file moved case key: %q -> %q", key, after[i+1])
		}
	}
	// Exercise actual modulo assignment and its union, with enough cases for all shards.
	keys := statementProofKeys(256)
	old, err := statementPartition(keys)
	if err != nil {
		t.Fatal(err)
	}
	added := append([]string{"stage1/new.ts:SourceFile:0"}, keys...)
	next, err := statementPartition(added)
	if err != nil {
		t.Fatal(err)
	}
	owners := map[string]int{}
	for n, ids := range old {
		for _, id := range ids {
			owners[keys[id]] = n
		}
	}
	for n, ids := range next {
		for _, id := range ids {
			if owner, exists := owners[added[id]]; exists && owner != n {
				t.Fatalf("added file moved %s from shard-%03d to shard-%03d", added[id], owner, n)
			}
		}
	}
	if _, err := statementPartition(append(keys, keys[0])); err == nil {
		t.Fatal("repeated stable key accepted")
	}
}

func statementUnion(count int, shards [][]int) error {
	if len(shards) != testStatementsAgainstGoAndPrettierShards {
		return fmt.Errorf("enumerated %d shards, want %d", len(shards), testStatementsAgainstGoAndPrettierShards)
	}
	seen := make([]bool, count)
	total := 0
	for number, ids := range shards {
		if len(ids) == 0 {
			return fmt.Errorf("shard-%03d is empty", number)
		}
		for _, id := range ids {
			if id < 0 || id >= count {
				return fmt.Errorf("shard-%03d: extra case ID %d", number, id)
			}
			if seen[id] {
				return fmt.Errorf("shard-%03d: repeated case ID %d", number, id)
			}
			seen[id] = true
			total++
		}
	}
	for id, found := range seen {
		if !found {
			return fmt.Errorf("missing case ID %d", id)
		}
	}
	if total != count {
		return fmt.Errorf("union count %d, want %d", total, count)
	}
	return nil
}

func statementShards(t *testing.T, cases, want, specs string) []statementShard {
	t.Helper()
	text, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(specs)
	if err != nil {
		t.Fatal(err)
	}
	var items []printerCase
	if err := json.Unmarshal(data, &items); err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimSuffix(string(text), "\n"), "\n")
	answers := strings.Split(strings.TrimSuffix(want, "\n"), "\n")
	if len(rows) != len(items) || len(answers) != len(items) {
		t.Fatalf("enumeration: %d inputs, %d specifications, %d answers", len(rows), len(items), len(answers))
	}
	// The shared comparator's full-corpus upstream-presence check belongs here;
	// each leaf then checks the exact pinned outcomes for its own cases.
	for _, record := range upstreamOutcomes(t) {
		if record.Family != "statements" {
			continue
		}
		found := false
		for _, item := range items {
			if strings.HasSuffix(item.Label, record.Label) && item.Source == record.Source {
				if item.Want != record.Go {
					t.Fatalf("%s: recorded Go outcome changed", record.Label)
				}
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("recorded upstream case absent: %s", record.Label)
		}
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	upstream, err := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
	if err != nil {
		t.Fatal(err)
	}
	keys, err := statementCaseKeys(items, root, upstream)
	if err != nil {
		t.Fatal(err)
	}
	partition, err := statementPartition(keys)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	result := make([]statementShard, len(partition))
	projected := make([][]int, len(partition))
	for number, ids := range partition {
		shard := statementShard{indices: ids, text: filepath.Join(dir, fmt.Sprintf("statement-shard-%03d.txt", number)), specs: filepath.Join(dir, fmt.Sprintf("statement-shard-%03d.json", number))}
		var input, expected strings.Builder
		subset := make([]printerCase, 0, len(ids))
		for _, id := range ids {
			input.WriteString(rows[id] + "\n")
			expected.WriteString(answers[id] + "\n")
			subset = append(subset, items[id])
			shard.labels = append(shard.labels, fmt.Sprintf("case %d %s", id, items[id].Label))
			projected[number] = append(projected[number], id)
		}
		shard.want = expected.String()
		encoded, err := json.Marshal(subset)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(shard.text, []byte(input.String()), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(shard.specs, encoded, 0644); err != nil {
			t.Fatal(err)
		}
		result[number] = shard
	}
	if err := statementUnion(len(items), projected); err != nil {
		t.Fatal(err)
	}
	return result
}

func statementShardSelection(t *testing.T) func(int) bool {
	t.Helper()
	selected, err := statementSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	return selected
}

func statementSelection(value string) (func(int) bool, error) {
	if value == "" {
		return func(int) bool { return true }, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("ADAMIC_TEST_SHARD must be i/n, got %q", value)
	}
	i, e1 := strconv.Atoi(parts[0])
	n, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || i < 0 || n <= 0 || i >= n {
		return nil, fmt.Errorf("ADAMIC_TEST_SHARD must have 0 <= i < n, got %q", value)
	}
	return func(shard int) bool { return shard%n == i }, nil
}

func TestStatementsShardSelection(t *testing.T) {
	t.Parallel()
	for n := 1; n <= 2*testStatementsAgainstGoAndPrettierShards; n++ {
		counts := make([]int, testStatementsAgainstGoAndPrettierShards)
		for i := 0; i < n; i++ {
			selected, err := statementSelection(fmt.Sprintf("%d/%d", i, n))
			if err != nil {
				t.Fatal(err)
			}
			for shard := range counts {
				if selected(shard) {
					counts[shard]++
				}
			}
		}
		for shard, count := range counts {
			if count != 1 {
				t.Fatalf("%d boxes: shard-%03d selected %d times", n, shard, count)
			}
		}
	}
	selected, err := statementSelection("")
	if err != nil {
		t.Fatal(err)
	}
	for shard := 0; shard < testStatementsAgainstGoAndPrettierShards; shard++ {
		if !selected(shard) {
			t.Fatalf("unset selection omitted shard-%03d", shard)
		}
	}
	for _, value := range []string{"bad", "0", "-1/2", "2/2", "0/0", "0/2/3"} {
		if _, err := statementSelection(value); err == nil {
			t.Fatalf("accepted invalid selector %q", value)
		}
	}
}

func statementDisagreement(name string, result run, want string, labels []string) error {
	if result.exitCode != 0 || len(result.stderr) != 0 {
		return fmt.Errorf("%s exit %d stderr %s", name, result.exitCode, result.stderr)
	}
	if string(result.stdout) == want {
		return nil
	}
	gotLines, wantLines := strings.Split(string(result.stdout), "\n"), strings.Split(want, "\n")
	for i := 0; i < max(len(gotLines), len(wantLines)); i++ {
		got, expected := "", ""
		if i < len(gotLines) {
			got = gotLines[i]
		}
		if i < len(wantLines) {
			expected = wantLines[i]
		}
		if got != expected {
			label := fmt.Sprintf("line %d", i+1)
			if i < len(labels) {
				label = labels[i]
			}
			return fmt.Errorf("%s %s: %s", name, label, firstDifference(got, expected))
		}
	}
	return fmt.Errorf("%s: %s", name, firstDifference(string(result.stdout), want))
}

// A real Node run passes first. A child test process then plants one changed
// answer and takes the same fatal comparison path as the production leaves.
func TestStatementsShardDisagreement(t *testing.T) {
	t.Parallel()
	const count, plantedID = 256, 17
	partition, err := statementPartition(statementProofKeys(count))
	if err != nil {
		t.Fatal(err)
	}
	target := -1
	for number, ids := range partition {
		for _, id := range ids {
			if id == plantedID {
				target = number
			}
		}
	}
	if target < 0 {
		t.Fatal("planted case absent")
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	result := statementExecute(t, []string{"ADAMIC_STATEMENTS_SHARD_PROOF=1", "ADAMIC_TEST_SHARD="}, binary, "-test.run=^TestStatementsAgainstGoAndPrettier_[0-9]{3}$", "-test.v", "-test.parallel=4", "-test.timeout=75s")
	name := fmt.Sprintf("TestStatementsAgainstGoAndPrettier_%03d", target)
	if result.exitCode != 1 || len(result.stderr) != 0 || strings.Count(string(result.stdout), "--- FAIL: TestStatementsAgainstGoAndPrettier_") != 1 || !strings.Contains(string(result.stdout), "--- FAIL: "+name) || !strings.Contains(string(result.stdout), "planted case 17") {
		t.Fatalf("expected exactly %s to catch planted disagreement: exit %d stdout %s stderr %s", name, result.exitCode, result.stdout, result.stderr)
	}
	t.Logf("planted case %d caught by exactly %s", plantedID, name)
	return
}

func statementRunShardProof(t *testing.T, number int) {
	t.Helper()
	const count, plantedID = 256, 17
	partition, err := statementPartition(statementProofKeys(count))
	if err != nil {
		t.Fatal(err)
	}
	ids := partition[number]

	port, _ := filepath.Abs("statementsMain.ts")
	var input, want strings.Builder
	var labels []string
	local := -1
	for position, id := range ids {
		fmt.Fprintf(&input, ">const x=%d;\n", id)
		fmt.Fprintf(&want, "ok\tconst x = %d;\\n\n", id)
		labels = append(labels, fmt.Sprintf("planted case %d", id))
		if id == plantedID {
			local = position
		}
	}
	path := filepath.Join(t.TempDir(), "proof.txt")
	if err := os.WriteFile(path, []byte(input.String()), 0644); err != nil {
		t.Fatal(err)
	}
	result := statementOnNode(t, port, "--cases", path, "80")
	if err := statementDisagreement("Node", result, want.String(), labels); err != nil {
		t.Fatal(err)
	}
	if local >= 0 {
		rows := strings.Split(string(result.stdout), "\n")
		rows[local] = "ok\tplanted disagreement"
		result.stdout = []byte(strings.Join(rows, "\n"))
	}
	if err := statementDisagreement("Node", result, want.String(), labels); err != nil {
		t.Fatal(err)
	}
}

func TestStatementsShardUnionRejectsMissingAndRepeatedIDs(t *testing.T) {
	t.Parallel()
	keys := statementProofKeys(256)
	shards, err := statementPartition(keys)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"missing", "repeated", "extra", "shard-count"} {
		t.Run(name, func(t *testing.T) {
			changed := make([][]int, len(shards))
			for i, ids := range shards {
				changed[i] = append([]int(nil), ids...)
			}
			switch name {
			case "missing":
				changed[0] = changed[0][1:]
			case "repeated":
				changed[0] = append(changed[0], changed[1][0])
			case "extra":
				changed[0] = append(changed[0], len(keys))
			case "shard-count":
				changed = changed[1:]
			}
			if err := statementUnion(len(keys), changed); err == nil {
				t.Fatal("invalid union accepted")
			}
		})
	}
}

// Full-enumeration presence is checked in statementShards; leaves compare all their outcomes.
func statementPrinterLibrary(t *testing.T, name string, result run, specs string, embedded bool) {
	t.Helper()
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("%s: exit %d stderr %s", name, result.exitCode, result.stderr)
	}
	data, err := os.ReadFile(specs)
	if err != nil {
		t.Fatal(err)
	}
	var cases []printerCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(result.stdout), "\n"), "\n")
	if len(lines) != len(cases) {
		t.Fatalf("%s: %d answers, want %d", name, len(lines), len(cases))
	}
	records := upstreamOutcomes(t)

	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	differences, refusals := 0, 0
	for index, item := range cases {
		status, want := "ok", item.Want
		for _, record := range records {
			if record.Family != "statements" || !strings.HasSuffix(item.Label, record.Label) || record.Source != item.Source {
				continue
			}
			if record.Go != item.Want {
				t.Fatalf("%s: recorded Go outcome changed", record.Label)
			}

			errText := record.PrettierError
			want = record.Prettier
			if embedded {
				want, errText = record.Embedded, record.EmbeddedError
			}
			if errText != "" {
				status, want = "error", errText
				refusals++
			} else {
				differences++
			}
			if want == "" {
				t.Fatalf("%s: missing %s outcome", record.Label, name)
			}
			break
		}
		expected := status + "\t" + escape.Replace(want)
		if lines[index] != expected {
			t.Errorf("%s %s: %s", name, item.Label, firstDifference(lines[index], expected))
		}
	}
	t.Logf("%s: %d cases; %d exact upstream text differences, %d exact parser refusals, all other bytes match Go", name, len(cases), differences, refusals)
}

func statementShardProofEnabled() bool { return os.Getenv("ADAMIC_STATEMENTS_SHARD_PROOF") == "1" }

// Command deadlines use only Go APIs and kill the entire child process group.
func statementCommand(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	if value, ok := statementCaseContexts.Load(t); ok {
		return statementContextCommand(value.(context.Context), name, arguments...)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	return statementContextCommand(ctx, name, arguments...)
}

func statementContextCommand(ctx context.Context, name string, arguments ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Setup's nested commands join its group, so its outer deadline also kills
	// Go compilers and their descendants if the whole setup process is cooked.
	if os.Getenv("ADAMIC_STATEMENTS_SETUP_CHILD") == "1" {
		command.SysProcAttr.Pgid = syscall.Getpgrp()
	}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		group := command.Process.Pid
		if command.SysProcAttr.Pgid != 0 {
			group = command.SysProcAttr.Pgid
		}
		err := syscall.Kill(-group, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 5 * time.Second
	return command
}

func statementExecute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := statementCommand(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

func statementOnNode(t *testing.T, path string, arguments ...string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return statementExecute(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}
