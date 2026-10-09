package cssnumbers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
)

const repository = "../../.."

const testCSSNumbersShards = 357

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
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	var texts []string
	paths := corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit, []string{"internal/format/css/testdata/prettier", "internal/lint/rules/tailwind"}, []string{"*.css", "*.scss", "*.less"})
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		texts = append(texts, string(data))
	}
	files := len(texts)
	alphabet := []string{"a", "0", "1", ".", "e", "-", "\"", "\\"}
	var generate func(string, int)
	generate = func(s string, n int) {
		texts = append(texts, s)
		if n > 0 {
			for _, c := range alphabet {
				generate(s+c, n-1)
			}
		}
	}
	generate("", 5)
	texts = append(texts, "\r\t\u2028\u2029", strings.Repeat("'😀\\\"x' ", 10000), "'never closed\\", `"\'"`, "\x00'null'")
	for _, mantissa := range []string{"0", "01", ".0", ".00100", "1.", "1.000", "12.34000"} {
		for _, exponent := range []string{"", "e0", "E+000", "e-000", "e+001", "e-002", "E00020", "e+", "e-"} {
			for _, unit := range strings.Split("|n|EM|Q|HZ|kHZ|px|cQmAX|unknown|foo|é|Σ|İ", "|") {
				for _, prefix := range []string{"", "a", "$x", "@foo", "-", "+", "😀", "'", "\""} {
					texts = append(texts, prefix+mantissa+exponent+unit)
				}
			}
		}
	}
	for _, unit := range strings.Split("em|rem|ex|rex|cap|rcap|ch|rch|ic|ric|lh|rlh|vw|svw|lvw|dvw|vh|svh|lvh|dvh|vi|svi|lvi|dvi|vb|svb|lvb|dvb|vmin|svmin|lvmin|dvmin|vmax|svmax|lvmax|dvmax|cm|mm|Q|in|pt|pc|px|deg|grad|rad|turn|s|ms|Hz|kHz|dpi|dpcm|dppx|x|cqw|cqh|cqi|cqb|cqmin|cqmax|fr", "|") {
		texts = append(texts, ".1000E+002"+strings.ToUpper(unit))
	}
	texts = append(texts, cssAuditNumberInputs...)
	texts = append(texts, "'1.000px'", "1.000unknown", "a1.000px", "😀1.000px", strings.Repeat(".000100E-002KHZ ", 10000))
	raw := append(append([]string{}, texts[:files]...), "", "a", "a\n", "a\r\n\n", "'😀'", "\"\\'\"")
	return numbersCorpus{texts: texts, paths: paths, raw: raw}
}

// TestCSSNumbers builds each input once before the parallel units. ADAMIC_TEST_SHARD=i/n
// (zero-based i) selects units whose stable ordinal modulo n equals i; unset runs all.
// Stable shard-NNN names use the same zero-based ordinals for direct -run selection.
// No corpus is sampled: batch ranges, raw files, mutants and throughput sides form the census.
// TestCSSNumbers_NNN are top-level parallel units. ADAMIC_TEST_SHARD=i/n
// selects stable ordinals modulo n locally; unset runs every unit. The gate uses
// -run '^TestCSSNumbers_NNN$'. Every original side and sanitizer/leak check stays.
// Contiguous ranges are fixed by cohere commit 7945d102a6c18dd36adf9114a758ce646e8b2359
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
		goBinary := numbersPreparedOracle(t)
		main, _ := filepath.Abs("main.ts")
		build := numbersPreparedProgram(t, -1)
		sanitized := numbersPreparedNative(t, -1, true)
		fast := numbersPreparedNative(t, -1, false)
		counted := numbersCounted(t, build)
		code, err := os.ReadFile(build.c)
		if err != nil {
			t.Fatal(err)
		}
		// macOS's leaks tool needs the unsanitized product; Linux reuses sanitized.
		mutants := make(map[string]numbersProgram)
		mutantBinaries := make(map[string]string)
		for index, mutation := range numbersMutations {
			product := numbersPreparedProgram(t, index)
			mutants[mutation.name] = product
			mutantBinaries[mutation.name] = numbersPreparedNative(t, index, true)
		}
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
				path := filepath.Join(t.TempDir(), "cases.txt")
				write(t, path, numbersInput(corpus.texts[unit.lo:unit.hi]))
				want := numbersOracleAnswers(t, "Go", goBinary, nil, numbersInput(corpus.texts[unit.lo:unit.hi]), goIdentity)
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
				numbersLeaks(t, string(code), sanitized, counted, "--batch", path)
				if library != "" {
					answer := numbersOracleAnswers(t, "Prettier", "node", []string{script, library}, numbersInput(corpus.texts[unit.lo:unit.hi]), libraryIdentity)
					clean(t, "Prettier", answer)
					checkNumbersOutput(t, unit, "Prettier", answer.stdout, want.stdout)
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
					write(t, singleCase, []byte(pref+numbersEncode.Replace(text)+"\n"))
					expected := numbersOracleAnswers(t, "Go", goBinary, nil, []byte(pref+numbersEncode.Replace(text)+"\n"), goIdentity)
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
					numbersExecute(t, numbersSanitizerEnvironment(false), mutantBinaries[mutation.name], "--batch", cases),
					numbersOnNode(t, mutants[mutation.name].main, "--batch", cases),
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
					command = fast
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
	// Fetch shared products once per process before starting this shard's clock.
	prepareCSSNumbersSetup(t)
	if numbersTopSetup.run == nil {
		t.Fatal("shared setup did not complete")
	}
	timer := time.AfterFunc(90*time.Second, func() { panic(fmt.Sprintf("cooked TestCSSNumbers_%03d: case deadline exceeded 90s", ordinal)) })
	defer timer.Stop()
	started := time.Now()
	defer func() {
		if elapsed := time.Since(started); elapsed > 60*time.Second {
			t.Errorf("invalid top-level unit: %s exceeds 60s", elapsed)
		}
	}()
	if numbersTopSetup.run == nil {
		t.Fatal("shared setup did not complete")
	}
	selected := selectedNumbersUnits(t, numbersTopSetup.units)
	for _, unit := range selected {
		if unit.name == numbersTopSetup.units[ordinal].name {
			numbersTopSetup.run(t, unit)
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

func numbersPreparedOracle(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	bridge, _ := filepath.Abs("testdata/bridge.go")
	driver, _ := filepath.Abs("testdata/go_driver.go")
	cohere := filepath.Join(root, "cohere")
	oracleDir := numbersGoProduct(t, "Go oracle", func(dir string) error {
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
		command := numbersSetupCommand(t, "go", "build", "-overlay="+overlayPath, "-o", filepath.Join(dir, "go-printer"), filepath.Join(cohere, "cmd/adamic_stage_one/main.go"))
		command.Dir = cohere
		if output, err := childguard.CombinedOutput(command, childguard.Options{}); err != nil {
			return fmt.Errorf("Go bridge: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(oracleDir, "go-printer")
}
