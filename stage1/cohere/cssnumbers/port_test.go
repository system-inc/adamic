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

// numbersCorpus preserves the original enumeration, including duplicate texts as distinct cases.
type numbersCorpus struct {
	texts []string
	paths []string
	raw   []string
}

func enumerateNumbers(t *testing.T) numbersCorpus {
	c := oracles.Enumerate(t)
	return numbersCorpus{texts: c.Texts, paths: c.Paths, raw: c.Raw}
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

func prepareCSSNumbersSetup(t *testing.T) {
	t.Helper()
	numbersTopSetup.once.Do(func() {
		corpus := enumerateNumbers(t)
		units := numbersUnits(corpus)
		if len(units) != testCSSNumbersShards {
			t.Fatalf("enumerated %d shards, declared %d", len(units), testCSSNumbersShards)
		}
		verifyNumbersUnion(t, corpus, units)
		dir := numbersTopTempDir(t)
		cases := filepath.Join(dir, "cases.txt")
		write(t, cases, numbersInput(corpus.texts))
		if keep := os.Getenv("ADAMIC_CSSNUMBERS_KEEP"); keep != "" {
			write(t, keep, numbersInput(corpus.texts))
			manifest, err := json.MarshalIndent(corpus.paths, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			write(t, keep+".files.json", manifest)
		}
		// Only what every unit reads is fetched here; the port's and each mutant's products wait for the units that
		// read them (numbersPrepareUnitProducts).
		goBinary := numbersPreparedOracle(t)
		main, _ := filepath.Abs("main.ts")
		library := os.Getenv("ADAMIC_CSSNUMBERS_LIBRARY")
		script, _ := filepath.Abs("testdata/library.mjs")
		goIdentity := numbersOracleIdentity(t, goBinary)
		libraryIdentity := ""
		if library != "" {
			libraryIdentity = numbersOracleIdentity(t, library)
		}
		if library == "" {
			t.Log("external library not checked: set ADAMIC_CSSNUMBERS_LIBRARY")
		}
		fullWant := numbersSetupOracleAnswers(t, "Go", goBinary, nil, numbersInput(corpus.texts), goIdentity)
		clean(t, "Go", fullWant)
		numbersTopSetup.units = units
		numbersTopSetup.run = func(t *testing.T, unit numbersUnit) {
			t.Logf("%s: %s range [%d,%d), side %s, %d case IDs", unit.name, unit.kind, unit.lo, unit.hi, unit.side, len(unit.ids))
			start := time.Now()
			defer func() {
				if elapsed := time.Since(start); elapsed > 60*time.Second {
					t.Errorf("invalid test unit %s: wall %s exceeds 60s", unit.name, elapsed)
				}
			}()
			switch unit.kind {
			case "batch":
				build := numbersPreparedProgram(t, -1)
				sanitized := numbersPreparedNative(t, -1, true)
				// macOS's leaks tool needs the unsanitized product; Linux reuses sanitized.
				fast := numbersPreparedNative(t, -1, false)
				path := filepath.Join(t.TempDir(), "cases.txt")
				write(t, path, numbersInput(corpus.texts[unit.lo:unit.hi]))
				want := numbersBatchGoAnswers(t, goBinary, goIdentity, corpus, unit)
				clean(t, "Go", want)
				for _, side := range []struct {
					name   string
					result run
				}{
					{"native", numbersExecute(t, numbersSanitizerEnvironment(false), sanitized, "--batch", path)},
					{"Node", numbersOnNode(t, main, "--batch", path)},
					{"JavaScript backend", numbersOnNode(t, build.javascript, "--batch", path)},
				} {
					clean(t, side.name, side.result)
					checkNumbersOutput(t, unit, side.name, side.result.stdout, want.stdout)
				}
				numbersLeaks(t, sanitized, fast, "--batch", path)
				if library != "" {
					answer := numbersBatchPrettierAnswers(t, library, libraryIdentity, corpus, unit)
					clean(t, "Prettier", answer)
					checkNumbersOutput(t, unit, "Prettier", answer.stdout, want.stdout)
				}
			case "raw":
				sanitized := numbersPreparedNative(t, -1, true)
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
					write(t, singleCase, []byte(pref+numbersEncode.Replace(text)+"\n"))
					expected := numbersRawGoAnswer(t, goBinary, goIdentity, corpus, unit.lo, pref)
					clean(t, "Go raw", expected)
					decoded := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`).Replace(strings.TrimSuffix(string(expected.stdout), "\n"))
					for _, r := range []run{numbersExecute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, args...), numbersOnNode(t, main, args...)} {
						clean(t, "raw", r)
						equal(t, "raw "+pref, r.stdout, []byte(decoded))
					}
				}
			case "mutant":
				want := fullWant
				clean(t, "Go", want)
				mutation := numbersMutations[unit.lo]
				for _, r := range []run{
					numbersExecute(t, numbersSanitizerEnvironment(false), numbersPreparedNative(t, unit.lo, true), "--batch", cases),
					numbersOnNode(t, numbersPreparedProgram(t, unit.lo).main, "--batch", cases),
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
				want := fullWant
				clean(t, "Go", want)
				side := unit.side
				command := goBinary
				args := []string{cases}
				switch side {
				case "native":
					command = numbersPreparedNative(t, -1, false)
					args = []string{"--batch", cases}
				case "Node":
					runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
					command = "node"
					args = []string{"--disable-warning=ExperimentalWarning", runner, main, "--batch", cases}
				case "Prettier":
					if library == "" {
						t.Log("external library not checked: set ADAMIC_CSSNUMBERS_LIBRARY")
						return
					}
					command = "node"
					args = []string{script, library, cases}
				}
				var elapsed time.Duration
				for round := 0; round < 3; round++ {
					start := time.Now()
					answer := numbersExecute(t, nil, command, args...)
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
}

func runCSSNumbersShard(t *testing.T, ordinal int) {
	t.Helper()
	// Fetch shared products once per process, and this unit's own, before starting this shard's clock.
	prepareCSSNumbersSetup(t)
	if numbersTopSetup.run == nil {
		t.Fatal("shared setup did not complete")
	}
	unit, chosen := numbersTopSetup.units[ordinal], false
	for _, selected := range selectedNumbersUnits(t, numbersTopSetup.units) {
		chosen = chosen || selected.name == unit.name
	}
	if !chosen {
		t.Skip("excluded by ADAMIC_TEST_SHARD")
	}
	numbersPrepareUnitProducts(t, unit)
	timer := time.AfterFunc(90*time.Second, func() { panic(fmt.Sprintf("cooked TestCSSNumbers_%03d: case deadline exceeded 90s", ordinal)) })
	defer timer.Stop()
	started := time.Now()
	defer func() {
		if elapsed := time.Since(started); elapsed > 60*time.Second {
			t.Errorf("invalid top-level unit: %s exceeds 60s", elapsed)
		}
	}()
	numbersTopSetup.run(t, unit)
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

// numbersPreparedOracle is cohere's printer with testdata's bridge and driver laid over it, a product
// (buildcache.GoBuildIn) built ahead. Its key is go's package list, the overlay's map and each replacement's content,
// and the build environment and toolchains go resolves.
func numbersPreparedOracle(t *testing.T) string { return oracles.Printer(t) }
