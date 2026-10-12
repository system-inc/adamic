// Package toolchain says which program a build runs for a tool's name, and the environment it runs with, so that what a
// product's bytes depend on is what its key reads and nothing of the machine slips past it (#eehjrv6, the static census
// of what adamic reads beyond its declared inputs). It imports nothing of adamic's: every package that runs go or git
// for a key or a product asks here.
//
// # Which program
//
// Path resolves a bare name the way a box's env.sh lays its tools out (cloud/setup.sh, cloud/write-env.sh): under
// ADAMIC_TOOLS, first <tools>/bin/<name> (clang, clang++, llvm-symbolizer and node are linked or installed there), then
// <tools>/go/bin/<name> (go and gofmt, when setup installed Go itself), which is the order env.sh's PATH puts them in.
// A name not there is looked up on PATH, as before, since a developer's Mac has no tools directory and a box whose
// installed go answered setup's check has none under it either. Under ADAMIC_SANDBOX=1 that fallback is gone: a tool
// is found under ADAMIC_TOOLS or the build fails, naming it, rather than running whatever PATH holds.
//
// # Which environment
//
// The rule: a variable that can change what a build produces is either read by the key or not passed. A variable
// passed that the key doesn't read must provably not change a product's bytes. Nothing else of the caller's
// environment reaches a tool: not GIT_DIR, GIT_INDEX_FILE or GIT_WORK_TREE from a hook that runs the tests, not
// GOROOT, GODEBUG, GOFIPS140, GO111MODULE, GOARM, GO386 or the other per-architecture variables the key doesn't ask go
// about, not SDKROOT, LD_LIBRARY_PATH, DYLD_*, PKG_CONFIG, AR or CGO_FFLAGS, and none of adamic's own ADAMIC_* variables
// but the shared cache's.
//
// Env, for go (and any other tool a build runs), passes these from the caller when they are set:
//
//   - GoSettings, every variable buildcache's key reads through go env (GOOS, GOARCH, GOAMD64, GOARM64, GOEXPERIMENT,
//     GOFLAGS, GOWORK, CGO_ENABLED, CC, CXX, CGO_CFLAGS, CGO_CPPFLAGS, CGO_CXXFLAGS, CGO_LDFLAGS, GOTOOLCHAIN): passed
//     because the key asks go itself, under this same environment, what each resolves to. GOTOOLCHAIN passes as it is
//     set (auto on the cloud boxes, so a box whose go is older than go.mod's fetches the release go.mod names, and
//     local on Workshop's gate): the key names the release it selects, never the policy, so forcing local would fail
//     an auto box's build without making any key truer.
//   - Where go keeps and fetches things, which buildcache's goSetting names as locations and never values (GOCACHE,
//     GOMODCACHE, GOPATH, GOTMPDIR, GOCACHEPROG, GOPROXY, GOSUMDB, GONOSUMDB, GONOSUMCHECK, GONOPROXY, GOPRIVATE,
//     GOINSECURE, GOVCS, GOAUTH): a cached action is go's own content-addressed answer and a module is checked against
//     go.sum, which the key hashes, so neither can change bytes. GOCACHEPROG's own ADAMIC_GOCACHE_* settings pass with
//     it; cloud/gocacheprog-identity.sh proves a build through that cache is byte for byte go's own.
//   - HOME, XDG_CACHE_HOME and XDG_CONFIG_HOME: go finds its build cache and GOPATH under them when GOCACHE and GOPATH
//     are unset (it refuses to build with neither GOCACHE nor a home), and its telemetry under the configuration
//     directory, which a trace classes as go's bookkeeping (buildcache's settleFacts) only where go put it. Its
//     configuration file under them is never read (GOENV=off below), so a home holds nothing a build reads.
//   - TMPDIR: where go and the C compiler write scratch files, removed when they finish. -trimpath and cgo's own
//     prefix maps keep the name out of the bytes.
//   - PWD: how the caller spells its working directory (/var or /private/var on a Mac). A program trusts it only once
//     it has checked it names its own directory (os.Getwd), so a tool run elsewhere ignores it, and one run where the
//     caller stands spells its directory as the caller does, which a workspace file naming a module by that spelling
//     needs (go matches go.work's use lines against it). exec.Cmd sets it to the command's Dir only for a command
//     with no Env of its own, which Command's never is, so a command run in another directory goes through CommandIn
//     or GitCommandIn, which set Dir and PWD together: with the caller's PWD a C compiler records getcwd's resolved
//     spelling (/private/tmp on a Mac) as the debug info's directory, which a prefix map of the unresolved one misses
//     (native's TestSplitBuildsNameNoPlaceOfTheirInputs, from Compiler's conversion).
//   - PATH, with ADAMIC_TOOLS's bin and go/bin directories put first when they exist, as env.sh puts them: go finds
//     the C compiler a cgo build runs on PATH, and Path must find the same one, or the key's report of the compiler
//     would name another than the build ran. PATH is a selector whose selection the key reads (the tool's report).
//   - DEVELOPER_DIR: on macOS it selects the Xcode whose clang /usr/bin/clang runs. Like PATH it selects, and the
//     compiler's report the key holds is the selected one's: dropping it would build with one clang and key another.
//   - GOMAXPROCS, GOGC and GOMEMLIMIT: how hard the go command works (its -p and the compiler's -c), never what it
//     writes. Go builds the same bytes at any parallelism, which is why the key already drops GOFLAGS' -p.
//   - HTTP_PROXY, HTTPS_PROXY, NO_PROXY, ALL_PROXY (either case), SSL_CERT_FILE and SSL_CERT_DIR: how a box reaches
//     the module proxy and the toolchain download, whose content go.sum and the release's own checksum fix.
//
// and sets these, whatever the caller had:
//
//   - GOENV=off: go's configuration file (~/Library/Application Support/go/env on a Mac, ~/.config/go/env on Linux)
//     is a second source of every go setting, under the home, which go env reports merged so the key read it, but no
//     two machines share. The release's own $GOROOT/go.env (GOPROXY, GOSUMDB, GOTOOLCHAIN=auto) is still read: it is
//     part of the release the key names.
//   - ZERO_AR_DATE=1: Apple's ar stamps the time into an archive's symbol table, so a c-archive built twice a second
//     apart differs (cloud/gocacheprog-identity.sh, measured Oct 9). With it, ar writes zero, and the bytes stop
//     depending on the clock. Other archivers ignore it.
//
// GitEnv, for git, passes only PATH (as Env spells it), TMPDIR and DEVELOPER_DIR (/usr/bin/git on a Mac is Xcode's),
// and sets GIT_CONFIG_GLOBAL=/dev/null and GIT_CONFIG_NOSYSTEM=1 (no ~/.gitconfig, no /etc/gitconfig: their
// core.autocrlf, filters and aliases change what git answers), core.excludesFile and core.attributesFile to /dev/null
// (both default under the home even with no global configuration), and LC_ALL=C. No HOME: git reads nothing there it
// needs, and without it there is nothing of the home for it to read. A git answer then depends on the repository
// alone.
//
// Either takes extra settings, appended last so they win, for a caller whose request names its own environment: a
// product's request keys every one of those itself.
package toolchain

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// GoSettings are the go env variables a product's key reads (buildcache's goSettings, each as goSetting keys it), and
// Env passes each from the caller: the key asks go, under Env, what each one resolves to.
var GoSettings = []string{"GOOS", "GOARCH", "GOAMD64", "GOARM64", "GOEXPERIMENT", "GOFLAGS", "GOWORK", "CGO_ENABLED", "CC", "CXX",
	"CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOTOOLCHAIN"}

// goLocations are where go keeps and fetches things: passed, never keyed (the package doc says why).
var goLocations = []string{"GOCACHE", "GOMODCACHE", "GOPATH", "GOTMPDIR", "GOCACHEPROG", "GOPROXY", "GOSUMDB", "GONOSUMDB",
	"GONOSUMCHECK", "GONOPROXY", "GOPRIVATE", "GOINSECURE", "GOVCS", "GOAUTH"}

// carried are the rest of what Env passes from the caller (the package doc says why for each). PATH is Env's own.
var carried = []string{"HOME", "XDG_CACHE_HOME", "XDG_CONFIG_HOME", "TMPDIR", "PWD", "DEVELOPER_DIR", "GOMAXPROCS", "GOGC", "GOMEMLIMIT",
	"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "no_proxy", "all_proxy", "SSL_CERT_FILE", "SSL_CERT_DIR"}

// gocacheSettings is the prefix of the shared Go cache's own settings (cmd/adamic-gocacheprog), passed with GOCACHEPROG.
const gocacheSettings = "ADAMIC_GOCACHE_"

// fixed are the settings Env gives every tool, whatever the caller had.
var fixed = []string{"GOENV=off", "ZERO_AR_DATE=1"}

// gitFixed are the settings GitEnv gives git, whatever the caller had: no configuration but the repository's own.
var gitFixed = []string{"GIT_CONFIG_GLOBAL=" + os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
	"GIT_CONFIG_COUNT=2", "GIT_CONFIG_KEY_0=core.excludesFile", "GIT_CONFIG_VALUE_0=" + os.DevNull,
	"GIT_CONFIG_KEY_1=core.attributesFile", "GIT_CONFIG_VALUE_1=" + os.DevNull, "LC_ALL=C"}

// Sandboxed is whether this process runs in a sandbox (ADAMIC_SANDBOX=1), where every tool and location must be
// declared: nothing falls back to PATH or the home directory.
func Sandboxed() bool {
	return os.Getenv("ADAMIC_SANDBOX") == "1"
}

// directories are the tools directory's places a tool is found, in env.sh's order, those that exist.
func directories() []string {
	tools := os.Getenv("ADAMIC_TOOLS")
	if tools == "" {
		return nil
	}
	var found []string
	for _, directory := range []string{filepath.Join(tools, "bin"), filepath.Join(tools, "go", "bin")} {
		if info, err := os.Stat(directory); err == nil && info.IsDir() {
			found = append(found, directory)
		}
	}
	return found
}

// Path is the program a build runs for name: name itself when it names a path, else the first of
// <ADAMIC_TOOLS>/bin/<name> and <ADAMIC_TOOLS>/go/bin/<name> that is executable, else (outside a sandbox) name on
// PATH. Under ADAMIC_SANDBOX=1 a tool not under ADAMIC_TOOLS is an error.
func Path(name string) (string, error) {
	if strings.ContainsRune(name, '/') || strings.ContainsRune(name, filepath.Separator) {
		return name, nil
	}
	for _, directory := range directories() {
		if path, err := exec.LookPath(filepath.Join(directory, name)); err == nil {
			return path, nil
		}
	}
	if Sandboxed() {
		tools := os.Getenv("ADAMIC_TOOLS")
		if tools == "" {
			return "", fmt.Errorf("%s: ADAMIC_SANDBOX=1 runs only tools under ADAMIC_TOOLS, and ADAMIC_TOOLS is unset", name)
		}
		return "", fmt.Errorf("%s: ADAMIC_SANDBOX=1 runs only tools under ADAMIC_TOOLS, and neither %s nor %s holds it", name,
			filepath.Join(tools, "bin"), filepath.Join(tools, "go", "bin"))
	}
	return exec.LookPath(name)
}

// Command is exec.Command for the program Path resolves name to, with Env. A name Path can't resolve runs as itself,
// so the command fails as it starts, with Path's error.
func Command(name string, arguments ...string) *exec.Cmd {
	path, err := Path(name)
	if err != nil {
		command := exec.Command(name, arguments...)
		command.Err = err
		return command
	}
	command := exec.Command(path, arguments...)
	command.Env = Env()
	return command
}

// GitCommand is Command for git, with GitEnv.
func GitCommand(arguments ...string) *exec.Cmd {
	command := Command("git", arguments...)
	if command.Err == nil {
		command.Env = GitEnv()
	}
	return command
}

// CommandIn is Command run in directory: its Dir, and PWD set to directory's absolute path, last, so it wins over the
// caller's, as exec.Cmd sets it for a command with no Env.
func CommandIn(directory, name string, arguments ...string) *exec.Cmd {
	return in(directory, Command(name, arguments...))
}

// GitCommandIn is GitCommand run in directory, as CommandIn runs one.
func GitCommandIn(directory string, arguments ...string) *exec.Cmd {
	return in(directory, GitCommand(arguments...))
}

// in sets command's Dir to directory and its PWD to directory's absolute path, unless it failed to resolve.
func in(directory string, command *exec.Cmd) *exec.Cmd {
	command.Dir = directory
	if command.Err != nil {
		return command
	}
	absolute, err := filepath.Abs(directory)
	if err != nil {
		command.Err = err
		return command
	}
	command.Env = append(command.Env, "PWD="+absolute)
	return command
}

// Env is the environment a tool runs with: what the package doc lists, from the caller, the fixed settings, then extra.
func Env(extra ...string) []string {
	environment := []string{"PATH=" + searchPath()}
	for _, setting := range os.Environ() {
		name, _, _ := strings.Cut(setting, "=")
		if slices.Contains(GoSettings, name) || slices.Contains(goLocations, name) || slices.Contains(carried, name) || strings.HasPrefix(name, gocacheSettings) {
			environment = append(environment, setting)
		}
	}
	environment = append(environment, fixed...)
	return append(environment, extra...)
}

// GitEnv is the environment git runs with: PATH, TMPDIR and DEVELOPER_DIR, git's configuration out of reach, then
// extra.
func GitEnv(extra ...string) []string {
	environment := []string{"PATH=" + searchPath()}
	for _, name := range []string{"TMPDIR", "DEVELOPER_DIR"} {
		if value, ok := os.LookupEnv(name); ok {
			environment = append(environment, name+"="+value)
		}
	}
	environment = append(environment, gitFixed...)
	return append(environment, extra...)
}

// searchPath is the caller's PATH with the tools directory's places first, each once.
func searchPath() string {
	first := directories()
	path := slices.Clone(first)
	for _, entry := range filepath.SplitList(os.Getenv("PATH")) {
		if !slices.Contains(first, entry) {
			path = append(path, entry)
		}
	}
	return strings.Join(path, string(filepath.ListSeparator))
}
