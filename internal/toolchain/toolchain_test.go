package toolchain

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// executableAt writes a script that prints what it is at path.
func executableAt(test *testing.T, path string) {
	test.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho "+path+"\n"), 0o755); err != nil {
		test.Fatal(err)
	}
}

// A tool is found under ADAMIC_TOOLS as env.sh lays it out, bin before go/bin, ahead of one first on PATH; a tool not
// there is PATH's outside a sandbox, and an error naming where it looked inside one. Mutants: PATH asked first; go/bin
// before bin; the sandbox falling back to PATH.
// Not parallel: test.Setenv.
func TestAToolIsFoundUnderTheToolsDirectoryAndOnlyThereInASandbox(test *testing.T) {
	tools, elsewhere := test.TempDir(), test.TempDir()
	executableAt(test, filepath.Join(tools, "bin", "adamic-fake-tool"))
	executableAt(test, filepath.Join(tools, "go", "bin", "adamic-fake-tool"))
	executableAt(test, filepath.Join(tools, "go", "bin", "adamic-fake-go"))
	executableAt(test, filepath.Join(elsewhere, "adamic-fake-tool"))
	executableAt(test, filepath.Join(elsewhere, "adamic-fake-go"))
	executableAt(test, filepath.Join(elsewhere, "adamic-fake-other"))
	test.Setenv("PATH", elsewhere)
	test.Setenv("ADAMIC_TOOLS", tools)
	test.Setenv("ADAMIC_SANDBOX", "")
	for name, want := range map[string]string{
		"adamic-fake-tool":  filepath.Join(tools, "bin", "adamic-fake-tool"),
		"adamic-fake-go":    filepath.Join(tools, "go", "bin", "adamic-fake-go"),
		"adamic-fake-other": filepath.Join(elsewhere, "adamic-fake-other"),
		"/bin/sh":           "/bin/sh",
	} {
		if got, err := Path(name); err != nil || got != want {
			test.Errorf("Path(%q) = %q, %v; want %q", name, got, err, want)
		}
	}
	test.Setenv("ADAMIC_SANDBOX", "1")
	if got, err := Path("adamic-fake-tool"); err != nil || got != filepath.Join(tools, "bin", "adamic-fake-tool") {
		test.Errorf("in a sandbox, Path(adamic-fake-tool) = %q, %v", got, err)
	}
	if got, err := Path(
		"adamic-fake-other",
	); err == nil || !strings.Contains(err.Error(), filepath.Join(tools, "bin")) {
		test.Errorf("in a sandbox, a tool only on PATH resolved: %q, %v", got, err)
	}
	if err := Command("adamic-fake-other").Run(); err == nil || !strings.Contains(err.Error(), "ADAMIC_SANDBOX=1") {
		test.Errorf("in a sandbox, a command for a tool only on PATH ran: %v", err)
	}
	test.Setenv("ADAMIC_TOOLS", "")
	if got, err := Path("adamic-fake-tool"); err == nil || !strings.Contains(err.Error(), "ADAMIC_TOOLS is unset") {
		test.Errorf("in a sandbox with no tools directory, Path(adamic-fake-tool) = %q, %v", got, err)
	}
	test.Setenv("ADAMIC_SANDBOX", "")
	if got, err := Path("adamic-fake-tool"); err != nil || got != filepath.Join(elsewhere, "adamic-fake-tool") {
		test.Errorf("with no tools directory, Path(adamic-fake-tool) = %q, %v; want PATH's", got, err)
	}
}

// values is an environment as a map, the last setting of a name winning, as exec.Cmd takes it.
func values(environment []string) map[string]string {
	found := map[string]string{}
	for _, setting := range environment {
		name, value, _ := strings.Cut(setting, "=")
		found[name] = value
	}
	return found
}

// Env passes what the key reads, the locations a build needs and nothing else; it turns go's configuration file off,
// and the request's own settings win. PATH puts the tools directory first, once. Mutants: the whole environment
// passed; GOENV left as the caller had it; a setting of GoSettings dropped; the tools directory not put first.
// Not parallel: test.Setenv.
func TestAToolsEnvironmentIsWhatTheKeyReadsAndNothingElse(test *testing.T) {
	tools := test.TempDir()
	if err := os.MkdirAll(filepath.Join(tools, "bin"), 0o755); err != nil {
		test.Fatal(err)
	}
	test.Setenv("ADAMIC_TOOLS", tools)
	test.Setenv(
		"PATH",
		"/usr/bin"+string(filepath.ListSeparator)+filepath.Join(tools, "bin")+string(filepath.ListSeparator)+"/bin",
	)
	for _, name := range GoSettings {
		test.Setenv(name, "set-"+name)
	}
	for name, value := range map[string]string{
		"HOME":                 "/home/someone",
		"TMPDIR":               "/scratch",
		"GOCACHE":              "/cache",
		"GOCACHEPROG":          "/bin/cacheprog",
		"ADAMIC_GOCACHE_STORE": "off",
		"GOCOVERDIR":           "/scratch/cover",
		// None of these may reach a tool.
		"ADAMIC_JUNK":     "/tmp",
		"GOENV":           "/home/someone/go.env",
		"GIT_DIR":         "/nonexistent",
		"GIT_INDEX_FILE":  "/nonexistent",
		"GOROOT":          "/other/go",
		"GODEBUG":         "x=1",
		"GO111MODULE":     "off",
		"SDKROOT":         "/sdk",
		"LD_LIBRARY_PATH": "/lib",
		"ZERO_AR_DATE":    "",
	} {
		test.Setenv(name, value)
	}
	environment := values(Env("GOFLAGS=-tags=request", "EXTRA=1"))
	for _, name := range GoSettings {
		if want := "set-" + name; name != "GOFLAGS" && environment[name] != want {
			test.Errorf("%s = %q, want %q: the key reads it, so the build must see it", name, environment[name], want)
		}
	}
	for name, want := range map[string]string{
		"GOFLAGS":              "-tags=request",
		"EXTRA":                "1",
		"HOME":                 "/home/someone",
		"TMPDIR":               "/scratch",
		"GOCACHE":              "/cache",
		"GOCACHEPROG":          "/bin/cacheprog",
		"ADAMIC_GOCACHE_STORE": "off",
		"GOENV":                "off",
		"ZERO_AR_DATE":         "1",
		"GOCOVERDIR":           "/scratch/cover",
		"PATH": filepath.Join(tools, "bin") +
			string(filepath.ListSeparator) +
			"/usr/bin" +
			string(filepath.ListSeparator) +
			"/bin",
	} {
		if environment[name] != want {
			test.Errorf("%s = %q, want %q", name, environment[name], want)
		}
	}
	for _, name := range []string{
		"ADAMIC_JUNK",
		"GIT_DIR",
		"GIT_INDEX_FILE",
		"GOROOT",
		"GODEBUG",
		"GO111MODULE",
		"SDKROOT",
		"LD_LIBRARY_PATH",
		"ADAMIC_TOOLS",
	} {
		if value, ok := environment[name]; ok {
			test.Errorf("%s=%s reached a tool's environment", name, value)
		}
	}
	git := values(GitEnv())
	for name, want := range map[string]string{
		"GIT_CONFIG_GLOBAL":   os.DevNull,
		"GIT_CONFIG_NOSYSTEM": "1",
		"LC_ALL":              "C",
		"TMPDIR":              "/scratch",
	} {
		if git[name] != want {
			test.Errorf("git's %s = %q, want %q", name, git[name], want)
		}
	}
	for _, name := range []string{"HOME", "GIT_DIR", "GIT_INDEX_FILE", "ADAMIC_JUNK", "GOFLAGS"} {
		if _, ok := git[name]; ok {
			test.Errorf("%s reached git's environment", name)
		}
	}
	if !slices.Contains(GitEnv("GIT_DIR=/request"), "GIT_DIR=/request") {
		test.Error("a request's own setting didn't reach git")
	}
}

// A command run in another directory (CommandIn, GitCommandIn) is told it stands there, as exec.Cmd tells one with no
// Env of its own: its PWD is the directory, however the caller's PWD reads, and a relative directory is made absolute.
// Mutant: in setting Dir alone (the tool reads the caller's PWD).
// Not parallel: test.Setenv and test.Chdir.
func TestACommandRunInADirectoryIsToldItStandsThere(test *testing.T) {
	directory := test.TempDir()
	test.Setenv("PWD", "/somewhere/else")
	test.Setenv("ADAMIC_SANDBOX", "")
	if got, err := CommandIn(
		directory,
		"sh",
		"-c",
		`printf %s "$PWD"`,
	).Output(); err != nil || string(got) != directory {
		test.Errorf("CommandIn ran with PWD %q (%v), want %q", got, err, directory)
	}
	if command := GitCommandIn(
		directory,
		"status",
	); command.Err != nil || command.Dir != directory || !slices.Contains(command.Env, "PWD="+directory) {
		test.Errorf("GitCommandIn's Dir %q and environment %v (%v)", command.Dir, command.Env, command.Err)
	}
	test.Chdir(filepath.Dir(directory))
	if got, err := CommandIn(
		filepath.Base(directory),
		"sh",
		"-c",
		`printf %s "$PWD"`,
	).Output(); err != nil || string(got) != directory {
		test.Errorf("a relative directory ran with PWD %q (%v), want %q", got, err, directory)
	}
}

// A tool's C and C++ compilers are clang and clang++ by name when the caller names neither, never go's default of gcc
// from PATH (#eehjrv6: 66 units' cgo ran the machine's gcc, as and ld); a caller's own compiler is kept, and a
// request's wins. go itself, under Env, reports CC as clang. Mutant: Env leaving CC unset (go reports gcc on Linux, and
// on a Mac the test's own CC=clang still fails the unset check).
// Not parallel: test.Setenv.
func TestAToolsCompilersAreClangUnlessTheCallerNamesOne(test *testing.T) {
	// Empty is unnamed, as go reads it.
	test.Setenv("CC", "")
	test.Setenv("CXX", "")
	environment := values(Env())
	if environment["CC"] != "clang" || environment["CXX"] != "clang++" {
		test.Fatalf("with no compiler named, CC=%q and CXX=%q", environment["CC"], environment["CXX"])
	}
	if output, err := Command(
		"go",
		"env",
		"CC",
		"CXX",
	).Output(); err != nil || strings.Fields(string(output))[0] != "clang" {
		test.Fatalf("go under Env reports %q (%v)", output, err)
	}
	test.Setenv("CC", "/usr/bin/cc")
	if environment := values(Env("CXX=g++")); environment["CC"] != "/usr/bin/cc" || environment["CXX"] != "g++" {
		test.Fatalf("a caller's CC and a request's CXX: CC=%q, CXX=%q", environment["CC"], environment["CXX"])
	}
}

// A tool keeps its caller's time zone, so a unit's TZ=UTC reaches every tool and none reads the machine's zone file
// (Server's /etc/localtime differed from the house's, Oct 12, and 742 products' read sets held it). Go names its local
// zone UTC when TZ=UTC reaches it and Local when it read the zone file, so the test runs itself as the tool with the
// environment Command gives one (Env). Mutant: TZ left out of carried.
// Not parallel: test.Setenv.
func TestAToolKeepsItsCallersTimeZone(test *testing.T) {
	if os.Getenv("ADAMIC_TOOLCHAIN_ZONE") == "print" {
		fmt.Print(time.Local.String())
		return
	}
	test.Setenv("TZ", "UTC")
	command := exec.Command(os.Args[0], "-test.run=^TestAToolKeepsItsCallersTimeZone$")
	command.Env = append(Env(), "ADAMIC_TOOLCHAIN_ZONE=print")
	output, err := command.Output()
	if err != nil || !strings.HasPrefix(string(output), "UTC") {
		test.Fatalf(
			"a tool under a caller's TZ=UTC named its zone %q (%v): it read the machine's zone file",
			output,
			err,
		)
	}
}
