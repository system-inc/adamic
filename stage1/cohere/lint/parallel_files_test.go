package lint

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Not parallel: ADAMIC_THREADS is the runtime's process-wide worker override.
func TestParallelFilesAgree(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	rows := parallelFilesRows(t)
	path := manifest(t, rows)
	want := node(t, directory, path, false).output
	binary := buildPort(t, directory, false)
	for _, workers := range []int{1, 2, runtime.NumCPU()} {
		t.Setenv("ADAMIC_THREADS", fmt.Sprint(workers))
		got := execute(t, "", binary, "--manifest", path).output
		if !bytes.Equal(got, want) {
			t.Fatalf("N=%d: %s", workers, difference(got, want))
		}
		t.Logf("N=%d cases=%d findings=%d fixed=%d bytes=%d sha256=%x", workers, len(rows), bytes.Count(got, []byte("\nrange ")), bytes.Count(got, []byte("\nfixed\t")), len(got), sha256.Sum256(got))
	}
}

// The original serial source is an independent rendering control. Node implements
// parallelMap sequentially; this test makes no claim about native concurrency.
func TestParallelFilesSourceControl(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	rows := parallelFilesRows(t)
	path := manifest(t, rows)
	want := node(t, directory, path, false).output
	serial := copyPort(t, t.TempDir(), "", "")
	command := exec.Command("git", "show", "9156bf5c579a44d687c9955d13e44f9ad8bbb6f8:stage1/cohere/lint/main.ts")
	original, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(serial, "main.ts"), []byte(rewritePortImports(t, "main.ts", string(original))), 0644); err != nil {
		t.Fatal(err)
	}
	got := node(t, serial, path, false).output
	if !bytes.Equal(got, want) {
		t.Fatal(difference(got, want))
	}
	t.Logf("serial and prepared source: cases=%d findings=%d fixed=%d bytes=%d sha256=%x", len(rows), bytes.Count(got, []byte("\nrange ")), bytes.Count(got, []byte("\nfixed\t")), len(got), sha256.Sum256(got))
	mutant := copyPort(t, t.TempDir(), "for(const result of results)", "for(const result of results.reverse())", "main.ts")
	changed := node(t, mutant, path, false).output
	if bytes.Equal(changed, want) {
		t.Fatal("result-order mutant survived")
	}
	t.Log("result-order source mutant caught by byte agreement")
}

// Not parallel: worker and sanitizer environment overrides belong to this run.
func TestParallelFilesThreadSanitizer(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	rows := parallelFilesRows(t)
	path := manifest(t, rows)
	want := node(t, directory, path, false).output
	built := checkerCompile(t, directory)
	binary := filepath.Join(t.TempDir(), "lint-tsan")
	options := native.Options{ThreadSanitize: true}
	if built.bridge {
		err = buildCheckerWithRuntime(built.c, binary, checkerArchive(t, false), options)
	} else {
		err = native.Build(built.c, binary, options)
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TSAN_OPTIONS", "halt_on_error=1:history_size=4:report_atomic_races=1")
	t.Setenv("ADAMIC_TSAN_PERTURB", "1")
	t.Setenv("ADAMIC_THREADS", fmt.Sprint(runtime.NumCPU()))
	got := execute(t, "", binary, "--manifest", path).output
	if !bytes.Equal(got, want) {
		t.Fatal(difference(got, want))
	}
	t.Logf("TSan clean: N=%d cases=%d", runtime.NumCPU(), len(rows))
}

func parallelFilesRows(t *testing.T) []string {
	t.Helper()
	rows := generated(t)
	for _, descriptor := range prepareRegistry(t, ".") {
		if descriptor.Name == "nexus/consistency-no-shouting" {
			rows = append(rows, ownedWitnessRows(t, ".", descriptor)...)
		}
	}
	return rows
}

func TestParallelFilesRegexControl(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, descriptor := range prepareRegistry(t, ".") {
		if descriptor.Name != "nexus/consistency-no-shouting" {
			continue
		}
		rows := ownedWitnessRows(t, ".", descriptor)
		path := manifest(t, rows)
		want := execute(t, "", goOracle(t), "--manifest", path).output
		got := node(t, directory, path, false).output
		if !bytes.Equal(got, want) {
			t.Fatal(difference(got, want))
		}
		t.Logf("local regex instances agree with Go: %d cases, %d bytes", len(rows), len(got))
		return
	}
	t.Fatal("shouting descriptor missing")
}
