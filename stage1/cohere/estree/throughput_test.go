package estree

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// Not parallel: native.Build writes the shared user cache directory adamic/runtime
func TestThroughput(test *testing.T) {
	if os.Getenv("ADAMIC_ESTREE_BENCHMARK") != "1" {
		test.Skip("set ADAMIC_ESTREE_BENCHMARK=1 for full-output throughput")
	}
	path, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	list := manifest(test, generated())
	listing, readError := os.ReadFile(list)
	if readError != nil {
		test.Fatal(readError)
	}
	if writeError := os.WriteFile(list, []byte(strings.Repeat(string(listing), 20)), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	count := len(generated()) * 20
	binary, _ := build(test, path, false)
	oracle := goOracle(test)
	commands := [][]string{{oracle, "--manifest", list}, {
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(root(test), "oracle/node.mjs"),
		path,
		"--manifest",
		list,
	}, {binary, "--manifest", list}}
	names := []string{"Go", "source Node", "native"}
	want := execute(test, "", commands[0][0], commands[0][1:]...)
	samples := make([][]float64, 3)
	for round := 0; round < 6; round++ {
		for offset := 0; offset < 3; offset++ {
			index := (round + offset) % 3
			start := time.Now()
			got := execute(test, "", commands[index][0], commands[index][1:]...)
			elapsed := time.Since(start).Seconds()
			if !bytes.Equal(want, got) {
				test.Fatalf("%s: %s", names[index], firstDifference(want, got))
			}
			if round > 0 {
				samples[index] = append(samples[index], float64(count)/elapsed)
			}
		}
	}
	for index, name := range names {
		unsorted := fmt.Sprint(samples[index])
		sort.Float64s(samples[index])
		test.Logf(
			"%s: %d texts, %d checked bytes/run; five samples texts/s %s; median %.0f texts/s",
			name,
			count,
			len(want),
			unsorted,
			samples[index][2],
		)
	}
	test.Log(
		"Includes startup, file reading, complete canonical serialization, file-backed stdout and verification read; release native, one warm-up and five rotated runs",
	)
}
