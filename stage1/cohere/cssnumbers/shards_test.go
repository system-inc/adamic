package cssnumbers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/cssnumbers/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var numbersEncode = strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

func numbersInput(texts []string) []byte { return oracles.Input(texts) }

type numbersMutation struct{ name, from, to string }

var numbersMutations = []numbersMutation{
	{"quoted-numbers", "quotedEnd >= 0", "quotedEnd < -1"},
	{"unknown-unit", "unit !== ''", "unit === ''"},
	{"word-prefix", "!match.hasWordPart", "match.hasWordPart || !match.hasWordPart"},
}

type numbersUnit struct {
	name, kind, side string
	lo, hi           int
	ids              []string
}

// Case IDs name enumeration positions, not text hashes: identical texts remain separate cases.
func numbersCaseID(kind string, index int, preference string) string {
	return fmt.Sprintf("%s-%06d-%s", kind, index, preference)
}

func numbersUnits(corpus numbersCorpus) []numbersUnit {
	var units []numbersUnit
	for _, bounds := range oracles.BatchRanges(len(corpus.texts)) {
		lo, hi := bounds[0], bounds[1]
		unit := numbersUnit{name: fmt.Sprintf("batch-%06d-%06d", lo, hi-1), kind: "batch", lo: lo, hi: hi}
		for i := lo; i < hi; i++ {
			for _, pref := range []string{"d", "s"} {
				unit.ids = append(unit.ids, numbersCaseID("batch", i, pref))
			}
		}
		units = append(units, unit)
	}
	for i := range corpus.raw {
		units = append(units, numbersUnit{name: fmt.Sprintf("raw-%06d", i), kind: "raw", lo: i, hi: i + 1,
			ids: []string{numbersCaseID("raw", i, "d"), numbersCaseID("raw", i, "s")}})
	}
	for i, mutation := range numbersMutations {
		units = append(units, numbersUnit{
			name: "mutant-" + mutation.name,
			kind: "mutant",
			lo:   i,
			hi:   i + 1,
			ids:  []string{"mutant-" + mutation.name},
		})
	}
	for _, side := range []string{"Go", "native", "Node", "Prettier"} {
		units = append(units, numbersUnit{
			name: "throughput-" + side,
			kind: "throughput",
			side: side,
			ids:  []string{"throughput-" + side},
		})
	}
	for ordinal := range units {
		units[ordinal].name = fmt.Sprintf("shard-%03d", ordinal)
	}

	return units
}

// Build the unsplit census independently of numbersUnits. Mutant and throughput cases
// each retain their full-corpus checks; the batch and raw preferences are counted individually.
func numbersUnionError(corpus numbersCorpus, units []numbersUnit) error {
	expected := make(map[string]bool)
	for kind, count := range map[string]int{"batch": len(corpus.texts), "raw": len(corpus.raw)} {
		for i := 0; i < count; i++ {
			for _, pref := range []string{"d", "s"} {
				expected[numbersCaseID(kind, i, pref)] = true
			}
		}
	}
	for _, m := range numbersMutations {
		expected["mutant-"+m.name] = true
	}
	for _, side := range []string{"Go", "native", "Node", "Prettier"} {
		expected["throughput-"+side] = true
	}
	seen := make(map[string]string)
	names := make(map[string]bool)
	total := 0
	for _, unit := range units {
		if names[unit.name] {
			return fmt.Errorf("repeated unit %s", unit.name)
		}
		names[unit.name] = true
		var actualIDs []string
		switch unit.kind {
		case "batch", "raw":
			limit := len(corpus.texts)
			if unit.kind == "raw" {
				limit = len(corpus.raw)
			}
			if unit.lo < 0 || unit.hi > limit || unit.lo >= unit.hi {
				return fmt.Errorf("invalid range in %s", unit.name)
			}
			if unit.kind == "raw" && unit.hi != unit.lo+1 {
				return fmt.Errorf("raw unit %s must contain one file", unit.name)
			}
			for i := unit.lo; i < unit.hi; i++ {
				for _, pref := range []string{"d", "s"} {
					actualIDs = append(actualIDs, numbersCaseID(unit.kind, i, pref))
				}
			}
		case "mutant":
			if unit.lo < 0 || unit.lo >= len(numbersMutations) {
				return fmt.Errorf("invalid mutant in %s", unit.name)
			}
			actualIDs = []string{"mutant-" + numbersMutations[unit.lo].name}
		case "throughput":
			actualIDs = []string{"throughput-" + unit.side}
		default:
			return fmt.Errorf("unknown unit kind %q", unit.kind)
		}
		if strings.Join(actualIDs, "\n") != strings.Join(unit.ids, "\n") {
			return fmt.Errorf("case IDs differ from executable range in %s", unit.name)
		}
		for _, id := range unit.ids {
			total++
			if !expected[id] {
				return fmt.Errorf("unexpected case %s in %s", id, unit.name)
			}
			if first, ok := seen[id]; ok {
				return fmt.Errorf("repeated case %s in %s and %s", id, first, unit.name)
			}
			seen[id] = unit.name
		}
	}
	if total != len(expected) {
		return fmt.Errorf("union count %d, unsplit count %d", total, len(expected))
	}
	for id := range expected {
		if _, ok := seen[id]; !ok {
			return fmt.Errorf("missing case %s", id)
		}
	}
	return nil
}

func verifyNumbersUnion(test *testing.T, corpus numbersCorpus, units []numbersUnit) {
	test.Helper()
	if err := numbersUnionError(corpus, units); err != nil {
		test.Fatal(err)
	}
	test.Logf(
		"union: %d batch preferences + %d raw preferences + %d mutants + 4 throughput sides = %d unique case IDs across %d units; equals unsplit enumeration",
		len(corpus.texts)*2,
		len(corpus.raw)*2,
		len(numbersMutations),
		len(corpus.texts)*2+len(corpus.raw)*2+len(numbersMutations)+4,
		len(units),
	)
}

func numbersSelection(value string, units []numbersUnit) ([]numbersUnit, error) {
	if value == "" {
		return units, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("ADAMIC_TEST_SHARD must be zero-based i/n, got %q", value)
	}
	i, errI := strconv.Atoi(parts[0])
	n, errN := strconv.Atoi(parts[1])
	if errI != nil || errN != nil || n <= 0 || i < 0 || i >= n {
		return nil, fmt.Errorf("ADAMIC_TEST_SHARD must satisfy 0 <= i < n, got %q", value)
	}
	var selected []numbersUnit
	for ordinal, unit := range units {
		if ordinal%n == i {
			selected = append(selected, unit)
		}
	}
	return selected, nil
}

func selectedNumbersUnits(test *testing.T, units []numbersUnit) []numbersUnit {
	test.Helper()
	selected, err := numbersSelection(os.Getenv("ADAMIC_TEST_SHARD"), units)
	if err != nil {
		test.Fatal(err)
	}
	test.Logf("selection: %d/%d units, ADAMIC_TEST_SHARD=%q", len(selected), len(units), os.Getenv("ADAMIC_TEST_SHARD"))
	return selected
}

// Every non-Go product uses the shared cache, including temporary mutants.
// The Go oracle is buildcache.GoBuildIn's product (numbersPreparedOracle); products are immutable.
type numbersInputs = oracles.Inputs

func numbersProduct(test *testing.T, inputs numbersInputs, build func(dir string) error) string {
	test.Helper()
	keyed := numbersCacheInputs(test, inputs)
	return buildcache.Product(test, keyed, build)
}

// Temporary generated inputs are keyed by content and logical filename. Their
// directory names never enter a key; repository inputs stay repository-relative.
func numbersCacheInputs(test *testing.T, inputs numbersInputs) buildcache.Inputs {
	return oracles.CacheInputs(test, inputs)
}

// numbersCompilerFiles is compilerinputs.Lowering, the compiler's exact repository paths, as numbersInputs names a
// repository file: through repository, from this package's directory.
func numbersCompilerFiles() []string {
	var files []string
	for _, path := range compilerinputs.Lowering() {
		files = append(files, repository+"/"+path)
	}
	return files
}

type numbersProgram struct{ dir, main, c, javascript string }

// numbersProgramInputs keys the port lowered from main: the compiler, and main with the numbers.ts beside it and the
// cssstrings strings.ts it imports.
func numbersProgramInputs(name, main string) numbersInputs {
	return numbersInputs{
		Name: name + " lowering and backends", Files: append(
			numbersCompilerFiles(),
			main,
			filepath.Join(filepath.Dir(main), "numbers.ts"),
			filepath.Join(filepath.Dir(main), "../cssstrings/strings.ts"),
		),
		Toolchain: runtime.Version() + "; checker " + "d92d9bfee114c80be2c375d72edae966176e3a4f",
	}
}

func buildNumbersProgram(test *testing.T, name, main string) numbersProgram {
	test.Helper()
	dir := numbersProduct(test, numbersProgramInputs(name, main), func(dir string) error {
		loaded, err := load.Load([]string{main})
		if err != nil {
			return err
		}
		program, err := lower.Lower(context.Background(), loaded)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	return numbersProgram{
		dir:        dir,
		main:       main,
		c:          filepath.Join(dir, "program.c"),
		javascript: filepath.Join(dir, "program.mjs"),
	}
}

// numbersClangToolchain is clang by its report, without where this machine installed it (#t37sw0f).
func numbersClangToolchain(test *testing.T) string {
	test.Helper()
	// Read through childguard.Capture for numbersExecute's reason: a WaitDelay could fail a run that had finished.
	command, ctx := numbersSetupCommand(test, "clang", "--version")
	if _, _, err := childguard.Capture(ctx, command); err != nil {
		test.Fatal(err)
	}
	return buildcache.Tool("clang", "--version")
}

// numbersNativeInputs keys the native binary native.Build makes from c: the compiler, whose native.Build links more
// than internal/native (internal/ir, internal/flow, internal/toolchain, bridge/tsgo), and c by content.
func numbersNativeInputs(name, c string, options native.Options, clang string) numbersInputs {
	return numbersInputs{
		Name:      name,
		Files:     append(numbersCompilerFiles(), c),
		Flags:     native.Flags(options),
		Toolchain: clang,
	}
}

func buildNumbersNative(test *testing.T, name string, program numbersProgram, sanitize bool) string {
	test.Helper()
	options := native.Options{Sanitize: sanitize}
	dir := numbersProduct(test, numbersNativeInputs(
		name,
		program.c,
		options,
		numbersClangToolchain(test),
	), func(dir string) error {
		source, err := os.ReadFile(program.c)
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(dir, "port"), options)
	})
	return filepath.Join(dir, "port")
}

func prepareNumbersMutant(test *testing.T, mutation numbersMutation) string {
	test.Helper()
	parent := numbersTopTempDir(test)
	scratch := filepath.Join(parent, "cssnumbers")
	if err := os.Mkdir(scratch, 0755); err != nil {
		test.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(parent, "cssstrings"), 0755); err != nil {
		test.Fatal(err)
	}
	dependency, err := os.ReadFile("../cssstrings/strings.ts")
	if err != nil {
		test.Fatal(err)
	}
	write(test, filepath.Join(parent, "cssstrings/strings.ts"), dependency)
	source, err := os.ReadFile("numbers.ts")
	if err != nil {
		test.Fatal(err)
	}
	if strings.Count(string(source), mutation.from) != 1 {
		test.Fatal("mutant must change one site")
	}
	write(
		test,
		filepath.Join(scratch, "numbers.ts"),
		[]byte(strings.Replace(string(source), mutation.from, mutation.to, 1)),
	)
	source, err = os.ReadFile("main.ts")
	if err != nil {
		test.Fatal(err)
	}
	main := filepath.Join(scratch, "main.ts")
	write(test, main, source)
	return main
}

func numbersSanitizerEnvironment(leak bool) []string {
	if runtime.GOOS != "linux" {
		return nil
	}
	if leak {
		return []string{"ASAN_OPTIONS=detect_leaks=1"}
	}
	return []string{"ASAN_OPTIONS=detect_leaks=0"}
}

func numbersLeaks(test *testing.T, sanitized, unsanitized string, args ...string) {
	test.Helper()
	switch runtime.GOOS {
	case "linux":
		r := numbersExecute(test, numbersSanitizerEnvironment(true), sanitized, args...)
		if r.exitCode != 0 {
			test.Fatalf("leaks: exit %d\n%s", r.exitCode, r.stderr)
		}
	case "darwin":
		r := numbersExecute(test, nil, "leaks", append([]string{"--atExit", "--", unsanitized}, args...)...)
		if r.exitCode != 0 {
			test.Fatalf("leaks: %s", r.stdout)
		}
	default:
		test.Fatalf("no leak check for %s", runtime.GOOS)
	}
}

func numbersOutputError(unit numbersUnit, side string, got, want []byte) error {
	lines := func(output []byte) [][]byte {
		if len(output) == 0 {
			return nil
		}
		return bytes.Split(bytes.TrimSuffix(output, []byte("\n")), []byte("\n"))
	}
	a, b := lines(got), lines(want)
	if len(a) != len(unit.ids) || len(b) != len(unit.ids) {
		return fmt.Errorf(
			"%s %s: output count got %d want %d census %d",
			unit.name,
			side,
			len(a),
			len(b),
			len(unit.ids),
		)
	}
	for i := range b {
		if !bytes.Equal(a[i], b[i]) {
			return fmt.Errorf("%s %s: disagreement in case %s", unit.name, side, unit.ids[i])
		}
	}
	// Preserve the original byte-for-byte check, including the final newline.
	if !bytes.Equal(got, want) {
		return fmt.Errorf("%s %s: raw stdout disagreement", unit.name, side)
	}
	return nil
}

func checkNumbersOutput(test *testing.T, unit numbersUnit, side string, got, want []byte) {
	test.Helper()
	if err := numbersOutputError(unit, side, got, want); err != nil {
		test.Fatal(err)
	}
}

func TestCSSNumbersUnion(test *testing.T) {
	test.Parallel()
	corpus := enumerateNumbers(test)
	units := numbersUnits(corpus)
	verifyNumbersUnion(test, corpus, units)
	if err := numbersUnionError(corpus, units[1:]); err == nil {
		test.Fatal("missing range survived")
	}
	repeated := append(append([]numbersUnit{}, units...), units[0])
	repeated[len(repeated)-1].name = "repeated-range"
	if err := numbersUnionError(corpus, repeated); err == nil || !strings.Contains(err.Error(), "repeated case") {
		test.Fatalf("repeated range: %v", err)
	}
	// Gate selections partition every unit once, even when n exceeds the unit count.
	for _, n := range []int{1, 2, 7, len(units) + 1} {
		var union []numbersUnit
		for i := 0; i < n; i++ {
			selected, err := numbersSelection(fmt.Sprintf("%d/%d", i, n), units)
			if err != nil {
				test.Fatal(err)
			}
			union = append(union, selected...)
		}
		if err := numbersUnionError(corpus, union); err != nil {
			test.Fatal(err)
		}
	}
	for _, bad := range []string{"0/0", "-1/2", "2/2", "1", "a/2", "1/2/3"} {
		if _, err := numbersSelection(bad, units); err == nil {
			test.Fatalf("accepted %q", bad)
		}
	}
}

// The child uses the same comparator and batch-unit IDs as the real oracle checks.
// A disagreement is planted in exactly one preference case. All units execute;
// the parent requires exactly one failed unit, with its case ID and shard name.
func TestCSSNumbersPlantedDisagreement(test *testing.T) {
	test.Parallel()
	corpus := enumerateNumbers(test)
	units := numbersUnits(corpus)
	verifyNumbersUnion(test, corpus, units)
	const planted = "batch-002049-s"
	var owner string
	for _, unit := range units {
		for _, id := range unit.ids {
			if id == planted {
				owner = unit.name
			}
		}
	}
	if owner == "" {
		test.Fatal("planted case absent")
	}
	if os.Getenv("ADAMIC_CSSNUMBERS_DISAGREEMENT_PROBE") == "1" {
		for _, unit := range units {
			if unit.kind != "batch" {
				continue
			}
			test.Run(unit.name, func(test *testing.T) {
				test.Parallel()
				var want, got strings.Builder
				for _, id := range unit.ids {
					want.WriteString(id + "\n")
					if id == planted {
						got.WriteString("planted disagreement\n")
					} else {
						got.WriteString(id + "\n")
					}
				}
				checkNumbersOutput(test, unit, "planted", []byte(got.String()), []byte(want.String()))
			})
		}
		return
	}
	binary, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	r := numbersExecute(
		test,
		[]string{"ADAMIC_CSSNUMBERS_DISAGREEMENT_PROBE=1"},
		binary,
		"-test.run=^TestCSSNumbersPlantedDisagreement$",
		"-test.v",
		"-test.parallel=4",
	)
	if r.exitCode != 1 || len(r.stderr) != 0 {
		test.Fatalf("probe exit %d stderr %s", r.exitCode, r.stderr)
	}
	failedPrefix := "--- FAIL: TestCSSNumbersPlantedDisagreement/"
	if strings.Count(string(r.stdout), failedPrefix) != 1 ||
		!strings.Contains(string(r.stdout), failedPrefix+owner+" (") ||
		!strings.Contains(string(r.stdout), "disagreement in case "+planted) {
		test.Fatalf("expected only shard %s to catch %s:\n%s", owner, planted, r.stdout)
	}
	test.Logf("planted disagreement %s caught by exactly shard %s", planted, owner)
}

// Oracle answers are immutable inputs; throughput still executes every original round.
func numbersOracleAnswers(test *testing.T, name, command string, arguments []string, data []byte, identity string) run {
	return numbersOracleAnswersWithCommand(test, name, command, arguments, data, identity, numbersCommand)
}

func numbersSetupOracleAnswers(
	test *testing.T,
	name, command string,
	arguments []string,
	data []byte,
	identity string,
) run {
	return numbersOracleAnswersWithCommand(test, name, command, arguments, data, identity, numbersSetupCommand)
}

func numbersOracleAnswersWithCommand(
	test *testing.T,
	name, command string,
	arguments []string,
	data []byte,
	identity string,
	makeCommand func(*testing.T, string, ...string) (*exec.Cmd, context.Context),
) run {
	test.Helper()
	if name == "Go" {
		return run{stdout: oracles.FullAnswers(test, command, arguments, data, identity)}
	}
	dir := numbersProduct(test, numbersInputs{
		Name:      name + " oracle answers",
		Files:     []string{"testdata/library.mjs"},
		Flags:     []string{identity, fmt.Sprintf("input-sha256=%x", sha256.Sum256(data))},
		Toolchain: buildcache.Tool("node", "--version"),
	}, func(dir string) error {
		path := filepath.Join(dir, "input.txt")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
		// Read through childguard.Capture for numbersExecute's reason: a WaitDelay could fail a run that had
		// finished. A run made by numbersCommand ends at its 90 s deadline; one made by numbersSetupCommand has
		// none, and the unit's own ceiling bounds it.
		cmd, ctx := makeCommand(test, command, append(append([]string{}, arguments...), path)...)
		stdout, stderr, err := childguard.Capture(ctx, cmd)
		if err != nil || len(stderr) != 0 {
			return fmt.Errorf("%s oracle: %v; %s", name, err, stderr)
		}
		return os.WriteFile(filepath.Join(dir, "answers.txt"), stdout, 0644)
	})
	answer, err := os.ReadFile(filepath.Join(dir, "answers.txt"))
	if err != nil {
		test.Fatal(err)
	}
	return run{stdout: answer}
}

func numbersOracleIdentity(test *testing.T, path string) string { return oracles.Identity(test, path) }
