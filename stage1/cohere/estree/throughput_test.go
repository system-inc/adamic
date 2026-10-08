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

func TestThroughput(t *testing.T) {
	if os.Getenv("ADAMIC_ESTREE_BENCHMARK") != "1" {
		// census: measurement Opt-in full-output throughput of the ESTree port (ADAMIC_ESTREE_BENCHMARK=1); timing only, never a required input of the gate.
		t.Skip("set ADAMIC_ESTREE_BENCHMARK=1 for full-output throughput")
	}
	path, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	list := manifest(t, generated())
	listing, err := os.ReadFile(list)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(list, []byte(strings.Repeat(string(listing), 20)), 0644); err != nil {
		t.Fatal(err)
	}
	count := len(generated()) * 20
	binary, _ := build(t, path, false)
	oracle := goOracle(t)
	commands := [][]string{{oracle, "--manifest", list}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), path, "--manifest", list}, {binary, "--manifest", list}}
	names := []string{"Go", "source Node", "native"}
	want := execute(t, "", commands[0][0], commands[0][1:]...)
	samples := make([][]float64, 3)
	for round := 0; round < 6; round++ {
		for offset := 0; offset < 3; offset++ {
			index := (round + offset) % 3
			start := time.Now()
			got := execute(t, "", commands[index][0], commands[index][1:]...)
			elapsed := time.Since(start).Seconds()
			if !bytes.Equal(want, got) {
				t.Fatalf("%s: %s", names[index], firstDifference(want, got))
			}
			if round > 0 {
				samples[index] = append(samples[index], float64(count)/elapsed)
			}
		}
	}
	for index, name := range names {
		unsorted := fmt.Sprint(samples[index])
		sort.Float64s(samples[index])
		t.Logf("%s: %d texts, %d checked bytes/run; five samples texts/s %s; median %.0f texts/s", name, count, len(want), unsorted, samples[index][2])
	}
	t.Log("Includes startup, file reading, complete canonical serialization, file-backed stdout and verification read; release native, one warm-up and five rotated runs")
}
