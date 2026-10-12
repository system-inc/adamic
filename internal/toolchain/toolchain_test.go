package toolchain

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// executableAt writes a script that prints what it is at path.
func executableAt(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho "+path+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// A tool is found under ADAMIC_TOOLS as env.sh lays it out, bin before go/bin, ahead of one first on PATH; a tool not
// there is PATH's outside a sandbox, and an error naming where it looked inside one. Mutants: PATH asked first; go/bin
// before bin; the sandbox falling back to PATH.
// Not parallel: t.Setenv.
func TestAToolIsFoundUnderTheToolsDirectoryAndOnlyThereInASandbox(t *testing.T) {
	tools, elsewhere := t.TempDir(), t.TempDir()
	executableAt(t, filepath.Join(tools, "bin", "adamic-fake-tool"))
	executableAt(t, filepath.Join(tools, "go", "bin", "adamic-fake-tool"))
	executableAt(t, filepath.Join(tools, "go", "bin", "adamic-fake-go"))
	executableAt(t, filepath.Join(elsewhere, "adamic-fake-tool"))
	executableAt(t, filepath.Join(elsewhere, "adamic-fake-go"))
	executableAt(t, filepath.Join(elsewhere, "adamic-fake-other"))
	t.Setenv("PATH", elsewhere)
	t.Setenv("ADAMIC_TOOLS", tools)
	t.Setenv("ADAMIC_SANDBOX", "")
	for name, want := range map[string]string{
		"adamic-fake-tool":  filepath.Join(tools, "bin", "adamic-fake-tool"),
		"adamic-fake-go":    filepath.Join(tools, "go", "bin", "adamic-fake-go"),
		"adamic-fake-other": filepath.Join(elsewhere, "adamic-fake-other"),
		"/bin/sh":           "/bin/sh",
	} {
		if got, err := Path(name); err != nil || got != want {
			t.Errorf("Path(%q) = %q, %v; want %q", name, got, err, want)
		}
	}
	t.Setenv("ADAMIC_SANDBOX", "1")
	if got, err := Path("adamic-fake-tool"); err != nil || got != filepath.Join(tools, "bin", "adamic-fake-tool") {
		t.Errorf("in a sandbox, Path(adamic-fake-tool) = %q, %v", got, err)
	}
	if got, err := Path("adamic-fake-other"); err == nil || !strings.Contains(err.Error(), filepath.Join(tools, "bin")) {
		t.Errorf("in a sandbox, a tool only on PATH resolved: %q, %v", got, err)
	}
	if err := Command("adamic-fake-other").Run(); err == nil || !strings.Contains(err.Error(), "ADAMIC_SANDBOX=1") {
		t.Errorf("in a sandbox, a command for a tool only on PATH ran: %v", err)
	}
	t.Setenv("ADAMIC_TOOLS", "")
	if got, err := Path("adamic-fake-tool"); err == nil || !strings.Contains(err.Error(), "ADAMIC_TOOLS is unset") {
		t.Errorf("in a sandbox with no tools directory, Path(adamic-fake-tool) = %q, %v", got, err)
	}
	t.Setenv("ADAMIC_SANDBOX", "")
	if got, err := Path("adamic-fake-tool"); err != nil || got != filepath.Join(elsewhere, "adamic-fake-tool") {
		t.Errorf("with no tools directory, Path(adamic-fake-tool) = %q, %v; want PATH's", got, err)
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
// Not parallel: t.Setenv.
func TestAToolsEnvironmentIsWhatTheKeyReadsAndNothingElse(t *testing.T) {
	tools := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tools, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_TOOLS", tools)
	t.Setenv("PATH", "/usr/bin"+string(filepath.ListSeparator)+filepath.Join(tools, "bin")+string(filepath.ListSeparator)+"/bin")
	for _, name := range GoSettings {
		t.Setenv(name, "set-"+name)
	}
	for name, value := range map[string]string{
		"HOME": "/home/someone", "TMPDIR": "/scratch", "GOCACHE": "/cache", "GOCACHEPROG": "/bin/cacheprog", "ADAMIC_GOCACHE_STORE": "off",
		// None of these may reach a tool.
		"ADAMIC_JUNK": "/tmp", "GOENV": "/home/someone/go.env", "GIT_DIR": "/nonexistent", "GIT_INDEX_FILE": "/nonexistent", "GOROOT": "/other/go",
		"GODEBUG": "x=1", "GO111MODULE": "off", "SDKROOT": "/sdk", "LD_LIBRARY_PATH": "/lib", "ZERO_AR_DATE": "",
	} {
		t.Setenv(name, value)
	}
	environment := values(Env("GOFLAGS=-tags=request", "EXTRA=1"))
	for _, name := range GoSettings {
		if want := "set-" + name; name != "GOFLAGS" && environment[name] != want {
			t.Errorf("%s = %q, want %q: the key reads it, so the build must see it", name, environment[name], want)
		}
	}
	for name, want := range map[string]string{
		"GOFLAGS": "-tags=request", "EXTRA": "1", "HOME": "/home/someone", "TMPDIR": "/scratch", "GOCACHE": "/cache", "GOCACHEPROG": "/bin/cacheprog",
		"ADAMIC_GOCACHE_STORE": "off", "GOENV": "off", "ZERO_AR_DATE": "1",
		"PATH": filepath.Join(tools, "bin") + string(filepath.ListSeparator) + "/usr/bin" + string(filepath.ListSeparator) + "/bin",
	} {
		if environment[name] != want {
			t.Errorf("%s = %q, want %q", name, environment[name], want)
		}
	}
	for _, name := range []string{"ADAMIC_JUNK", "GIT_DIR", "GIT_INDEX_FILE", "GOROOT", "GODEBUG", "GO111MODULE", "SDKROOT", "LD_LIBRARY_PATH", "ADAMIC_TOOLS"} {
		if value, ok := environment[name]; ok {
			t.Errorf("%s=%s reached a tool's environment", name, value)
		}
	}
	git := values(GitEnv())
	for name, want := range map[string]string{"GIT_CONFIG_GLOBAL": os.DevNull, "GIT_CONFIG_NOSYSTEM": "1", "LC_ALL": "C", "TMPDIR": "/scratch"} {
		if git[name] != want {
			t.Errorf("git's %s = %q, want %q", name, git[name], want)
		}
	}
	for _, name := range []string{"HOME", "GIT_DIR", "GIT_INDEX_FILE", "ADAMIC_JUNK", "GOFLAGS"} {
		if _, ok := git[name]; ok {
			t.Errorf("%s reached git's environment", name)
		}
	}
	if !slices.Contains(GitEnv("GIT_DIR=/request"), "GIT_DIR=/request") {
		t.Error("a request's own setting didn't reach git")
	}
}

// A command run in another directory (CommandIn, GitCommandIn) is told it stands there, as exec.Cmd tells one with no
// Env of its own: its PWD is the directory, however the caller's PWD reads, and a relative directory is made absolute.
// Mutant: in setting Dir alone (the tool reads the caller's PWD).
// Not parallel: t.Setenv and t.Chdir.
func TestACommandRunInADirectoryIsToldItStandsThere(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("PWD", "/somewhere/else")
	t.Setenv("ADAMIC_SANDBOX", "")
	if got, err := CommandIn(directory, "sh", "-c", `printf %s "$PWD"`).Output(); err != nil || string(got) != directory {
		t.Errorf("CommandIn ran with PWD %q (%v), want %q", got, err, directory)
	}
	if command := GitCommandIn(directory, "status"); command.Err != nil || command.Dir != directory || !slices.Contains(command.Env, "PWD="+directory) {
		t.Errorf("GitCommandIn's Dir %q and environment %v (%v)", command.Dir, command.Env, command.Err)
	}
	t.Chdir(filepath.Dir(directory))
	if got, err := CommandIn(filepath.Base(directory), "sh", "-c", `printf %s "$PWD"`).Output(); err != nil || string(got) != directory {
		t.Errorf("a relative directory ran with PWD %q (%v), want %q", got, err, directory)
	}
}

// A tool's C and C++ compilers are clang and clang++ by name when the caller names neither, never go's default of gcc
// from PATH (#eehjrv6: 66 units' cgo ran the machine's gcc, as and ld); a caller's own compiler is kept, and a
// request's wins. go itself, under Env, reports CC as clang. Mutant: Env leaving CC unset (go reports gcc on Linux, and
// on a Mac the test's own CC=clang still fails the unset check).
// Not parallel: t.Setenv.
func TestAToolsCompilersAreClangUnlessTheCallerNamesOne(t *testing.T) {
	// Empty is unnamed, as go reads it.
	t.Setenv("CC", "")
	t.Setenv("CXX", "")
	environment := values(Env())
	if environment["CC"] != "clang" || environment["CXX"] != "clang++" {
		t.Fatalf("with no compiler named, CC=%q and CXX=%q", environment["CC"], environment["CXX"])
	}
	if output, err := Command("go", "env", "CC", "CXX").Output(); err != nil || strings.Fields(string(output))[0] != "clang" {
		t.Fatalf("go under Env reports %q (%v)", output, err)
	}
	t.Setenv("CC", "/usr/bin/cc")
	if environment := values(Env("CXX=g++")); environment["CC"] != "/usr/bin/cc" || environment["CXX"] != "g++" {
		t.Fatalf("a caller's CC and a request's CXX: CC=%q, CXX=%q", environment["CC"], environment["CXX"])
	}
}
