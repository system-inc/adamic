package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Not parallel: compiler mutations are bounded subprocesses. The corpus pin
// and point selection are deterministic and no diagnostic input is filtered.
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
	binary := buildPort(t, directory, true)
	artifacts := os.Getenv("ADAMIC_RECOVERY_ARTIFACTS")
	if artifacts == "" {
		artifacts = "/tmp/adamic-parser-incomplete"
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
	t.Logf("planned all %d compiler files: %d inputs", files, planned)
	checked := 0
	start := 0
	if value := os.Getenv("ADAMIC_RECOVERY_START"); value != "" {
		start, err = strconv.Atoi(value)
		if err != nil || start < 0 || start >= planned {
			t.Fatalf("invalid ADAMIC_RECOVERY_START %q", value)
		}
		t.Logf("debug continuation skips first %d inputs; this is not a full corpus gate", start)
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
			if position <= start {
				return
			}
			path := filepath.Join(artifacts, fmt.Sprintf("%s-%s-%d.ts", filepath.Base(sourcePath), mode, index))
			if err := os.WriteFile(path, input, 0644); err != nil {
				t.Fatal(err)
			}
			want, err := recoveryRun(t, path+".go", oracle, path, "--whole", "--recovery")
			if err != nil {
				t.Fatalf("Go %s: %v; saved input %s", sourcePath, err, path)
			}
			for _, side := range []struct {
				name, command string
				args          []string
			}{
				{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), path, "--whole", "--recovery"}},
				{"native", binary, []string{path, "--whole", "--recovery"}},
			} {
				got, err := recoveryRun(t, path+"."+side.name, side.command, side.args...)
				if err != nil {
					t.Errorf("%s %s %s token %d: %v; saved input %s", side.name, sourcePath, mode, index, err, path)
				} else if !bytes.Equal(got, want) {
					t.Errorf("%s %s %s token %d: %s; saved input %s", side.name, sourcePath, mode, index, difference(got, want), path)
				}
			}
			checked++
			if t.Failed() {
				t.Fatalf("comparison stopped at input %d after checking %d inputs; first failure retained", position, checked)
			}
			// Successful artifacts are reproducible from the pin and point selection.
			// Retain every failed input and both output streams, not gigabytes of successes.
			for _, suffix := range []string{"", ".go.stdout", ".go.stderr", ".Node.stdout", ".Node.stderr", ".native.stdout", ".native.stderr"} {
				if err := os.Remove(path + suffix); err != nil {
					t.Fatal(err)
				}
			}
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
	if start != 0 {
		t.Logf("debug continuation: %d inputs checked, %d skipped; full corpus gate not run", checked, start)
	} else {
		t.Logf("%d compiler files, %d incomplete inputs: Go, Node and sanitized native identical", files, checked)
	}
}
