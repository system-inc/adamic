package native

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"
)

func parallelHarness(t *testing.T, file string, options Options) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "parallel", file))
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(string(source), binary, options); err != nil {
		t.Fatal(err)
	}
	return binary
}

func parallelRun(t *testing.T, binary string, threads string, arguments ...string) (string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, binary, arguments...)
	command.Env = append(os.Environ(), "ADAMIC_THREADS="+threads, "ASAN_OPTIONS=detect_leaks=1", "TSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("threads=%s: %v\n%s\n%s", threads, err, stdout.String(), stderr.String())
	}
	if strings.Contains(stderr.String(), "WARNING: ThreadSanitizer") || strings.Contains(stderr.String(), "ERROR:") {
		t.Fatal(stderr.String())
	}
	return stdout.String(), stderr.String()
}

func parallelBuilds() []struct {
	name    string
	options Options
} {
	builds := []struct {
		name    string
		options Options
	}{{"asan", Options{Sanitize: true}}, {"count", Options{Count: true}}}
	if goruntime.GOOS == "linux" {
		builds = append(builds, struct {
			name    string
			options Options
		}{"tsan", Options{ThreadSanitize: true}})
	}
	return builds
}

func TestParallelMemory(t *testing.T) {
	t.Parallel()
	for _, build := range parallelBuilds() {
		t.Run(build.name, func(t *testing.T) {
			binary := parallelHarness(t, "memory.c", build.options)
			stdout, _ := parallelRun(t, binary, "4")
			if stdout != "memory clean\n" {
				t.Fatalf("got %q", stdout)
			}
		})
	}
}
