package oracle

import (
	"bytes"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// Opt-in sequential measurement. Compilation is excluded; each binary starts a fresh process.
func TestCyclesDecisionMeasurements(t *testing.T) {
	if runtime.GOOS != "linux" || os.Getenv("ADAMIC_CYCLES_MEASURE") != "1" {
		t.Skip("set ADAMIC_CYCLES_MEASURE=1")
	}
	load, _ := os.ReadFile("/proc/loadavg")
	t.Logf("load-before %s", load)
	runnerSource := filepath.Join(t.TempDir(), "rss.c")
	runnerCode := `#define _GNU_SOURCE
#include <sys/resource.h>
#include <sys/wait.h>
#include <unistd.h>
#include <stdio.h>
int main(int argc, char **argv) {
 if (argc != 2) return 2;
 pid_t child = fork();
 if (child < 0) return 2;
 if (child == 0) { execl(argv[1], argv[1], (char *)0); _exit(127); }
 int status; struct rusage usage;
 if (wait4(child, &status, 0, &usage) < 0) return 2;
 fprintf(stderr, "rss_kib %ld\n", usage.ru_maxrss);
 return WIFEXITED(status) ? WEXITSTATUS(status) : 128 + WTERMSIG(status);
}`
	if err := os.WriteFile(runnerSource, []byte(runnerCode), 0644); err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(t.TempDir(), "rss")
	if output, err := exec.Command("clang", "-O2", runnerSource, "-o", runner).CombinedOutput(); err != nil {
		t.Fatalf("rss runner: %v %s", err, output)
	}
	compare := os.Getenv("ADAMIC_CYCLES_PROGRAM_COMPARE") == "1"
	paths := []string{}
	for _, fixture := range fixtures {
		path := fixture.path
		if (strings.Contains(path, "graph_regions_") || strings.Contains(path, "/weak") || strings.Contains(path, "/cycles_") || strings.HasSuffix(path, "graph_regions/million.a")) && fixture.lowers && !fixture.checked {
			if !compare || strings.Contains(path, "/cycles_") || strings.HasSuffix(path, "graph_regions/million.a") {
				paths = append(paths, path)
			}
		}
	}
	probes, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/weak/*.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range probes {
		relative, _ := filepath.Rel(repository, probe)
		if !compare {
			paths = append(paths, relative)
		}
	}
	modes := []bool{false}
	if compare {
		modes = append(modes, true)
	}
	for _, path := range paths {
		for _, enabled := range modes {
			mode := "graph"
			if enabled {
				mode = "program"
			}
			expectedExit := 0
			if strings.HasSuffix(path, "/weak/narrowed.a") {
				expectedExit = 70
			}

			absolute, _ := filepath.Abs(filepath.Join(repository, path))
			var program *ir.Program
			if compare {
				program = programRegionLowered(t, absolute, enabled)
			} else {
				var err error
				program, err = lowered(t, absolute)
				if err != nil {
					t.Fatal(err)
				}
			}

			code := native.C(program)
			binary := filepath.Join(t.TempDir(), "release")
			options := native.Options{Release: true, ProgramRegion: enabled}
			if err := native.Build(code, binary, options); err != nil {
				t.Fatal(err)
			}
			best := time.Duration(1<<63 - 1)
			rss := ""
			samples := []string{}
			for i := 0; i < 3; i++ {
				var stdout, stderr bytes.Buffer
				cmd := exec.Command(runner, binary)
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				started := time.Now()
				err := cmd.Run()
				elapsed := time.Since(started)
				if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != expectedExit {
					t.Fatalf("%s: %v %s", path, err, stderr.String())
				}
				rssAt := strings.LastIndex(stderr.String(), "rss_kib ")
				if rssAt < 0 {
					t.Fatal("no RSS")
				}
				measuredRSS := strings.TrimSpace(stderr.String()[rssAt:])
				samples = append(samples, fmt.Sprintf("%.9f %s", elapsed.Seconds(), measuredRSS))
				if elapsed < best {
					best = elapsed
					rss = measuredRSS
				}
			}
			t.Logf("MEASURE %s mode %s seconds %.9f %s samples %q flags %s", path, mode, best.Seconds(), rss, samples, strings.Join(native.Flags(options), " "))
			countedBinary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(code, countedBinary, native.Options{Count: true, ProgramRegion: enabled}); err != nil {
				t.Fatal(err)
			}
			name, args := pinnedStack(countedBinary)
			result := execute(t, name, args...)
			if result.exitCode != expectedExit {
				t.Fatalf("%s counted: %d %s", path, result.exitCode, result.stderr)
			}
			t.Logf("COUNTS %s mode %s\n%s", path, mode, result.stderr)
		}
	}
	load, _ = os.ReadFile("/proc/loadavg")
	t.Logf("load-after %s", load)
}
