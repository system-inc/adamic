package cssnumbers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/stage1/cohere/cssnumbers/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const repository = "../../.."

const testCSSNumbersShards = 361

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

// numbersCorpus preserves the original enumeration, including duplicate texts as distinct cases.
type numbersCorpus struct {
	texts []string
	paths []string
	raw   []string
}

func enumerateNumbers(test *testing.T) numbersCorpus {
	corpus := oracles.Enumerate(test)
	return numbersCorpus{texts: corpus.Texts, paths: corpus.Paths, raw: corpus.Raw}
}

// TestCSSNumbers builds each input once before the parallel units. ADAMIC_TEST_SHARD=i/n
// (zero-based i) selects units whose stable ordinal modulo n equals i; unset runs all.
// Stable shard-NNN names use the same zero-based ordinals for direct -run selection.
// No corpus is sampled: batch ranges, raw files, mutants and throughput sides form the census.
// TestCSSNumbers_NNN are top-level parallel units. ADAMIC_TEST_SHARD=i/n
// selects stable ordinals modulo n locally; unset runs every unit. The gate uses
// -run '^TestCSSNumbers_NNN$'. Every original side and sanitizer/leak check stays.
// Contiguous ranges are fixed by cohere commit 0183ccf8c57a775f874ca73be1adfc06e9ad1562
// and the generated grammar in enumerateNumbers.

func prepareCSSNumbersSetup(test *testing.T) {
	test.Helper()
	numbersTopSetup.once.Do(func() {
		corpus := enumerateNumbers(test)
		units := numbersUnits(corpus)
		if len(units) != testCSSNumbersShards {
			test.Fatalf("enumerated %d shards, declared %d", len(units), testCSSNumbersShards)
		}
		verifyNumbersUnion(test, corpus, units)
		dir := numbersTopTempDir(test)
		cases := filepath.Join(dir, "cases.txt")
		write(test, cases, numbersInput(corpus.texts))
		if keep := os.Getenv("ADAMIC_CSSNUMBERS_KEEP"); keep != "" {
			write(test, keep, numbersInput(corpus.texts))
			manifest, encodeError := json.MarshalIndent(corpus.paths, "", "  ")
			if encodeError != nil {
				test.Fatal(encodeError)
			}
			write(test, keep+".files.json", manifest)
		}
		// Only what every unit reads is fetched here; the port's and each mutant's products wait for the units that
		// read them (numbersPrepareUnitProducts).
		goBinary := numbersPreparedOracle(test)
		main, _ := filepath.Abs("main.ts")
		library := os.Getenv("ADAMIC_CSSNUMBERS_LIBRARY")
		script, _ := filepath.Abs("testdata/library.mjs")
		goIdentity := numbersOracleIdentity(test, goBinary)
		libraryIdentity := ""
		if library != "" {
			libraryIdentity = numbersOracleIdentity(test, library)
		}
		if library == "" {
			test.Log("external library not checked: set ADAMIC_CSSNUMBERS_LIBRARY")
		}
		fullWant := numbersSetupOracleAnswers(test, "Go", goBinary, nil, numbersInput(corpus.texts), goIdentity)
		clean(test, "Go", fullWant)
		numbersTopSetup.units = units
		numbersTopSetup.run = func(test *testing.T, unit numbersUnit) {
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
				build := numbersPreparedProgram(test, -1)
				sanitized := numbersPreparedNative(test, -1, true)
				// macOS's leaks tool needs the unsanitized product; Linux reuses sanitized.
				fast := numbersPreparedNative(test, -1, false)
				path := filepath.Join(test.TempDir(), "cases.txt")
				write(test, path, numbersInput(corpus.texts[unit.lo:unit.hi]))
				want := numbersBatchGoAnswers(test, goBinary, goIdentity, corpus, unit)
				clean(test, "Go", want)
				for _, side := range []struct {
					name   string
					result run
				}{
					{"native", numbersExecute(test, numbersSanitizerEnvironment(false), sanitized, "--batch", path)},
					{"Node", numbersOnNode(test, main, "--batch", path)},
					{"JavaScript backend", numbersOnNode(test, build.javascript, "--batch", path)},
				} {
					clean(test, side.name, side.result)
					checkNumbersOutput(test, unit, side.name, side.result.stdout, want.stdout)
				}
				numbersLeaks(test, sanitized, fast, "--batch", path)
				if library != "" {
					answer := numbersBatchPrettierAnswers(test, library, libraryIdentity, corpus, unit)
					clean(test, "Prettier", answer)
					checkNumbersOutput(test, unit, "Prettier", answer.stdout, want.stdout)
				}
			case "raw":
				sanitized := numbersPreparedNative(test, -1, true)
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
					write(test, singleCase, []byte(pref+numbersEncode.Replace(text)+"\n"))
					expected := numbersRawGoAnswer(test, goBinary, goIdentity, corpus, unit.lo, pref)
					clean(test, "Go raw", expected)
					decoded := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`).Replace(
						strings.TrimSuffix(string(expected.stdout), "\n"),
					)
					for _, run := range []run{
						numbersExecute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, args...),
						numbersOnNode(test, main, args...),
					} {
						clean(test, "raw", run)
						equal(test, "raw "+pref, run.stdout, []byte(decoded))
					}
				}
			case "mutant":
				want := fullWant
				clean(test, "Go", want)
				mutation := numbersMutations[unit.lo]
				for _, run := range []run{
					numbersExecute(
						test,
						numbersSanitizerEnvironment(false),
						numbersPreparedNative(test, unit.lo, true),
						"--batch",
						cases,
					),
					numbersOnNode(test, numbersPreparedProgram(test, unit.lo).main, "--batch", cases),
				} {
					clean(test, "mutant", run)
					if bytes.Equal(run.stdout, want.stdout) {
						test.Fatal("mutant survived")
					}
					a, b := strings.Split(string(run.stdout), "\n"), strings.Split(string(want.stdout), "\n")
					for index := range b {
						if index >= len(a) || a[index] != b[index] {
							test.Logf("mutant %s caught by %s holding batch case %d", mutation.name, unit.name, index)
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
					command = numbersPreparedNative(test, -1, false)
					args = []string{"--batch", cases}
				case "Node":
					runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
					command = "node"
					args = []string{"--disable-warning=ExperimentalWarning", runner, main, "--batch", cases}
				case "Prettier":
					if library == "" {
						test.Log("external library not checked: set ADAMIC_CSSNUMBERS_LIBRARY")
						return
					}
					command = "node"
					args = []string{script, library, cases}
				}
				var elapsed time.Duration
				for round := 0; round < 3; round++ {
					start := time.Now()
					answer := numbersExecute(test, nil, command, args...)
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

func runCSSNumbersShard(test *testing.T, ordinal int) {
	test.Helper()
	// Fetch shared products once per process, and this unit's own, before starting this shard's clock.
	prepareCSSNumbersSetup(test)
	if numbersTopSetup.run == nil {
		test.Fatal("shared setup did not complete")
	}
	unit, chosen := numbersTopSetup.units[ordinal], false
	for _, selected := range selectedNumbersUnits(test, numbersTopSetup.units) {
		chosen = chosen || selected.name == unit.name
	}
	if !chosen {
		test.Skip("excluded by ADAMIC_TEST_SHARD")
	}
	numbersPrepareUnitProducts(test, unit)
	timer := time.AfterFunc(
		90*time.Second,
		func() { panic(fmt.Sprintf("cooked TestCSSNumbers_%03d: case deadline exceeded 90s", ordinal)) },
	)
	defer timer.Stop()
	started := time.Now()
	defer func() {
		if elapsed := time.Since(started); elapsed > 60*time.Second {
			test.Errorf("invalid top-level unit: %s exceeds 60s", elapsed)
		}
	}()
	numbersTopSetup.run(test, unit)
}

func write(test *testing.T, path string, b []byte) {
	test.Helper()
	if writeError := os.WriteFile(path, b, 0644); writeError != nil {
		test.Fatal(writeError)
	}
}
func clean(test *testing.T, name string, run run) {
	test.Helper()
	if run.exitCode != 0 || len(run.stderr) > 0 {
		test.Fatalf("%s exit %d stderr %s", name, run.exitCode, run.stderr)
	}
}
func equal(test *testing.T, name string, left, right []byte) {
	test.Helper()
	if !bytes.Equal(left, right) {
		index := 0
		for index < len(left) && index < len(right) && left[index] == right[index] {
			index++
		}
		test.Fatalf("%s first byte difference at %d (lengths %d/%d)", name, index, len(left), len(right))
	}
}

// numbersPreparedOracle is cohere's printer with testdata's bridge and driver laid over it, a product
// (buildcache.GoBuildIn) built ahead. Its key is go's package list, the overlay's map and each replacement's content,
// and the build environment and toolchains go resolves.
func numbersPreparedOracle(test *testing.T) string { return oracles.Printer(test) }
