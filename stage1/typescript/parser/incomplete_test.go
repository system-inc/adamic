package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// The deadline covers parsing and printing. Keep the 10s floor for small inputs,
// with 10s more per started MiB so large trees have scheduling and output margin.
func incompleteDeadline(inputBytes int) time.Duration {
	return 10*time.Second + time.Duration((inputBytes+(1<<20)-1)/(1<<20))*10*time.Second
}

// Workers run short manifests to amortize process startup. The corpus pin and
// point selection are deterministic; every planned case must return once.
// Prepare shared builds and start the corpus workers before the parent pauses.
// The workers overlap other tests even while the parent waits in the queue.
func TestIncompleteCompilerAgrees(t *testing.T) {
	manifest, files := compilerManifest(t)
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	batchRunner, err := filepath.Abs("testdata/batch_node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	binary := buildPort(t, directory, true)
	artifacts := os.Getenv("ADAMIC_RECOVERY_ARTIFACTS")
	if artifacts == "" {
		artifacts = "/tmp/adamic-parser-incomplete"
	}
	artifacts, err = filepath.Abs(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(artifacts, 0755); err != nil {
		t.Fatal(err)
	}
	type span struct{ start, end int }
	tokensByPath := map[string][]span{}
	planned := 0
	// Enumerate every file before comparing, so an early failure cannot hide
	// missing files or change the deterministic selection of later points.
	for _, sourcePath := range strings.Fields(string(data)) {
		info, err := os.Stat(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		artifact := filepath.Join(artifacts, filepath.Base(sourcePath)+".tokens")
		spans, err := recoveryRun(t, artifact, oracle, "--token-spans", sourcePath)
		if err != nil {
			t.Fatalf("token boundaries for %s: %v", sourcePath, err)
		}
		fields := strings.Fields(string(spans))
		if len(fields)%2 != 0 {
			t.Fatalf("odd token span fields for %s", sourcePath)
		}
		tokens := []span{}
		for i := 0; i < len(fields); i += 2 {
			start, e1 := strconv.Atoi(fields[i])
			end, e2 := strconv.Atoi(fields[i+1])
			if e1 != nil || e2 != nil || start < 0 || end <= start || int64(end) > info.Size() {
				t.Fatalf("bad token span for %s: %s %s", sourcePath, fields[i], fields[i+1])
			}
			tokens = append(tokens, span{start, end})
		}
		tokensByPath[sourcePath] = tokens
		stride := (len(tokens) + 255) / 256
		if stride == 0 {
			stride = 1
		}
		editStride := (len(tokens) + 31) / 32
		if editStride == 0 {
			editStride = 1
		}
		count := 1 + len(tokens)/stride + 2*((len(tokens)+editStride-1)/editStride)
		planned += count
		t.Logf("planned %s: %d tokens, cutoff stride %d, edit stride %d, %d inputs", sourcePath, len(tokens), stride, editStride, count)
	}
	t.Logf("planned all %d compiler files: %d inputs; each shard child has a %s CPU budget", files, planned, incompleteShardCPUBudget)
	if planned != 22497 {
		t.Fatalf("missing or extra pinned corpus cases: planned %d, expected 22497", planned)
	}
	var checked atomic.Int64
	type comparison struct {
		position, index  int
		sourcePath, mode string
		input            []byte
	}
	jobs := []comparison{}
	if value := os.Getenv("ADAMIC_RECOVERY_START"); value != "" {
		t.Fatal("ADAMIC_RECOVERY_START is not supported by the full corpus gate")
	}
	position := 0
	for _, sourcePath := range strings.Fields(string(data)) {
		source, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		tokens := tokensByPath[sourcePath]
		// Every Nth token, approximately 256 cuts per file; short files use N=1.
		stride := (len(tokens) + 255) / 256
		if stride == 0 {
			stride = 1
		}
		t.Logf("%s: %d tokens, cutoff stride %d", sourcePath, len(tokens), stride)
		compare := func(mode string, index int, input []byte) {
			position++
			jobs = append(jobs, comparison{position, index, sourcePath, mode, input})
		}
		compare("cut", 0, nil)
		for i := stride - 1; i < len(tokens); i += stride {
			compare("cut", i+1, source[:tokens[i].end])
		}
		// Up to 32 regularly spaced edit points per file, in each mode.
		editStride := (len(tokens) + 31) / 32
		if editStride == 0 {
			editStride = 1
		}
		for i := 0; i < len(tokens); i += editStride {
			token := tokens[i]
			removed := append([]byte(nil), source[:token.start]...)
			removed = append(removed, source[token.end:]...)
			compare("remove", i+1, removed)
			duplicated := append([]byte(nil), source[:token.end]...)
			duplicated = append(duplicated, ' ')
			duplicated = append(duplicated, source[token.start:]...)
			compare("duplicate", i+1, duplicated)
		}
	}
	if position != planned || len(jobs) != planned {
		t.Fatalf("missing or extra planned cases: generated %d, scheduled %d, expected %d", position, len(jobs), planned)
	}
	// Bound each command by both case count and input bytes. A slow or broken
	// parser cannot turn a whole-corpus command into a package-long hang.
	const maxCases = 64
	const maxBytes = 8 << 20
	batches := [][]comparison{}
	for begin := 0; begin < len(jobs); {
		end, inputBytes := begin, 0
		for end < len(jobs) && end-begin < maxCases {
			if end > begin && inputBytes+len(jobs[end].input) > maxBytes {
				break
			}
			inputBytes += len(jobs[end].input)
			end++
		}
		batches = append(batches, jobs[begin:end])
		begin = end
	}
	workerCount := min(runtime.NumCPU(), 8)
	t.Logf("%d workers, %d short manifests; maximum %d cases or %d input bytes; child CPU budget %s", workerCount, len(batches), maxCases, maxBytes, incompleteShardCPUBudget)
	type timing struct {
		duration time.Duration
		path     string
	}
	slowest := map[string]timing{}
	seen := make([]bool, planned)
	var resultLock sync.Mutex
	// Large successful answers stay in local test scratch. Only failures are
	// copied to the persistent artifact directory, with rebased input manifests.
	workDirectory := t.TempDir()
	runBatch := func(batchIndex int) {
		batch := batches[batchIndex]
		workBatch := filepath.Join(workDirectory, fmt.Sprintf("shard-%04d", batchIndex))
		if err := os.Mkdir(workBatch, 0755); err != nil {
			t.Error(err)
			return
		}
		defer func() {
			if err := os.RemoveAll(workBatch); err != nil {
				t.Error(err)
			}
		}()
		prefix := filepath.Join(workBatch, "output")
		savedPrefix := filepath.Join(artifacts, fmt.Sprintf("shard-%04d", batchIndex))
		paths, names := []string{}, []string{}
		for _, job := range batch {
			name := fmt.Sprintf("%05d-%s-%s-%d.ts", job.position, filepath.Base(job.sourcePath), job.mode, job.index)
			path := filepath.Join(workBatch, name)
			names = append(names, filepath.Join(artifacts, name))
			if err := os.WriteFile(path, job.input, 0644); err != nil {
				t.Error(err)
				break
			}
			paths = append(paths, path)
		}
		if len(paths) != len(batch) {
			return
		}
		manifestPath := prefix + ".manifest"
		if err := os.WriteFile(manifestPath, []byte(strings.Join(paths, "\n")+"\n"), 0644); err != nil {
			t.Error(err)
			return
		}

		saveFailure := func() {
			for index, path := range paths {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Error(err)
					continue
				}
				if err := os.WriteFile(names[index], data, 0644); err != nil {
					t.Error(err)
				}
			}
			if err := os.WriteFile(savedPrefix+".manifest", []byte(strings.Join(names, "\n")+"\n"), 0644); err != nil {
				t.Error(err)
			}
			for _, label := range []string{"Go", "Node", "native"} {
				for _, suffix := range []string{".stdout", ".stderr"} {
					data, err := os.ReadFile(prefix + "." + label + suffix)
					if err != nil {
						t.Error(err)
						continue
					}
					if err := os.WriteFile(savedPrefix+"."+label+suffix, data, 0644); err != nil {
						t.Error(err)
					}
				}
			}
		}
		answers := make([][]byte, 3)
		failed := false
		for sideIndex, side := range []struct {
			name, command string
			args          []string
		}{
			{"Go", oracle, []string{"--manifest", manifestPath, "--whole", "--recovery"}},
			{"Node", "node", []string{"--disable-warning=ExperimentalWarning", batchRunner, runner, filepath.Join(directory, "main.ts"), "--manifest", manifestPath, "--whole", "--recovery"}},
			{"native", binary, []string{"--manifest", manifestPath, "--whole", "--recovery"}},
		} {
			started := time.Now()
			answer, err := recoveryRunLimit(t, incompleteShardCPUBudget, prefix+"."+side.name, side.command, side.args...)
			elapsed := time.Since(started)
			resultLock.Lock()
			if elapsed > slowest[side.name].duration {
				slowest[side.name] = timing{elapsed, manifestPath}
			}
			resultLock.Unlock()
			if err != nil {
				t.Errorf("%s shard %d (%s through %s): %v; saved manifest %s", side.name, batchIndex, names[0], names[len(names)-1], err, savedPrefix+".manifest")
				failed = true
				continue
			}
			answers[sideIndex] = answer
			if err := checkIncompleteFrames(answer, names); err != nil {
				t.Errorf("%s shard %d: %v", side.name, batchIndex, err)
				failed = true
			}
		}
		if failed {
			saveFailure()
			return
		}
		for sideIndex, side := range []string{"Node", "native"} {
			if err := compareIncompleteFrames(answers[0], answers[sideIndex+1], names); err != nil {
				t.Errorf("%s shard %d: %v; saved manifest %s", side, batchIndex, err, savedPrefix+".manifest")
				failed = true
			}
		}
		resultLock.Lock()
		for _, job := range batch {
			if seen[job.position-1] {
				t.Errorf("extra case %s-%s-%d", job.sourcePath, job.mode, job.index)
			}
			seen[job.position-1] = true
			checked.Add(1)
		}
		resultLock.Unlock()
		if failed {
			saveFailure()
			return
		}
		t.Logf("shard %d: %d cases compared; total %d of %d", batchIndex, len(batch), checked.Load(), planned)
	}
	var cursor atomic.Int64
	var workers sync.WaitGroup
	for worker := 0; worker < workerCount; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				batchIndex := int(cursor.Add(1)) - 1
				if batchIndex >= len(batches) {
					return
				}
				runBatch(batchIndex)
			}
		}()
	}
	t.Parallel()
	workers.Wait()
	for _, label := range []string{"Go", "Node", "native"} {
		sample := slowest[label]
		t.Logf("slowest %s child: %s; manifest %s", label, sample.duration, sample.path)
	}
	for index, present := range seen {
		if !present {
			job := jobs[index]
			t.Errorf("missing case %s-%s-%d", job.sourcePath, job.mode, job.index)
		}
	}
	if checked.Load() != int64(planned) {
		t.Errorf("missing or extra case count: compared %d, expected %d", checked.Load(), planned)
	}
	if t.Failed() {
		t.Fatalf("comparison completed after checking %d inputs; failures retained", checked.Load())
	}
	t.Logf("%d compiler files, %d incomplete inputs: Go, Node and sanitized native identical", files, checked.Load())
}
