package buildcache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hostileHome is a home whose git configuration, read, would run its own programs inside every git a key asks and
// change what git compares: an fsmonitor hook and a clean filter (each touches marker), core.autocrlf, an attributes
// file that sends every path through the filter, and an alias. In ~/.gitconfig and again in $XDG_CONFIG_HOME/git, the
// two places git looks.
func hostileHome(t *testing.T) (home, marker string) {
	t.Helper()
	home = t.TempDir()
	marker = filepath.Join(home, "git-ran-the-homes-program")
	program := filepath.Join(home, "program.sh")
	if err := os.WriteFile(program, []byte("#!/bin/sh\ntouch "+marker+"\nprintf 'token\\0'\ncat\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	configuration := "[core]\n\tautocrlf = true\n\tfsmonitor = " + program + "\n" +
		"[filter \"hostile\"]\n\tclean = " + program + "\n\tsmudge = " + program + "\n" +
		"[alias]\n\tls-files = status\n"
	write(t, home, ".gitconfig", configuration)
	write(t, home, ".config/git/config", configuration)
	write(t, home, ".config/git/attributes", "* filter=hostile\n")
	return home, marker
}

// A home's git configuration never reaches what Tracked says of a checkout or what a declared key values: git lists
// the tree under toolchain's git environment, so the hostile home's programs never run and every answer is the empty
// home's. The checkout has a changed file, so git compares content where a filter or line ending rule would bend it.
// Mutant: repositoryAt's git run with the process's whole environment again.
// Not parallel: t.Setenv.
func TestAHomesGitConfigurationReachesNeitherTrackedNorAKey(t *testing.T) {
	root := checkout(t)
	write(t, root, "source/b.c", "int B;\r\n")
	answers := func() (string, map[string]bool) {
		forgetTracked()
		tracked := map[string]bool{}
		for _, name := range []string{"source/a.c", "source/b.c", "source/nested/run.sh", "flags.txt", "untracked.txt"} {
			tracked[name] = Tracked(root, filepath.Join(root, name), false)
		}
		return key(t, root, declared), tracked
	}
	write(t, root, "untracked.txt", "u\n")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	wantKey, wantTracked := answers()
	home, marker := hostileHome(t)
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	gotKey, gotTracked := answers()
	t.Cleanup(forgetTracked)
	if _, err := os.Stat(marker); err == nil {
		t.Error("git ran a program the home's configuration names")
	}
	if gotKey != wantKey {
		t.Errorf("under a hostile home the key is %s, under an empty one %s", gotKey, wantKey)
	}
	for name, want := range wantTracked {
		if gotTracked[name] != want {
			t.Errorf("under a hostile home Tracked(%s) = %v, under an empty one %v", name, gotTracked[name], want)
		}
	}
}

// A bare name in Tool is the program toolchain.Path resolves, the one a build asking toolchain runs: under
// ADAMIC_TOOLS ahead of PATH, its install prefix still read as <name>, so the report is the one PATH's copy of the same
// tool gives. In a sandbox, a tool only on PATH is reported by toolchain's error, so a product needing it fails rather
// than keying PATH's. Mutants: Tool running the name from PATH; installedAs looking the bare name up on PATH.
// Not parallel: t.Setenv.
func TestAToolIsTheProgramToolchainResolves(t *testing.T) {
	directory, elsewhere := t.TempDir(), t.TempDir()
	fake := func(bin, version string) {
		script := "#!/bin/sh\necho 'adamic-fake-tool " + version + "'\necho \"resources: " + filepath.Dir(bin) + "/lib\"\n"
		write(t, bin, "adamic-fake-tool", script)
		if err := os.Chmod(filepath.Join(bin, "adamic-fake-tool"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	fake(filepath.Join(directory, "bin"), "1.0")
	fake(filepath.Join(elsewhere, "bin"), "2.0")
	t.Setenv("PATH", filepath.Join(elsewhere, "bin")+string(filepath.ListSeparator)+os.Getenv("PATH"))
	t.Setenv("ADAMIC_SANDBOX", "")
	forget := func() { tools.Range(func(name, _ any) bool { tools.Delete(name); return true }) }
	t.Cleanup(forget)
	t.Setenv("ADAMIC_TOOLS", "")
	forget()
	if got, want := Tool("adamic-fake-tool", "--version"), "adamic-fake-tool --version: adamic-fake-tool 2.0\nresources: <adamic-fake-tool>/lib"; got != want {
		t.Errorf("with no tools directory, Tool = %q, want PATH's %q", got, want)
	}
	t.Setenv("ADAMIC_TOOLS", directory)
	forget()
	if got, want := Tool("adamic-fake-tool", "--version"), "adamic-fake-tool --version: adamic-fake-tool 1.0\nresources: <adamic-fake-tool>/lib"; got != want {
		t.Errorf("Tool = %q, want the tools directory's %q", got, want)
	}
	t.Setenv("ADAMIC_TOOLS", t.TempDir())
	t.Setenv("ADAMIC_SANDBOX", "1")
	forget()
	if got := Tool("adamic-fake-tool", "--version"); !strings.Contains(got, "ADAMIC_SANDBOX=1") || strings.Contains(got, "2.0") {
		t.Errorf("in a sandbox, a tool only on PATH reported %q", got)
	}
}
