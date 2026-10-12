package buildcache

import (
	"bytes"
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

// A GoBuild product's key and bytes are the same whatever the caller's environment holds beyond what the key reads:
// a go configuration file (GOENV) setting GOFLAGS and the architecture level, which reach the bytes, an ADAMIC_JUNK, and
// a GIT_DIR naming nothing, which broke the key's go list before (its build stamping asked git). Mutants: the build's go
// run with the whole environment (the bytes move); the key's go list or go env run with it (the key fails or moves).
// Not parallel: t.Setenv, cached.
func TestAGoProductIsTheSameWhateverElseTheEnvironmentHolds(t *testing.T) {
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	// No GOFLAGS of the caller's, so the configuration file's would be go's, and go list would stamp a build through git.
	t.Setenv("GOFLAGS", "")
	build := func() (string, []byte) {
		t.Helper()
		cached(t)
		forgetInputs()
		inputs, err := goRequest{verb: "build", output: "hello", pkg: "./internal/buildcache/testdata/hello", arguments: reproducible(nil)}.inputs()
		if err != nil {
			t.Fatal(err)
		}
		content, err := os.ReadFile(GoBuild(t, "hello", "./internal/buildcache/testdata/hello", nil))
		if err != nil {
			t.Fatal(err)
		}
		return key(t, root, inputs), content
	}
	wantKey, wantBytes := build()
	configuration := filepath.Join(t.TempDir(), "env")
	write(t, filepath.Dir(configuration), "env", "GOFLAGS=-tags=junk\nGOARM64=v8.1\nGOAMD64=v3\n")
	t.Setenv("GOENV", configuration)
	t.Setenv("ADAMIC_JUNK", t.TempDir())
	t.Setenv("GIT_DIR", filepath.Join(t.TempDir(), "nothing"))
	t.Cleanup(forgetInputs)
	gotKey, gotBytes := build()
	if gotKey != wantKey {
		t.Errorf("with junk in the environment the key is %s, without %s", gotKey, wantKey)
	}
	if !bytes.Equal(gotBytes, wantBytes) {
		t.Errorf("with junk in the environment the product is %d other bytes", len(gotBytes))
	}
}

// Outside a sandbox the tools directory and the product cache default under the home, as before; in one
// (ADAMIC_SANDBOX=1) each is its variable or an error naming it, so a tool entry fails to value rather than resolve
// under the home, and no product is written there. Mutants: toolsRoot's home default in a sandbox; cacheLocation's.
// Not parallel: t.Setenv.
func TestASandboxDeclaresWhereToolsAndProductsLive(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("ADAMIC_TOOLS", "")
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", "")
	t.Setenv("ADAMIC_SANDBOX", "")
	user, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if tools, err := toolsRoot(); err != nil || tools != filepath.Join(home, "adamic-tools") {
		t.Errorf("outside a sandbox, toolsRoot = %q, %v", tools, err)
	}
	if cache, err := cacheLocation(); err != nil || cache != filepath.Join(user, "adamic-build") {
		t.Errorf("outside a sandbox, cacheLocation = %q, %v", cache, err)
	}
	t.Setenv("ADAMIC_SANDBOX", "1")
	if tools, err := toolsRoot(); err == nil || !strings.Contains(err.Error(), "ADAMIC_TOOLS is unset") {
		t.Errorf("in a sandbox with no ADAMIC_TOOLS, toolsRoot = %q, %v", tools, err)
	}
	if value, err := entryValue(t.TempDir(), t.TempDir(), readEntry{Kind: "tool", Path: "bin/adamic-fake-tool"}, 0, nil); err == nil {
		t.Errorf("in a sandbox with no ADAMIC_TOOLS, a tool entry valued %q", value)
	}
	if cache, err := cacheLocation(); err == nil || !strings.Contains(err.Error(), "ADAMIC_BUILD_CACHE_DIR is unset") {
		t.Errorf("in a sandbox with no ADAMIC_BUILD_CACHE_DIR, cacheLocation = %q, %v", cache, err)
	}
	tools, cache := t.TempDir(), t.TempDir()
	t.Setenv("ADAMIC_TOOLS", tools)
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", cache)
	if got, err := toolsRoot(); err != nil || got != tools {
		t.Errorf("in a sandbox, toolsRoot = %q, %v; want %q", got, err, tools)
	}
	if got, err := cacheLocation(); err != nil || got != cache {
		t.Errorf("in a sandbox, cacheLocation = %q, %v; want %q", got, err, cache)
	}
}

// A tool reports under toolchain.Env, the environment the build it keys runs it with, never the caller's whole one: a
// variable the build doesn't see can't change the report (Tests, from Lint's review of Compiler's rows: native keyed
// clang --version under the caller's environment, and refusalrewrite a go env file's GOTOOLCHAIN the build never saw).
// Mutant: Tool under the caller's environment (the report names the caller's ADAMIC_FAKE_SETTING).
// Not parallel: t.Setenv.
func TestAToolReportsUnderTheEnvironmentItsBuildRunsIn(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "bin")
	write(t, bin, "adamic-fake-tool", "#!/bin/sh\necho \"adamic-fake-tool, setting ${ADAMIC_FAKE_SETTING:-unset}\"\n")
	if err := os.Chmod(filepath.Join(bin, "adamic-fake-tool"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_TOOLS", filepath.Dir(bin))
	t.Setenv("ADAMIC_SANDBOX", "")
	t.Setenv("ADAMIC_FAKE_SETTING", "the caller's")
	forget := func() { tools.Range(func(name, _ any) bool { tools.Delete(name); return true }) }
	forget()
	t.Cleanup(forget)
	if got, want := Tool("adamic-fake-tool", "--version"), "adamic-fake-tool --version: adamic-fake-tool, setting unset"; got != want {
		t.Errorf("Tool = %q, want %q", got, want)
	}
}
