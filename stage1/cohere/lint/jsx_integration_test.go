package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestJsxLintNode(t *testing.T) {
	directory := batch8Snapshot(t)
	oracle := batch8Oracle(t)
	rows := batch8Upstream(t)
	for i, row := range rows {
		fields := strings.Split(row, "\t")
		if strings.HasSuffix(fields[0], ".tsx") || strings.HasSuffix(fields[0], ".jsx") {
			data, err := os.ReadFile(fields[0])
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("case %d %s: %q", i, fields[1], data)
		}
		rows[i] = fields[0] + "\t" + fields[1]
	}
	path := manifest(t, rows)
	want := execute(t, "", oracle, "--manifest", path)
	got := batch8Node(t, directory, path, false)
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatal(diff)
	}
	t.Logf("%d original upstream sources parse directly; %d identical finding/fix/suggestion bytes", len(rows), len(want.output))
}

// Not parallel: fresh-process throughput is measured after correctness on the same sources.
func TestJsxLintReleaseAndThroughput(t *testing.T) {
	if os.Getenv("ADAMIC_LINT_BENCH") != "1" {
		t.Skip("set ADAMIC_LINT_BENCH=1 for JSX throughput")
	}
	oracle := batch8Oracle(t)
	var rows []string
	for _, row := range batch8Upstream(t) {
		path := strings.Split(row, "\t")[0]
		if batch8Spans(t, oracle, path) != "" {
			rows = append(rows, path+"\tall")
		}
	}
	if len(rows) != 54 {
		t.Fatalf("JSX fixture count %d", len(rows))
	}
	directory := batch8Snapshot(t)
	binary := batch8Build(t, directory, false)
	batch8Compare(t, oracle, binary, directory, rows)
	input := manifest(t, rows)
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
				result = batch8Node(t, directory, input, true)
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
		t.Logf("best of 5 JSX %s: %.6fs, %.2f findings/s (54 files, %d findings)", name, best[name].Seconds(), float64(count)/best[name].Seconds(), count)
	}
}

// Not parallel: this replays all JSX fixture trees before native throughput.
func TestJsxLintTrees(t *testing.T) {
	lintOracle := batch8Oracle(t)
	var paths []string
	for _, row := range batch8Upstream(t) {
		path := strings.Split(row, "\t")[0]
		if batch8Spans(t, lintOracle, path) != "" {
			paths = append(paths, path)
		}
	}
	if len(paths) != 54 {
		t.Fatalf("JSX fixture count %d", len(paths))
	}
	root, _ := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
	side, _ := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser/testdata/oracle.go"))
	virtual := filepath.Join(root, "adamic_jsx_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	overlayPath := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(t.TempDir(), "parser-oracle")
	execute(t, root, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	path := manifest(t, paths)
	want := execute(t, "", oracle, "--manifest", path, "--whole", "--jsx-recovery")
	directory, _ := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser"))
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	node := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path, "--whole")
	if diff := difference(node.output, want.output); diff != "" {
		t.Fatal(diff)
	}
	binary := batch8Build(t, directory, true)
	got := execute(t, "", binary, "--manifest", path, "--whole")
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatal(diff)
	}
	t.Logf("54 original cohere JSX sources: %d identical whole-tree bytes", len(want.output))
}
