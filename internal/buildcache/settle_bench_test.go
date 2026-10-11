package buildcache

import (
	"fmt"
	"strings"
	"testing"
)

// syntheticTrace is lines lines of strace -f output in cmd/traced's form, in the mix a product test's go build and
// tools make: opens decoded to their paths, lookups that miss, listings, threads and processes started and ended,
// calls split across threads (unfinished, then resumed).
func syntheticTrace(root string, lines int) string {
	var trace strings.Builder
	pid, next := 1000, 1001
	for line := range lines {
		name := fmt.Sprintf("%s/internal/package%d/file%d.go", root, line%97, line%13)
		switch line % 10 {
		case 0, 1, 2:
			fmt.Fprintf(&trace, "%d openat(AT_FDCWD<%s>, %q, O_RDONLY|O_CLOEXEC) = 3<%s>\n", pid, root, name, name)
		case 3, 4:
			fmt.Fprintf(&trace, "%d newfstatat(AT_FDCWD<%s>, %q, 0xc0001a2b38, 0) = -1 ENOENT (No such file or directory)\n", pid, root, name+".missing")
		case 5:
			fmt.Fprintf(&trace, "%d newfstatat(AT_FDCWD<%s>, %q, {st_mode=S_IFREG|0644, st_size=1234, ...}, 0) = 0\n", pid, root, name)
		case 6:
			fmt.Fprintf(&trace, "%d getdents64(3<%s/internal/package%d>, 0xc000400000 /* 24 entries */, 8192) = 784\n", pid, root, line%97)
		case 7:
			fmt.Fprintf(&trace, "%d openat(AT_FDCWD<%s>, %q, O_RDONLY|O_CLOEXEC <unfinished ...>\n", pid, root, name)
			fmt.Fprintf(&trace, "%d clone(child_stack=0xc000200000, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM|CLONE_SETTLS, tls=0xc000100090) = %d\n", pid, next)
			fmt.Fprintf(&trace, "%d <... openat resumed>) = 4<%s>\n", pid, name)
			next++
		case 8:
			fmt.Fprintf(&trace, "%d readlinkat(AT_FDCWD<%s>, \"/proc/self/exe\", \"/usr/lib/go/pkg/tool/linux_amd64/compile\", 128) = 40\n", pid, root)
		case 9:
			fmt.Fprintf(&trace, "%d clone3({flags=CLONE_VM|CLONE_VFORK|CLONE_CLEAR_SIGHAND, exit_signal=SIGCHLD, stack=0x7f, stack_size=0x9000}, 88) = %d\n", pid, next)
			fmt.Fprintf(&trace, "%d execve(\"/usr/lib/go/pkg/tool/linux_amd64/compile\", [\"compile\", \"-o\"], 0x7ffd /* 40 vars */) = 0\n", next)
			fmt.Fprintf(&trace, "%d +++ exited with 0 +++\n", next)
			next++
		}
	}
	return trace.String()
}

// Settle reads the trace on one thread while the run goes on (#vt46geg step 3: about 96% of a core per traced test on
// Home's dry run), so its cost per line is the cost of tracing.
func BenchmarkTraceRead(b *testing.B) {
	root := "/home/ahra/levers/adamic"
	trace := syntheticTrace(root, 200_000)
	lines := strings.Count(trace, "\n")
	b.SetBytes(int64(len(trace)))
	b.ResetTimer()
	for range b.N {
		state := newTraceState(&traceFacts{root: root, roots: []string{root}, ran: map[string]bool{}, executables: map[string]bool{}, searched: map[string]bool{}})
		if err := state.read(strings.NewReader(trace)); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*lines), "ns/line")
}
