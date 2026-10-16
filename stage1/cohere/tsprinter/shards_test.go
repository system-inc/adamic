package tsprinter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

func executeShards(t *testing.T, environment []string, name string, arguments []string) (oracles.Run, bool) {
	t.Helper()
	position := -1
	for i, argument := range arguments {
		switch filepath.Base(argument) {
		case "cases.txt", "cases.json", "tsc-cases.txt", "docs.txt", "docs.json":
			position = i
		}
	}
	if position < 0 {
		return oracles.Run{}, false
	}
	// Tiny pinned upstream probes have JSON only and stay a single process.
	path := arguments[position]
	if _, err := os.Stat(strings.TrimSuffix(path, filepath.Ext(path)) + ".txt"); os.IsNotExist(err) {
		return oracles.Run{}, false
	}
	plan := oracles.PlanCorpus(t, path)
	outputs := make([][]byte, len(plan.Shards))
	results := make([]oracles.Run, len(plan.Shards))
	var group sync.WaitGroup
	for number, shard := range plan.Shards {
		group.Add(1)
		go func(number int, shard oracles.CorpusShard) {
			defer group.Done()
			args := append([]string(nil), arguments...)
			args[position] = shard.Text
			if filepath.Ext(path) == ".json" {
				args[position] = shard.Specs
			}
			results[number] = executeOne(t, environment, name, args...)
			outputs[number] = results[number].Stdout
		}(number, shard)
	}
	group.Wait()
	// macOS leaks adds its own report to stdout. Each invocation must pass;
	// the ordinary native pass independently checks every printed answer.
	if name == "leaks" {
		for number, result := range results {
			if result.ExitCode != 0 {
				t.Fatalf("leaks shard %d (%s): exit %d stdout %s stderr %s", number, plan.Labels[plan.Shards[number].Indices[0]], result.ExitCode, result.Stdout, result.Stderr)
			}
		}
		return oracles.Run{}, true
	}
	for number, result := range results {
		if result.ExitCode != 0 || len(result.Stderr) != 0 {
			t.Fatalf("%s shard %d (%s): exit %d stderr %s", name, number, plan.Labels[plan.Shards[number].Indices[0]], result.ExitCode, result.Stderr)
		}
	}
	stdout, err := oracles.MergeCorpus(plan, outputs)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return oracles.Run{Stdout: stdout}, true
}

func corpusDifference(t *testing.T, cases, got, want string) string {
	t.Helper()
	if got == want {
		return ""
	}
	plan := oracles.PlanCorpus(t, cases)
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
			if i < len(plan.Labels) {
				label = plan.Labels[i]
			}
			return fmt.Sprintf("%s: %s", label, firstDifference(a, b))
		}
	}
	return firstDifference(got, want)
}

func TestCorpusShardAssignment(t *testing.T) {
	t.Parallel()
	got := oracles.AssignCorpus([]int{1, 10, 2, 10, 3, 4}, 2)
	want := [][]int{{1, 5, 0}, {3, 4, 2}}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("largest-first assignment %v, want %v", got, want)
	}
}

func TestCorpusShardCoverage(t *testing.T) {
	t.Parallel()
	plan := &oracles.CorpusPlan{Labels: []string{"first.ts:0", "second.ts:4", "third.ts:8"}, Shards: []oracles.CorpusShard{{Indices: []int{2, 0}}, {Indices: []int{1}}}}
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
			if _, err := oracles.MergeCorpus(plan, test.output); err == nil || !strings.Contains(err.Error(), test.want) {
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
	var cases []oracles.PrinterCase
	var input, expected strings.Builder
	for i := 0; i < 23; i++ {
		source := strings.Repeat("x", 24-i)
		cases = append(cases, oracles.PrinterCase{Label: fmt.Sprintf("fixture-%02d.ts:0", i), Source: source, Want: source})
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
	if result.ExitCode != 0 || len(result.Stderr) != 0 || string(result.Stdout) != expected.String() {
		t.Fatalf("transport: exit %d stderr %s diff %s", result.ExitCode, result.Stderr, corpusDifference(t, path, string(result.Stdout), expected.String()))
	}
}
