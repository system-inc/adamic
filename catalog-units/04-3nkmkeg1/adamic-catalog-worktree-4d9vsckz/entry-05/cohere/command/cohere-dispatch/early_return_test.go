//go:build !windows

package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"syscall"
	"testing"
	"time"
)

// A run that returns early gives its caller exactly what a run that waits gives: every byte of stdout and
// stderr and the exit code, on a clean tree, a tree with findings, one with a type error, and a run whose
// fixes rewrite a file. Only what the caller does not wait for moves: the engine's cache write, which this
// also waits for and reads back (#zqsdzbq, early return).
func TestReturningEarlyChangesNothingTheCallerSees(t *testing.T) {
	t.Parallel()
	dispatcher, engine := buildDispatcherAndEngine(t)
	scenarios := []struct {
		name      string
		files     map[string]string
		arguments []string
	}{
		{"clean", map[string]string{"Clean.ts": "export const clean = 1;\n"}, []string{"--no-fix"}},
		{"findings", map[string]string{"Producer.ts": "export function value(): number {\n    debugger;\n    return 1;\n}\n"}, []string{"--no-fix"}},
		{"a type error", map[string]string{"Broken.ts": "export const broken: number = \"text\";\n"}, []string{"--no-fix"}},
		{"a rewrite", map[string]string{"Producer.ts": "export function value(): number {\n    debugger;\n    return 1;\n}\n"}, nil},
	}
	durations := regexp.MustCompile(`\d+(\.\d+)?(ms|s|µs| GB)\b`)
	clock := regexp.MustCompile(`\d\d:\d\d:\d\d`)
	invocationLines := regexp.MustCompile(`(?m)^ *(total|memory:) .*\n`)

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			type outcome struct {
				stdout, stderr string
				code           int
				root           string
			}
			run := func(wait bool) outcome {
				root := writeEarlyReturnProject(t, scenario.files)
				command := earlyReturnCommand(dispatcher, engine, root, scenario.arguments...)
				if wait {
					command.Env = append(command.Env, "COHERE_WAIT=1")
				}
				var stdout, stderr bytes.Buffer
				command.Stdout, command.Stderr = &stdout, &stderr
				code := 0
				if err := command.Run(); err != nil {
					exitError, isExit := err.(*exec.ExitError)
					if !isExit {
						t.Fatalf("running the dispatcher: %v", err)
					}
					code = exitError.ExitCode()
				}
				normalize := func(text string) string {
					text = invocationLines.ReplaceAllString(text, "")
					text = bytes.NewBufferString(text).String()
					resolved, _ := filepath.EvalSymlinks(root)
					text = regexp.MustCompile(regexp.QuoteMeta(resolved)).ReplaceAllString(text, "ROOT")
					text = regexp.MustCompile(regexp.QuoteMeta(root)).ReplaceAllString(text, "ROOT")
					return clock.ReplaceAllString(durations.ReplaceAllString(text, "D"), "T")
				}
				return outcome{normalize(stdout.String()), normalize(stderr.String()), code, root}
			}

			waited := run(true)
			early := run(false)
			if early.stdout != waited.stdout || early.stderr != waited.stderr || early.code != waited.code {
				t.Fatalf("returning early changed what the caller sees:\n--- waited (exit %d)\n%s\n%s\n--- early (exit %d)\n%s\n%s",
					waited.code, waited.stdout, waited.stderr, early.code, early.stdout, early.stderr)
			}

			// The engine finishes alone after an early return, and what it finishes is the cache table: a run that
			// rewrote files records no run of its own, and still writes the sections it produced.
			deadline := time.Now().Add(30 * time.Second)
			for {
				if files, _ := filepath.Glob(filepath.Join(early.root, ".cache", "cohere", "*.gob")); len(files) > 0 {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("the engine that returned early never wrote its cache table")
				}
				time.Sleep(20 * time.Millisecond)
			}
			if _, err := os.Stat(filepath.Join(early.root, ".cache", "cohere", "notes-after-return.txt")); err == nil {
				t.Error("the engine left a note after returning, so something it did in the background failed")
			}
		})
	}
}

// buildDispatcherAndEngine builds this dispatcher and the engine it runs, from the working tree.
func buildDispatcherAndEngine(t *testing.T) (dispatcher string, engine string) {
	binaries := t.TempDir()
	build := func(output string, packagePath string) string {
		binary := filepath.Join(binaries, output)
		if combined, err := exec.Command("go", "build", "-o", binary, packagePath).CombinedOutput(); err != nil {
			t.Fatalf("building %s: %v\n%s", packagePath, err, combined)
		}
		return binary
	}
	return build("cohere-dispatch", "."), build("cohere", "../cohere")
}

// writeEarlyReturnProject writes a small project with one lint rule on, plus the given files, and returns
// its root.
//
// The settings are written formatted, as the formatter leaves them: a bare run formats (#b1sjy7b), so a
// fixture whose own JSON the formatter would rewrite turns every clean run into one that exits 1 on it.
func writeEarlyReturnProject(t *testing.T, extra map[string]string) string {
	root := t.TempDir()
	files := map[string]string{
		"tsconfig.json": "{\n" +
			"  \"compilerOptions\": {\n" +
			"    \"target\": \"ES2022\",\n" +
			"    \"module\": \"esnext\",\n" +
			"    \"moduleResolution\": \"bundler\",\n" +
			"    \"strict\": true,\n" +
			"    \"noEmit\": true\n" +
			"  },\n" +
			"  \"include\": [\"**/*.ts\"]\n" +
			"}\n",
		"CohereSettings.json":      "{ \"extends\": \"./NexusCohereSettings.json\", \"rules\": { \"no-debugger\": \"error\" } }\n",
		"NexusCohereSettings.json": "{ \"format\": {} }\n",
	}
	for name, contents := range extra {
		files[name] = contents
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// earlyReturnCommand runs the dispatcher on root the way a terminal would, returning early unless the
// caller adds COHERE_WAIT.
func earlyReturnCommand(dispatcher string, engine string, root string, arguments ...string) *exec.Cmd {
	command := exec.Command(dispatcher, arguments...)
	command.Dir = root
	command.Env = append(os.Environ(), "COHERE_BINARY="+engine, "CI=", "COHERE_WAIT=")
	return command
}

// A profiled run does not return early. The profile is something the caller reads, so the moment the
// dispatcher returns it must be whole: a gzip stream that reads to its end. Stopped after the verdict, it
// could be read half-written, or empty, with a failure to close it swallowed.
func TestAProfiledRunIsWholeWhenTheCallerHasItsAnswer(t *testing.T) {
	t.Parallel()
	dispatcher, engine := buildDispatcherAndEngine(t)
	for round := range 5 {
		root := writeEarlyReturnProject(t, map[string]string{"Clean.ts": "export const clean = 1;\n"})
		profilePath := filepath.Join(root, "cpu.pprof")
		// Output to a file rather than a buffer, as a terminal takes it: a buffer is fed through a pipe the
		// engine holds open, so the test would wait for the engine to exit and see a whole profile whenever
		// it was written.
		output, err := os.Create(filepath.Join(t.TempDir(), "output"))
		if err != nil {
			t.Fatal(err)
		}
		command := earlyReturnCommand(dispatcher, engine, root, "--no-fix", "--profile", profilePath)
		command.Stdout, command.Stderr = output, output
		err = command.Run()
		output.Close()
		if err != nil {
			written, _ := os.ReadFile(output.Name())
			t.Fatalf("round %d: the profiled run failed: %v\n%s", round, err, written)
		}
		file, err := os.Open(profilePath)
		if err != nil {
			t.Fatalf("round %d: no profile when the caller had its answer: %v", round, err)
		}
		reader, err := gzip.NewReader(file)
		if err != nil {
			file.Close()
			t.Fatalf("round %d: the profile is not a whole gzip stream when the caller had its answer: %v", round, err)
		}
		read, err := io.Copy(io.Discard, reader)
		file.Close()
		if err != nil || read == 0 {
			t.Fatalf("round %d: the profile read %d bytes and then %v when the caller had its answer", round, read, err)
		}
	}
}

// writeRepositoryProject writes a project that records its runs: a repository with its cache ignored.
func writeRepositoryProject(t *testing.T) string {
	root := writeEarlyReturnProject(t, map[string]string{"Clean.ts": "export const clean = 1;\n", ".gitignore": ".cache/\n"})
	for _, arguments := range [][]string{{"init", "--quiet"}, {"add", "-A"}, {"-c", "user.email=fixture@example.com", "-c", "user.name=fixture", "commit", "--quiet", "-m", "initial"}} {
		git := exec.Command("git", arguments...)
		git.Dir = root
		if output, err := git.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", arguments, err, output)
		}
	}
	return root
}

// A run holds off reading the table while another run holds the table's lock, and goes ahead the moment
// the lock is released. `s c` runs its fix step and then its check back to back, and the first of a pair
// like that writes the table after its caller has its answer, so without the wait the second would read
// around the write and miss what it recorded. The test stands in for a writer still finishing in the
// background, which makes the wait certain rather than a race the writer usually wins on a fixture this
// small.
//
// Whether the second run then replays is deliberately not asserted: the run cache records the project's
// ancestor directories, and the shared temporary directory above a test's moves under other processes.
func TestARunWaitsForAWriterStillHoldingTheTable(t *testing.T) {
	t.Parallel()
	dispatcher, engine := buildDispatcherAndEngine(t)
	root := writeRepositoryProject(t)
	if output, err := earlyReturnCommand(dispatcher, engine, root, "--no-fix").CombinedOutput(); err != nil {
		t.Fatalf("the first run failed: %v\n%s", err, output)
	}

	// The first run's engine may still be finishing; taking the lock waits for it.
	lock, err := os.OpenFile(filepath.Join(root, ".cache", "cohere", "table.lock"), os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("the first run left no table lock: %v", err)
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatal(err)
	}

	second := earlyReturnCommand(dispatcher, engine, root, "--no-fix")
	var output bytes.Buffer
	second.Stdout, second.Stderr = &output, &output
	if err := second.Start(); err != nil {
		t.Fatal(err)
	}
	finished := make(chan error, 1)
	go func() { finished <- second.Wait() }()
	select {
	case err := <-finished:
		t.Fatalf("the second run finished while the table's lock was held (%v):\n%s", err, output.String())
	case <-time.After(1500 * time.Millisecond):
	}
	syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	select {
	case err := <-finished:
		if err != nil {
			t.Fatalf("the second run failed: %v\n%s", err, output.String())
		}
	case <-time.After(30 * time.Second):
		t.Fatal("the second run did not finish after the lock was released")
	}
}
