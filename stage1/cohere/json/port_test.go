package json

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

const repository = "../../.."

type run struct {
	stdout, stderr []byte
	exitCode       int
}

func execute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	result, err := executeResult(environment, name, arguments...)
	if err != nil {
		t.Fatalf("running %s: %v", name, err)
	}
	return result
}

func executeResult(environment []string, name string, arguments ...string) (run, error) {
	command := exec.Command(name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := childguard.Run(command, jsonGuard)
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return run{}, err
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}, nil
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
func onJavaScriptBackend(t *testing.T, program jsonProducts, arguments ...string) run {
	t.Helper()
	return onNode(t, program.script, arguments...)
}

// nativelyRun is natively's run alone.
func nativelyRun(t *testing.T, program jsonProducts, arguments ...string) run {
	t.Helper()
	result, _ := natively(t, program, arguments...)
	return result
}

// natively builds the lowered port under the address and undefined-behavior sanitizers and runs it,
// returning the binary too, for the leak check. Leak detection is off here, as in the oracle; leaks is
// its own run.
func natively(t *testing.T, program jsonProducts, arguments ...string) (run, string) {
	t.Helper()
	tools, err := jsonClangToolchain()
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(jsonBuildUnit(t, "build-sanitized", jsonNativePortInputs(program.c, true, tools), buildJSONSanitizedPort(program.c)), "port")
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return execute(t, environment, binary, arguments...), binary
}

// leaks returns a report of everything the finished port never let go of, or "": macOS's leaks tool on
// an unsanitized build, or LeakSanitizer on Linux running the sanitized binary again, as the oracle
// checks every fixture.
func leaks(t *testing.T, program jsonProducts, sanitized string, arguments ...string) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		tools, err := jsonClangToolchain()
		if err != nil {
			t.Fatal(err)
		}
		binary := filepath.Join(jsonBuildUnit(t, "build-release", jsonNativePortInputs(program.c, false, tools), buildJSONReleasePort(program.c)), "port")
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

var portFiles = []string{"parser.ts", "doc.ts", "formatter.ts", "main.ts", "width.ts", "widthTables.ts", "identifierTables.ts"}

func escape(text string) string {
	return strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(text)
}
func protocol(cases []textCase, answers []answer) (string, string) {
	var inputs, outputs strings.Builder
	for index, item := range cases {
		fmt.Fprintf(&inputs, "%s\t%s\n", item.Name, escape(item.Text))
		value := answers[index]
		if value.Error != "" {
			fmt.Fprintf(&outputs, "error\t%s\n", escape(value.Error))
		} else {
			fmt.Fprintf(&outputs, "ok\t%s\n", escape(value.Output))
		}
	}
	return inputs.String(), outputs.String()
}
func compare(t *testing.T, name string, result run, expected string, cases []textCase) {
	t.Helper()
	if err := comparisonError(name, result, expected, cases); err != nil {
		t.Fatal(err)
	}
}
func comparisonError(name string, result run, expected string, cases []textCase) error {
	if result.exitCode != 0 || len(result.stderr) != 0 {
		return fmt.Errorf("%s exits %d: %s", name, result.exitCode, result.stderr)
	}
	if string(result.stdout) == expected {
		return nil
	}
	got, want := strings.Split(string(result.stdout), "\n"), strings.Split(expected, "\n")
	for index := 0; index < len(got) && index < len(want); index++ {
		if got[index] != want[index] {
			label := "end"
			if index < len(cases) {
				label = cases[index].Name
			}
			a, b := got[index], want[index]
			first := 0
			for first < len(a) && first < len(b) && a[first] == b[first] {
				first++
			}
			start := max(0, first-80)
			endA := min(len(a), first+200)
			endB := min(len(b), first+200)
			return fmt.Errorf("%s case %d %s byte %d: got %q; Go %q", name, index, label, first, a[start:endA], b[start:endB])
		}
	}
	return fmt.Errorf("%s output length %d, Go %d", name, len(result.stdout), len(expected))
}

const testPortMatchesGoCohereShards = 2168

// TestPortMatchesGoCohere checks every side, sanitizer and leak detector on every
// shard. ADAMIC_TEST_SHARD=i/n (zero-based i) selects shard ordinals modulo n;
// unset runs all shards. Gate selectors use direct shard-0000 through shard-2071.
// Units 0..517 compare all sides; 518..2071 hold three benchmark rounds per range.
// ADAMIC_JSON_BENCH=1 enables those benchmark comparisons. Validate the complete union before selecting any work.
// Build subtests prepare shared inputs once; shard clocks start after preparation.
// Measure isolated units with -parallel=1; normal runs schedule shards in parallel.
func TestPortMatchesGoCohere(t *testing.T) {
	t.Parallel()
	parentStart := time.Now()
	cases := sampledCorpusCases(t, 32)
	shards := jsonPortShards(cases)
	enumeratedUnits := len(shards) + 3*len(shards) // agreement plus three benchmark rounds
	if got := enumeratedUnits; got != testPortMatchesGoCohereShards {
		t.Fatalf("enumerated %d shards, declared %d; corpus sampling is incompatible with the gate census", got, testPortMatchesGoCohereShards)
	}
	if err := jsonPortUnion(cases, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("shard union: %d cases exactly once across %d shards", len(cases), len(shards))
	index, count, err := jsonPortSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	setup := time.Now()
	goTools, err := jsonGoToolchain()
	if err != nil {
		t.Fatal(err)
	}
	clangTools, err := jsonClangToolchain()
	if err != nil {
		t.Fatal(err)
	}
	oracleDir := jsonGoOracleUnit(t)
	loweredDir := jsonBuildUnit(t, "build-lowered-port", jsonLoweredPortInputs(goTools), buildJSONLoweredPort)
	cBytes, err := os.ReadFile(filepath.Join(loweredDir, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	c := string(cBytes)
	releaseDir := jsonBuildUnit(t, "build-release", jsonNativePortInputs(c, false, clangTools), buildJSONReleasePort(c))
	sanitizedDir := jsonBuildUnit(t, "build-sanitized", jsonNativePortInputs(c, true, clangTools), buildJSONSanitizedPort(c))
	entry, script := filepath.Join(loweredDir, "main.ts"), filepath.Join(loweredDir, "program.mjs")
	release, sanitized, oracle := filepath.Join(releaseDir, "port"), filepath.Join(sanitizedDir, "port"), filepath.Join(oracleDir, "go-cohere")
	if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
		if err := os.MkdirAll(artifacts, 0755); err != nil {
			t.Fatal(err)
		}
		for name, product := range map[string]string{"release-port": release, "sanitized-port": sanitized, "go-cohere": oracle} {
			data, err := os.ReadFile(product)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(artifacts, name), data, 0755); err != nil {
				t.Fatal(err)
			}
		}
	}
	setupTime := time.Since(setup)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	references := make([]string, len(shards))
	for ordinal, shard := range shards {
		if ordinal%count != index {
			continue
		}
		t.Run(fmt.Sprintf("shard-%04d", ordinal), func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			items := cases[shard.start:shard.end]
			input, _ := protocol(items, make([]answer, len(items)))
			path := filepath.Join(t.TempDir(), "cases.txt")
			if err := os.WriteFile(path, []byte(input), 0644); err != nil {
				t.Fatal(err)
			}
			sides := []struct {
				name, binary           string
				arguments, environment []string
			}{
				{"Go", oracle, []string{"--cases", path}, nil},
				{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, entry, "--cases", path}, nil},
				{"release", release, []string{"--cases", path}, nil},
				{"ASan/UBSan", sanitized, []string{"--cases", path}, []string{"ASAN_OPTIONS=detect_leaks=0"}},
				{"JavaScript backend", "node", []string{"--disable-warning=ExperimentalWarning", runner, script, "--cases", path}, nil},
			}
			if runtime.GOOS == "linux" {
				sides = append(sides, struct {
					name, binary           string
					arguments, environment []string
				}{"LeakSanitizer", sanitized, []string{"--cases", path}, []string{"ASAN_OPTIONS=detect_leaks=1"}})
			}
			results := make([]chunkResult, len(sides))
			timings := make([]time.Duration, len(sides))
			// The sanitizer processes retain substantially more memory than the
			// other sides. Run them together after the first wave has exited,
			// so an isolated four-CPU unit does not compete with Node/Go heaps.
			for _, wave := range [][]int{{0, 1, 2, 4}, {3, 5}} {
				var workers sync.WaitGroup
				for _, side := range wave {
					if side >= len(sides) {
						continue
					}
					workers.Add(1)
					go func(side int) {
						defer workers.Done()
						s := sides[side]
						began := time.Now()
						results[side].result, results[side].err = executeResult(s.environment, s.binary, s.arguments...)
						timings[side] = time.Since(began)
					}(side)
				}
				workers.Wait()
			}
			for side, timing := range timings {
				t.Logf("%s %.3fs", sides[side].name, timing.Seconds())
			}
			reference := results[0]
			if reference.err != nil || reference.result.exitCode != 0 || len(reference.result.stderr) != 0 {
				t.Fatalf("%s Go oracle: %v exit %d: %s", t.Name(), reference.err, reference.result.exitCode, reference.result.stderr)
			}
			expected := string(reference.result.stdout)
			references[ordinal] = expected
			if len(strings.Split(strings.TrimSuffix(expected, "\n"), "\n")) != len(items) {
				t.Fatalf("%s Go oracle did not answer all %d cases", t.Name(), len(items))
			}
			for side := 1; side < len(sides); side++ {
				if results[side].err != nil {
					t.Fatalf("%s %s: %v", t.Name(), sides[side].name, results[side].err)
				}
				compare(t, t.Name()+" "+sides[side].name, results[side].result, expected, items)
				if side >= 3 {
					compare(t, t.Name()+" "+sides[side].name+" vs release", results[side].result, string(results[2].result.stdout), items)
				}
			}
			if runtime.GOOS != "linux" {
				report := execute(t, nil, "leaks", "--atExit", "--", release, "--cases", path)
				if report.exitCode != 0 {
					t.Fatalf("%s leaks: %s", t.Name(), report.stdout)
				}
			}
			if os.Getenv("ADAMIC_JSON_GUARD_CALIBRATE") == "1" {
				// Calibration uses the same per-case Go answers as mandatory comparisons.
				answers, err := jsonPortAnswers(expected, len(items))
				if err != nil {
					t.Fatal(err)
				}
				calibrateNativeCases(t, sanitized, items, answers)
			}
			if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
				directory := filepath.Join(artifacts, fmt.Sprintf("shard-%04d", ordinal))
				if err := os.MkdirAll(directory, 0755); err != nil {
					t.Fatal(err)
				}
				writeJSON(t, filepath.Join(directory, "cases.json"), items)
				for side, s := range sides {
					name := strings.NewReplacer("/", "-", " ", "-").Replace(s.name)
					if err := os.WriteFile(filepath.Join(directory, name+".txt"), results[side].result.stdout, 0644); err != nil {
						t.Fatal(err)
					}
				}
			}
			t.Logf("case range [%d:%d]; %d cases; execution %.3fs; shared setup waited %.3fs; execution plus locally measured builds %.3fs", shard.start, shard.end, len(items), time.Since(start).Seconds(), setupTime.Seconds(), (time.Since(start) + setupTime).Seconds())
		})
	}
	var setupBeforeShards time.Duration
	t.Cleanup(func() {
		cleanupStart := time.Now()
		defer func() {
			t.Logf("setup outside shards: %.3fs (including local builds)", (setupBeforeShards + time.Since(cleanupStart)).Seconds())
		}()
		allReferences := true
		for _, reference := range references {
			allReferences = allReferences && reference != ""
		}
		if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" && count == 1 && allReferences {
			joined := strings.Join(references, "")
			answers, err := jsonPortAnswers(joined, len(cases))
			if err != nil {
				t.Fatal(err)
			}
			input, _ := protocol(cases, answers)
			writeJSON(t, filepath.Join(artifacts, "cases.json"), cases)
			writeJSON(t, filepath.Join(artifacts, "go.json"), answers)
			for name, value := range map[string]string{"cases.txt": input, "node.txt": joined} {
				if err := os.WriteFile(filepath.Join(artifacts, name), []byte(value), 0644); err != nil {
					t.Fatal(err)
				}
			}
		}

	})
	{
		for ordinal, shard := range shards {
			for round := 0; round < 3; round++ {
				unit := len(shards) + ordinal*3 + round
				if unit%count != index {
					continue
				}
				t.Run(fmt.Sprintf("shard-%04d", unit), func(t *testing.T) {
					if os.Getenv("ADAMIC_JSON_BENCH") != "1" {
						t.Skip("ADAMIC_JSON_BENCH=1 enables benchmark comparisons")
					}
					t.Parallel()
					items := cases[shard.start:shard.end]
					input, _ := protocol(items, make([]answer, len(items)))
					path := filepath.Join(t.TempDir(), "cases.txt")
					if err := os.WriteFile(path, []byte(input), 0644); err != nil {
						t.Fatal(err)
					}
					expected := ""      // A gate unit runs independently of the agreement shard.
					if expected == "" { // -run may select a benchmark without its agreement sibling.
						reference := execute(t, nil, oracle, "--cases", path)
						if reference.exitCode != 0 || len(reference.stderr) != 0 {
							t.Fatalf("%s Go oracle: %+v", t.Name(), reference)
						}
						expected = string(reference.stdout)
						if _, err := jsonPortAnswers(expected, len(items)); err != nil {
							t.Fatal(err)
						}
					}
					for _, side := range []struct {
						name, binary string
						args         []string
					}{
						{"release", release, []string{"--cases", path}},
						{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, entry, "--cases", path}},
						{"Go", oracle, []string{"--cases", path}},
					} {
						began := time.Now()
						result := execute(t, nil, side.binary, side.args...)
						compare(t, t.Name()+" "+side.name, result, expected, items)
						t.Logf("case range [%d:%d]; round %d %s %d cases %.3fs", shard.start, shard.end, round, side.name, len(items), time.Since(began).Seconds())
					}
				})
			}
		}
	}
	setupBeforeShards = time.Since(parentStart)
}

func TestThreePortMutantsAreCaught(t *testing.T) {
	t.Parallel()
	cases := []textCase{
		{"probe.json", `{"a":1,"b":[2,3]}`},
		{"package.json", `{"a":1,"b":[2,3]}`},
		{"package.json", `{"text":"é😀","empty":[]}`},
		{"probe.json", `[1__0]`},
	}
	answers, _ := cohereAnswers(t, cases, false)
	input, expected := protocol(cases, answers)
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []printerMutation{
		{"missing final newline", "formatter.ts", "printer.docs.concat([document, printer.docs.line(false, true)])", "document"},
		{"missing colon space", "formatter.ts", "documents.text(': ')", "documents.text(':')"},
		{"wrong filename parser", "formatter.ts", "base === 'package.json'", "base === 'other-package.json'"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			t.Parallel()
			program := jsonPrepareMutation(t, mutation)
			entry := program.entry
			for _, side := range []struct {
				name   string
				result run
			}{
				{"Node", onNode(t, entry, "--cases", path)},
				{"native", nativelyRun(t, program, "--cases", path)},
			} {
				if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
					t.Fatalf("%s mutant must succeed: exit %d stderr %s", side.name, side.result.exitCode, side.result.stderr)
				}
				if string(side.result.stdout) == expected {
					t.Fatalf("%s failed to catch %s", side.name, mutation.name)
				}
				got, want := strings.Split(string(side.result.stdout), "\n"), strings.Split(expected, "\n")
				count := 0
				for index := range cases {
					if got[index] != want[index] {
						count++
						t.Logf("%s caught %s on %s: %q; Go %q", side.name, mutation.name, cases[index].Name, got[index], want[index])
					}
				}
				if count == 0 {
					t.Fatal("difference must be in formatter answers")
				}
			}
		})
	}
}

const testAdditionalJSONBoundariesShards = 5

// ADAMIC_TEST_SHARD=i/n selects ordinal modulo n; unset runs every shard.
func TestAdditionalJSONBoundaries(t *testing.T) {
	t.Parallel()
	texts := []string{
		"[\n// dangling\n]",
		`[1.0,1.000,.000,1e0_0,1_0e+001,1.0_0,1.e0,1.000e10]`,
		`[1e]`, `[1e+]`, `[0x]`, `[0o8]`, `[.]`, `[01]`, `[1n]`,
		`[,]`, `[1,,]`, `[,,2]`, `{1:2,1.50:3,0x10:4}`, `{'a':'can\'t',b:'"hi"'}`,
		"`raw`", "`\\u0041\\n\\x42`", "`line\nline`", "{é:1,𐌘:2}",
		"// header\n\n{\n// before\n a:1, // after\n b:[1,2]\n}\n// footer",
		"{a:1,/* before */b:2}", "{/* dangling */}", "[1,// after\n2]",
		`"\xG0"`, `"\u{}"`, `"\u{110000}"`, `"\uD800"`, `"\uD83D\uDE00"`,
	}
	for _, literal := range []string{"中文", "ｅ́", "🇺🇸", "👨‍👩‍👧", "1️⃣", "❤️", "©", "👍🏽", "𐌘"} {
		texts = append(texts, `{"text":"`+strings.Repeat(literal, 25)+`","items":[1,2,3]}`)
	}
	var cases []textCase
	for index, text := range texts {
		for _, name := range []string{"probe.json", "package.json"} {
			cases = append(cases, textCase{fmt.Sprintf("boundary/%d/%s", index, name), text})
		}
	}
	shards := jsonPortShards(cases)
	if len(shards) != testAdditionalJSONBoundariesShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testAdditionalJSONBoundariesShards)
	}
	if err := jsonPortUnion(cases, shards); err != nil {
		t.Fatal(err)
	}
	index, count, err := jsonPortSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	products := jsonPreparePort(t, runtime.GOOS == "darwin", true)
	references := make([][]answer, len(shards))
	for ordinal, shard := range shards {
		if ordinal%count != index {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", ordinal), func(t *testing.T) {
			t.Parallel()
			items := cases[shard.start:shard.end]
			answers := jsonOracleAnswers(t, products.oracle, items)
			references[ordinal] = answers
			input, expected := protocol(items, answers)
			path := filepath.Join(t.TempDir(), "cases.txt")
			if err := os.WriteFile(path, []byte(input), 0644); err != nil {
				t.Fatal(err)
			}
			node := onNode(t, products.entry, "--cases", path)
			compare(t, t.Name()+" Node boundaries", node, expected, items)
			var environment []string
			if runtime.GOOS == "linux" {
				environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
			}
			compare(t, t.Name()+" sanitized boundaries", execute(t, environment, products.sanitized, "--cases", path), expected, items)
			switch runtime.GOOS {
			case "linux":
				compare(t, t.Name()+" LeakSanitizer boundaries", execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, "--cases", path), expected, items)
			case "darwin":
				result := execute(t, nil, "leaks", "--atExit", "--", products.release, "--cases", path)
				if result.exitCode != 0 {
					t.Fatalf("%s leaks: %s", t.Name(), result.stdout)
				}
			default:
				t.Fatalf("no leak check for %s", runtime.GOOS)
			}
			if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
				directory := filepath.Join(artifacts, "boundaries", fmt.Sprintf("shard-%03d", ordinal))
				if err := os.MkdirAll(directory, 0755); err != nil {
					t.Fatal(err)
				}
				writeJSON(t, filepath.Join(directory, "cases.json"), items)
				writeJSON(t, filepath.Join(directory, "go.json"), answers)
				if err := os.WriteFile(filepath.Join(directory, "node.txt"), node.stdout, 0644); err != nil {
					t.Fatal(err)
				}
			}
			t.Logf("case range [%d:%d]; %d cases", shard.start, shard.end, len(items))
		})
	}
	t.Cleanup(func() {
		if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" && count == 1 {
			var answers []answer
			for _, reference := range references {
				if reference == nil {
					return
				}
				answers = append(answers, reference...)
			}
			input, expected := protocol(cases, answers)
			writeJSON(t, filepath.Join(artifacts, "boundary.json"), cases)
			writeJSON(t, filepath.Join(artifacts, "boundary-go.json"), answers)
			for name, text := range map[string]string{"boundary.txt": input, "boundary-node.txt": expected} {
				if err := os.WriteFile(filepath.Join(artifacts, name), []byte(text), 0644); err != nil {
					t.Error(err)
				}
			}
		}
	})
	t.Logf("exact union: %d cases", len(cases))
}

func buildGoDriver(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := buildJSONGoOracle(dir); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(dir, "go-cohere")
}

const testSingleFileStdoutDriverShards = 2

// ADAMIC_TEST_SHARD=i/n selects shard ordinals modulo n; unset runs every shard.
func TestSingleFileStdoutDriver(t *testing.T) {
	t.Parallel()
	cases := []textCase{{"probe.json", `{"text":"é😀","items":[1,2,3]}`}, {"package.json", `{"text":"é😀","items":[1,2,3]}`}}
	shards := []nativeChunk{{start: 0, end: 1}, {start: 1, end: 2}}
	if len(shards) != testSingleFileStdoutDriverShards {
		t.Fatal("single-file shard count changed")
	}
	if err := jsonPortUnion(cases, shards); err != nil {
		t.Fatal(err)
	}
	index, count, err := jsonPortSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	products := jsonPreparePort(t, false, true)
	for ordinal, shard := range shards {
		if ordinal%count != index {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", ordinal), func(t *testing.T) {
			t.Parallel()
			items := cases[shard.start:shard.end]
			expected := jsonOracleAnswers(t, products.oracle, items)
			path := filepath.Join(t.TempDir(), items[0].Name)
			if err := os.WriteFile(path, []byte(items[0].Text), 0644); err != nil {
				t.Fatal(err)
			}
			compare(t, t.Name()+" Node file driver", onNode(t, products.entry, path), expected[0].Output, items)
			compare(t, t.Name()+" native file driver", execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, path), expected[0].Output, items)
			t.Logf("case range [%d:%d]; %s", shard.start, shard.end, items[0].Name)
		})
	}
}
