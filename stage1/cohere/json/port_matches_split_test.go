package json

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestPortMatchesGoCohere(t *testing.T) {
	t.Parallel()
	cases := sampledCorpusCases(t, 32)
	// The Go oracle is mandatory even without an optional Prettier installation.
	goAnswers, _ := cohereAnswers(t, cases, false)
	input, expected := protocol(cases, goAnswers)
	path := filepath.Join(t.TempDir(), "cases.txt")
	if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
		path = filepath.Join(artifacts, "cases.txt")
		writeJSON(t, filepath.Join(artifacts, "cases.json"), cases)
		writeJSON(t, filepath.Join(artifacts, "go.json"), goAnswers)
	}
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	source := portDirectory(t, nil)
	entry := filepath.Join(source, "main.ts")
	started := time.Now()
	nodeRun := onNode(t, entry, "--cases", path)
	nodeTime := time.Since(started)
	if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
		os.WriteFile(filepath.Join(artifacts, "node.txt"), nodeRun.stdout, 0644)
	}
	compare(t, "Node", nodeRun, expected, cases)
	if os.Getenv("ADAMIC_JSON_NODE_ONLY") != "" {
		t.Skip("debug run requested Node only; no native parity claim")
	}
	program := lowered(t, entry)
	release := filepath.Join(t.TempDir(), "release")
	if err := native.Build(native.C(program), release, native.Options{}); err != nil {
		t.Fatal(err)
	}
	single := execute(t, nil, release, "--cases", path)
	compare(t, "release", single, expected, cases)
	binary := filepath.Join(t.TempDir(), "sanitized")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	chunks := nativeChunks(t, cases, goAnswers)
	sanitized := runNativeChunks(t, binary, []string{"ASAN_OPTIONS=detect_leaks=0"}, chunks, single, cases, "native ASan/UBSan")
	compare(t, "native ASan/UBSan", sanitized, expected, cases)
	if runtime.GOOS == "linux" {
		leaked := runNativeChunks(t, binary, []string{"ASAN_OPTIONS=detect_leaks=1"}, chunks, single, cases, "LeakSanitizer")
		compare(t, "LeakSanitizer", leaked, expected, cases)
	} else {
		// Preserve macOS's allocation-counting/leaks tool check for every chunk.
		for index, chunk := range chunks {
			if report := leaks(t, program, binary, "--cases", chunk.path); report != "" {
				t.Fatalf("leaks chunk %d: %s", index, report)
			}
		}
	}
	compare(t, "JavaScript backend", onJavaScriptBackend(t, program, "--cases", path), expected, cases)
	if os.Getenv("ADAMIC_JSON_GUARD_CALIBRATE") == "1" {
		calibrateNativeCases(t, binary, cases, goAnswers)
	}
	if os.Getenv("ADAMIC_JSON_BENCH") == "1" {
		goDriver := buildGoDriver(t)
		for round := 0; round < 3; round++ {
			started = time.Now()
			result := execute(t, nil, release, "--cases", path)
			elapsed := time.Since(started)
			compare(t, "release", result, expected, cases)
			t.Logf("round %d native %d texts in %.3fs: %.2f texts/s", round, len(cases), elapsed.Seconds(), float64(len(cases))/elapsed.Seconds())
			started = time.Now()
			result = onNode(t, entry, "--cases", path)
			elapsed = time.Since(started)
			compare(t, "Node timed", result, expected, cases)
			t.Logf("round %d Node %d texts in %.3fs: %.2f texts/s", round, len(cases), elapsed.Seconds(), float64(len(cases))/elapsed.Seconds())
			started = time.Now()
			result = execute(t, nil, goDriver, "--cases", path)
			elapsed = time.Since(started)
			compare(t, "Go timed", result, expected, cases)
			t.Logf("round %d Go %d texts in %.3fs: %.2f texts/s", round, len(cases), elapsed.Seconds(), float64(len(cases))/elapsed.Seconds())
		}
	} else {
		t.Log("timing skipped; set ADAMIC_JSON_BENCH=1 for three timing rounds")
	}
	t.Logf("initial Node %.3fs; %d texts identical on Go, Node, native and JS backend", nodeTime.Seconds(), len(cases))
}
