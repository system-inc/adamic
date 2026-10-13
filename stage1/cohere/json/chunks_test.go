package json

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

type nativeChunk struct {
	start, end int
}
type chunkResult struct {
	result run
	err    error
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
	for worker := 0; worker < min(runtime.GOMAXPROCS(0), 8); worker++ {
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
