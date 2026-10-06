package oracle

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/oracle/fixturedata"
)

// updateCounts writes only the measured fixtures whose recorded counts changed.
var updateCounts = flag.Bool("update-counts", false, "update changed per-fixture counts files")

// countsLine is the line a counted build writes last to stderr (runtime/count.c).
var countsLine = regexp.MustCompile(`(?m)^adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n\z`)

// pinnedStack is a command run with an 8 MiB stack, whatever the stack of whoever runs the test: a
// fixture that recurses until the stack runs out is counted at the depth it reached, which moves with
// the stack's size (ulimit -s), and the table has to be the same on every machine. The shell's ulimit
// is POSIX's, on Linux and macOS alike; where the hard limit won't allow 8 MiB, it fails out loud.
func pinnedStack(name string, arguments ...string) (string, []string) {
	return "/bin/sh", append([]string{"-c", `ulimit -s 8192 && exec "$0" "$@"`, name}, arguments...)
}

// counted builds a lowered fixture counted, runs it, and returns its row of the table. An input
// fixture runs as TestInputAgreesWithNode runs it: from its own directory, with its arguments, a
// directory of its own to write in when it writes, and as nobody when the test is root. It must
// finish, as it must there, so its row counts the whole program and not where it stopped.
func counted(t *testing.T, path string, input bool, arguments []string, unreadable bool, writes bool) string {
	t.Helper()
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, absolute)
	if err != nil {
		t.Fatalf("Lower: %v", err)
	}
	given := identity(t)
	context := cacheKey(given.context, path, fmt.Sprint(input, unreadable, writes), cacheKey(arguments...))
	if input {
		context = cacheKey(context, inputIdentity(t))
	}
	key := nativeResultKey(native.C(program), given.libraries[2], given.nodeVersion, context)
	executeCounted := func() recordedRun {
		var result run
		if input {
			shared := sharedDirectory(t)
			how := inputRun{directory: filepath.Dir(absolute), arguments: arguments}
			if os.Geteuid() == 0 {
				how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
			}
			if unreadable {
				path := filepath.Join(shared, "unreadable.txt")
				if err := os.WriteFile(path, []byte("secret\n"), 0o000); err != nil {
					t.Fatal(err)
				}
				how.arguments = append(append([]string{}, how.arguments...), path)
			}
			if writes {
				how.arguments = append([]string{writable(t, shared, "counted")}, how.arguments...)
			}
			binary := filepath.Join(shared, "program")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			name, pinned := pinnedStack(binary, how.arguments...)
			result = executeInput(t, how, nil, name, pinned...)
			if result.exitCode != 0 {
				t.Fatalf("want an input fixture's counted run to finish, as TestInputAgreesWithNode wants it to on Node: exit %d, stderr %q", result.exitCode, result.stderr)
			}
		} else {
			binary := filepath.Join(t.TempDir(), "program")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			name, pinned := pinnedStack(binary)
			result = execute(t, name, pinned...)
		}
		return record(result)
	}
	var result run
	if *updateCounts {
		result = executeCounted().run()
	} else {
		result = cachedResult(t, given.cache, nativeResults, key, executeCounted).run()
	}
	match := countsLine.FindSubmatch(result.stderr)
	if match == nil {
		t.Fatalf("no counts at the end of stderr (exit %d): %q", result.exitCode, result.stderr)
	}
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", path, match[1], match[2], match[3], match[4], match[5], match[6])
}

// Every fixture's counts are recorded; filtered updates touch only fixtures that ran.
func TestCountsAreRecorded(t *testing.T) {
	t.Parallel()
	t.Run("fixtures", func(t *testing.T) {
		for _, input := range []bool{false, true} {
			for _, fixture := range oracleFixtures(t, input) {
				if !fixture.lowers || fixture.uncounted {
					continue
				}
				t.Run(fixture.path, func(t *testing.T) {
					t.Parallel()
					row := counted(t, fixture.path, input, fixture.arguments, fixture.unreadable, fixture.writes)
					if err := fixturedata.CheckCounts(fixture.countsPath, fixture.path, row, *updateCounts); err != nil {
						t.Error(err)
					}
				})
			}
		}
	})
}
