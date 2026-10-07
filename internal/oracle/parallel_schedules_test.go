package oracle

import (
	"bytes"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// Give this test its own gate shard, with Go's result cache disabled:
// ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestParallelSchedules$' -count=1 -v -timeout 30m
// Node and native observations below bypass the oracle result cache even without that environment.
func TestParallelSchedules(t *testing.T) {
	t.Parallel()
	executions := 0
	defer func() { t.Logf("schedule sweep: %d native executions", executions) }()
	for _, fixture := range fixtures {
		if !strings.HasPrefix(fixture.path, "internal/oracle/testdata/concurrency/accepted/") && fixture.path != "bench/parallel_files.a" {
			continue
		}
		// Not parallel: each process can start 64 threads; keep fixture pools from competing.
		t.Run(fixture.path, func(t *testing.T) {
			path := filepath.Join(repository, fixture.path)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if !usesParallelMap(program) {
				t.Fatal("schedule fixture emits no parallel map")
			}
			oracle := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			builds := []struct {
				name    string
				options native.Options
			}{{"release", native.Options{}}}
			if runtime.GOOS == "linux" {
				builds = append(builds, struct {
					name    string
					options native.Options
				}{"tsan", native.Options{ThreadSanitize: true}})
			}
			for _, build := range builds {
				t.Run(build.name, func(t *testing.T) {
					binary := filepath.Join(t.TempDir(), "program")
					if err := native.Build(native.C(program), binary, build.options); err != nil {
						t.Fatal(err)
					}
					t.Logf("build flags: %s", strings.Join(native.Flags(build.options), " "))
					completed := 0
					defer func() { t.Logf("%s: %d native executions", build.name, completed) }()
					for _, threads := range []string{"1", "2", "3", "7", "64", ""} {
						name, repetitions := threads, 20
						if threads == "1" {
							repetitions = 1
						}
						if threads == "" {
							name = "unset"
						}
						t.Run(name, func(t *testing.T) {
							for repetition := 1; repetition <= repetitions; repetition++ {
								// A fresh command.Run on the same binary, never a cached observation or a rebuild.
								// Match the runtime oracle's bounded TSan allowance under gate load.
								deadline := time.Minute
								if build.options.ThreadSanitize {
									deadline = 5 * time.Minute
								}
								observed := executeParallelDeadline(t, threads, false, deadline, binary)
								completed++
								executions++
								t.Logf("execution=%d threads=%s repetition=%d/%d exit=%d", executions, name, repetition, repetitions, observed.exitCode)
								if bytes.Contains(observed.stderr, []byte("ThreadSanitizer")) {
									t.Fatalf("threads=%s repetition=%d: ThreadSanitizer report:\n%s", name, repetition, observed.stderr)
								}
								if difference := disagreement(oracle, observed); difference != "" {
									t.Fatalf("threads=%s repetition=%d: %s: Node exit %d stdout %q stderr %q; native exit %d stdout %q stderr %q", name, repetition, difference, oracle.exitCode, oracle.stdout, oracle.stderr, observed.exitCode, observed.stdout, observed.stderr)
								}
							}
						})
					}
				})
			}
		})
	}
}
