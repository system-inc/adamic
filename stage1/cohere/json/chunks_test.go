package json

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type nativeChunk struct {
	path       string
	start, end int
}
type chunkResult struct {
	result run
	err    error
}

func chunkRanges(total, cores int) []nativeChunk {
	count := min(cores, 8, total)
	chunks := make([]nativeChunk, count)
	for index := range chunks {
		chunks[index].start = index * total / count
		chunks[index].end = (index + 1) * total / count
	}
	return chunks
}

func nativeChunks(t *testing.T, cases []textCase, answers []answer) []nativeChunk {
	t.Helper()
	chunks := chunkRanges(len(cases), runtime.NumCPU())
	directory := t.TempDir()
	for index := range chunks {
		chunk := &chunks[index]
		chunk.path = filepath.Join(directory, fmt.Sprintf("chunk-%d.txt", index))
		input, _ := protocol(cases[chunk.start:chunk.end], answers[chunk.start:chunk.end])
		if err := os.WriteFile(chunk.path, []byte(input), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("native corpus: %d contiguous chunks, %d cores, %d cases", len(chunks), runtime.NumCPU(), len(cases))
	return chunks
}

func runNativeChunks(t *testing.T, binary string, environment []string, chunks []nativeChunk, single run, cases []textCase, side string) run {
	t.Helper()
	results := make([]chunkResult, len(chunks))
	var workers sync.WaitGroup
	for index, chunk := range chunks {
		workers.Add(1)
		go func(index int, chunk nativeChunk) {
			defer workers.Done()
			results[index].result, results[index].err = executeResult(environment, binary, "--cases", chunk.path)
		}(index, chunk)
	}
	workers.Wait()
	joined, err := joinNativeChunks(side, chunks, results, single.stdout, cases)
	if err != nil {
		t.Fatal(err)
	}
	return joined
}

// Compare every chunk with the corresponding bytes of the single-process
// release run, then join in input order. The caller still compares with Go.
func joinNativeChunks(side string, chunks []nativeChunk, results []chunkResult, single []byte, cases []textCase) (run, error) {
	lines := strings.Split(strings.TrimSuffix(string(single), "\n"), "\n")
	if len(lines) != len(cases) {
		return run{}, fmt.Errorf("%s single-process reference has %d/%d cases", side, len(lines), len(cases))
	}
	var stdout bytes.Buffer
	for index, chunk := range chunks {
		name := fmt.Sprintf("%s chunk %d [%d:%d]", side, index, chunk.start, chunk.end)
		if results[index].err != nil {
			return run{}, fmt.Errorf("%s: %w", name, results[index].err)
		}
		expected := strings.Join(lines[chunk.start:chunk.end], "\n") + "\n"
		if err := comparisonError(name, results[index].result, expected, cases[chunk.start:chunk.end]); err != nil {
			return run{}, err
		}
		stdout.Write(results[index].result.stdout)
	}
	if !bytes.Equal(stdout.Bytes(), single) {
		return run{}, fmt.Errorf("%s joined stdout differs from single-process release stdout", side)
	}
	return run{stdout: stdout.Bytes()}, nil
}

func TestNativeChunkPlan(t *testing.T) {
	t.Parallel()
	for _, total := range []int{1, 2, 17, 2492} {
		for _, cores := range []int{1, 5, 16, 64} {
			chunks := chunkRanges(total, cores)
			if len(chunks) != min(total, cores, 8) {
				t.Fatal("wrong shard count")
			}
			end := 0
			for _, chunk := range chunks {
				if chunk.start != end || chunk.end <= chunk.start {
					t.Fatalf("lost/repeated cases: %v", chunks)
				}
				end = chunk.end
			}
			if end != total {
				t.Fatal("lost final cases")
			}
		}
	}
}

func TestNativeChunkDifferenceNamesCase(t *testing.T) {
	t.Parallel()
	cases := []textCase{{Name: "alpha.json"}, {Name: "beta.json"}, {Name: "gamma.json"}}
	chunks := chunkRanges(3, 3)
	results := []chunkResult{{result: run{stdout: []byte("ok\ta\n")}}, {result: run{stdout: []byte("ok\tb\n")}}, {result: run{stdout: []byte("ok\tc\n")}}}
	single := []byte("ok\ta\nok\tb\nok\tc\n")
	joined, err := joinNativeChunks("fixture", chunks, results, single, cases)
	if err != nil || !bytes.Equal(joined.stdout, single) {
		t.Fatalf("ordered control: %v", err)
	}
	results[1].result.stdout[3] = 'X'
	_, err = joinNativeChunks("fixture", chunks, results, single, cases)
	if err == nil || !strings.Contains(err.Error(), "chunk 1") || !strings.Contains(err.Error(), "beta.json") || !strings.Contains(err.Error(), "byte 3") {
		t.Fatalf("one-byte mutant survived or lost its case: %v", err)
	}
	t.Logf("caught planted one-byte difference: %v", err)
}

// Opt-in calibration isolates every case under ASan, including startup. The
// observed maximum is a conservative per-case stall estimate for this box.
func calibrateNativeCases(t *testing.T, binary string, cases []textCase, answers []answer) {
	t.Helper()
	directory := t.TempDir()
	durations := make([]time.Duration, len(cases))
	failures := make([]error, len(cases))
	jobs := make(chan int)
	var workers sync.WaitGroup
	for worker := 0; worker < min(runtime.NumCPU(), 8); worker++ {
		workers.Add(1)
		go func(worker int) {
			defer workers.Done()
			path := filepath.Join(directory, fmt.Sprintf("single-%d.txt", worker))
			for index := range jobs {
				input, expected := protocol(cases[index:index+1], answers[index:index+1])
				if err := os.WriteFile(path, []byte(input), 0644); err != nil {
					failures[index] = err
					continue
				}
				started := time.Now()
				result, err := executeResult([]string{"ASAN_OPTIONS=detect_leaks=0"}, binary, "--cases", path)
				durations[index] = time.Since(started)
				if err == nil {
					err = comparisonError("isolated ASan calibration", result, expected, cases[index:index+1])
				}
				failures[index] = err
			}
		}(worker)
	}
	for index := range cases {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	slowest := 0
	for index, duration := range durations {
		if failures[index] != nil {
			t.Fatalf("calibration case %d %s: %v", index, cases[index].Name, failures[index])
		}
		if duration > durations[slowest] {
			slowest = index
		}
	}
	t.Logf("ASan slowest isolated case %d %s: %.3fs over %d cases; stall %.0fs gives %.1fx headroom", slowest, cases[slowest].Name, durations[slowest].Seconds(), len(cases), jsonGuard.Stall.Seconds(), jsonGuard.Stall.Seconds()/durations[slowest].Seconds())
}
