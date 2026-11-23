package oracle

import (
	"debug/elf"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
	"github.com/system-inc/adamic/internal/toolchain"
)

// mathLinkBuilds are the builds a Linux binary is linked as, each of which names libm through
// native_runtime.MathLinkFlags: plain, sanitized, counted, a shipped release whose code is generated at the link by
// ThinLTO, and a split build, which links its units in units.go rather than native.go.
var mathLinkBuilds = []struct {
	name    string
	options native_runtime.Options
}{
	{"plain", native_runtime.Options{}},
	{"sanitized", native_runtime.Options{Sanitize: true}},
	{"counted", native_runtime.Options{Count: true}},
	{"release", native_runtime.Options{Release: true}},
	{"split", native_runtime.Options{Split: true}},
}

// A Linux binary loads libm at start only when something it links calls libm (native_runtime.MathLinkFlags, #7z36qn6).
// A program whose % on doubles lowers to fmod keeps libm and imports fmod in every build, the release's included, whose
// calls are made at the link by ThinLTO. Hello world's release build, the one adamic build ships, holds no libm in its
// NEEDED. Its other builds keep libm with reason: the runtime's archive is linked whole (RuntimeLinkFlags), so every
// runtime object is in the binary, some of them call libm, and only the release's link-time optimization drops the ones
// hello never reaches; the test logs what each of them imports. Each binary's output is held to Node's.
func TestMathLinksLibmOnlyWhereCalled(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("libm is libSystem's on macOS, so there is no load to keep or drop")
	}
	node, err := toolchain.Path("node")
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct {
		path string
		// libm is whether the binary must keep libm, and the fmod it must import with it.
		libm bool
	}{
		{"internal/oracle/testdata/math_link/hello.a", false},
		{"internal/oracle/testdata/math_link/fmod.a", true},
	} {
		t.Run(filepath.Base(fixture.path), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(repository, fixture.path)
			expected, err := exec.Command(node, "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path).Output()
			if err != nil {
				t.Fatalf("Node: %v", err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			for _, build := range mathLinkBuilds {
				t.Run(build.name, func(t *testing.T) {
					binary := filepath.Join(t.TempDir(), "program")
					if err := native.Build(code, binary, build.options); err != nil {
						t.Fatal(err)
					}
					command := exec.Command(binary)
					command.Env = []string{"ASAN_OPTIONS=detect_leaks=0"}
					printed, err := command.Output()
					if err != nil {
						t.Fatalf("running the %s build: %v", build.name, err)
					}
					if string(printed) != string(expected) {
						t.Errorf("the %s build printed %q, and Node %q", build.name, printed, expected)
					}
					file, err := elf.Open(binary)
					if err != nil {
						t.Fatal(err)
					}
					defer file.Close()
					libraries, err := file.ImportedLibraries()
					if err != nil {
						t.Fatal(err)
					}
					symbols, err := file.ImportedSymbols()
					if err != nil {
						t.Fatal(err)
					}
					keepsLibm := slices.Contains(libraries, "libm.so.6")
					importsFmod := slices.ContainsFunc(symbols, func(symbol elf.ImportedSymbol) bool { return symbol.Name == "fmod" })
					switch {
					case fixture.libm && (!keepsLibm || !importsFmod):
						t.Errorf("the %s build calls fmod, so it must keep libm and import fmod: NEEDED %v, fmod imported %v", build.name, libraries, importsFmod)
					case !fixture.libm && keepsLibm && build.options.Release:
						t.Errorf("the %s build calls nothing of libm and still loads it at start: NEEDED %v", build.name, libraries)
					case !fixture.libm && keepsLibm:
						t.Logf("the %s build keeps libm for the runtime objects linked whole: %d imported symbols, NEEDED %v", build.name, len(symbols), libraries)
					}
				})
			}
		})
	}
}
