package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Not parallel: timing the three implementations together would measure their
// contention. The opt-in benchmark runs after correctness and sanitizers.
func TestPerformance(t *testing.T)      { performance(t, false) }
func TestWholePerformance(t *testing.T) { performance(t, true) }
func performance(t *testing.T, whole bool) {
	if os.Getenv("ADAMIC_PARSER_BENCH") != "1" {
		t.Skip("set ADAMIC_PARSER_BENCH=1 for best-of-five whole compiler parsing")
	}
	manifest, files := compilerManifest(t)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	binary := buildPort(t, directory, false)
	if whole {
		want := execute(t, "", oracle, "--manifest", manifest, "--whole").output
		for _, side := range []struct {
			name string
			got  execution
		}{
			{"Node", wholeNode(t, directory, manifest, false)},
			{"release native", execute(t, "", binary, "--manifest", manifest, "--whole")},
		} {
			if diff := difference(side.got.output, want); diff != "" {
				t.Fatalf("%s preflight: %s", side.name, diff)
			}
			t.Logf("%s preflight: %d identical whole-tree bytes", side.name, len(want))
		}
	}
	t.Logf("machine: %s/%s logical processors=%d", runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
	for _, path := range []string{"/proc/cpuinfo", "/sys/fs/cgroup/cpu.max", "/sys/fs/cgroup/memory.max", "/proc/loadavg"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Logf("machine %s: %v", path, err)
			continue
		}
		text := strings.TrimSpace(string(data))
		if path == "/proc/cpuinfo" {
			for _, line := range strings.Split(text, "\n") {
				if strings.HasPrefix(line, "model name") {
					text = line
					break
				}
			}
		}
		t.Logf("machine %s: %s", path, text)
	}
	runners := []struct {
		name string
		run  func() execution
	}{
		{"Go", func() execution {
			args := []string{"--manifest", manifest, "--count"}
			if whole {
				args = append(args, "--whole")
			}
			return execute(t, "", oracle, args...)
		}},
		{"Node", func() execution {
			if whole {
				return wholeNode(t, directory, manifest, true)
			}
			return node(t, directory, manifest, true)
		}},
		{"native", func() execution {
			args := []string{"--manifest", manifest, "--count"}
			if whole {
				args = append(args, "--whole")
			}
			return execute(t, "", binary, args...)
		}},
	}
	var want []byte
	for _, runner := range runners {
		result := runner.run()
		if want == nil {
			want = result.output
		}
		if !bytes.Equal(result.output, want) {
			t.Fatalf("%s warm-up count %q differs from %q", runner.name, result.output, want)
		}
	}
	count, err := strconv.Atoi(strings.TrimSpace(string(want)))
	if err != nil || count <= 0 {
		t.Fatalf("bad count %q: %v", want, err)
	}
	best := []time.Duration{time.Hour, time.Hour, time.Hour}
	for sample := 0; sample < 5; sample++ {
		// Rotate the first implementation to avoid a fixed ordering advantage.
		for offset := 0; offset < len(runners); offset++ {
			index := (sample + offset) % len(runners)
			runner := runners[index]
			result := runner.run()
			if !bytes.Equal(result.output, want) {
				t.Fatalf("%s sample %d count %q differs", runner.name, sample+1, result.output)
			}
			if result.duration < best[index] {
				best[index] = result.duration
			}
			load, _ := os.ReadFile("/proc/loadavg")
			t.Logf("sample=%d side=%s nodes=%d seconds=%.9f nodes/s=%.0f load=%s", sample+1, runner.name, count, result.duration.Seconds(), float64(count)/result.duration.Seconds(), strings.TrimSpace(string(load)))
		}
	}
	for index, runner := range runners {
		t.Logf("BEST side=%s files=%d nodes=%d seconds=%.9f nodes/s=%.0f", runner.name, files, count, best[index].Seconds(), float64(count)/best[index].Seconds())
	}
	t.Logf("native/Go time ratio %.3fx; native/Node time ratio %.3fx", best[2].Seconds()/best[0].Seconds(), best[2].Seconds()/best[1].Seconds())
	if destination := os.Getenv("ADAMIC_PARSER_ARTIFACTS"); destination != "" {
		if err := os.MkdirAll(destination, 0755); err != nil {
			t.Fatal(err)
		}
		for _, file := range []struct{ from, name string }{{binary, "parser-native"}, {oracle, "parser-go"}, {manifest, "compiler.manifest"}} {
			data, err := os.ReadFile(file.from)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(destination, file.name), data, 0755); err != nil {
				t.Fatal(err)
			}
		}
		t.Log(fmt.Sprintf("artifacts: %s", destination))
	}
}

// This mutant leaves every printed AST byte unchanged. Only the node-count
// check used by the benchmark can catch it, not the ordinary tree comparison.
func TestNodeCountCheckCatchesMutant(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "expressions.ts")
	if err := os.WriteFile(source, []byte("x + y * z; a?.b(x); (x) => x;"), 0644); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(directory, "manifest")
	if err := os.WriteFile(manifest, []byte(source+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	wantCount := execute(t, "", oracle, "--manifest", manifest, "--count").output
	wantTree := execute(t, "", oracle, "--manifest", manifest).output
	absolute, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	baseline := buildPort(t, absolute, true)
	for _, result := range []execution{execute(t, "", baseline, "--manifest", manifest, "--count"), node(t, absolute, manifest, true)} {
		if !bytes.Equal(result.output, wantCount) {
			t.Fatalf("baseline count %q differs from Go %q", result.output, wantCount)
		}
	}
	from := "export function countTree(nodes: readonly ParseNode[], root: number): number {\n    const node = nodes[root] ?? panic('missing parse node');\n    let count = 1;"
	to := strings.Replace(from, "let count = 1;", "let count = 0;", 1)
	mutant := copyPort(t, "nodes.ts", from, to)
	binary := buildPort(t, mutant, true)
	for _, runner := range []struct {
		name        string
		tree, count execution
	}{
		{"native", execute(t, "", binary, "--manifest", manifest), execute(t, "", binary, "--manifest", manifest, "--count")},
		{"Node", node(t, mutant, manifest, false), node(t, mutant, manifest, true)},
	} {
		if !bytes.Equal(runner.tree.output, wantTree) {
			t.Fatalf("%s count mutant also changed AST output", runner.name)
		}
		if bytes.Equal(runner.count.output, wantCount) {
			t.Fatalf("%s count mutant survived count comparison", runner.name)
		}
		t.Logf("%s count mutant finished with identical AST bytes; count %q differs from Go %q", runner.name, runner.count.output, wantCount)
	}
}
