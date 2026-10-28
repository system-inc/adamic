package json

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testguard"
	"github.com/system-inc/adamic/internal/toolchain"

	"github.com/system-inc/adamic/stage1/cohere/json/oracles"
)

const repository = "../../.."

type run struct {
	stdout, stderr []byte
	exitCode       int
	// admissionWait is how long a sanitized run waited for json's memory budget before it started.
	admissionWait time.Duration
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

func execute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	result, err := executeResult(environment, name, arguments...)
	if err != nil {
		t.Fatalf("running %s: %v", name, err)
	}
	return result
}

func executeResult(environment []string, name string, arguments ...string) (run, error) {
	// All sanitized execution paths share admission, including parallel leaves,
	// chunk workers, leak checks and optional calibration.
	var admissionWait time.Duration
	for _, setting := range environment {
		if strings.HasPrefix(setting, "ASAN_OPTIONS=") {
			budget, err := jsonSanitizedMemory()
			if err != nil {
				return run{}, err
			}
			admissionWait = budget.acquire(jsonMemorySanitizedWeight)
			defer budget.release(jsonMemorySanitizedWeight)
			break
		}
	}
	command := exec.Command(name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node, macOS's leaks) is a tool: the one internal/toolchain resolves, with a tool's environment
		// plus the caller's settings, not PATH's. A program a test built (the Go oracle, a port) runs as named, with this
		// process's environment.
		command = toolchain.Command(name, arguments...)
		command.Env = append(command.Env, environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := childguard.Run(command, jsonGuard)
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return run{}, err
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode(), admissionWait: admissionWait}, nil
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

var portFiles = []string{"parser.ts", "doc.ts", "formatter.ts", "main.ts", "width.ts", "widthTables.ts", "identifierTables.ts"}

func portDirectory(t *testing.T, mutation *printerMutation) string {
	t.Helper()
	directory := t.TempDir()
	for _, name := range portFiles {
		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		text := string(source)
		if mutation != nil && mutation.file == name {
			if strings.Count(text, mutation.from) != 1 {
				t.Fatalf("mutant %s must change one place", mutation.name)
			}
			text = strings.Replace(text, mutation.from, mutation.to, 1)
		}
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func protocol(cases []textCase, answers []answer) (string, string) {
	return oracles.Protocol(oracleCases(cases), answers)
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

// jsonPortBinaries are what a port agreement runs: the Go oracle, the lowered port's TypeScript entry and its
// JavaScript backend (both through the Node runner), and the release and sanitized native ports.
type jsonPortBinaries struct {
	oracle, entry, script, release, sanitized, runner string
}

// jsonPortPart is which of a port agreement's sides one test runs. A top shard and the inventory owner run the whole.
// A heavy case's owner is three tests, its unsanitized, sanitized and JavaScript backend parts, which go test runs in
// parallel, so the case's work isn't one test running five sides one after another. Each part runs the Go oracle, so
// each holds its own sides to it.
type jsonPortPart int

const (
	jsonPortWhole jsonPortPart = iota
	jsonPortUnsanitized
	jsonPortSanitized
	jsonPortJavaScript
)

// jsonPortPartSides are the sides each part runs, by jsonPortAgree's numbering: 0 the Go oracle, 1 Node, 2 release, 3
// the sanitized port, 4 the JavaScript backend. The three parts are the whole between them
// (TestJSONPortPartsAreTheWhole).
var jsonPortPartSides = map[jsonPortPart][]int{
	jsonPortWhole:       {0, 1, 2, 3, 4},
	jsonPortUnsanitized: {0, 1, 2},
	jsonPortSanitized:   {0, 3},
	jsonPortJavaScript:  {0, 4},
}

// jsonPortParts are a heavy case's parts, one owner test each.
var jsonPortParts = []jsonPortPart{jsonPortUnsanitized, jsonPortSanitized, jsonPortJavaScript}

// jsonPortAgree holds the port to the Go oracle over items, as a top shard and a dedicated owner both check them, with
// the sides part names, and returns the Go oracle's answers. With ADAMIC_JSON_ARTIFACTS set, the cases and each side's
// output go under artifactName.
func jsonPortAgree(t *testing.T, items []textCase, binaries jsonPortBinaries, artifactName string, part jsonPortPart) string {
	t.Helper()
	input, _ := protocol(items, make([]answer, len(items)))
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	// Sides 0 to 2 are the Go oracle, Node and release; side 3 on is checked against release too.
	type side struct {
		name, binary           string
		arguments, environment []string
	}
	sanitizedSide := side{"ASan/UBSan", binaries.sanitized, []string{"--cases", path}, []string{"ASAN_OPTIONS=detect_leaks=0"}}
	if runtime.GOOS == "linux" {
		sanitizedSide = side{"ASan/UBSan/LeakSanitizer", binaries.sanitized, []string{"--cases", path}, []string{"ASAN_OPTIONS=detect_leaks=1"}}
	}
	sides := []side{
		{"Go", binaries.oracle, []string{"--cases", path}, nil},
		{"Node", "node", []string{"--disable-warning=ExperimentalWarning", binaries.runner, binaries.entry, "--cases", path}, nil},
		{"release", binaries.release, []string{"--cases", path}, nil},
		sanitizedSide,
		{"JavaScript backend", "node", []string{"--disable-warning=ExperimentalWarning", binaries.runner, binaries.script, "--cases", path}, nil},
	}
	runs := map[int]bool{}
	for _, side := range jsonPortPartSides[part] {
		runs[side] = true
	}
	waves := [][]int{{0, 1, 2, 4}, {3}}
	results := make([]chunkResult, len(sides))
	timings := make([]time.Duration, len(sides))
	// The sanitizer processes retain substantially more memory than the
	// other sides. Run them together after the first wave has exited,
	// so an isolated four-CPU unit does not compete with Node/Go heaps.
	// Within a wave, at most the test's share at once (testguard.Each).
	for _, wave := range waves {
		var present []int
		for _, side := range wave {
			if side < len(sides) && runs[side] {
				present = append(present, side)
			}
		}
		testguard.Each(len(present), func(index int) {
			side := present[index]
			s := sides[side]
			began := time.Now()
			results[side].result, results[side].err = executeResult(s.environment, s.binary, s.arguments...)
			timings[side] = time.Since(began)
		})
	}
	// A sanitized side's time is its run alone, and its wait for the memory budget is said beside it: the wait
	// is other shards' work, the one part of this test's 90 s it doesn't control (#q2qjes8).
	for side, timing := range timings {
		if sides[side].name == "" || !runs[side] {
			continue
		}
		if wait := results[side].result.admissionWait; wait > 0 {
			t.Logf("%s %.3fs; admission waited %.3fs", sides[side].name, (timing - wait).Seconds(), wait.Seconds())
			continue
		}
		t.Logf("%s %.3fs", sides[side].name, timing.Seconds())
	}
	reference := results[0]
	if reference.err != nil || reference.result.exitCode != 0 || len(reference.result.stderr) != 0 {
		t.Fatalf("%s Go oracle: %v exit %d: %s", t.Name(), reference.err, reference.result.exitCode, reference.result.stderr)
	}
	expected := string(reference.result.stdout)
	if len(strings.Split(strings.TrimSuffix(expected, "\n"), "\n")) != len(items) {
		t.Fatalf("%s Go oracle did not answer all %d cases", t.Name(), len(items))
	}
	for side := 1; side < len(sides); side++ {
		if sides[side].name == "" || !runs[side] {
			continue
		}
		if results[side].err != nil {
			t.Fatalf("%s %s: %v", t.Name(), sides[side].name, results[side].err)
		}
		compare(t, t.Name()+" "+sides[side].name, results[side].result, expected, items)
		// The sanitized part runs no release, and needs none: its output equal to the Go oracle's, byte for byte, is
		// equal to release's, which the unsanitized part holds to the same answer.
		if side >= 3 && runs[2] {
			compare(t, t.Name()+" "+sides[side].name+" vs release", results[side].result, string(results[2].result.stdout), items)
		}
	}
	if runtime.GOOS != "linux" && runs[2] {
		report := execute(t, nil, "leaks", "--atExit", "--", binaries.release, "--cases", path)
		if report.exitCode != 0 {
			t.Fatalf("%s leaks: %s", t.Name(), report.stdout)
		}
	}
	if os.Getenv("ADAMIC_JSON_GUARD_CALIBRATE") == "1" && runs[3] {
		// Calibration uses the same per-case Go answers as mandatory comparisons.
		answers, err := jsonPortAnswers(expected, len(items))
		if err != nil {
			t.Fatal(err)
		}
		calibrateNativeCases(t, binaries.sanitized, items, answers)
	}
	if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
		directory := filepath.Join(artifacts, artifactName)
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
		writeJSON(t, filepath.Join(directory, "cases.json"), items)
		for side, s := range sides {
			if !runs[side] {
				continue
			}
			name := strings.NewReplacer("/", "-", " ", "-").Replace(s.name)
			if err := os.WriteFile(filepath.Join(directory, name+".txt"), results[side].result.stdout, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return expected
}

// jsonPortTopShard is a top shard's agreement: the Go oracle, Node, release, the sanitized port and the JavaScript
// backend. The sanitized port runs once, with LeakSanitizer on where it runs (Linux): one detect_leaks=1 run reports
// every ASan/UBSan error a detect_leaks=0 run does, plus leaks, so the separate leak test that ran it again, and the Go
// oracle and release with it, is gone (#fqvxk72). Off Linux, macOS's leaks --atExit checks release instead. With both
// passes in one test, _1345 was 43.4 s alone on one core (ASan 13.6 s, LeakSanitizer 17.8 s; Workshop, Oct 11); with
// the one run it does LeakSanitizer's 17.8 s alone.
func jsonPortTopShard(t *testing.T, target int) {
	portMatchesPrepare(t)
	deadline := portMatchesDeadline(t.Name())
	defer deadline.Stop()
	parentStart := time.Now()
	allCases, allShards := jsonTopCorpus(t)
	cases, shards, dedicated := jsonPortDedicatedPartition(allCases)
	if err := jsonPortDedicatedUnion(allCases, cases, shards, dedicated); err != nil {
		t.Fatal(err)
	}
	enumeratedUnits := len(shards)
	if got := enumeratedUnits; got != testPortMatchesGoCohereShards {
		t.Fatalf("enumerated %d shards, declared %d", got, testPortMatchesGoCohereShards)
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
	oracle := jsonTopOracle(t)
	loweredDir := jsonTopProduct(t, "lowered", jsonLoweredPortInputs(), buildJSONLoweredPort)
	cBytes, err := os.ReadFile(filepath.Join(loweredDir, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	c := string(cBytes)
	releaseDir := jsonTopProduct(t, "release", jsonNativePortInputs(c, false, jsonTopState.clangTools), buildJSONReleasePort(c))
	sanitizedDir := jsonTopProduct(t, "sanitized", jsonNativePortInputs(c, true, jsonTopState.clangTools), buildJSONSanitizedPort(c))
	entry, script := filepath.Join(loweredDir, "main.ts"), filepath.Join(loweredDir, "program.mjs")
	release, sanitized := filepath.Join(releaseDir, "port"), filepath.Join(sanitizedDir, "port")
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
	binaries := jsonPortBinaries{oracle: oracle, entry: entry, script: script, release: release, sanitized: sanitized, runner: runner}
	references := make([]string, len(shards))
	for ordinal, shard := range shards {
		if ordinal != target || ordinal%count != index {
			continue
		}
		func(t *testing.T) {
			start := time.Now()
			items := cases[shard.start:shard.end]
			if len(items) == 0 {
				t.Log("empty hash bucket")
				return
			}
			references[ordinal] = jsonPortAgree(t, items, binaries, fmt.Sprintf("shard-%04d", ordinal), jsonPortWhole)
			t.Logf("case range [%d:%d]; %d cases; execution %.3fs; shared setup waited %.3fs; execution plus locally measured builds %.3fs", shard.start, shard.end, len(items), time.Since(start).Seconds(), setupTime.Seconds(), (time.Since(start) + setupTime).Seconds())
		}(t)
	}
	var setupBeforeShards time.Duration
	t.Cleanup(func() {
		cleanupStart := time.Now()
		defer func() {
			t.Logf("setup outside shards: %.3fs (including local builds)", (setupBeforeShards + time.Since(cleanupStart)).Seconds())
		}()
		allReferences := true
		for ordinal, reference := range references {
			allReferences = allReferences && (reference != "" || shards[ordinal].start == shards[ordinal].end)
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
	// Opt-in measurements reuse the explicitly selected agreement bucket.
	// Benchmark rounds are never independent gate units.
	if os.Getenv("ADAMIC_JSON_BENCH") == "1" {
		for ordinal, shard := range allShards {
			for round := 0; round < 3; round++ {
				unit := ordinal
				if unit != target || unit%count != index {
					continue
				}
				func(t *testing.T) {
					items := allCases[shard.start:shard.end]
					if len(items) == 0 {
						t.Log("empty hash bucket")
						return
					}
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
				}(t)
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
	for _, mutation := range formatterMutants {
		t.Run(mutation.name, func(t *testing.T) {
			t.Parallel()
			directory := portDirectory(t, &mutation)
			entry := filepath.Join(directory, "main.ts")
			binary := filepath.Join(mutantPortProduct(t, mutation, native.Options{Sanitize: true}), "port")
			for _, side := range []struct {
				name   string
				result run
			}{
				{"Node", onNode(t, entry, "--cases", path)},
				{"native", sanitizedRun(t, binary, "--cases", path)},
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
	answers, _ := cohereAnswers(t, cases, false)
	input, expected := protocol(cases, answers)
	path := filepath.Join(t.TempDir(), "cases.txt")
	if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
		path = filepath.Join(artifacts, "boundary.txt")
		writeJSON(t, filepath.Join(artifacts, "boundary.json"), cases)
		writeJSON(t, filepath.Join(artifacts, "boundary-go.json"), answers)
	}
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	source := portDirectory(t, nil)
	entry := filepath.Join(source, "main.ts")
	result := onNode(t, entry, "--cases", path)
	if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
		os.WriteFile(filepath.Join(artifacts, "boundary-node.txt"), result.stdout, 0644)
	}
	compare(t, "Node boundaries", result, expected, cases)
	if os.Getenv("ADAMIC_JSON_NODE_ONLY") != "" {
		t.Skip("debug Node only")
	}
	// The sanitized port is portMatchesProduct's, built ahead; the lowered program is only for leaks off Linux.
	program := lowered(t, entry)
	binary := filepath.Join(portMatchesProduct(t, "sanitized"), "port")
	result = sanitizedRun(t, binary, "--cases", path)
	compare(t, "native boundaries", result, expected, cases)
	if report := leaks(t, program, binary, "--cases", path); report != "" {
		t.Fatal(report)
	}
}

func TestSingleFileStdoutDriver(t *testing.T) {
	t.Parallel()
	cases := []textCase{{"probe.json", `{"text":"é😀","items":[1,2,3]}`}, {"package.json", `{"text":"é😀","items":[1,2,3]}`}}
	answers, _ := cohereAnswers(t, cases, false)
	directory := portDirectory(t, nil)
	entry := filepath.Join(directory, "main.ts")
	program := lowered(t, entry)
	binary := filepath.Join(t.TempDir(), "port")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for index, item := range cases {
		path := filepath.Join(t.TempDir(), item.Name)
		if err := os.WriteFile(path, []byte(item.Text), 0644); err != nil {
			t.Fatal(err)
		}
		compare(t, "Node file driver", onNode(t, entry, path), answers[index].Output, cases)
		compare(t, "native file driver", execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, path), answers[index].Output, cases)
	}
}
