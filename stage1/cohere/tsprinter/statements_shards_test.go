package tsprinter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

// Leak checks consume cached executable products, so a cache hit never needs
// an in-memory lowered program or a build inside a leaf.
func statementLeaks(t *testing.T, sanitized, release string, arguments ...string) string {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		report := oracles.StatementExecute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.ExitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.ExitCode, report.Stderr)
	case "darwin":
		report := oracles.StatementExecute(t, nil, "leaks", append([]string{"--atExit", "--", release}, arguments...)...)
		if report.ExitCode == 0 {
			return ""
		}
		return string(report.Stdout)
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
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
	items := []oracles.PrinterCase{
		{Label: root + "/stage1/a.ts:0:SourceFile"},
		{Label: root + "/stage1/a.ts:3:VariableStatement"},
		{Label: root + "/stage1/a.ts:9:VariableStatement"},
		{Label: upstream + "/src/compiler/b.ts:0:SourceFile"},
		{Label: "program-sequence"}, {Label: "program-sequence"},
	}
	before, err := oracles.StatementCaseKeys(items, root, upstream)
	if err != nil {
		t.Fatal(err)
	}
	grown := append([]oracles.PrinterCase{{Label: root + "/bench/new.ts:0:SourceFile"}}, items...)
	grown = append(grown, oracles.PrinterCase{Label: root + "/stage1/last.ts:0:SourceFile"})
	after, err := oracles.StatementCaseKeys(grown, root, upstream)
	if err != nil {
		t.Fatal(err)
	}
	for i, key := range before {
		if key != after[i+1] {
			t.Fatalf("added file moved case key: %q -> %q", key, after[i+1])
		}
	}
	// Workshop builds the oracle's outputs with its checkout and TypeScript source; a runner reads that product with
	// both somewhere else, and Workshop's paths are gone (#tqrqx60). Every case keys as it did on Workshop.
	product := t.TempDir()
	data, err := json.Marshal(items)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(product, "cases.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(product, "coverage.json"), []byte(`{"file_refusals":{"`+root+`/stage1/bad.ts":"`+root+`/stage1/bad.ts: refused"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err = oracles.TSPrinterRewriteMetadata(product, root, upstream, true); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cases.json", "coverage.json"} {
		if stored, _ := os.ReadFile(filepath.Join(product, name)); bytes.Contains(stored, []byte(root+"/")) || bytes.Contains(stored, []byte(upstream+"/")) {
			t.Fatalf("the product's %s names Workshop's paths: %s", name, stored)
		}
	}
	data, err = os.ReadFile(filepath.Join(product, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var stored []oracles.PrinterCase
	if err = json.Unmarshal(data, &stored); err != nil {
		t.Fatal(err)
	}
	runner, err := oracles.StatementCaseKeys(stored, "/srv/runner/adamic", "/opt/adamic-tools/typescript")
	if err != nil {
		t.Fatalf("a runner reading Workshop's product: %v", err)
	}
	if strings.Join(runner, "\n") != strings.Join(before, "\n") {
		t.Fatalf("a runner keys Workshop's product %q, Workshop %q", runner, before)
	}
	// Exercise actual modulo assignment and its union, with enough cases for all shards.
	keys := statementProofKeys(256)
	old, err := oracles.StatementPartition(keys)
	if err != nil {
		t.Fatal(err)
	}
	added := append([]string{"stage1/new.ts:SourceFile:0"}, keys...)
	next, err := oracles.StatementPartition(added)
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
	if _, err := oracles.StatementPartition(append(keys, keys[0])); err == nil {
		t.Fatal("repeated stable key accepted")
	}
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
	for n := 1; n <= 2*oracles.StatementsAgainstGoAndPrettierShards; n++ {
		counts := make([]int, oracles.StatementsAgainstGoAndPrettierShards)
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
	for shard := 0; shard < oracles.StatementsAgainstGoAndPrettierShards; shard++ {
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

func statementDisagreement(name string, result oracles.Run, want string, labels []string) error {
	if result.ExitCode != 0 || len(result.Stderr) != 0 {
		return fmt.Errorf("%s exit %d stderr %s", name, result.ExitCode, result.Stderr)
	}
	if string(result.Stdout) == want {
		return nil
	}
	gotLines, wantLines := strings.Split(string(result.Stdout), "\n"), strings.Split(want, "\n")
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
	return fmt.Errorf("%s: %s", name, firstDifference(string(result.Stdout), want))
}

// A real Node run passes first. A child test process then plants one changed
// answer and takes the same fatal comparison path as the production leaves.
func TestStatementsShardDisagreement(t *testing.T) {
	t.Parallel()
	const count, plantedID = 256, 17
	partition, err := oracles.StatementPartition(statementProofKeys(count))
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
	result := oracles.StatementExecute(t, []string{"ADAMIC_STATEMENTS_SHARD_PROOF=1", "ADAMIC_TEST_SHARD="}, binary, "-test.run=^TestStatementsAgainstGoAndPrettier_[0-9]{3}$", "-test.v", "-test.parallel=4", "-test.timeout=75s")
	name := fmt.Sprintf("TestStatementsAgainstGoAndPrettier_%03d", target)
	if result.ExitCode != 1 || len(result.Stderr) != 0 || strings.Count(string(result.Stdout), "--- FAIL: TestStatementsAgainstGoAndPrettier_") != 1 || !strings.Contains(string(result.Stdout), "--- FAIL: "+name) || !strings.Contains(string(result.Stdout), "planted case 17") {
		t.Fatalf("expected exactly %s to catch planted disagreement: exit %d stdout %s stderr %s", name, result.ExitCode, result.Stdout, result.Stderr)
	}
	t.Logf("planted case %d caught by exactly %s", plantedID, name)
	return
}

func statementRunShardProof(t *testing.T, number int) {
	t.Helper()
	const count, plantedID = 256, 17
	partition, err := oracles.StatementPartition(statementProofKeys(count))
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
		rows := strings.Split(string(result.Stdout), "\n")
		rows[local] = "ok\tplanted disagreement"
		result.Stdout = []byte(strings.Join(rows, "\n"))
	}
	if err := statementDisagreement("Node", result, want.String(), labels); err != nil {
		t.Fatal(err)
	}
}

func TestStatementsShardUnionRejectsMissingAndRepeatedIDs(t *testing.T) {
	t.Parallel()
	keys := statementProofKeys(256)
	shards, err := oracles.StatementPartition(keys)
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
			if err := oracles.StatementUnion(len(keys), changed); err == nil {
				t.Fatal("invalid union accepted")
			}
		})
	}
}

// Full-enumeration presence is checked in statementShards; leaves compare all their outcomes.
func statementPrinterLibrary(t *testing.T, name string, result oracles.Run, specs string, embedded bool) {
	t.Helper()
	if result.ExitCode != 0 || len(result.Stderr) != 0 {
		t.Fatalf("%s: exit %d stderr %s", name, result.ExitCode, result.Stderr)
	}
	data, err := os.ReadFile(specs)
	if err != nil {
		t.Fatal(err)
	}
	var cases []oracles.PrinterCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(result.Stdout), "\n"), "\n")
	if len(lines) != len(cases) {
		t.Fatalf("%s: %d answers, want %d", name, len(lines), len(cases))
	}
	records := oracles.UpstreamOutcomes(t)

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

func statementOnNode(t *testing.T, path string, arguments ...string) oracles.Run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(oracles.Repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return oracles.StatementExecute(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}
