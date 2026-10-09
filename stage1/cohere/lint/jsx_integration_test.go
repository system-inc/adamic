package lint

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// jsxSpansOracle builds the batch JSX-membership oracle once per test run.
func jsxSpansOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs(filepath.Join(repository, "cohere"))
	side, _ := filepath.Abs("testdata/jsx_inventory.go")
	return jsxGoOracle(t, "jsx-membership", root, side, "adamic_jsx_inventory.go")
}

// jsxSources preserves the complete captured enumeration, checking membership in
// one process instead of starting a parser process for every captured case.
func jsxSources(t *testing.T) []string {
	t.Helper()
	rows := upstream(t)
	all := make([]string, len(rows))
	for i, row := range rows {
		all[i] = strings.Split(row, "\t")[0]
	}
	output := execute(t, "", jsxSpansOracle(t), manifest(t, all)).output
	membership := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) != 2 || (fields[1] != "0" && fields[1] != "1") {
			t.Fatalf("malformed JSX membership %q", line)
		}
		if _, exists := membership[fields[0]]; exists {
			t.Fatalf("repeated JSX membership %q", fields[0])
		}
		membership[fields[0]] = fields[1] == "1"
	}
	if len(membership) != len(all) {
		t.Fatalf("JSX membership count %d, want %d", len(membership), len(all))
	}
	paths, byRule, err := discoverJsxInventory(prepareRegistry(t, "."), rows, func(path string) bool {
		value, exists := membership[path]
		if !exists {
			t.Fatalf("missing JSX membership %q", path)
		}
		return value
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("discovered captured JSX cases by rule %v", byRule)
	return paths
}

func buildNative(t *testing.T, entry string, sanitize bool) string {
	t.Helper()
	inputs := jsxProductInputs{Name: "lowered " + entry, Files: jsxInputFiles(t, entry, filepath.Join(repository, "stage1"), filepath.Join(repository, "internal"), filepath.Join(repository, "cohere")), Toolchain: runtime.Version()}
	lowered := jsxProduct(t, inputs, func(dir string) error {
		program, err := load.Load([]string{entry})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(lowered)), 0644)
	})
	source := filepath.Join(lowered, "program.c")
	inputs = jsxProductInputs{Name: "native " + entry, Files: jsxInputFiles(t, source, filepath.Join(repository, "internal/native/runtime")), Flags: native.Flags(native.Options{Sanitize: sanitize}), Toolchain: string(execute(t, "", "clang", "--version").output)}
	built := jsxProduct(t, inputs, func(dir string) error {
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		return nativeBuild(func() error {
			return native.Build(string(data), filepath.Join(dir, "native"), native.Options{Sanitize: sanitize})
		})
	})
	return filepath.Join(built, "native")
}

// Not parallel: fresh-process throughput is measured after correctness on the same sources.
func TestJsxLintReleaseAndThroughput(t *testing.T) {
	if os.Getenv("ADAMIC_LINT_BENCH") != "1" {
		t.Skip("set ADAMIC_LINT_BENCH=1 for JSX throughput")
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	var rows []string
	for _, path := range jsxSources(t) {
		rows = append(rows, path+"\tall")
	}
	oracle := goOracle(t)
	binary := buildPort(t, directory, false)
	// Captured JSX includes parser-recovery witnesses, just like TestRulesAgree.
	// Mark those rows before asking the oracle to compare or count their findings.
	input := manifest(t, recoveryRows(t, oracle, rows))
	compare(t, oracle, binary, directory, input)
	best := map[string]time.Duration{}
	var answer []byte
	names := []string{"Go", "native", "Node"}
	for round := 0; round < 5; round++ {
		for offset := 0; offset < len(names); offset++ {
			name := names[(round+offset)%len(names)]
			var result execution
			switch name {
			case "Go":
				result = execute(t, "", oracle, "--manifest", input, "--count")
			case "native":
				result = execute(t, "", binary, "--manifest", input, "--count")
			case "Node":
				result = node(t, directory, input, true)
			}
			if answer == nil {
				answer = result.output
			}
			if !bytes.Equal(answer, result.output) {
				t.Fatalf("JSX finding count differs on %s", name)
			}
			if best[name] == 0 || result.duration < best[name] {
				best[name] = result.duration
			}
			t.Logf("round %d %s %s findings=%s", round+1, name, result.duration, strings.TrimSpace(string(answer)))
		}
	}
	var count int
	if _, err := fmt.Sscan(string(answer), &count); err != nil || count == 0 {
		t.Fatalf("positive count missing: %s %v", answer, err)
	}
	for _, name := range names {
		t.Logf("best of 5 JSX %s: %.6fs, %.2f findings/s (%d files, %d findings)", name, best[name].Seconds(), float64(count)/best[name].Seconds(), len(rows), count)
	}
}

const testJsxLintTreesShards = 16

// ADAMIC_TEST_SHARD=i/n selects shard indices congruent to i modulo n;
// unset runs all shards. Builds are shared inputs, made once per run until
// internal/buildcache is available. Every shard compares Go, Node and ASan/UBSan
// native whole trees, including native leak checks, for its complete assigned cases.
func TestJsxLintTrees(t *testing.T) {
	t.Parallel()
	skipWhenRuleScoped(t)
	started := time.Now()
	paths := jsxSources(t)
	shards, err := jsxTreeShards(paths)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := jsxShardSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	root, _ := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
	side, _ := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser/testdata/oracle.go"))
	oracle := jsxGoOracle(t, "jsx-parser", root, side, "adamic_jsx_oracle.go")
	directory, _ := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser"))
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	binary := buildNative(t, filepath.Join(directory, "main.ts"), true)
	t.Logf("setup including builds: %s; union: %d cases in %d shards", time.Since(started), len(paths), len(shards))
	for i, cases := range shards {
		if !selected(i) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			path := manifest(t, cases)
			want := execute(t, "", oracle, "--manifest", path, "--whole", "--jsx-recovery")
			node := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path, "--whole")
			jsxCheckTree(t, "Node", node.output, want.output)
			got := execute(t, "", binary, "--manifest", path, "--whole")
			// The failure witness is opt-in and only changes one case in shard-003.
			if os.Getenv("ADAMIC_JSX_TREE_DISAGREEMENT") == "1" && i == 3 {
				t.Logf("planted disagreement in case id %s", cases[0])
				got.output = jsxPlantDisagreement(got.output)
			}
			jsxCheckTree(t, "native", got.output, want.output)
			t.Logf("%d captured cohere JSX sources: %d identical whole-tree bytes", len(cases), len(want.output))
		})
	}
}
