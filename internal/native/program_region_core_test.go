package native_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

func programCoreCode(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/program_region/core.c")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
func programCoreRun(environment []string, command string, arguments ...string) leakcheck.Run {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	process := exec.CommandContext(ctx, command, arguments...)
	process.Env = append(os.Environ(), environment...)
	var stdout, stderr bytes.Buffer
	process.Stdout, process.Stderr = &stdout, &stderr
	err := process.Run()
	exit := 0
	if err != nil {
		exit = -1
		if process.ProcessState != nil {
			exit = process.ProcessState.ExitCode()
		}
	}
	return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: exit}
}
func programCoreLeaks(t *testing.T, code, binary, mode string, build func(string, string) error) string {
	t.Helper()
	report, err := leakcheck.Check(leakcheck.Program{C: code, Sanitized: binary,
		Counted: filepath.Join(t.TempDir(), "counted"), BuildCounted: build,
		Arguments: func() []string { return []string{mode} }, Execute: programCoreRun})
	if err != nil {
		t.Fatal(err)
	}
	return report
}
func TestProgramRegionCore(t *testing.T) {
	// Not parallel: compile and compare the opt-in allocator variants and mutations in bounded subprocesses.
	code := programCoreCode(t)
	node := programCoreRun(nil, "node", "--input-type=commonjs", "-e",
		`require('node:vm').runInThisContext(require('node:fs').readFileSync(process.argv[1], 'utf8'))`,
		"testdata/program_region/oracle.a")
	if node.ExitCode != 0 || len(node.Stderr) != 0 {
		t.Fatalf("Node: exit %d stderr %s", node.ExitCode, node.Stderr)
	}
	for _, options := range []native.Options{
		{ProgramRegion: true, Sanitize: true, Count: true},
		{ProgramRegion: true, Sanitize: true, Slabs: true, Count: true},
		{ProgramRegion: true, Count: true},
	} {
		t.Run(fmt.Sprintf("asan=%t-slabs=%t", options.Sanitize, options.Slabs), func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "program-region")
			if err := native.Build(code, binary, options); err != nil {
				t.Fatal(err)
			}
			for _, mode := range []string{"end", "exit"} {
				run := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, mode)
				if run.ExitCode != 0 {
					t.Fatalf("%s: exit %d stderr %s", mode, run.ExitCode, run.Stderr)
				}
				if !bytes.Equal(run.Stdout, node.Stdout) {
					t.Fatalf("Node differs: native %q Node %q", run.Stdout, node.Stdout)
				}
				if report := leakcheck.Unbalanced(run); report != "" {
					t.Fatal(report)
				}
				counts := leakcheck.CountsLine.FindSubmatch(run.Stderr)
				if len(counts) != 7 {
					t.Fatalf("missing counts: %s", run.Stderr)
				}
				if string(counts[1]) != "11" || string(counts[2]) != "7" || string(counts[6]) != "4" {
					t.Fatalf("unexpected membership: %s", run.Stderr)
				}
				t.Logf("%s %s", mode, strings.TrimSpace(string(run.Stderr)))
			}
			// ASan malloc owns individual storage, so it is the independent Linux leak proof.
			if options.Sanitize && !options.Slabs {
				if report := programCoreLeaks(t, code, binary, "end", func(code, output string) error {
					return native.Build(code, output, native.Options{ProgramRegion: true, Count: true})
				}); report != "" {
					t.Fatal(report)
				}
				if report := programCoreLeaks(t, code, binary, "exit", func(code, output string) error {
					return native.Build(code, output, native.Options{ProgramRegion: true, Count: true})
				}); report != "" {
					t.Fatal(report)
				}
				for _, mode := range []string{"share", "share-cell", "cached-adoption", "weak-before-adoption", "weak-cell-before-adoption"} {
					refused := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0"}, binary, mode)
					message := "Program region members cannot cross into parallel work"
					if mode == "cached-adoption" {
						message = "Program adoption must precede canonical caching"
					}
					if strings.HasPrefix(mode, "weak-") {
						message = "Program adoption needs a fresh counted allocation"
					}
					if refused.ExitCode != 70 || !strings.Contains(string(refused.Stderr), message) || strings.Contains(string(refused.Stderr), "ERROR: AddressSanitizer") {
						t.Fatalf("%s: exit %d stderr %s", mode, refused.ExitCode, refused.Stderr)
					}
				}
			}
		})
	}
}

func TestProgramRegionTarget(t *testing.T) {
	t.Setenv("ADAMIC_PROGRAM_REGION", "")
	const refusal = "native: Program regions require a one-shot native CLI"
	for _, options := range []native.Options{{ProgramRegion: true, Target: "wasm32-wasi"}, {ProgramRegion: true, Request: true}} {
		if err := native.Build("int main(void) {return 0;}", filepath.Join(t.TempDir(), "refused"), options); err == nil || err.Error() != refusal {
			t.Fatalf("want %q, got %v", refusal, err)
		}
	}
	if !strings.Contains(strings.Join(native.Flags(native.Options{ProgramRegion: true}), " "), "-DADAMIC_PROGRAM_REGION") {
		t.Fatal("missing opt-in flag")
	}
	if strings.Contains(strings.Join(native.Flags(native.Options{}), " "), "-DADAMIC_PROGRAM_REGION") {
		t.Fatal("default enabled Program region")
	}
	t.Setenv("ADAMIC_PROGRAM_REGION", "1")
	if err := native.ValidateOptions(native.Options{Target: "wasm32-wasi"}); err == nil || err.Error() != refusal {
		t.Fatalf("environment switch bypassed refusal: %v", err)
	}
}

type programCoreRuntime struct{ directory string }

func (snapshot programCoreRuntime) build(code, output string, options native.Options) error {
	library, err := native.RuntimeLibraryForSource(snapshot.directory, code, options)
	if err != nil {
		return err
	}
	source := output + ".c"
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		return err
	}
	args := append(native.Flags(options), "-I", filepath.Dir(library), "-o", output, source)
	args = append(args, native.RuntimeLinkFlags(library)...)
	args = append(args, "-lm")
	result, err := exec.Command("clang", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("mutant must compile: %w\n%s", err, result)
	}
	return nil
}
func programCoreSnapshot(t *testing.T, file, before, after string) programCoreRuntime {
	t.Helper()
	snapshot := programCoreRuntime{t.TempDir()}
	entries, err := os.ReadDir("runtime")
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, entry := range entries {
		if entry.IsDir() || (filepath.Ext(entry.Name()) != ".c" && filepath.Ext(entry.Name()) != ".h") {
			continue
		}
		data, err := os.ReadFile(filepath.Join("runtime", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() == file {
			if strings.Count(string(data), before) != 1 {
				t.Fatalf("mutation must name one site in %s", file)
			}
			data = []byte(strings.Replace(string(data), before, after, 1))
			changed = true
		}
		if err := os.WriteFile(filepath.Join(snapshot.directory, entry.Name()), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if !changed {
		t.Fatal("missing mutation file")
	}
	return snapshot
}
func TestProgramRegionCoreMutants(t *testing.T) {
	for _, mutant := range []struct{ name, file, before, after, caught string }{
		{"member-release-frees", "heap.c", "if (heap == NULL || adamic_program_is(heap)) { return NULL; }", "if (heap == NULL) { return NULL; }\n    if (adamic_program_is(heap)) { return heap; }", "heap-use-after-free"},
		{"weak-not-forgotten", "program_region.c", "  adamic_weak_forget(each);", "  /* forgot the member's weak handles */", "adamic_weak_target(weak) == NULL"},
		{"cell-weak-not-forgotten", "program_region.c", "adamic_weak_forget(&environment->cells[i]);", "(void)i;", "adamic_weak_target(cell_weak) == NULL"},
		{"short-adoption", "adamic.h", "count * (sizeof(adamic_value) + sizeof(size_t) + 2)", "count * (sizeof(adamic_value) + 2)", "heap-buffer-overflow"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			snapshot := programCoreSnapshot(t, mutant.file, mutant.before, mutant.after)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := snapshot.build(programCoreCode(t), binary, native.Options{ProgramRegion: true, Sanitize: true, Count: true}); err != nil {
				t.Fatal(err)
			}
			run := programCoreRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, "end")
			if run.ExitCode == 0 || !strings.Contains(string(run.Stderr), mutant.caught) {
				t.Fatalf("want %s: exit %d stderr %s", mutant.caught, run.ExitCode, run.Stderr)
			}
			t.Logf("caught by %s", mutant.caught)
		})
	}
}
