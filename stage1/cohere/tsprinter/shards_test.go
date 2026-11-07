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

func executeShards(test *testing.T, environment []string, name string, arguments []string) (oracles.Run, bool) {
	test.Helper()
	position := -1
	for index, argument := range arguments {
		switch filepath.Base(argument) {
		case "cases.txt", "cases.json", "tsc-cases.txt", "docs.txt", "docs.json":
			position = index
		}
	}
	if position < 0 {
		return oracles.Run{}, false
	}
	// Tiny pinned upstream probes have JSON only and stay a single process.
	path := arguments[position]
	if _, statError := os.Stat(strings.TrimSuffix(path, filepath.Ext(path)) + ".txt"); os.IsNotExist(statError) {
		return oracles.Run{}, false
	}
	plan := oracles.PlanCorpus(test, path)
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
			results[number] = executeOne(test, environment, name, args...)
			outputs[number] = results[number].Stdout
		}(number, shard)
	}
	group.Wait()
	// macOS leaks adds its own report to stdout. Each invocation must pass;
	// the ordinary native pass independently checks every printed answer.
	if name == "leaks" {
		for number, result := range results {
			if result.ExitCode != 0 {
				test.Fatalf(
					"leaks shard %d (%s): exit %d stdout %s stderr %s",
					number,
					plan.Labels[plan.Shards[number].Indices[0]],
					result.ExitCode,
					result.Stdout,
					result.Stderr,
				)
			}
		}
		return oracles.Run{}, true
	}
	for number, result := range results {
		if result.ExitCode != 0 || len(result.Stderr) != 0 {
			test.Fatalf(
				"%s shard %d (%s): exit %d stderr %s",
				name,
				number,
				plan.Labels[plan.Shards[number].Indices[0]],
				result.ExitCode,
				result.Stderr,
			)
		}
	}
	stdout, mergeCorpusError := oracles.MergeCorpus(plan, outputs)
	if mergeCorpusError != nil {
		test.Fatalf("%s: %v", name, mergeCorpusError)
	}
	return oracles.Run{Stdout: stdout}, true
}

func corpusDifference(test *testing.T, cases, got, want string) string {
	test.Helper()
	if got == want {
		return ""
	}
	plan := oracles.PlanCorpus(test, cases)
	actual, expected := strings.Split(got, "\n"), strings.Split(want, "\n")
	for index := 0; index < max(len(actual), len(expected)); index++ {
		a, b := "", ""
		if index < len(actual) {
			a = actual[index]
		}
		if index < len(expected) {
			b = expected[index]
		}
		if a != b {
			label := cases
			if index < len(plan.Labels) {
				label = plan.Labels[index]
			}
			return fmt.Sprintf("%s: %s", label, firstDifference(a, b))
		}
	}
	return firstDifference(got, want)
}

func TestCorpusShardAssignment(test *testing.T) {
	test.Parallel()
	got := oracles.AssignCorpus([]int{1, 10, 2, 10, 3, 4}, 2)
	want := [][]int{{1, 5, 0}, {3, 4, 2}}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		test.Fatalf("largest-first assignment %v, want %v", got, want)
	}
}

func TestCorpusShardCoverage(test *testing.T) {
	test.Parallel()
	plan := &oracles.CorpusPlan{
		Labels: []string{"first.ts:0", "second.ts:4", "third.ts:8"},
		Shards: []oracles.CorpusShard{{Indices: []int{2, 0}}, {Indices: []int{1}}},
	}
	for _, testCase := range []struct {
		name   string
		output [][]byte
		want   string
	}{
		{"missing", [][]byte{[]byte("ok\tthird\n"), []byte("ok\tsecond\n")}, "first.ts:0: missing or extra case"},
		{
			"extra",
			[][]byte{[]byte("ok\tthird\nok\tfirst\nok\textra\n"), []byte("ok\tsecond\n")},
			"first.ts:0: missing or extra case",
		},
		{
			"unterminated",
			[][]byte{[]byte("ok\tthird\nok\tfirst"), []byte("ok\tsecond\n")},
			"first.ts:0: unterminated answer",
		},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			if _, mergeCorpusError := oracles.MergeCorpus(
				plan,
				testCase.output,
			); mergeCorpusError == nil || !strings.Contains(mergeCorpusError.Error(), testCase.want) {
				test.Fatalf("error %v, want %s", mergeCorpusError, testCase.want)
			}
		})
	}
}

// A real child reads the shard inputs; its independent expected bytes establish
// both original-order reconstruction and named comparisons across shard boundaries.
func TestCorpusShardTransport(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	path := filepath.Join(directory, "cases.txt")
	var cases []oracles.PrinterCase
	var input, expected strings.Builder
	for index := 0; index < 23; index++ {
		source := strings.Repeat("x", 24-index)
		cases = append(
			cases,
			oracles.PrinterCase{Label: fmt.Sprintf("fixture-%02d.ts:0", index), Source: source, Want: source},
		)
		input.WriteString(">" + source + "\n")
		expected.WriteString("ok\t" + source + "\n")
	}
	data, encodeError := json.Marshal(cases)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(path, []byte(input.String()), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	if writeError := os.WriteFile(filepath.Join(directory, "cases.json"), data, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	script := filepath.Join(directory, "driver.mjs")
	source := `import {readFileSync} from 'node:fs';
for(const row of readFileSync(process.argv[2],'utf8').trimEnd().split('\n')) {
 process.stdout.write('ok\t'+row.slice(1)+'\n');
}
`
	if writeError := os.WriteFile(script, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	result := execute(test, nil, "node", script, path)
	if result.ExitCode != 0 || len(result.Stderr) != 0 || string(result.Stdout) != expected.String() {
		test.Fatalf(
			"transport: exit %d stderr %s diff %s",
			result.ExitCode,
			result.Stderr,
			corpusDifference(test, path, string(result.Stdout), expected.String()),
		)
	}
}
