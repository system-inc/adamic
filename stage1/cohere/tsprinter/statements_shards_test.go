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
func statementLeaks(test *testing.T, sanitized, release string, arguments ...string) string {
	test.Helper()
	switch runtime.GOOS {
	case "linux":
		report := oracles.StatementExecute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.ExitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.ExitCode, report.Stderr)
	case "darwin":
		report := oracles.StatementExecute(
			test,
			nil,
			"leaks",
			append([]string{"--atExit", "--", release}, arguments...)...,
		)
		if report.ExitCode == 0 {
			return ""
		}
		return string(report.Stdout)
	}
	test.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

func statementProofKeys(count int) []string {
	keys := make([]string, count)
	for index := range keys {
		keys[index] = fmt.Sprintf("stage1/cohere/tsprinter/statements_shards_test.go:proof:%d", index)
	}
	return keys
}

func TestStatementsShardAssignmentStable(test *testing.T) {
	test.Parallel()
	root, upstream := "/repository", "/typescript"
	items := []oracles.PrinterCase{
		{Label: root + "/stage1/a.ts:0:SourceFile"},
		{Label: root + "/stage1/a.ts:3:VariableStatement"},
		{Label: root + "/stage1/a.ts:9:VariableStatement"},
		{Label: upstream + "/src/compiler/b.ts:0:SourceFile"},
		{Label: "program-sequence"}, {Label: "program-sequence"},
	}
	before, statementCaseKeysError := oracles.StatementCaseKeys(items, root, upstream)
	if statementCaseKeysError != nil {
		test.Fatal(statementCaseKeysError)
	}
	grown := append([]oracles.PrinterCase{{Label: root + "/bench/new.ts:0:SourceFile"}}, items...)
	grown = append(grown, oracles.PrinterCase{Label: root + "/stage1/last.ts:0:SourceFile"})
	after, statementCaseKeysError := oracles.StatementCaseKeys(grown, root, upstream)
	if statementCaseKeysError != nil {
		test.Fatal(statementCaseKeysError)
	}
	for index, key := range before {
		if key != after[index+1] {
			test.Fatalf("added file moved case key: %q -> %q", key, after[index+1])
		}
	}
	// Workshop builds the oracle's outputs with its checkout and TypeScript source; a runner reads that product with
	// both somewhere else, and Workshop's paths are gone (#tqrqx60). Every case keys as it did on Workshop.
	product := test.TempDir()
	data, encodeError := json.Marshal(items)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(filepath.Join(product, "cases.json"), data, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	if writeError := os.WriteFile(
		filepath.Join(product, "coverage.json"),
		[]byte(`{"file_refusals":{"`+root+`/stage1/bad.ts":"`+root+`/stage1/bad.ts: refused"}}`),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	if tsPrinterRewriteMetadataError := oracles.TSPrinterRewriteMetadata(
		product,
		root,
		upstream,
		true,
	); tsPrinterRewriteMetadataError != nil {
		test.Fatal(tsPrinterRewriteMetadataError)
	}
	for _, name := range []string{"cases.json", "coverage.json"} {
		if stored, _ := os.ReadFile(
			filepath.Join(product, name),
		); bytes.Contains(stored, []byte(root+"/")) || bytes.Contains(stored, []byte(upstream+"/")) {
			test.Fatalf("the product's %s names Workshop's paths: %s", name, stored)
		}
	}
	data, readError := os.ReadFile(filepath.Join(product, "cases.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var stored []oracles.PrinterCase
	if decodeError := json.Unmarshal(data, &stored); decodeError != nil {
		test.Fatal(decodeError)
	}
	runner, statementCaseKeysError := oracles.StatementCaseKeys(
		stored,
		"/srv/runner/adamic",
		"/opt/adamic-tools/typescript",
	)
	if statementCaseKeysError != nil {
		test.Fatalf("a runner reading Workshop's product: %v", statementCaseKeysError)
	}
	if strings.Join(runner, "\n") != strings.Join(before, "\n") {
		test.Fatalf("a runner keys Workshop's product %q, Workshop %q", runner, before)
	}
	// Exercise actual modulo assignment and its union, with enough cases for all shards.
	keys := statementProofKeys(256)
	old, statementPartitionError := oracles.StatementPartition(keys)
	if statementPartitionError != nil {
		test.Fatal(statementPartitionError)
	}
	added := append([]string{"stage1/new.ts:SourceFile:0"}, keys...)
	next, statementPartitionError := oracles.StatementPartition(added)
	if statementPartitionError != nil {
		test.Fatal(statementPartitionError)
	}
	owners := map[string]int{}
	for index, ids := range old {
		for _, id := range ids {
			owners[keys[id]] = index
		}
	}
	for index, ids := range next {
		for _, id := range ids {
			if owner, exists := owners[added[id]]; exists && owner != index {
				test.Fatalf("added file moved %s from shard-%03d to shard-%03d", added[id], owner, index)
			}
		}
	}
	if _, statementPartitionError := oracles.StatementPartition(append(keys, keys[0])); statementPartitionError == nil {
		test.Fatal("repeated stable key accepted")
	}
}

func statementShardSelection(test *testing.T) func(int) bool {
	test.Helper()
	selected, statementSelectionError := statementSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if statementSelectionError != nil {
		test.Fatal(statementSelectionError)
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
	index, e1 := strconv.Atoi(parts[0])
	n, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || index < 0 || n <= 0 || index >= n {
		return nil, fmt.Errorf("ADAMIC_TEST_SHARD must have 0 <= i < n, got %q", value)
	}
	return func(shard int) bool { return shard%n == index }, nil
}

func TestStatementsShardSelection(test *testing.T) {
	test.Parallel()
	for n := 1; n <= 2*oracles.StatementsAgainstGoAndPrettierShards; n++ {
		counts := make([]int, oracles.StatementsAgainstGoAndPrettierShards)
		for i := 0; i < n; i++ {
			selected, statementSelectionError := statementSelection(fmt.Sprintf("%d/%d", i, n))
			if statementSelectionError != nil {
				test.Fatal(statementSelectionError)
			}
			for shard := range counts {
				if selected(shard) {
					counts[shard]++
				}
			}
		}
		for shard, count := range counts {
			if count != 1 {
				test.Fatalf("%d boxes: shard-%03d selected %d times", n, shard, count)
			}
		}
	}
	selected, statementSelectionError := statementSelection("")
	if statementSelectionError != nil {
		test.Fatal(statementSelectionError)
	}
	for shard := 0; shard < oracles.StatementsAgainstGoAndPrettierShards; shard++ {
		if !selected(shard) {
			test.Fatalf("unset selection omitted shard-%03d", shard)
		}
	}
	for _, value := range []string{"bad", "0", "-1/2", "2/2", "0/0", "0/2/3"} {
		if _, statementSelectionError := statementSelection(value); statementSelectionError == nil {
			test.Fatalf("accepted invalid selector %q", value)
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
	for index := 0; index < max(len(gotLines), len(wantLines)); index++ {
		got, expected := "", ""
		if index < len(gotLines) {
			got = gotLines[index]
		}
		if index < len(wantLines) {
			expected = wantLines[index]
		}
		if got != expected {
			label := fmt.Sprintf("line %d", index+1)
			if index < len(labels) {
				label = labels[index]
			}
			return fmt.Errorf("%s %s: %s", name, label, firstDifference(got, expected))
		}
	}
	return fmt.Errorf("%s: %s", name, firstDifference(string(result.Stdout), want))
}

// A real Node run passes first. A child test process then plants one changed
// answer and takes the same fatal comparison path as the production leaves.
func TestStatementsShardDisagreement(test *testing.T) {
	test.Parallel()
	const count, plantedID = 256, 17
	partition, statementPartitionError := oracles.StatementPartition(statementProofKeys(count))
	if statementPartitionError != nil {
		test.Fatal(statementPartitionError)
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
		test.Fatal("planted case absent")
	}
	binary, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	result := oracles.StatementExecute(
		test,
		[]string{"ADAMIC_STATEMENTS_SHARD_PROOF=1", "ADAMIC_TEST_SHARD="},
		binary,
		"-test.run=^TestStatementsAgainstGoAndPrettier_[0-9]{3}$",
		"-test.v",
		"-test.parallel=4",
		"-test.timeout=75s",
	)
	name := fmt.Sprintf("TestStatementsAgainstGoAndPrettier_%03d", target)
	if result.ExitCode != 1 ||
		len(result.Stderr) != 0 ||
		strings.Count(string(result.Stdout), "--- FAIL: TestStatementsAgainstGoAndPrettier_") != 1 ||
		!strings.Contains(string(result.Stdout), "--- FAIL: "+name) ||
		!strings.Contains(string(result.Stdout), "planted case 17") {
		test.Fatalf(
			"expected exactly %s to catch planted disagreement: exit %d stdout %s stderr %s",
			name,
			result.ExitCode,
			result.Stdout,
			result.Stderr,
		)
	}
	test.Logf("planted case %d caught by exactly %s", plantedID, name)
	return
}

func statementRunShardProof(test *testing.T, number int) {
	test.Helper()
	const count, plantedID = 256, 17
	partition, statementPartitionError := oracles.StatementPartition(statementProofKeys(count))
	if statementPartitionError != nil {
		test.Fatal(statementPartitionError)
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
	path := filepath.Join(test.TempDir(), "proof.txt")
	if writeError := os.WriteFile(path, []byte(input.String()), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	result := statementOnNode(test, port, "--cases", path, "80")
	if statementDisagreementError := statementDisagreement(
		"Node",
		result,
		want.String(),
		labels,
	); statementDisagreementError != nil {
		test.Fatal(statementDisagreementError)
	}
	if local >= 0 {
		rows := strings.Split(string(result.Stdout), "\n")
		rows[local] = "ok\tplanted disagreement"
		result.Stdout = []byte(strings.Join(rows, "\n"))
	}
	if statementDisagreementError := statementDisagreement(
		"Node",
		result,
		want.String(),
		labels,
	); statementDisagreementError != nil {
		test.Fatal(statementDisagreementError)
	}
}

func TestStatementsShardUnionRejectsMissingAndRepeatedIDs(test *testing.T) {
	test.Parallel()
	keys := statementProofKeys(256)
	shards, statementPartitionError := oracles.StatementPartition(keys)
	if statementPartitionError != nil {
		test.Fatal(statementPartitionError)
	}
	for _, name := range []string{"missing", "repeated", "extra", "shard-count"} {
		test.Run(name, func(test *testing.T) {
			changed := make([][]int, len(shards))
			for index, ids := range shards {
				changed[index] = append([]int(nil), ids...)
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
			if statementUnionError := oracles.StatementUnion(len(keys), changed); statementUnionError == nil {
				test.Fatal("invalid union accepted")
			}
		})
	}
}

// Full-enumeration presence is checked in statementShards; leaves compare all their outcomes.
func statementPrinterLibrary(test *testing.T, name string, result oracles.Run, specs string, embedded bool) {
	test.Helper()
	if result.ExitCode != 0 || len(result.Stderr) != 0 {
		test.Fatalf("%s: exit %d stderr %s", name, result.ExitCode, result.Stderr)
	}
	data, readError := os.ReadFile(specs)
	if readError != nil {
		test.Fatal(readError)
	}
	var cases []oracles.PrinterCase
	if decodeError := json.Unmarshal(data, &cases); decodeError != nil {
		test.Fatal(decodeError)
	}
	lines := strings.Split(strings.TrimSuffix(string(result.Stdout), "\n"), "\n")
	if len(lines) != len(cases) {
		test.Fatalf("%s: %d answers, want %d", name, len(lines), len(cases))
	}
	records := oracles.UpstreamOutcomes(test)

	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	differences, refusals := 0, 0
	for index, item := range cases {
		status, want := "ok", item.Want
		for _, record := range records {
			if record.Family != "statements" ||
				!strings.HasSuffix(item.Label, record.Label) ||
				record.Source != item.Source {
				continue
			}
			if record.Go != item.Want {
				test.Fatalf("%s: recorded Go outcome changed", record.Label)
			}

			textError := record.PrettierError
			want = record.Prettier
			if embedded {
				want, textError = record.Embedded, record.EmbeddedError
			}
			if textError != "" {
				status, want = "error", textError
				refusals++
			} else {
				differences++
			}
			if want == "" {
				test.Fatalf("%s: missing %s outcome", record.Label, name)
			}
			break
		}
		expected := status + "\t" + escape.Replace(want)
		if lines[index] != expected {
			test.Errorf("%s %s: %s", name, item.Label, firstDifference(lines[index], expected))
		}
	}
	test.Logf(
		"%s: %d cases; %d exact upstream text differences, %d exact parser refusals, all other bytes match Go",
		name,
		len(cases),
		differences,
		refusals,
	)
}

func statementShardProofEnabled() bool { return os.Getenv("ADAMIC_STATEMENTS_SHARD_PROOF") == "1" }

func statementOnNode(test *testing.T, path string, arguments ...string) oracles.Run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(oracles.Repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return oracles.StatementExecute(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}
