package oracle

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// jobserverFixtures is how many of the oracle's fixtures that lower the jobserver proof builds: enough to compile
// programs of several shapes and their runtime archives, few enough to keep the proof under a minute.
const jobserverFixtures = 4

// The jobserver never reaches the emitted C or a clang argument (Compiler's condition 4 on #yzwm4rj): oracle fixtures
// are lowered, emitted and built three times at once, each run a process of its own with its own build cache and its
// own tools directory, under no jobserver, one of 1 token and one of 8, named in ADAMIC_CC_JOBSERVER as Loom names it,
// so ccjobserver's init opens it. The emitted C, every binary, and every clang command line (clang reached through a
// wrapper that records each one, with each run's directories named alike) are the same in all three. The runs go at
// once, compared after all three end, since one after another they took about 31 s of a unit's 30 (#s3dce49,
// Correctness's measure). The runtime archives the programs link come from runtimeproducts' supplier, so the archive's
// build runs under the handed jobserver and every other compile under the process's. Mutants: a flag added to a
// compile when a jobserver is open; the token count reaching an argument.
func TestTheJobserverNeverReachesTheEmittedCOrClang(test *testing.T) {
	test.Parallel()
	clang, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Skip("no clang: ", pathError)
	}
	var mutex sync.Mutex
	runs := map[string][2]string{}
	// Serial: a group, which returns only once its three parallel runs have ended, so the comparison below reads them all.
	test.Run("runs", func(test *testing.T) {
		for _, jobserver := range []struct {
			name   string
			tokens int
		}{{"none", 0}, {"1 token", 1}, {"8 tokens", 8}} {
			test.Run(jobserver.name, func(test *testing.T) {
				test.Parallel()
				report, commands := jobserverRun(test, clang, jobserver.tokens)
				mutex.Lock()
				defer mutex.Unlock()
				runs[jobserver.name] = [2]string{report, commands}
			})
		}
	})
	if test.Failed() {
		return
	}
	if strings.Count(runs["none"][0], "\n") != 2*jobserverFixtures ||
		strings.Count(runs["none"][1], "end\n") < jobserverFixtures {
		test.Fatalf(
			"the runs recorded too little:\n%s\n%d commands",
			runs["none"][0],
			strings.Count(runs["none"][1], "end\n"),
		)
	}
	for _, name := range []string{"1 token", "8 tokens"} {
		if runs[name][0] != runs["none"][0] {
			test.Errorf(
				"under a jobserver of %s the emitted C or the binaries differ:\n%s\nfrom none's:\n%s",
				name,
				runs[name][0],
				runs["none"][0],
			)
		}
		if runs[name][1] != runs["none"][1] {
			test.Errorf(
				"under a jobserver of %s clang was run otherwise:\n%s",
				name,
				firstCommandMissing(runs[name][1], runs["none"][1]),
			)
		}
	}
}

// jobserverRun builds the child's fixtures in a process of its own under a jobserver of tokens (none at 0), and returns
// its report and every clang command line it ran, its directories spelled alike. Its tools directory is its own, a
// clang wrapper writing to its own log, so runs at once never share a wrapper; the directory is spelled alike too.
func jobserverRun(test *testing.T, clang string, tokens int) (report, commands string) {
	test.Helper()
	tools := filepath.Join(test.TempDir(), "tools")
	if createDirectoryError := os.MkdirAll(filepath.Join(tools, "bin"), 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if archiver := filepath.Join(filepath.Dir(clang), "llvm-ar"); fileRunnable(archiver) {
		if symlinkError := os.Symlink(archiver, filepath.Join(tools, "bin", "llvm-ar")); symlinkError != nil {
			test.Fatal(symlinkError)
		}
	}
	fifo := ""
	if tokens > 0 {
		fifo = tokenFIFO(test, tokens)
	}
	out, cache, log := test.TempDir(), test.TempDir(), filepath.Join(test.TempDir(), "clang")
	// The log's directory is written into the wrapper, since clang runs under toolchain.Env, which passes no variable
	// of ours, and each clang writes its own file in it, named by its pid, so compiles at once never interleave.
	wrapper := "#!/bin/sh\n{ printf 'directory %s\\n' \"$PWD\"; for argument in \"$@\"; do printf 'argument %s\\n' \"$argument\"; done; printf 'end\\n'; } > " +
		strconv.Quote(log) + "/$$\nexec " + strconv.Quote(clang) + " \"$@\"\n"
	if createDirectoryError := os.MkdirAll(log, 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if writeError := os.WriteFile(filepath.Join(tools, "bin", "clang"), []byte(wrapper), 0o755); writeError != nil {
		test.Fatal(writeError)
	}
	runContext, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	command := exec.CommandContext(
		runContext,
		os.Args[0],
		"-test.run",
		"^TestTheJobserverNeverReachesTheEmittedCOrClangChild$",
		"-test.count=1",
		"-test.v",
	)
	// Each run builds what it links, the runtime archive among them, into its own empty cache, so the cache builds:
	// a Loom unit's ADAMIC_BUILD_CACHE=read, which only reads what the tree's product tests built, is cleared (the
	// last value of a variable is the one a child sees). Under it, chg_j1h9dkkk's unit refused the archive as not
	// built (#yzwm4rj).
	command.Env = append(
		os.Environ(),
		"ADAMIC_ORACLE_JOBSERVER_CHILD="+out,
		ccjobserver.Variable+"="+fifo,
		"ADAMIC_TOOLS="+tools,
		"ADAMIC_BUILD_CACHE_DIR="+cache,
		"ADAMIC_BUILD_STORE=off",
		"ADAMIC_BUILD_CACHE=",
		location.OutputDirectoryVariable+"=",
	)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil ||
		!strings.Contains(string(output), "--- PASS: TestTheJobserverNeverReachesTheEmittedCOrClangChild") {
		test.Fatalf("the run under a jobserver of %d tokens: %v\n%s", tokens, combinedOutputError, output)
	}
	written, readError := os.ReadFile(filepath.Join(out, "report"))
	if readError != nil {
		test.Fatal(readError)
	}
	recorded, readCommandsError := readCommands(log)
	if readCommandsError != nil {
		test.Fatal(readCommandsError)
	}
	return string(
		written,
	), alikeCommands(recorded, []runDirectory{{out, "<out>"}, {cache, "<cache>"}, {tools, "<tools>"}})
}

// TestTheJobserverNeverReachesTheEmittedCOrClangChild lowers, emits and builds the first jobserverFixtures fixtures
// that lower, writing each one's emitted C's sha256 and its binary's to ADAMIC_ORACLE_JOBSERVER_CHILD's report.
func TestTheJobserverNeverReachesTheEmittedCOrClangChild(test *testing.T) {
	test.Parallel()
	out := os.Getenv("ADAMIC_ORACLE_JOBSERVER_CHILD")
	// Only a child of its own, whose jobserver ccjobserver's init opened, builds; the parent's run returns at once, as
	// the house's child tests do, never a skip the census would refuse (chg_zkj3gd91's red).
	if out == "" {
		return
	}
	var report strings.Builder
	built := 0
	for _, fixture := range fixtures {
		if built == jobserverFixtures {
			break
		}
		if !fixture.lowers || refusedAdamicNonNullFixture(fixture.path) {
			continue
		}
		path, absolutePathError := filepath.Abs(filepath.Join(repository, fixture.path))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		program, loweredError := lowered(test, path)
		if loweredError != nil {
			test.Fatal(fixture.path, loweredError)
		}
		code := native.C(program)
		binary := filepath.Join(out, fmt.Sprintf("program-%d", built))
		if buildError := native.Build(code, binary, native.Options{}); buildError != nil {
			test.Fatal(fixture.path, buildError)
		}
		bytes, readError := os.ReadFile(binary)
		if readError != nil {
			test.Fatal(readError)
		}
		fmt.Fprintf(
			&report,
			"%s C %x\n%s binary %x\n",
			fixture.path,
			sha256.Sum256([]byte(code)),
			fixture.path,
			sha256.Sum256(bytes),
		)
		built++
	}
	if writeError := os.WriteFile(filepath.Join(out, "report"), []byte(report.String()), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
}

// tokenFIFO makes a jobserver FIFO holding count tokens, held open for the test so its tokens stay, and returns its
// path.
func tokenFIFO(test *testing.T, count int) string {
	test.Helper()
	path := filepath.Join(test.TempDir(), "jobserver")
	if mkfifoError := syscall.Mkfifo(path, 0o600); mkfifoError != nil {
		test.Fatal(mkfifoError)
	}
	filler, openError := os.OpenFile(path, os.O_RDWR, 0)
	if openError != nil {
		test.Fatal(openError)
	}
	test.Cleanup(func() { filler.Close() })
	if _, writeError := filler.Write([]byte(strings.Repeat("+", count))); writeError != nil {
		test.Fatal(writeError)
	}
	return path
}

// fileRunnable says whether path is a file anyone may run.
func fileRunnable(path string) bool {
	info, statError := os.Stat(path)
	return statError == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0
}

// randomDirectory is a random name a build gives its own directories (os.MkdirTemp's, buildcache's .building- ones),
// which differ run to run.
var randomDirectory = regexp.MustCompile(`(adamic-[a-z-]+-|\.build-|\.unit-|\.publish-|\.building-[0-9a-f]+-)[0-9]+`)

// A runDirectory is a directory of one run's own, and the name it's spelled by in that run's commands (alikeCommands).
type runDirectory struct{ path, name string }

// alikeCommands is commands, clang's recorded command lines, with a run's own directories and every random directory
// name spelled alike, sorted, since compiles start in whatever order they're scheduled. The longest path is spelled
// first, so a directory whose path leads another's never takes part of it.
func alikeCommands(commands string, directories []runDirectory) string {
	directories = slices.Clone(directories)
	slices.SortFunc(directories, func(left, right runDirectory) int { return len(right.path) - len(left.path) })
	for _, directory := range directories {
		commands = strings.ReplaceAll(commands, directory.path, directory.name)
		if resolved, resolveSymlinksError := filepath.EvalSymlinks(directory.path); resolveSymlinksError == nil {
			commands = strings.ReplaceAll(commands, resolved, directory.name)
		}
	}
	commands = randomDirectory.ReplaceAllString(commands, "${1}<random>")
	lines := strings.SplitAfter(commands, "end\n")
	slices.Sort(lines)
	return strings.Join(lines, "")
}

// firstCommandMissing is the first command line in got that want doesn't hold, for a failure's message.
func firstCommandMissing(got, want string) string {
	held := map[string]bool{}
	for _, line := range strings.SplitAfter(want, "end\n") {
		held[line] = true
	}
	for _, line := range strings.SplitAfter(got, "end\n") {
		if !held[line] {
			return line
		}
	}
	return "(the same lines, other counts)"
}

// readCommands is every clang command line a run recorded, one file each in directory.
func readCommands(directory string) (string, error) {
	entries, readDirectoryError := os.ReadDir(directory)
	if readDirectoryError != nil {
		return "", readDirectoryError
	}
	var commands strings.Builder
	for _, entry := range entries {
		content, readError := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readError != nil {
			return "", readError
		}
		commands.Write(content)
	}
	return commands.String(), nil
}
