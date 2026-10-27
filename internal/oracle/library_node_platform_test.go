package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// process.platform is a host read (#t90fcv5). compare.a is tsc's getNodeSystem shape, four reads each
// against a name, and print.a prints the read alone; both are registered fixtures, held to Node on
// native and the JavaScript backend by the families, so each box's Node answer is the expectation and
// nothing here names a platform.
func init() {
	for _, name := range []string{"compare.a", "print.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/process_platform/" + name, true, false})
	}
}

// The fixtures that read process.platform, which a wasm32-wasi module can't see (Platforms' ruling on
// #t90fcv5, docs/wasm-limits.md): under the WASI oracles each stops with wasiHostPlatformPanic where
// Node prints its host's platform. wasiHostPlatform asserts the table both ways, so a fixture that
// starts or stops reading the platform without its row here goes red.
var wasiHostPlatformFixtures = map[string]bool{
	"internal/oracle/testdata/process_platform/compare.a": true,
	"internal/oracle/testdata/process_platform/print.a":   true,
}

const wasiHostPlatformPanic = "adamic: panic: process.platform: a wasm32-wasi module can't see its host's platform\n"

// readsHostPlatform is whether program reads process.platform anywhere.
func readsHostPlatform(program *ir.Program) bool {
	reads := false
	look := func(node any) any {
		if call, ok := node.(ir.NodeFSFile); ok && call.Operation == "host_platform" {
			reads = true
		}
		return node
	}
	mutateReadiness(program.Main, look)
	for _, function := range program.Functions {
		mutateReadiness(function.Body, look)
	}
	return reads
}

// wasiHostPlatform is a WASI oracle's expectation for the fixture at path: the named panic, when it
// reads the platform.
func wasiHostPlatform(t *testing.T, path string, program *ir.Program) (run, bool) {
	t.Helper()
	reads := readsHostPlatform(program)
	if wasiHostPlatformFixtures[path] != reads {
		t.Fatalf("reads process.platform: %v, but listed in wasiHostPlatformFixtures: %v; list a fixture exactly when it reads it", reads, wasiHostPlatformFixtures[path])
	}
	return run{stderr: []byte(wasiHostPlatformPanic), exitCode: 70}, reads
}

// Every box is Linux, so a constant "linux" in place of the read agrees with Node on every one. The
// shape catches it: print.a's lowered program holds the host read and no platform name among its
// strings. Mutant: the read replaced by that constant fails both.
func TestProcessPlatformIsReadNeverFolded(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/process_platform/print.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if problem := hostPlatformShape(program); problem != "" {
		t.Fatal(problem)
	}
	program.Strings = append(program.Strings, "linux")
	constant := ir.StringConstant{Index: len(program.Strings) - 1}
	program.Main = mutateReadiness(program.Main, func(node any) any {
		if call, ok := node.(ir.NodeFSFile); ok && call.Operation == "host_platform" {
			return constant
		}
		return node
	})
	if hostPlatformShape(program) == "" {
		t.Fatal("the read replaced by a constant \"linux\" kept the shape")
	}
}

func hostPlatformShape(program *ir.Program) string {
	if !readsHostPlatform(program) {
		return "no host_platform read"
	}
	for _, constant := range program.Strings {
		switch constant {
		case "linux", "darwin", "win32", "freebsd", "openbsd", "sunos", "aix":
			return "a platform name among the program's strings: " + constant
		}
	}
	return ""
}
