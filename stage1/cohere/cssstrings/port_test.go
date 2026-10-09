package cssstrings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const repository = "../../.."

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// lowered checks and lowers a program, failing the test with stage 0's refusal if it can't.
func lowered(t *testing.T, path string) *ir.Program {
	t.Helper()
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	result, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatalf("Lower: %v", err)
	}
	return result
}

// bounded configures a child command; Run and CombinedOutput below guard its output progress.
// Silent builds and buffered children use childguard's 30-minute FirstOutput window;
// after output starts, the default two-minute Stall allows over 3x the measured loaded gaps.
func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	command := exec.Command(name, arguments...)
	t.Cleanup(func() {
		if command.Process != nil && command.ProcessState == nil {
			_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		}
	})
	return command
}

func execute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := bounded(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := childguard.Run(command, childguard.Options{})
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

// onNode runs a program's source on Node, through the oracle's runner, with its arguments.
func onNode(t *testing.T, path string, arguments ...string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return execute(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}

// onJavaScriptBackend runs the lowered port through the JavaScript backend, on Node.
func onJavaScriptBackend(t *testing.T, program *ir.Program, arguments ...string) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path, arguments...)
}

// nativelyRun is natively's run alone.
func nativelyRun(t *testing.T, program *ir.Program, arguments ...string) run {
	t.Helper()
	result, _ := natively(t, program, arguments...)
	return result
}

// natively builds the lowered port under the address and undefined-behavior sanitizers and runs it,
// returning the binary too, for the leak check. Leak detection is off here, as in the oracle; leaks is
// its own run.
func natively(t *testing.T, program *ir.Program, arguments ...string) (run, string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "port")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return execute(t, environment, binary, arguments...), binary
}

// leaks returns a report of everything the finished port never let go of, or "": macOS's leaks tool on
// an unsanitized build, or LeakSanitizer on Linux running the sanitized binary again, as the oracle
// checks every fixture.
func leaks(t *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		binary := filepath.Join(t.TempDir(), "port")
		if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
			t.Fatal(err)
		}
		report := execute(t, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

// TestCSSStrings_NNN are top-level parallel units. ADAMIC_TEST_SHARD=i/n
// selects stable ordinals modulo n locally; unset runs every unit. The gate uses
// -run '^TestCSSStrings_NNN$'. Every original side and sanitizer/leak check stays.
// Contiguous ranges are fixed by cohere commit 7945d102a6c18dd36adf9114a758ce646e8b2359
// and the generated grammar in enumerateStrings.

func runCSSStringsShard(t *testing.T, ordinal int) {
	t.Helper()
	started := time.Now()
	defer func() {
		if elapsed := time.Since(started); elapsed > 60*time.Second {
			t.Errorf("invalid top-level unit: %s exceeds 60s", elapsed)
		}
	}()
	stringsTopSetup.once.Do(func() {
		corpus := enumerateStrings(t)
		units := stringsUnits(corpus)
		if len(units) != testCSSStringsShards {
			t.Fatalf("enumerated %d shards, declared %d", len(units), testCSSStringsShards)
		}
		verifyStringsUnion(t, corpus, units)
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		dir := stringsTopTempDir(t)
		cases := filepath.Join(dir, "cases.txt")
		write(t, cases, stringsInput(corpus.texts))
		if keep := os.Getenv("ADAMIC_CSSSTRINGS_KEEP"); keep != "" {
			write(t, keep, stringsInput(corpus.texts))
			manifest, err := json.MarshalIndent(corpus.paths, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			write(t, keep+".files.json", manifest)
		}
		bridge, _ := filepath.Abs("testdata/bridge.go")
		driver, _ := filepath.Abs("testdata/go_driver.go")
		cohere := filepath.Join(root, "cohere")
		oracleDir := stringsGoProduct(t, "Go oracle", func(dir string) error {
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
				filepath.Join(cohere, "internal/format/css/adamic_stage_one.go"): bridge,
				filepath.Join(cohere, "cmd/adamic_stage_one/main.go"):            driver,
			}})
			if err != nil {
				return err
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
				return err
			}
			command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", filepath.Join(dir, "go-printer"), filepath.Join(cohere, "cmd/adamic_stage_one/main.go"))
			command.Dir = cohere
			if output, err := childguard.CombinedOutput(command, childguard.Options{}); err != nil {
				return fmt.Errorf("Go bridge: %w\n%s", err, output)
			}
			return nil
		})
		goBinary := filepath.Join(oracleDir, "go-printer")
		main, _ := filepath.Abs("main.ts")
		build := buildStringsProgram(t, "port", main)
		sanitized := buildStringsNative(t, "sanitized", build, true)
		fast := buildStringsNative(t, "native-fast", build, false)
		// macOS's leaks tool needs the unsanitized product; Linux reuses sanitized.
		mutants := make(map[string]stringsProgram)
		mutantBinaries := make(map[string]string)
		for _, mutation := range stringsMutations {
			mutated := prepareStringsMutant(t, mutation)
			product := buildStringsProgram(t, mutation.name, mutated)
			mutants[mutation.name] = product
			mutantBinaries[mutation.name] = buildStringsNative(t, mutation.name+"-sanitized", product, true)
		}
		library := os.Getenv("ADAMIC_CSSSTRINGS_LIBRARY")
		script, _ := filepath.Abs("testdata/library.mjs")
		goIdentity := stringsOracleIdentity(t, goBinary)
		libraryIdentity := ""
		if library != "" {
			libraryIdentity = stringsOracleIdentity(t, library)
		}
		if library == "" {
			t.Log("external library not checked: set ADAMIC_CSSSTRINGS_LIBRARY")
		}
		stringsTopSetup.units = units
		stringsTopSetup.run = func(t *testing.T, unit stringsUnit) {
			t.Logf("%s: %s range [%d,%d), side %s, %d case IDs", unit.name, unit.kind, unit.lo, unit.hi, unit.side, len(unit.ids))
			start := time.Now()
			defer func() {
				if elapsed := time.Since(start); elapsed > 60*time.Second {
					t.Errorf("invalid test unit %s: wall %s exceeds 60s", unit.name, elapsed)
				}
			}()
			switch unit.kind {
			case "batch":
				path := filepath.Join(t.TempDir(), "cases.txt")
				write(t, path, stringsInput(corpus.texts[unit.lo:unit.hi]))
				want := stringsOracleAnswers(t, "Go", goBinary, nil, stringsInput(corpus.texts[unit.lo:unit.hi]), goIdentity)
				clean(t, "Go", want)
				for _, side := range []struct {
					name   string
					result run
				}{
					{"native", execute(t, stringsSanitizerEnvironment(false), sanitized, "--batch", path)},
					{"Node", onNode(t, main, "--batch", path)},
					{"JavaScript backend", onNode(t, build.javascript, "--batch", path)},
				} {
					clean(t, side.name, side.result)
					checkStringsOutput(t, unit, side.name, side.result.stdout, want.stdout)
				}
				stringsLeaks(t, sanitized, fast, "--batch", path)
				if library != "" {
					answer := stringsOracleAnswers(t, "Prettier", "node", []string{script, library}, stringsInput(corpus.texts[unit.lo:unit.hi]), libraryIdentity)
					clean(t, "Prettier", answer)
					checkStringsOutput(t, unit, "Prettier", answer.stdout, want.stdout)
				}
			case "raw":
				path := filepath.Join(t.TempDir(), "raw.css")
				text := corpus.raw[unit.lo]
				write(t, path, []byte(text))
				for _, single := range []bool{false, true} {
					args := []string{path}
					pref := "d"
					if single {
						args = append(args, "--single")
						pref = "s"
					}
					singleCase := filepath.Join(t.TempDir(), "single.txt")
					write(t, singleCase, []byte(pref+stringsEncode.Replace(text)+"\n"))
					expected := stringsOracleAnswers(t, "Go", goBinary, nil, []byte(pref+stringsEncode.Replace(text)+"\n"), goIdentity)
					clean(t, "Go raw", expected)
					decoded := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`).Replace(strings.TrimSuffix(string(expected.stdout), "\n"))
					for _, r := range []run{execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, args...), onNode(t, main, args...)} {
						clean(t, "raw", r)
						equal(t, "raw "+pref, r.stdout, []byte(decoded))
					}
				}
			case "mutant":
				want := stringsOracleAnswers(t, "Go", goBinary, nil, stringsInput(corpus.texts), goIdentity)
				clean(t, "Go", want)
				mutation := stringsMutations[unit.lo]
				for _, r := range []run{
					execute(t, stringsSanitizerEnvironment(false), mutantBinaries[mutation.name], "--batch", cases),
					onNode(t, mutants[mutation.name].main, "--batch", cases),
				} {
					clean(t, "mutant", r)
					if bytes.Equal(r.stdout, want.stdout) {
						t.Fatal("mutant survived")
					}
					a, b := strings.Split(string(r.stdout), "\n"), strings.Split(string(want.stdout), "\n")
					for i := range b {
						if i >= len(a) || a[i] != b[i] {
							t.Logf("mutant %s caught by %s holding batch case %d", mutation.name, unit.name, i)
							break
						}
					}
				}
			case "throughput":
				want := stringsOracleAnswers(t, "Go", goBinary, nil, stringsInput(corpus.texts), goIdentity)
				clean(t, "Go", want)
				side := unit.side
				command := goBinary
				args := []string{cases}
				switch side {
				case "native":
					command = fast
					args = []string{"--batch", cases}
				case "Node":
					runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
					command = "node"
					args = []string{"--disable-warning=ExperimentalWarning", runner, main, "--batch", cases}
				case "Prettier":
					if library == "" {
						t.Log("external library not checked: set ADAMIC_CSSSTRINGS_LIBRARY")
						return
					}
					command = "node"
					args = []string{script, library, cases}
				}
				var elapsed time.Duration
				for round := 0; round < 3; round++ {
					start := time.Now()
					answer := execute(t, nil, command, args...)
					elapsed += time.Since(start)
					clean(t, side, answer)
					equal(t, side, answer.stdout, want.stdout)
				}
				t.Logf("throughput %s %.1f texts/s, 3 process runs %.6fs (startup, input, escaping, output included; build excluded)", side, float64(len(corpus.texts)*2*3)/elapsed.Seconds(), elapsed.Seconds())
			default:
				t.Fatalf("unknown unit %q", unit.kind)
			}
		}
	})
	if stringsTopSetup.run == nil {
		t.Fatal("shared setup did not complete")
	}
	selected := selectedStringsUnits(t, stringsTopSetup.units)
	for _, unit := range selected {
		if unit.name == stringsTopSetup.units[ordinal].name {
			stringsTopSetup.run(t, unit)
			return
		}
	}
	t.Skip("excluded by ADAMIC_TEST_SHARD")
}
func write(t *testing.T, path string, b []byte) {
	t.Helper()
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
}
func clean(t *testing.T, name string, r run) {
	t.Helper()
	if r.exitCode != 0 || len(r.stderr) > 0 {
		t.Fatalf("%s exit %d stderr %s", name, r.exitCode, r.stderr)
	}
}
func equal(t *testing.T, name string, a, b []byte) {
	t.Helper()
	if !bytes.Equal(a, b) {
		i := 0
		for i < len(a) && i < len(b) && a[i] == b[i] {
			i++
		}
		t.Fatalf("%s first byte difference at %d (lengths %d/%d)", name, i, len(a), len(b))
	}
}
