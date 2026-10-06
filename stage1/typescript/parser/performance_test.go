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
func TestPerformance(t *testing.T) {
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
		{"Go", func() execution { return execute(t, "", oracle, "--manifest", manifest, "--count") }},
		{"Node", func() execution { return node(t, directory, manifest, true) }},
		{"native", func() execution { return execute(t, "", binary, "--manifest", manifest, "--count") }},
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
