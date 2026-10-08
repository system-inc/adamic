package tsprinter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
)

type corpusShard struct {
	indices     []int
	text, specs string
}
type corpusPlan struct {
	labels []string
	shards []corpusShard
}

var corpusPlans sync.Map

// Source bytes approximate parser and printer work. Ties use the original case
// number and then the shard number, making the assignment reproducible.
func assignCorpus(weights []int, count int) [][]int {
	count = min(count, len(weights))
	if count == 0 {
		return nil
	}
	order := make([]int, len(weights))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool {
		if weights[order[i]] != weights[order[j]] {
			return weights[order[i]] > weights[order[j]]
		}
		return order[i] < order[j]
	})
	result, totals := make([][]int, count), make([]int, count)
	for _, index := range order {
		shard := 0
		for i := 1; i < count; i++ {
			if totals[i] < totals[shard] {
				shard = i
			}
		}
		result[shard] = append(result[shard], index)
		totals[shard] += max(weights[index], 1)
	}
	return result
}

func planCorpus(t *testing.T, path string) *corpusPlan {
	t.Helper()
	if cached, ok := corpusPlans.Load(path); ok {
		return cached.(*corpusPlan)
	}
	base := strings.TrimSuffix(path, filepath.Ext(path))
	textPath, specsPath := base+".txt", base+".json"
	text, err := os.ReadFile(textPath)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(specsPath)
	if err != nil {
		t.Fatal(err)
	}
	var specs []json.RawMessage
	if err := json.Unmarshal(data, &specs); err != nil {
		t.Fatal(err)
	}
	var rows []string
	if filepath.Base(base) == "docs" {
		for _, part := range strings.Split(string(text), "reset\t")[1:] {
			rows = append(rows, "reset\t"+part)
		}
	} else {
		for _, row := range strings.Split(strings.TrimSuffix(string(text), "\n"), "\n") {
			rows = append(rows, row+"\n")
		}
	}
	if len(rows) != len(specs) || len(rows) == 0 {
		t.Fatalf("%s: %d protocol cases, %d specifications", path, len(rows), len(specs))
	}
	plan := &corpusPlan{labels: make([]string, len(rows))}
	weights := make([]int, len(rows))
	for i, raw := range specs {
		var item printerCase
		if err := json.Unmarshal(raw, &item); err != nil {
			t.Fatal(err)
		}
		plan.labels[i] = item.Label
		if item.Label == "" {
			plan.labels[i] = fmt.Sprintf("%s:case %d", specsPath, i)
		}
		weights[i] = len(rows[i])
	}
	directory := t.TempDir()
	for number, indices := range assignCorpus(weights, min(runtime.NumCPU(), 8)) {
		var batch strings.Builder
		subset := make([]json.RawMessage, 0, len(indices))
		for _, index := range indices {
			batch.WriteString(rows[index])
			subset = append(subset, specs[index])
		}
		shard := corpusShard{indices: indices, text: filepath.Join(directory, fmt.Sprintf("shard-%d.txt", number)), specs: filepath.Join(directory, fmt.Sprintf("shard-%d.json", number))}
		if err := os.WriteFile(shard.text, []byte(batch.String()), 0644); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(subset)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(shard.specs, encoded, 0644); err != nil {
			t.Fatal(err)
		}
		plan.shards = append(plan.shards, shard)
	}
	corpusPlans.Store(path, plan)
	// Either extension resolves to the same immutable plan, retained by the
	// corpus-owning test until all its parallel mutant subtests finish.
	corpusPlans.Store(textPath, plan)
	corpusPlans.Store(specsPath, plan)
	t.Cleanup(func() { corpusPlans.Delete(textPath); corpusPlans.Delete(specsPath) })
	t.Logf("%s: %d cases in %d largest-first processes", filepath.Base(base), len(rows), len(plan.shards))
	return plan
}

func mergeCorpus(plan *corpusPlan, outputs [][]byte) ([]byte, error) {
	if len(outputs) != len(plan.shards) {
		return nil, fmt.Errorf("corpus: %d shards, want %d", len(outputs), len(plan.shards))
	}
	merged := make([][]byte, len(plan.labels))
	for number, shard := range plan.shards {
		output := outputs[number]
		lines := bytes.Split(bytes.TrimSuffix(output, []byte("\n")), []byte("\n"))
		if len(output) == 0 {
			lines = nil
		}
		if len(lines) != len(shard.indices) {
			index := shard.indices[min(len(lines), len(shard.indices)-1)]
			return nil, fmt.Errorf("shard %d %s: missing or extra case: %d answers, want %d", number, plan.labels[index], len(lines), len(shard.indices))
		}
		if !bytes.HasSuffix(output, []byte("\n")) {
			return nil, fmt.Errorf("shard %d %s: unterminated answer", number, plan.labels[shard.indices[len(shard.indices)-1]])
		}
		for i, index := range shard.indices {
			if index < 0 || index >= len(merged) {
				return nil, fmt.Errorf("shard %d: extra case %d", number, index)
			}
			if merged[index] != nil {
				return nil, fmt.Errorf("shard %d %s: repeated case", number, plan.labels[index])
			}
			merged[index] = lines[i]
		}
	}
	var result bytes.Buffer
	for index, line := range merged {
		if line == nil {
			return nil, fmt.Errorf("%s: missing case", plan.labels[index])
		}
		result.Write(line)
		result.WriteByte('\n')
	}
	return result.Bytes(), nil
}

func executeShards(t *testing.T, environment []string, name string, arguments []string) (run, bool) {
	t.Helper()
	position := -1
	for i, argument := range arguments {
		switch filepath.Base(argument) {
		case "cases.txt", "cases.json", "tsc-cases.txt", "docs.txt", "docs.json":
			position = i
		}
	}
	if position < 0 {
		return run{}, false
	}
	// Tiny pinned upstream probes have JSON only and stay a single process.
	path := arguments[position]
	if _, err := os.Stat(strings.TrimSuffix(path, filepath.Ext(path)) + ".txt"); os.IsNotExist(err) {
		return run{}, false
	}
	plan := planCorpus(t, path)
	outputs := make([][]byte, len(plan.shards))
	results := make([]run, len(plan.shards))
	var group sync.WaitGroup
	for number, shard := range plan.shards {
		group.Add(1)
		go func(number int, shard corpusShard) {
			defer group.Done()
			args := append([]string(nil), arguments...)
			args[position] = shard.text
			if filepath.Ext(path) == ".json" {
				args[position] = shard.specs
			}
			results[number] = executeOne(t, environment, name, args...)
			outputs[number] = results[number].stdout
		}(number, shard)
	}
	group.Wait()
	// macOS leaks adds its own report to stdout. Each invocation must pass;
	// the ordinary native pass independently checks every printed answer.
	if name == "leaks" {
		for number, result := range results {
			if result.exitCode != 0 {
				t.Fatalf("leaks shard %d (%s): exit %d stdout %s stderr %s", number, plan.labels[plan.shards[number].indices[0]], result.exitCode, result.stdout, result.stderr)
			}
		}
		return run{}, true
	}
	for number, result := range results {
		if result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("%s shard %d (%s): exit %d stderr %s", name, number, plan.labels[plan.shards[number].indices[0]], result.exitCode, result.stderr)
		}
	}
	stdout, err := mergeCorpus(plan, outputs)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return run{stdout: stdout}, true
}

func corpusDifference(t *testing.T, cases, got, want string) string {
	t.Helper()
	if got == want {
		return ""
	}
	plan := planCorpus(t, cases)
	actual, expected := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < max(len(actual), len(expected)); i++ {
		a, b := "", ""
		if i < len(actual) {
			a = actual[i]
		}
		if i < len(expected) {
			b = expected[i]
		}
		if a != b {
			label := cases
			if i < len(plan.labels) {
				label = plan.labels[i]
			}
			return fmt.Sprintf("%s: %s", label, firstDifference(a, b))
		}
	}
	return firstDifference(got, want)
}

func TestCorpusShardAssignment(t *testing.T) {
	t.Parallel()
	got := assignCorpus([]int{1, 10, 2, 10, 3, 4}, 2)
	want := [][]int{{1, 5, 0}, {3, 4, 2}}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("largest-first assignment %v, want %v", got, want)
	}
}

func TestCorpusShardCoverage(t *testing.T) {
	t.Parallel()
	plan := &corpusPlan{labels: []string{"first.ts:0", "second.ts:4", "third.ts:8"}, shards: []corpusShard{{indices: []int{2, 0}}, {indices: []int{1}}}}
	for _, test := range []struct {
		name   string
		output [][]byte
		want   string
	}{
		{"missing", [][]byte{[]byte("ok\tthird\n"), []byte("ok\tsecond\n")}, "first.ts:0: missing or extra case"},
		{"extra", [][]byte{[]byte("ok\tthird\nok\tfirst\nok\textra\n"), []byte("ok\tsecond\n")}, "first.ts:0: missing or extra case"},
		{"unterminated", [][]byte{[]byte("ok\tthird\nok\tfirst"), []byte("ok\tsecond\n")}, "first.ts:0: unterminated answer"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := mergeCorpus(plan, test.output); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error %v, want %s", err, test.want)
			}
		})
	}
}

// A real child reads the shard inputs; its independent expected bytes establish
// both original-order reconstruction and named comparisons across shard boundaries.
func TestCorpusShardTransport(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "cases.txt")
	var cases []printerCase
	var input, expected strings.Builder
	for i := 0; i < 23; i++ {
		source := strings.Repeat("x", 24-i)
		cases = append(cases, printerCase{fmt.Sprintf("fixture-%02d.ts:0", i), source, source})
		input.WriteString(">" + source + "\n")
		expected.WriteString("ok\t" + source + "\n")
	}
	data, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(input.String()), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "cases.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(directory, "driver.mjs")
	source := `import {readFileSync} from 'node:fs';
for(const row of readFileSync(process.argv[2],'utf8').trimEnd().split('\n')) {
 process.stdout.write('ok\t'+row.slice(1)+'\n');
}
`
	if err := os.WriteFile(script, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	result := execute(t, nil, "node", script, path)
	if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != expected.String() {
		t.Fatalf("transport: exit %d stderr %s diff %s", result.exitCode, result.stderr, corpusDifference(t, path, string(result.stdout), expected.String()))
	}
}
