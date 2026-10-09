package json

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"
)

// Only buckets exceeding 30s in both isolated command walls are subdivided.
// Each child keeps every comparison for its cases. Shared products are fetched
// before portMatchesRunItems starts the child's own-work deadline.
func jsonGrain30Names(ordinal int) []string {
	switch ordinal {
	case 1, 4, 61, 110:
		names := []string{fmt.Sprintf("TestPortMatchesGoCohere_%03d_000", ordinal), fmt.Sprintf("TestPortMatchesGoCohere_%03d_001", ordinal)}
		if jsonGrain30SeparateLeak(ordinal) {
			names = append(names, fmt.Sprintf("TestPortMatchesGoCohere_%03d_002", ordinal))
		}
		return names
	default:
		return []string{fmt.Sprintf("TestPortMatchesGoCohere_%03d", ordinal)}
	}
}

// Greedy byte balancing isolates the dominant fixture without changing its text
// or filename. Name breaks ties so every independently selected process agrees.
func jsonGrain30Parts(items []textCase) [][]textCase {
	ordered := append([]textCase(nil), items...)
	sort.Slice(ordered, func(i, j int) bool {
		if len(ordered[i].Text) != len(ordered[j].Text) {
			return len(ordered[i].Text) > len(ordered[j].Text)
		}
		return ordered[i].Name < ordered[j].Name
	})
	parts := make([][]textCase, 2)
	weights := [2]int{}
	for _, item := range ordered {
		child := 0
		if weights[1] < weights[0] {
			child = 1
		}
		parts[child] = append(parts[child], item)
		weights[child] += len(item.Text) + 1
	}
	return parts
}

func jsonGrain30Union(items []textCase, parts [][]textCase) error {
	expected := make(map[string]textCase, len(items))
	for _, item := range items {
		if _, exists := expected[item.Name]; exists {
			return fmt.Errorf("duplicate input %s", item.Name)
		}
		expected[item.Name] = item
	}
	seen := make(map[string]bool, len(items))
	for child, cases := range parts {
		for _, item := range cases {
			want, exists := expected[item.Name]
			if !exists || want != item || seen[item.Name] {
				return fmt.Errorf("child %d changed, repeated or added %s", child, item.Name)
			}
			seen[item.Name] = true
		}
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("union contains %d of %d cases", len(seen), len(expected))
	}
	return nil
}

// The largest indivisible fixtures still exceed the wall budget when both
// sanitizer passes run serially. Match the existing class-order arrangement:
// each case has one agreement owner and one independent leak-pass owner.
func jsonGrain30SeparateLeak(ordinal int) bool {
	return ordinal == 1 || ordinal == 61 || ordinal == 110
}

type jsonGrain30Work struct {
	cases           []textCase
	agreement, leak bool
}

func jsonGrain30Plan(ordinal int, items []textCase) []jsonGrain30Work {
	parts := jsonGrain30Parts(items)
	separate := jsonGrain30SeparateLeak(ordinal)
	work := []jsonGrain30Work{{parts[0], true, !separate}, {parts[1], true, true}}
	if separate {
		work = append(work, jsonGrain30Work{parts[0], false, true})
	}
	return work
}

func jsonGrain30Coverage(items []textCase, work []jsonGrain30Work) error {
	expected := make(map[string]textCase, len(items))
	for _, item := range items {
		if _, exists := expected[item.Name]; exists {
			return fmt.Errorf("duplicate input %s", item.Name)
		}
		expected[item.Name] = item
	}
	seen := make(map[string][2]int, len(items))
	for _, child := range work {
		for _, item := range child.cases {
			if want, exists := expected[item.Name]; !exists || want != item {
				return fmt.Errorf("changed or added %s", item.Name)
			}
			count := seen[item.Name]
			if child.agreement {
				count[0]++
			}
			if child.leak {
				count[1]++
			}
			seen[item.Name] = count
		}
	}
	for name := range expected {
		if count := seen[name]; count != [2]int{1, 1} {
			return fmt.Errorf("%s has agreement/leak owners %v, want [1 1]", name, count)
		}
	}
	return nil
}

func jsonGrain30Run(t *testing.T, ordinal, child int) {
	t.Helper()
	portMatchesPrepare(t)
	items := portMatchesShared.shards[ordinal]
	work := jsonGrain30Plan(ordinal, items)
	if err := jsonGrain30Coverage(items, work); err != nil {
		t.Fatal(err)
	}
	t.Logf("exact child union: %d cases in %d children; child %d has %d cases; agreement=%t leak=%t", len(items), len(work), child, len(work[child].cases), work[child].agreement, work[child].leak)
	if work[child].agreement {
		portMatchesRunItems(t, ordinal, work[child].cases, true, work[child].leak)
	} else {
		jsonGrain30Leak(t, work[child].cases)
	}
}

func jsonGrain30Leak(t *testing.T, items []textCase) {
	t.Helper()
	deadline := portMatchesDeadline(t.Name())
	defer deadline.Stop()
	started := time.Now()
	input, _ := protocol(items, make([]answer, len(items)))
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	reference := execute(t, nil, portMatchesShared.oracle, "--cases", path)
	if reference.exitCode != 0 || len(reference.stderr) != 0 {
		t.Fatalf("Go oracle exit %d: %s", reference.exitCode, reference.stderr)
	}
	expected := string(reference.stdout)
	answers := portMatchesAnswers(t, expected, len(items))
	release := execute(t, nil, portMatchesShared.release, "--cases", path)
	compare(t, "release reference", release, expected, items)
	chunks := nativeChunks(t, items, answers)
	if runtime.GOOS == "linux" {
		leaked := runNativeChunks(t, portMatchesShared.sanitized, []string{"ASAN_OPTIONS=detect_leaks=1"}, chunks, release, items, "LeakSanitizer")
		compare(t, "LeakSanitizer", leaked, expected, items)
	} else {
		for _, chunk := range chunks {
			report := execute(t, nil, "leaks", "--atExit", "--", portMatchesShared.release, "--cases", chunk.path)
			if report.exitCode != 0 {
				t.Fatalf("leaks: %s", report.stdout)
			}
		}
	}
	t.Logf("leak child: %d cases; execution %.3fs", len(items), time.Since(started).Seconds())
}

// The serial sampled runner adds the costs of both Node backends. As in the
// 2048-bucket runner, execute non-sanitized sides together, then let the shared
// memory admission limit the sanitizer waves. No test failure runs on a worker.
func jsonGrain30Sides(t *testing.T, path string) []run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	sides := []struct {
		name, binary string
		args         []string
	}{
		{"Go", portMatchesShared.oracle, []string{"--cases", path}},
		{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, portMatchesShared.entry, "--cases", path}},
		{"release", portMatchesShared.release, []string{"--cases", path}},
		{"JavaScript backend", "node", []string{"--disable-warning=ExperimentalWarning", runner, portMatchesShared.script, "--cases", path}},
	}
	results := make([]run, len(sides))
	errors := make([]error, len(sides))
	elapsed := make([]time.Duration, len(sides))
	var workers sync.WaitGroup
	for i, side := range sides {
		workers.Add(1)
		go func(i int, binary string, args []string) {
			defer workers.Done()
			started := time.Now()
			results[i], errors[i] = executeResult(nil, binary, args...)
			elapsed[i] = time.Since(started)
		}(i, side.binary, side.args)
	}
	workers.Wait()
	for i, side := range sides {
		t.Logf("%s own work: %.3fs", side.name, elapsed[i].Seconds())
		if errors[i] != nil {
			t.Fatalf("%s: %v", side.name, errors[i])
		}
	}
	return results
}

func TestPortMatchesGoCohere_001_000(t *testing.T) { t.Parallel(); jsonGrain30Run(t, 1, 0) }
func TestPortMatchesGoCohere_001_001(t *testing.T) { t.Parallel(); jsonGrain30Run(t, 1, 1) }
func TestPortMatchesGoCohere_001_002(t *testing.T) { t.Parallel(); jsonGrain30Run(t, 1, 2) }
func TestPortMatchesGoCohere_004_000(t *testing.T) { t.Parallel(); jsonGrain30Run(t, 4, 0) }
func TestPortMatchesGoCohere_004_001(t *testing.T) { t.Parallel(); jsonGrain30Run(t, 4, 1) }
func TestPortMatchesGoCohere_061_000(t *testing.T) { t.Parallel(); jsonGrain30Run(t, 61, 0) }
func TestPortMatchesGoCohere_061_001(t *testing.T) { t.Parallel(); jsonGrain30Run(t, 61, 1) }
func TestPortMatchesGoCohere_061_002(t *testing.T) { t.Parallel(); jsonGrain30Run(t, 61, 2) }
func TestPortMatchesGoCohere_110_000(t *testing.T) { t.Parallel(); jsonGrain30Run(t, 110, 0) }
func TestPortMatchesGoCohere_110_001(t *testing.T) { t.Parallel(); jsonGrain30Run(t, 110, 1) }
func TestPortMatchesGoCohere_110_002(t *testing.T) { t.Parallel(); jsonGrain30Run(t, 110, 2) }

func TestJSONGrain30Union(t *testing.T) {
	t.Parallel()
	shards := portMatchesPartition(t)
	file, err := parser.ParseFile(token.NewFileSet(), "grain30_json_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	bindings := make(map[string][2]int)
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || function.Body == nil || len(function.Body.List) != 2 {
			continue
		}
		statement, ok := function.Body.List[1].(*ast.ExprStmt)
		if !ok {
			continue
		}
		call, ok := statement.X.(*ast.CallExpr)
		if !ok {
			continue
		}
		helper, ok := call.Fun.(*ast.Ident)
		if !ok || helper.Name != "jsonGrain30Run" || len(call.Args) != 3 {
			continue
		}
		var binding [2]int
		for i := range binding {
			literal, ok := call.Args[i+1].(*ast.BasicLit)
			if !ok || literal.Kind != token.INT {
				t.Fatalf("%s has a nonliteral child binding", function.Name.Name)
			}
			value, err := strconv.Atoi(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			binding[i] = value
		}
		bindings[function.Name.Name] = binding
	}
	for _, ordinal := range []int{1, 4, 61, 110} {
		items := shards[ordinal]
		parts := jsonGrain30Parts(items)
		work := jsonGrain30Plan(ordinal, items)
		if err := jsonGrain30Coverage(items, work); err != nil {
			t.Fatal(err)
		}
		if err := jsonGrain30Union(items, parts); err != nil {
			t.Fatal(err)
		}
		names := jsonGrain30Names(ordinal)
		if len(work) != len(names) {
			t.Fatal("child enumeration differs from partition")
		}
		for child, name := range names {
			if binding, exists := bindings[name]; !exists || binding != [2]int{ordinal, child} {
				t.Fatalf("%s binds to %v, want [%d %d]", name, binding, ordinal, child)
			}
			delete(bindings, name)
		}
		t.Logf("%03d: exact union of %d cases across %v", ordinal, len(items), names)
		// Each nonempty child owns a changed oracle answer exactly once, tested
		// through the same comparisonError that real backend comparisons use.
		for owner, owned := range work {
			if len(owned.cases) == 0 {
				continue
			}
			planted := owned.cases[0].Name
			caught := 0
			for child, candidate := range work {
				cases := candidate.cases
				answers := make([]answer, len(cases))
				_, expected := protocol(cases, answers)
				for i, item := range cases {
					if item.Name == planted && ((owned.agreement && candidate.agreement) || (!owned.agreement && candidate.leak)) {
						answers[i].Output = "planted disagreement"
					}
				}
				_, actual := protocol(cases, answers)
				if comparisonError("planted", run{stdout: []byte(actual)}, expected, cases) != nil {
					caught++
					if child != owner {
						t.Fatal("wrong child caught planted failure")
					}
					t.Logf("planted %q caught only by %s", planted, names[child])
				}
			}
			if caught != 1 {
				t.Fatalf("planted failure caught by %d children", caught)
			}
		}
		if err := jsonGrain30Coverage(items, work[1:]); err == nil {
			t.Fatal("missing comparison owner survived")
		}
		if err := jsonGrain30Coverage(items, append(append([]jsonGrain30Work(nil), work...), work[0])); err == nil {
			t.Fatal("repeated comparison owner survived")
		}
		flat := append(append([]textCase(nil), parts[0]...), parts[1]...)
		if len(flat) == 0 {
			t.Fatal("empty measured bucket")
		}
		missing := [][]textCase{flat[1:]}
		repeated := [][]textCase{append(append([]textCase(nil), flat...), flat[0])}
		changed := append([]textCase(nil), flat...)
		changed[0].Text += " "
		for _, mutant := range [][][]textCase{missing, repeated, {changed}} {
			if err := jsonGrain30Union(items, mutant); err == nil {
				t.Fatal("invalid child union survived")
			}
		}
	}
	if len(bindings) != 0 {
		t.Fatalf("unexpected child wrappers: %v", bindings)
	}
}
