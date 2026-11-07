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
func TestPerformance(test *testing.T) { performance(test, false) }

// Not parallel: the whole-tree timing would measure contention with its neighbors, as TestPerformance's would.
func TestWholePerformance(test *testing.T) { performance(test, true) }

func performance(test *testing.T, whole bool) {
	if os.Getenv("ADAMIC_PARSER_BENCH") != "1" {
		test.Skip("set ADAMIC_PARSER_BENCH=1 for best-of-five whole compiler parsing")
	}
	manifest, files := compilerManifest(test)
	directory, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	oracle := goOracle(test)
	binary := buildPort(test, directory, false)
	if whole {
		want := execute(test, "", oracle, "--manifest", manifest, "--whole").output
		for _, side := range []struct {
			name string
			got  execution
		}{
			{"Node", wholeNode(test, directory, manifest, false)},
			{"release native", execute(test, "", binary, "--manifest", manifest, "--whole")},
		} {
			if diff := difference(side.got.output, want); diff != "" {
				test.Fatalf("%s preflight: %s", side.name, diff)
			}
			test.Logf("%s preflight: %d identical whole-tree bytes", side.name, len(want))
		}
	}
	test.Logf("machine: %s/%s logical processors=%d", runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
	for _, path := range []string{
		"/proc/cpuinfo",
		"/sys/fs/cgroup/cpu.max",
		"/sys/fs/cgroup/memory.max",
		"/proc/loadavg",
	} {
		data, readError := os.ReadFile(path)
		if readError != nil {
			test.Logf("machine %s: %v", path, readError)
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
		test.Logf("machine %s: %s", path, text)
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
			return execute(test, "", oracle, args...)
		}},
		{"Node", func() execution {
			if whole {
				return wholeNode(test, directory, manifest, true)
			}
			return node(test, directory, manifest, true)
		}},
		{"native", func() execution {
			args := []string{"--manifest", manifest, "--count"}
			if whole {
				args = append(args, "--whole")
			}
			return execute(test, "", binary, args...)
		}},
	}
	var want []byte
	for _, runner := range runners {
		result := runner.run()
		if want == nil {
			want = result.output
		}
		if !bytes.Equal(result.output, want) {
			test.Fatalf("%s warm-up count %q differs from %q", runner.name, result.output, want)
		}
	}
	count, parseError := strconv.Atoi(strings.TrimSpace(string(want)))
	if parseError != nil || count <= 0 {
		test.Fatalf("bad count %q: %v", want, parseError)
	}
	best := []time.Duration{time.Hour, time.Hour, time.Hour}
	for sample := 0; sample < 5; sample++ {
		// Rotate the first implementation to avoid a fixed ordering advantage.
		for offset := 0; offset < len(runners); offset++ {
			index := (sample + offset) % len(runners)
			runner := runners[index]
			result := runner.run()
			if !bytes.Equal(result.output, want) {
				test.Fatalf("%s sample %d count %q differs", runner.name, sample+1, result.output)
			}
			if result.duration < best[index] {
				best[index] = result.duration
			}
			load, _ := os.ReadFile("/proc/loadavg")
			test.Logf(
				"sample=%d side=%s nodes=%d seconds=%.9f nodes/s=%.0f load=%s",
				sample+1,
				runner.name,
				count,
				result.duration.Seconds(),
				float64(count)/result.duration.Seconds(),
				strings.TrimSpace(string(load)),
			)
		}
	}
	for index, runner := range runners {
		test.Logf(
			"BEST side=%s files=%d nodes=%d seconds=%.9f nodes/s=%.0f",
			runner.name,
			files,
			count,
			best[index].Seconds(),
			float64(count)/best[index].Seconds(),
		)
	}
	test.Logf(
		"native/Go time ratio %.3fx; native/Node time ratio %.3fx",
		best[2].Seconds()/best[0].Seconds(),
		best[2].Seconds()/best[1].Seconds(),
	)
	if destination := os.Getenv("ADAMIC_PARSER_ARTIFACTS"); destination != "" {
		if createDirectoryError := os.MkdirAll(destination, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		for _, file := range []struct{ from, name string }{
			{binary, "parser-native"},
			{oracle, "parser-go"},
			{manifest, "compiler.manifest"},
		} {
			data, readError := os.ReadFile(file.from)
			if readError != nil {
				test.Fatal(readError)
			}
			if writeError := os.WriteFile(filepath.Join(destination, file.name), data, 0755); writeError != nil {
				test.Fatal(writeError)
			}
		}
		test.Log(fmt.Sprintf("artifacts: %s", destination))
	}
}

// This mutant leaves every printed AST byte unchanged. Only the node-count
// check used by the benchmark can catch it, not the ordinary tree comparison.
func TestNodeCountCheckCatchesMutant(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	source := filepath.Join(directory, "expressions.ts")
	if writeError := os.WriteFile(source, []byte("x + y * z; a?.b(x); (x) => x;"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	manifest := filepath.Join(directory, "manifest")
	if writeError := os.WriteFile(manifest, []byte(source+"\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	oracle := goOracle(test)
	wantCount := execute(test, "", oracle, "--manifest", manifest, "--count").output
	wantTree := execute(test, "", oracle, "--manifest", manifest).output
	absolute, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	baseline := buildPort(test, absolute, true)
	for _, result := range []execution{
		execute(test, "", baseline, "--manifest", manifest, "--count"),
		node(test, absolute, manifest, true),
	} {
		if !bytes.Equal(result.output, wantCount) {
			test.Fatalf("baseline count %q differs from Go %q", result.output, wantCount)
		}
	}
	mutant := countTreeMutant(test)
	binary := buildPort(test, mutant, true)
	for _, runner := range []struct {
		name        string
		tree, count execution
	}{
		{
			"native",
			execute(test, "", binary, "--manifest", manifest),
			execute(test, "", binary, "--manifest", manifest, "--count"),
		},
		{"Node", node(test, mutant, manifest, false), node(test, mutant, manifest, true)},
	} {
		if !bytes.Equal(runner.tree.output, wantTree) {
			test.Fatalf("%s count mutant also changed AST output", runner.name)
		}
		if bytes.Equal(runner.count.output, wantCount) {
			test.Fatalf("%s count mutant survived count comparison", runner.name)
		}
		test.Logf(
			"%s count mutant finished with identical AST bytes; count %q differs from Go %q",
			runner.name,
			runner.count.output,
			wantCount,
		)
	}
}
