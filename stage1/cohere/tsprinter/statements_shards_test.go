package tsprinter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Until internal/buildcache lands, products are built once per caller and
// shared by all its leaves. This helper deliberately has no package cache.
type statementInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func statementProduct(t *testing.T, inputs statementInputs, build func(dir string) error) string {
	t.Helper()
	dir := t.TempDir()
	start := time.Now()
	cpu := statementCPU()
	if err := build(dir); err != nil {
		t.Fatalf("build %s: %v", inputs.Name, err)
	}
	t.Logf("build %s: %.3fs, %.3f CPU s (toolchain %s)", inputs.Name, time.Since(start).Seconds(), statementCPU()-cpu, inputs.Toolchain)
	return dir
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

func statementPartition(weights []int) ([][]int, error) {
	shards := assignCorpus(weights, testStatementsAgainstGoAndPrettierShards)
	if err := statementUnion(len(weights), shards); err != nil {
		return nil, err
	}
	return shards, nil
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
	weights := make([]int, len(rows))
	for i, row := range rows {
		weights[i] = len(row)
	}
	partition, err := statementPartition(weights)
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
	const count, plantedID = 33, 17
	weights := make([]int, count)
	for id := range weights {
		weights[id] = len(fmt.Sprintf(">const x=%d;", id))
	}
	partition, err := statementPartition(weights)
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
	if os.Getenv("ADAMIC_STATEMENTS_SHARD_PROOF") != "1" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		result := executeOne(t, []string{"ADAMIC_STATEMENTS_SHARD_PROOF=1", "ADAMIC_TEST_SHARD="}, binary, "-test.run=^TestStatementsShardDisagreement$", "-test.v", "-test.parallel=4", "-test.timeout=30s")
		name := fmt.Sprintf("TestStatementsShardDisagreement/shard-%03d", target)
		if result.exitCode != 1 || len(result.stderr) != 0 || strings.Count(string(result.stdout), "--- FAIL: TestStatementsShardDisagreement/shard-") != 1 || !strings.Contains(string(result.stdout), "--- FAIL: "+name) || !strings.Contains(string(result.stdout), "planted case 17") {
			t.Fatalf("expected exactly %s to catch planted disagreement: exit %d stdout %s stderr %s", name, result.exitCode, result.stdout, result.stderr)
		}
		t.Logf("planted case %d caught by exactly %s", plantedID, name)
		return
	}
	port, _ := filepath.Abs("statementsMain.ts")
	for number, ids := range partition {
		t.Run(fmt.Sprintf("shard-%03d", number), func(t *testing.T) {
			t.Parallel()
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
			result := onNode(t, port, "--cases", path, "80")
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
		})
	}
}

func TestStatementsShardUnionRejectsMissingAndRepeatedIDs(t *testing.T) {
	weights := make([]int, 33)
	for i := range weights {
		weights[i] = 1
	}
	shards, err := statementPartition(weights)
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
				changed[0] = append(changed[0], 33)
			case "shard-count":
				changed = changed[1:]
			}
			if err := statementUnion(len(weights), changed); err == nil {
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
