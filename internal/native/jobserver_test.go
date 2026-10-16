package native

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// jobserverWith makes a jobserver FIFO holding count tokens and opens it, returning it and its path.
func jobserverWith(t *testing.T, count int) (*Jobserver, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "jobserver")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}
	// Opened for reading and writing, as native opens it, so filling it doesn't wait for a reader.
	filler, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { filler.Close() })
	if _, err := filler.Write([]byte(strings.Repeat("+", count))); err != nil {
		t.Fatal(err)
	}
	jobserver, err := OpenJobserver(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { jobserver.fifo.Close() })
	return jobserver, path
}

// With two tokens, never more than two compiles run at once however many ask, and both tokens are used (#cyew6p3).
// Mutant: the token never taken (eight at once).
func TestCompilesHoldAJobserverTokenEach(t *testing.T) {
	t.Parallel()
	jobserver, _ := jobserverWith(t, 2)
	var running, most atomic.Int32
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			err := jobserver.Hold(func() error {
				now := running.Add(1)
				for {
					seen := most.Load()
					if now <= seen || most.CompareAndSwap(seen, now) {
						break
					}
				}
				time.Sleep(20 * time.Millisecond)
				running.Add(-1)
				return nil
			})
			if err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()
	if most.Load() != 2 {
		t.Fatalf("%d compiles ran at once on two tokens", most.Load())
	}
}

// A compile that fails, or panics, gives its token back, so the next compile on a one-token jobserver still runs.
// Mutant: the token returned only on success (the next compile waits forever).
func TestAFailedOrPanickingCompileGivesItsTokenBack(t *testing.T) {
	t.Parallel()
	jobserver, _ := jobserverWith(t, 1)
	failed := errors.New("clang failed")
	if err := jobserver.Hold(func() error { return failed }); !errors.Is(err, failed) {
		t.Fatalf("the compile's error came back as %v", err)
	}
	func() {
		defer func() { recover() }()
		jobserver.Hold(func() error { panic("clang's caller panicked") })
	}()
	done := make(chan error, 1)
	go func() { done <- jobserver.Hold(func() error { return nil }) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the next compile is still waiting for the token a failed or panicking one held")
	}
}

// A nil jobserver runs a compile unbounded, as a process given none did before; one that can't be opened is an error
// naming it. Mutant: an unopenable jobserver ignored.
func TestAJobserverThatCantBeOpenedFailsNamingIt(t *testing.T) {
	t.Parallel()
	ran := false
	var none *Jobserver
	if err := none.Hold(func() error { ran = true; return nil }); err != nil || !ran {
		t.Fatalf("no jobserver: ran %v, %v", ran, err)
	}
	missing := filepath.Join(t.TempDir(), "absent")
	if jobserver, err := OpenJobserver(missing); err == nil || jobserver != nil || !strings.Contains(err.Error(), missing) {
		t.Fatalf("a missing jobserver: %v, %v", jobserver, err)
	}
}

// Native reads no environment for its jobserver (Compiler's condition 3 on #yzwm4rj): no file of internal/native, its
// tests included, names the variable, which internal/ccjobserver alone reads, and nothing in jobserver.go reads the
// environment at all. Mutant: withCompilerToken's os.Getenv of the variable back in jobserver.go.
func TestNativeReadsNoJobserverVariable(t *testing.T) {
	t.Parallel()
	// The name spelled in two pieces, so this file doesn't name it either.
	variable := "ADAMIC_CC_" + "JOBSERVER"
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatal(files, err)
	}
	for _, name := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.BasicLit:
				if node.Kind == token.STRING && strings.Contains(node.Value, variable) {
					t.Errorf("%s names %s: native reads no environment for its jobserver, internal/ccjobserver does", name, variable)
				}
			case *ast.SelectorExpr:
				if identifier, ok := node.X.(*ast.Ident); ok && name == "jobserver.go" && identifier.Name == "os" && slices.Contains([]string{"Getenv", "LookupEnv", "Environ"}, node.Sel.Name) {
					t.Errorf("jobserver.go reads the environment (os.%s)", node.Sel.Name)
				}
			}
			return true
		})
	}
}

// childVariable marks a run of this test binary as a jobserver child (inChild): the process's jobserver is set once, so
// each test that sets it runs in a process of its own.
const childVariable = "ADAMIC_NATIVE_JOBSERVER_CHILD"

// inChild runs this test binary again for the one test named, with environment added, and returns its output. A run
// that fails, or runs past bound, fails t, naming it.
func inChild(t *testing.T, test string, bound time.Duration, environment ...string) string {
	t.Helper()
	runContext, cancel := context.WithTimeout(context.Background(), bound)
	defer cancel()
	command := exec.CommandContext(runContext, os.Args[0], "-test.run", "^"+test+"$", "-test.count=1", "-test.v")
	command.Env = append(append(os.Environ(), childVariable+"=1"), environment...)
	output, err := command.CombinedOutput()
	if runContext.Err() != nil {
		t.Fatalf("%s ran past %s, waiting on a token it shouldn't need:\n%s", test, bound, output)
	}
	if err != nil || !strings.Contains(string(output), "--- PASS: "+test) {
		t.Fatalf("%s: %v\n%s", test, err, output)
	}
	return string(output)
}

// onlyAsChild skips a test meant to run in a child of its own (inChild).
func onlyAsChild(t *testing.T) {
	t.Helper()
	if os.Getenv(childVariable) == "" {
		t.Skip("runs only in a child of its own, which sets the process's jobserver")
	}
}

// The process's jobserver is set once (Compiler's addition 1 on #yzwm4rj): the same again changes nothing, and another
// is ErrJobserverSet, so nothing swaps the bound mid-run. Mutant: a second UseJobserver replacing the first.
func TestTheProcessJobserverIsSetOnce(t *testing.T) {
	t.Parallel()
	inChild(t, "TestTheProcessJobserverIsSetOnceChild", time.Minute)
}

func TestTheProcessJobserverIsSetOnceChild(t *testing.T) {
	t.Parallel()
	onlyAsChild(t)
	first, _ := jobserverWith(t, 1)
	second, _ := jobserverWith(t, 1)
	if err := UseJobserver(first, nil); err != nil {
		t.Fatal(err)
	}
	if err := UseJobserver(first, nil); err != nil {
		t.Fatalf("the same jobserver again: %v", err)
	}
	if err := UseJobserver(second, nil); !errors.Is(err, ErrJobserverSet) {
		t.Fatalf("another jobserver: %v", err)
	}
	if got, err := defaultJobserver(); got != first || err != nil {
		t.Fatalf("the process's jobserver is %v, %v, want the first", got, err)
	}
}

// A jobserver that couldn't be opened fails every compile under the process's (condition 2): Build returns why, never
// builds unbounded. Mutant: the open's error dropped by UseJobserver.
func TestAProcessJobserverThatCouldntBeOpenedFailsTheBuild(t *testing.T) {
	t.Parallel()
	inChild(t, "TestAProcessJobserverThatCouldntBeOpenedFailsTheBuildChild", time.Minute)
}

func TestAProcessJobserverThatCouldntBeOpenedFailsTheBuildChild(t *testing.T) {
	t.Parallel()
	onlyAsChild(t)
	_, failed := OpenJobserver(filepath.Join(t.TempDir(), "absent"))
	if err := UseJobserver(nil, failed); err != nil {
		t.Fatal(err)
	}
	err := Build("int main(void) { return 0; }\n", filepath.Join(t.TempDir(), "program"), Options{})
	if err == nil || !errors.Is(err, failed) && !strings.Contains(err.Error(), "absent") {
		t.Fatalf("a build under a jobserver that couldn't be opened: %v", err)
	}
}

// The archive's build never falls back to the process's jobserver (Node's rail and Compiler's addition 3 on #yzwm4rj):
// handed nil, RuntimeRequest.Build compiles unbounded, whatever the process's is, even one holding no token at all,
// which any compile that took it would wait on for ever. Mutant: compileRuntime taking the process's when handed nil.
func TestTheArchiveBuildNeverFallsBackToTheProcessJobserver(t *testing.T) {
	t.Parallel()
	inChild(t, "TestTheArchiveBuildNeverFallsBackToTheProcessJobserverChild", 3*time.Minute)
}

func TestTheArchiveBuildNeverFallsBackToTheProcessJobserverChild(t *testing.T) {
	t.Parallel()
	onlyAsChild(t)
	// The request first: its header scan runs under the process's jobserver, which is none yet.
	request, err := RuntimeRequestFor(Options{})
	if err != nil {
		t.Fatal(err)
	}
	empty, _ := jobserverWith(t, 0)
	if err := UseJobserver(empty, nil); err != nil {
		t.Fatal(err)
	}
	if err := request.Build(t.TempDir(), 4, nil); err != nil {
		t.Fatal(err)
	}
}

// The jobserver never reaches the bytes or a key (Compiler's condition 4 on #yzwm4rj, the proof lever 3's census leans
// on): the release runtime archive is built three times, each in a process of its own, under no jobserver, one of 1
// token and one of 8, with clang reached through a wrapper that records every argument it's given and the directory it
// runs in. Every file of the archive's directory, runtime.a included, is byte-identical, every clang command line is
// the same once each run's random build directory is named alike, and the request's Parts and Portable are the same.
// Mutants: a flag added to the compile when a jobserver is open (the objects and the arguments differ); the token count
// reaching an argument.
func TestTheJobserverNeverReachesTheBytes(t *testing.T) {
	t.Parallel()
	clang, err := toolchain.Path("clang")
	if err != nil {
		t.Skip("no clang: ", err)
	}
	tools := filepath.Join(t.TempDir(), "tools")
	if err := os.MkdirAll(filepath.Join(tools, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if archiver := filepath.Join(filepath.Dir(clang), "llvm-ar"); fileRuns(archiver) {
		if err := os.Symlink(archiver, filepath.Join(tools, "bin", "llvm-ar")); err != nil {
			t.Fatal(err)
		}
	}
	_, one := jobserverWith(t, 1)
	_, eight := jobserverWith(t, 8)
	type built struct{ files, commands, parts string }
	runs := map[string]built{}
	for _, jobserver := range []struct{ name, path string }{{"none", ""}, {"1 token", one}, {"8 tokens", eight}} {
		directory := t.TempDir()
		// The log's directory is written into the wrapper, since clang runs under toolchain.Env, which passes no variable
		// of ours, and each clang writes its own file in it, named by its pid, so compiles at once never interleave. The
		// wrapper stays at one path, so the compiler every run names is the same.
		log := filepath.Join(t.TempDir(), "clang")
		wrapper := "#!/bin/sh\n{ printf 'directory %s\\n' \"$PWD\"; for argument in \"$@\"; do printf 'argument %s\\n' \"$argument\"; done; printf 'end\\n'; } > " +
			strconv.Quote(log) + "/$$\nexec " + strconv.Quote(clang) + " \"$@\"\n"
		if err := os.MkdirAll(log, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tools, "bin", "clang"), []byte(wrapper), 0o755); err != nil {
			t.Fatal(err)
		}
		inChild(t, "TestTheJobserverNeverReachesTheBytesChild", 5*time.Minute, "ADAMIC_TOOLS="+tools,
			"ADAMIC_NATIVE_JOBSERVER_FIFO="+jobserver.path, "ADAMIC_NATIVE_ARCHIVE="+directory)
		files, err := os.ReadDir(directory)
		if err != nil || len(files) == 0 {
			t.Fatal(jobserver.name, files, err)
		}
		var listing strings.Builder
		for _, file := range files {
			content, err := os.ReadFile(filepath.Join(directory, file.Name()))
			if err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(&listing, "%s %x\n", file.Name(), sha256.Sum256(content))
		}
		commands, err := readCommands(log)
		if err != nil {
			t.Fatal(err)
		}
		parts, err := os.ReadFile(directory + ".parts")
		if err != nil {
			t.Fatal(err)
		}
		runs[jobserver.name] = built{listing.String(), sameDirectories(string(commands), directory), string(parts)}
	}
	if !strings.Contains(runs["none"].files, "runtime.a ") || strings.Count(runs["none"].commands, "end\n") < 2 {
		t.Fatalf("the archive or its commands weren't recorded:\n%s\n%s", runs["none"].files, runs["none"].commands)
	}
	for _, name := range []string{"1 token", "8 tokens"} {
		if runs[name].files != runs["none"].files {
			t.Errorf("under a jobserver of %s the archive's files differ:\n%s\nfrom none's:\n%s", name, runs[name].files, runs["none"].files)
		}
		if runs[name].commands != runs["none"].commands {
			t.Errorf("under a jobserver of %s clang was run otherwise:\n%s\nfrom none's:\n%s", name, runs[name].commands, runs["none"].commands)
		}
		if runs[name].parts != runs["none"].parts {
			t.Errorf("under a jobserver of %s the request's key parts differ", name)
		}
	}
}

// TestTheJobserverNeverReachesTheBytesChild builds the release archive into ADAMIC_NATIVE_ARCHIVE under the process's
// jobserver and the handed one, both the FIFO ADAMIC_NATIVE_JOBSERVER_FIFO names (none when it's empty), and writes the
// request's Parts and Portable beside it.
func TestTheJobserverNeverReachesTheBytesChild(t *testing.T) {
	t.Parallel()
	onlyAsChild(t)
	var jobserver *Jobserver
	if path := os.Getenv("ADAMIC_NATIVE_JOBSERVER_FIFO"); path != "" {
		var err error
		if jobserver, err = OpenJobserver(path); err != nil {
			t.Fatal(err)
		}
	}
	if err := UseJobserver(jobserver, nil); err != nil {
		t.Fatal(err)
	}
	request, err := RuntimeRequestFor(Options{})
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_NATIVE_ARCHIVE")
	if err := request.Build(directory, 4, jobserver); err != nil {
		t.Fatal(err)
	}
	parts := strings.Join(request.Parts, "\n") + "\n--\n" + strings.Join(request.Portable, "\n") + "\n"
	if err := os.WriteFile(directory+".parts", []byte(parts), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fileRuns says whether path is a file anyone may run.
func fileRuns(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0
}

// temporaryName is a random name a build gives its own directories (os.MkdirTemp's), which differ run to run.
var temporaryName = regexp.MustCompile(`(adamic-[a-z-]+-|\.build-|\.unit-)[0-9]+`)

// sameDirectories is commands, clang's recorded command lines, with archive (the directory the archive is built in)
// and every random temporary name spelled alike, so two runs' command lines compare. The commands are sorted, since
// the archive's units compile in whatever order they start.
func sameDirectories(commands, archive string) string {
	commands = strings.ReplaceAll(commands, archive, "<archive>")
	if resolved, err := filepath.EvalSymlinks(archive); err == nil {
		commands = strings.ReplaceAll(commands, resolved, "<archive>")
	}
	commands = temporaryName.ReplaceAllString(commands, "${1}<random>")
	lines := strings.SplitAfter(commands, "end\n")
	slices.Sort(lines)
	return strings.Join(lines, "")
}

// readCommands is every clang command line a run recorded, one file each in directory.
func readCommands(directory string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}
	var commands strings.Builder
	for _, entry := range entries {
		content, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return "", err
		}
		commands.Write(content)
	}
	return commands.String(), nil
}
