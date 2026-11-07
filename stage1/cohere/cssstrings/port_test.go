package cssstrings

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/stage1/cohere/cssstrings/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const repository = "../../.."

const testCSSStringsShards = 291

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// bounded configures a child command; Run and CombinedOutput below guard its output progress.
// Silent builds and buffered children use childguard's 30-minute FirstOutput window;
// after output starts, the default two-minute Stall allows over 3x the measured loaded gaps.
func bounded(test *testing.T, name string, arguments ...string) *exec.Cmd {
	test.Helper()
	command := exec.Command(name, arguments...)
	test.Cleanup(func() {
		if command.Process != nil && command.ProcessState == nil {
			_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		}
	})
	return command
}

func execute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	command := bounded(test, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runError := childguard.Run(command, childguard.Options{})
	var exitError *exec.ExitError
	if runError != nil && !errors.As(runError, &exitError) {
		test.Fatalf("running %s: %v", name, runError)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

// onNode runs a program's source on Node, through the oracle's runner, with its arguments.
func onNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return execute(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}

// stringsCorpus preserves the original enumeration, including duplicate texts as distinct cases.
type stringsCorpus struct {
	texts []string
	paths []string
	raw   []string
}

func enumerateStrings(test *testing.T) stringsCorpus {
	c := oracles.Enumerate(test)
	return stringsCorpus{texts: c.Texts, paths: c.Paths, raw: c.Raw}
}

// TestCSSStrings builds each input once before the parallel units. ADAMIC_TEST_SHARD=i/n
// (zero-based i) selects units whose stable ordinal modulo n equals i; unset runs all.
// Stable shard-NNN names use the same zero-based ordinals for direct -run selection.
// No corpus is sampled: batch ranges, raw files, mutants and throughput sides form the census.
// TestCSSStrings_NNN are top-level parallel units. ADAMIC_TEST_SHARD=i/n
// selects stable ordinals modulo n locally; unset runs every unit. The gate uses
// -run '^TestCSSStrings_NNN$'. Every original side and sanitizer/leak check stays.
// Contiguous ranges are fixed by cohere commit 0183ccf8c57a775f874ca73be1adfc06e9ad1562
// and the generated grammar in enumerateStrings.

func prepareCSSStringsSetup(test *testing.T) {
	test.Helper()
	stringsTopSetup.once.Do(func() {
		corpus := enumerateStrings(test)
		units := stringsUnits(corpus)
		if len(units) != testCSSStringsShards {
			test.Fatalf("enumerated %d shards, declared %d", len(units), testCSSStringsShards)
		}
		verifyStringsUnion(test, corpus, units)
		dir := stringsTopTempDir(test)
		cases := filepath.Join(dir, "cases.txt")
		write(test, cases, stringsInput(corpus.texts))
		if keep := os.Getenv("ADAMIC_CSSSTRINGS_KEEP"); keep != "" {
			write(test, keep, stringsInput(corpus.texts))
			manifest, encodeError := json.MarshalIndent(corpus.paths, "", "  ")
			if encodeError != nil {
				test.Fatal(encodeError)
			}
			write(test, keep+".files.json", manifest)
		}
		goBinary := stringsPreparedOracle(test)
		main, _ := filepath.Abs("main.ts")
		build := stringsPreparedProgram(test, -1)
		sanitized := stringsPreparedNative(test, -1, true)
		fast := stringsPreparedNative(test, -1, false)
		// macOS's leaks tool needs the unsanitized product; Linux reuses sanitized.
		mutants := make(map[string]stringsProgram)
		mutantBinaries := make(map[string]string)
		for index, mutation := range stringsMutations {
			product := stringsPreparedProgram(test, index)
			mutants[mutation.name] = product
			mutantBinaries[mutation.name] = stringsPreparedNative(test, index, true)
		}
		library := os.Getenv("ADAMIC_CSSSTRINGS_LIBRARY")
		script, _ := filepath.Abs("testdata/library.mjs")
		goIdentity := stringsOracleIdentity(test, goBinary)
		libraryIdentity := ""
		if library != "" {
			libraryIdentity = stringsOracleIdentity(test, library)
		}
		if library == "" {
			test.Log("external library not checked: set ADAMIC_CSSSTRINGS_LIBRARY")
		}
		fullWant := stringsSetupOracleAnswers(test, "Go", goBinary, nil, stringsInput(corpus.texts), goIdentity)
		clean(test, "Go", fullWant)
		stringsTopSetup.units = units
		stringsTopSetup.run = func(test *testing.T, unit stringsUnit) {
			test.Logf(
				"%s: %s range [%d,%d), side %s, %d case IDs",
				unit.name,
				unit.kind,
				unit.lo,
				unit.hi,
				unit.side,
				len(unit.ids),
			)
			start := time.Now()
			defer func() {
				if elapsed := time.Since(start); elapsed > 60*time.Second {
					test.Errorf("invalid test unit %s: wall %s exceeds 60s", unit.name, elapsed)
				}
			}()
			switch unit.kind {
			case "batch":
				path := filepath.Join(test.TempDir(), "cases.txt")
				write(test, path, stringsInput(corpus.texts[unit.lo:unit.hi]))
				want := stringsBatchGoAnswers(test, goBinary, goIdentity, corpus, unit)
				clean(test, "Go", want)
				for _, side := range []struct {
					name   string
					result run
				}{
					{"native", stringsExecute(test, stringsSanitizerEnvironment(false), sanitized, "--batch", path)},
					{"Node", stringsOnNode(test, main, "--batch", path)},
					{"JavaScript backend", stringsOnNode(test, build.javascript, "--batch", path)},
				} {
					clean(test, side.name, side.result)
					checkStringsOutput(test, unit, side.name, side.result.stdout, want.stdout)
				}
				stringsLeaks(test, sanitized, fast, "--batch", path)
				if library != "" {
					answer := stringsBatchPrettierAnswers(test, library, libraryIdentity, corpus, unit)
					clean(test, "Prettier", answer)
					checkStringsOutput(test, unit, "Prettier", answer.stdout, want.stdout)
				}
			case "raw":
				path := filepath.Join(test.TempDir(), "raw.css")
				text := corpus.raw[unit.lo]
				write(test, path, []byte(text))
				for _, single := range []bool{false, true} {
					args := []string{path}
					pref := "d"
					if single {
						args = append(args, "--single")
						pref = "s"
					}
					singleCase := filepath.Join(test.TempDir(), "single.txt")
					write(test, singleCase, []byte(pref+stringsEncode.Replace(text)+"\n"))
					expected := stringsRawGoAnswer(test, goBinary, goIdentity, corpus, unit.lo, pref)
					clean(test, "Go raw", expected)
					decoded := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`).Replace(
						strings.TrimSuffix(string(expected.stdout), "\n"),
					)
					for _, r := range []run{
						stringsExecute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, args...),
						stringsOnNode(test, main, args...),
					} {
						clean(test, "raw", r)
						equal(test, "raw "+pref, r.stdout, []byte(decoded))
					}
				}
			case "mutant":
				want := fullWant
				clean(test, "Go", want)
				mutation := stringsMutations[unit.lo]
				for _, r := range []run{
					stringsExecute(
						test,
						stringsSanitizerEnvironment(false),
						mutantBinaries[mutation.name],
						"--batch",
						cases,
					),
					stringsOnNode(test, mutants[mutation.name].main, "--batch", cases),
				} {
					clean(test, "mutant", r)
					if bytes.Equal(r.stdout, want.stdout) {
						test.Fatal("mutant survived")
					}
					a, b := strings.Split(string(r.stdout), "\n"), strings.Split(string(want.stdout), "\n")
					for i := range b {
						if i >= len(a) || a[i] != b[i] {
							test.Logf("mutant %s caught by %s holding batch case %d", mutation.name, unit.name, i)
							break
						}
					}
				}
			case "throughput":
				want := fullWant
				clean(test, "Go", want)
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
						test.Log("external library not checked: set ADAMIC_CSSSTRINGS_LIBRARY")
						return
					}
					command = "node"
					args = []string{script, library, cases}
				}
				var elapsed time.Duration
				for round := 0; round < 3; round++ {
					start := time.Now()
					answer := stringsExecute(test, nil, command, args...)
					elapsed += time.Since(start)
					clean(test, side, answer)
					equal(test, side, answer.stdout, want.stdout)
				}
				test.Logf(
					"throughput %s %.1f texts/s, 3 process runs %.6fs (startup, input, escaping, output included; build excluded)",
					side,
					float64(len(corpus.texts)*2*3)/elapsed.Seconds(),
					elapsed.Seconds(),
				)
			default:
				test.Fatalf("unknown unit %q", unit.kind)
			}
		}
	})
}

func runCSSStringsShard(test *testing.T, ordinal int) {
	test.Helper()
	// Fetch shared products once per process before starting this shard's clock.
	prepareCSSStringsSetup(test)
	if stringsTopSetup.run == nil {
		test.Fatal("shared setup did not complete")
	}
	timer := time.AfterFunc(
		90*time.Second,
		func() { panic(fmt.Sprintf("cooked TestCSSStrings_%03d: case deadline exceeded 90s", ordinal)) },
	)
	defer timer.Stop()
	started := time.Now()
	defer func() {
		if elapsed := time.Since(started); elapsed > 60*time.Second {
			test.Errorf("invalid top-level unit: %s exceeds 60s", elapsed)
		}
	}()
	if stringsTopSetup.run == nil {
		test.Fatal("shared setup did not complete")
	}
	selected := selectedStringsUnits(test, stringsTopSetup.units)
	for _, unit := range selected {
		if unit.name == stringsTopSetup.units[ordinal].name {
			stringsTopSetup.run(test, unit)
			return
		}
	}
	test.Skip("excluded by ADAMIC_TEST_SHARD")
}

func write(test *testing.T, path string, b []byte) {
	test.Helper()
	if writeError := os.WriteFile(path, b, 0644); writeError != nil {
		test.Fatal(writeError)
	}
}
func clean(test *testing.T, name string, r run) {
	test.Helper()
	if r.exitCode != 0 || len(r.stderr) > 0 {
		test.Fatalf("%s exit %d stderr %s", name, r.exitCode, r.stderr)
	}
}
func equal(test *testing.T, name string, a, b []byte) {
	test.Helper()
	if !bytes.Equal(a, b) {
		i := 0
		for i < len(a) && i < len(b) && a[i] == b[i] {
			i++
		}
		test.Fatalf("%s first byte difference at %d (lengths %d/%d)", name, i, len(a), len(b))
	}
}

// stringsPreparedOracle is the Go side, cohere with testdata/bridge.go and the driver testdata/go_driver.go its overlay
// lays over it, a product (buildcache.GoBuildIn) built ahead. The overlay names this checkout's paths, so it sits in a
// temporary directory and is never part of the product (#tqrqx60).
func stringsPreparedOracle(test *testing.T) string { return oracles.Printer(test) }
