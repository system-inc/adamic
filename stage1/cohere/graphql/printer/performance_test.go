package printer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// Not parallel: interleaved throughput measurements share the same CPUs.
func TestPrinterThroughput(t *testing.T) {
	if os.Getenv("ADAMIC_GRAPHQL_PRINTER_BENCH") == "" {
		// census: measurement Opt-in throughput of the verified GraphQL printer (ADAMIC_GRAPHQL_PRINTER_BENCH=1); timing only, never a required input of the gate.
		t.Skip("set ADAMIC_GRAPHQL_PRINTER_BENCH=1 to measure verified throughput")
	}
	library := os.Getenv("ADAMIC_GRAPHQL_PRETTIER")
	if library == "" {
		t.Fatal("set ADAMIC_GRAPHQL_PRETTIER to scratch prettier@3.9.6")
	}
	cases, want := printerCases(t, "defaults")
	inputs, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	lines, answers := strings.Split(string(inputs), "\n"), strings.Split(want, "\n")
	var selected, expected strings.Builder
	count := 0
	for i, answer := range answers {
		if strings.HasPrefix(answer, "ok\t") {
			selected.WriteString(lines[i] + "\n")
			expected.WriteString(answer + "\n")
			count++
		}
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "parsed.txt")
	if err = os.WriteFile(path, []byte(selected.String()), 0644); err != nil {
		t.Fatal(err)
	}
	port, _ := filepath.Abs("main.ts")
	binary := filepath.Join(directory, "native")
	if err = native.Build(native.C(lowered(t, port)), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	cohere, _ := filepath.Abs(filepath.Join(repository, "cohere"))
	driver, _ := filepath.Abs("testdata/cohere_driver.go")
	fake := filepath.Join(cohere, "adamic_graphql_printer_benchmark.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{fake: driver}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err = os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	goBinary := filepath.Join(directory, "cohere")
	command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, fake)
	command.Dir = cohere
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Go driver: %v\n%s", err, output)
	}
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	prettier, _ := filepath.Abs("testdata/prettier.mjs")
	commands := []struct {
		name, command string
		args          []string
	}{
		{"native", binary, []string{"--cases", path}},
		{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, port, "--cases", path}},
		{"Go cohere", goBinary, []string{path}},
		{"Prettier 3.9.6", "node", []string{prettier, library, path}},
	}
	times := map[string][]float64{}
	// An untimed warm-up, then five rounds with the first side rotating.
	for round := -1; round < 5; round++ {
		for offset := range commands {
			index := offset
			if round >= 0 {
				index = (offset + round) % len(commands)
			}
			side := commands[index]
			start := time.Now()
			result := execute(t, nil, side.command, side.args...)
			seconds := time.Since(start).Seconds()
			if result.exitCode != 0 || len(result.stderr) > 0 || string(result.stdout) != expected.String() {
				t.Fatalf("%s timed output: exit %d, stderr %s, difference %s", side.name, result.exitCode, result.stderr, firstDifference(string(result.stdout), expected.String()))
			}
			if round >= 0 {
				times[side.name] = append(times[side.name], seconds)
			}
		}
	}
	t.Logf("%d successfully formatted texts; process startup, input and escaped output included, compilation excluded", count)
	for _, side := range commands {
		samples := times[side.name]
		best := samples[0]
		for _, seconds := range samples {
			if seconds < best {
				best = seconds
			}
		}
		t.Logf("%s: seconds %v, best %.6f, %.1f texts/s", side.name, samples, best, float64(count)/best)
	}
}
