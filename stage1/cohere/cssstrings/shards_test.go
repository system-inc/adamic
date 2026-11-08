package cssstrings

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/cssstrings/oracles"
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

var stringsEncode = strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

func stringsInput(texts []string) []byte { return oracles.Input(texts) }

type stringsMutation struct{ name, from, to string }

var stringsMutations = []stringsMutation{
	{"quote-tie", "preferredCount > alternateCount", "preferredCount >= alternateCount"},
	{"escaped-closing-quote", "character === 92", "character === 0"},
	{"preserve-original-escape", "raw.slice(0, 1) === quote", "raw.length === 0"},
}

type stringsUnit struct {
	name, kind, side string
	lo, hi           int
	ids              []string
}

// Case IDs name enumeration positions, not text hashes: identical texts remain separate cases.
func stringsCaseId(kind string, index int, preference string) string {
	return fmt.Sprintf("%s-%06d-%s", kind, index, preference)
}

func stringsUnits(corpus stringsCorpus) []stringsUnit {
	var units []stringsUnit
	for _, bounds := range oracles.BatchRanges(len(corpus.texts)) {
		lo, hi := bounds[0], bounds[1]
		unit := stringsUnit{name: fmt.Sprintf("batch-%06d-%06d", lo, hi-1), kind: "batch", lo: lo, hi: hi}
		for index := lo; index < hi; index++ {
			for _, pref := range []string{"d", "s"} {
				unit.ids = append(unit.ids, stringsCaseId("batch", index, pref))
			}
		}
		units = append(units, unit)
	}
	for index := range corpus.raw {
		units = append(units, stringsUnit{name: fmt.Sprintf("raw-%06d", index), kind: "raw", lo: index, hi: index + 1,
			ids: []string{stringsCaseId("raw", index, "d"), stringsCaseId("raw", index, "s")}})
	}
	for index, mutation := range stringsMutations {
		units = append(units, stringsUnit{
			name: "mutant-" + mutation.name,
			kind: "mutant",
			lo:   index,
			hi:   index + 1,
			ids:  []string{"mutant-" + mutation.name},
		})
	}
	for _, side := range []string{"Go", "native", "Node", "Prettier"} {
		units = append(units, stringsUnit{
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

// Build the unsplit census independently of stringsUnits. Mutant and throughput cases
// each retain their full-corpus checks; the batch and raw preferences are counted individually.
func stringsUnionError(corpus stringsCorpus, units []stringsUnit) error {
	expected := make(map[string]bool)
	for kind, count := range map[string]int{"batch": len(corpus.texts), "raw": len(corpus.raw)} {
		for index := 0; index < count; index++ {
			for _, pref := range []string{"d", "s"} {
				expected[stringsCaseId(kind, index, pref)] = true
			}
		}
	}
	for _, stringsMutation := range stringsMutations {
		expected["mutant-"+stringsMutation.name] = true
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
		var actualIds []string
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
			for index := unit.lo; index < unit.hi; index++ {
				for _, pref := range []string{"d", "s"} {
					actualIds = append(actualIds, stringsCaseId(unit.kind, index, pref))
				}
			}
		case "mutant":
			if unit.lo < 0 || unit.lo >= len(stringsMutations) {
				return fmt.Errorf("invalid mutant in %s", unit.name)
			}
			actualIds = []string{"mutant-" + stringsMutations[unit.lo].name}
		case "throughput":
			actualIds = []string{"throughput-" + unit.side}
		default:
			return fmt.Errorf("unknown unit kind %q", unit.kind)
		}
		if strings.Join(actualIds, "\n") != strings.Join(unit.ids, "\n") {
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

func verifyStringsUnion(test *testing.T, corpus stringsCorpus, units []stringsUnit) {
	test.Helper()
	if err := stringsUnionError(corpus, units); err != nil {
		test.Fatal(err)
	}
	test.Logf(
		"union: %d batch preferences + %d raw preferences + %d mutants + 4 throughput sides = %d unique case IDs across %d units; equals unsplit enumeration",
		len(corpus.texts)*2,
		len(corpus.raw)*2,
		len(stringsMutations),
		len(corpus.texts)*2+len(corpus.raw)*2+len(stringsMutations)+4,
		len(units),
	)
}

func stringsSelection(value string, units []stringsUnit) ([]stringsUnit, error) {
	if value == "" {
		return units, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("ADAMIC_TEST_SHARD must be zero-based i/n, got %q", value)
	}
	index, iError := strconv.Atoi(parts[0])
	n, nError := strconv.Atoi(parts[1])
	if iError != nil || nError != nil || n <= 0 || index < 0 || index >= n {
		return nil, fmt.Errorf("ADAMIC_TEST_SHARD must satisfy 0 <= i < n, got %q", value)
	}
	var selected []stringsUnit
	for ordinal, unit := range units {
		if ordinal%n == index {
			selected = append(selected, unit)
		}
	}
	return selected, nil
}

func selectedStringsUnits(test *testing.T, units []stringsUnit) []stringsUnit {
	test.Helper()
	selected, stringsSelectionError := stringsSelection(os.Getenv("ADAMIC_TEST_SHARD"), units)
	if stringsSelectionError != nil {
		test.Fatal(stringsSelectionError)
	}
	test.Logf("selection: %d/%d units, ADAMIC_TEST_SHARD=%q", len(selected), len(units), os.Getenv("ADAMIC_TEST_SHARD"))
	return selected
}

// Every non-Go product uses the shared cache, including temporary mutants.
// The Go oracle is a buildcache.GoBuildIn product; products are immutable.
type stringsInputs = oracles.Inputs

func stringsProduct(test *testing.T, inputs stringsInputs, build func(dir string) error) string {
	test.Helper()
	keyed := stringsCacheInputs(test, inputs)
	return buildcache.Product(test, keyed, build)
}

// Temporary generated inputs are keyed by content and logical filename. Their
// directory names never enter a key; repository inputs stay repository-relative.
func stringsCacheInputs(test *testing.T, inputs stringsInputs) buildcache.Inputs {
	return oracles.CacheInputs(test, inputs)
}

type stringsProgram struct{ dir, main, c, javascript string }

// stringsLowering is compilerinputs.Lowering as stringsInputs names a file, relative to this package's directory.
func stringsLowering() []string {
	var files []string
	for _, entry := range compilerinputs.Lowering() {
		files = append(files, repository+"/"+entry)
	}
	return files
}

// stringsProgramInputs keys the port lowered from main: the compiler that lowers it (compilerinputs.Lowering) and the
// port's two .ts files, main and the strings.ts beside it.
func stringsProgramInputs(name, main string) stringsInputs {
	return stringsInputs{
		Name:      name + " lowering and backends",
		Files:     append(stringsLowering(), main, filepath.Join(filepath.Dir(main), "strings.ts")),
		Toolchain: runtime.Version() + "; checker " + "d92d9bfee114c80be2c375d72edae966176e3a4f",
	}
}

func buildStringsProgram(test *testing.T, name, main string) stringsProgram {
	test.Helper()
	dir := stringsProduct(test, stringsProgramInputs(name, main), func(dir string) error {
		loaded, loadError := load.Load([]string{main})
		if loadError != nil {
			return loadError
		}
		program, lowerError := lower.Lower(context.Background(), loaded)
		if lowerError != nil {
			return lowerError
		}
		if writeError := os.WriteFile(
			filepath.Join(dir, "program.c"),
			[]byte(native.C(program)),
			0644,
		); writeError != nil {
			return writeError
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	return stringsProgram{
		dir:        dir,
		main:       main,
		c:          filepath.Join(dir, "program.c"),
		javascript: filepath.Join(dir, "program.mjs"),
	}
}

// stringsClangToolchain is clang by its report, without where this machine installed it (#t37sw0f).
func stringsClangToolchain(test *testing.T) string {
	test.Helper()
	// Read through childguard.Capture for stringsExecute's reason: a WaitDelay could fail a run that had finished.
	command, ctx := stringsSetupCommand(test, "clang", "--version")
	if _, _, captureError := childguard.Capture(ctx, command); captureError != nil {
		test.Fatal(captureError)
	}
	return buildcache.Tool("clang", "--version")
}

// stringsNativeInputs keys the native port built from programC: the compiler that builds it (compilerinputs.Lowering,
// which holds internal/native) and the C itself.
func stringsNativeInputs(test *testing.T, name, programC string, sanitize bool) stringsInputs {
	test.Helper()
	return stringsInputs{
		Name:      name,
		Files:     append(stringsLowering(), programC),
		Flags:     native.Flags(native.Options{Sanitize: sanitize}),
		Toolchain: stringsClangToolchain(test),
	}
}

func buildStringsNative(test *testing.T, name string, program stringsProgram, sanitize bool) string {
	test.Helper()
	dir := stringsProduct(test, stringsNativeInputs(test, name, program.c, sanitize), func(dir string) error {
		source, readError := os.ReadFile(program.c)
		if readError != nil {
			return readError
		}
		return native.Build(string(source), filepath.Join(dir, "port"), native.Options{Sanitize: sanitize})
	})
	return filepath.Join(dir, "port")
}

// An edit to a package the compiler links must move every lowered and native product's key, and an edit to tsc's
// baselines, which nothing links, must keep it.
func TestCssStringsCompilerProductsKeyWhatTheCompilerLinks(test *testing.T) {
	test.Parallel()
	main, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	// A native product keys its C by content, so a stand-in C file keys it without lowering anything.
	programC := filepath.Join(test.TempDir(), "program.c")
	write(test, programC, nil)
	products := []stringsInputs{
		stringsProgramInputs("port", main),
		stringsNativeInputs(test, "native-fast", programC, false),
		stringsNativeInputs(test, "sanitized", programC, true),
	}
	for _, mutation := range stringsMutations {
		products = append(
			products,
			stringsProgramInputs(mutation.name, prepareStringsMutant(test, mutation)),
			stringsNativeInputs(test, mutation.name+"-sanitized", programC, true),
		)
	}
	covers := func(entry, path string) bool { return entry == path || strings.HasPrefix(path, entry+"/") }
	for _, product := range products {
		keyed := stringsCacheInputs(test, product)
		linked, baselines := false, false
		for _, entry := range keyed.Files {
			linked = linked || covers(entry, "cohere/TypeScript/tsc/internal/checker")
			baselines = baselines || covers(entry, "cohere/TypeScript/tsc/testdata")
		}
		if !linked {
			test.Errorf(
				"%s: no Files entry covers cohere/TypeScript/tsc/internal/checker, a package the compiler links",
				keyed.Name,
			)
		}
		if baselines {
			test.Errorf("%s: a Files entry covers cohere/TypeScript/tsc/testdata, baselines nothing links", keyed.Name)
		}
	}
}

func prepareStringsMutant(test *testing.T, mutation stringsMutation) string {
	test.Helper()
	// stringsCacheInputs keys a generated .ts by its directory's name and its own, so the mutant sits in cssstrings/, a
	// name every checkout gives it, never in the temporary directory's random one (#sgemgmn).
	scratch := filepath.Join(stringsTopTempDir(test), "cssstrings")
	if createDirectoryError := os.Mkdir(scratch, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	source, readError := os.ReadFile("strings.ts")
	if readError != nil {
		test.Fatal(readError)
	}
	if strings.Count(string(source), mutation.from) != 1 {
		test.Fatal("mutant must change one site")
	}
	write(
		test,
		filepath.Join(scratch, "strings.ts"),
		[]byte(strings.Replace(string(source), mutation.from, mutation.to, 1)),
	)
	source, readError = os.ReadFile("main.ts")
	if readError != nil {
		test.Fatal(readError)
	}
	main := filepath.Join(scratch, "main.ts")
	write(test, main, source)
	return main
}

func stringsSanitizerEnvironment(leak bool) []string {
	if runtime.GOOS != "linux" {
		return nil
	}
	if leak {
		return []string{"ASAN_OPTIONS=detect_leaks=1"}
	}
	return []string{"ASAN_OPTIONS=detect_leaks=0"}
}

func stringsLeaks(test *testing.T, sanitized, unsanitized string, args ...string) {
	test.Helper()
	switch runtime.GOOS {
	case "linux":
		run := stringsExecute(test, stringsSanitizerEnvironment(true), sanitized, args...)
		if run.exitCode != 0 {
			test.Fatalf("leaks: exit %d\n%s", run.exitCode, run.stderr)
		}
	case "darwin":
		run := stringsExecute(test, nil, "leaks", append([]string{"--atExit", "--", unsanitized}, args...)...)
		if run.exitCode != 0 {
			test.Fatalf("leaks: %s", run.stdout)
		}
	default:
		test.Fatalf("no leak check for %s", runtime.GOOS)
	}
}

func stringsOutputError(unit stringsUnit, side string, got, want []byte) error {
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
	for index := range b {
		if !bytes.Equal(a[index], b[index]) {
			return fmt.Errorf("%s %s: disagreement in case %s", unit.name, side, unit.ids[index])
		}
	}
	// Preserve the original byte-for-byte check, including the final newline.
	if !bytes.Equal(got, want) {
		return fmt.Errorf("%s %s: raw stdout disagreement", unit.name, side)
	}
	return nil
}

func checkStringsOutput(test *testing.T, unit stringsUnit, side string, got, want []byte) {
	test.Helper()
	if err := stringsOutputError(unit, side, got, want); err != nil {
		test.Fatal(err)
	}
}

func TestCssStringsUnion(test *testing.T) {
	test.Parallel()
	corpus := enumerateStrings(test)
	units := stringsUnits(corpus)
	verifyStringsUnion(test, corpus, units)
	if err := stringsUnionError(corpus, units[1:]); err == nil {
		test.Fatal("missing range survived")
	}
	repeated := append(append([]stringsUnit{}, units...), units[0])
	repeated[len(repeated)-1].name = "repeated-range"
	if err := stringsUnionError(corpus, repeated); err == nil || !strings.Contains(err.Error(), "repeated case") {
		test.Fatalf("repeated range: %v", err)
	}
	// Gate selections partition every unit once, even when n exceeds the unit count.
	for _, n := range []int{1, 2, 7, len(units) + 1} {
		var union []stringsUnit
		for index := 0; index < n; index++ {
			selected, stringsSelectionError := stringsSelection(fmt.Sprintf("%d/%d", index, n), units)
			if stringsSelectionError != nil {
				test.Fatal(stringsSelectionError)
			}
			union = append(union, selected...)
		}
		if err := stringsUnionError(corpus, union); err != nil {
			test.Fatal(err)
		}
	}
	for _, bad := range []string{"0/0", "-1/2", "2/2", "1", "a/2", "1/2/3"} {
		if _, stringsSelectionError := stringsSelection(bad, units); stringsSelectionError == nil {
			test.Fatalf("accepted %q", bad)
		}
	}
}

// The child uses the same comparator and batch-unit IDs as the real oracle checks.
// A disagreement is planted in exactly one preference case. All units execute;
// the parent requires exactly one failed unit, with its case ID and shard name.
func TestCSSStringsPlantedDisagreement(test *testing.T) {
	test.Parallel()
	corpus := enumerateStrings(test)
	units := stringsUnits(corpus)
	verifyStringsUnion(test, corpus, units)
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
	if os.Getenv("ADAMIC_CSSSTRINGS_DISAGREEMENT_PROBE") == "1" {
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
				checkStringsOutput(test, unit, "planted", []byte(got.String()), []byte(want.String()))
			})
		}
		return
	}
	binary, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	run := stringsExecute(
		test,
		[]string{"ADAMIC_CSSSTRINGS_DISAGREEMENT_PROBE=1"},
		binary,
		"-test.run=^TestCSSStringsPlantedDisagreement$",
		"-test.v",
		"-test.parallel=4",
	)
	if run.exitCode != 1 || len(run.stderr) != 0 {
		test.Fatalf("probe exit %d stderr %s", run.exitCode, run.stderr)
	}
	failedPrefix := "--- FAIL: TestCSSStringsPlantedDisagreement/"
	if strings.Count(string(run.stdout), failedPrefix) != 1 ||
		!strings.Contains(string(run.stdout), failedPrefix+owner+" (") ||
		!strings.Contains(string(run.stdout), "disagreement in case "+planted) {
		test.Fatalf("expected only shard %s to catch %s:\n%s", owner, planted, run.stdout)
	}
	test.Logf("planted disagreement %s caught by exactly shard %s", planted, owner)
}

// Oracle answers are immutable inputs; throughput still executes every original round.
func stringsOracleAnswers(test *testing.T, name, command string, arguments []string, data []byte, identity string) run {
	return stringsOracleAnswersWithCommand(test, name, command, arguments, data, identity, stringsCommand)
}

func stringsSetupOracleAnswers(
	test *testing.T,
	name, command string,
	arguments []string,
	data []byte,
	identity string,
) run {
	return stringsOracleAnswersWithCommand(test, name, command, arguments, data, identity, stringsSetupCommand)
}

func stringsOracleAnswersWithCommand(
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
	dir := stringsProduct(test, stringsInputs{
		Name:      name + " oracle answers",
		Files:     []string{"testdata/library.mjs"},
		Flags:     []string{identity, fmt.Sprintf("input-sha256=%x", sha256.Sum256(data))},
		Toolchain: buildcache.Tool("node", "--version"),
	}, func(dir string) error {
		path := filepath.Join(dir, "input.txt")
		if writeError := os.WriteFile(path, data, 0644); writeError != nil {
			return writeError
		}
		// Read through childguard.Capture for stringsExecute's reason: a WaitDelay could fail a run that had
		// finished. A run made by stringsCommand ends at its 90 s deadline; one made by stringsSetupCommand has
		// none, and the unit's own ceiling bounds it.
		osCommand, ctx := makeCommand(test, command, append(append([]string{}, arguments...), path)...)
		stdout, stderr, captureError := childguard.Capture(ctx, osCommand)
		if captureError != nil || len(stderr) != 0 {
			return fmt.Errorf("%s oracle: %v; %s", name, captureError, stderr)
		}
		return os.WriteFile(filepath.Join(dir, "answers.txt"), stdout, 0644)
	})
	answer, readError := os.ReadFile(filepath.Join(dir, "answers.txt"))
	if readError != nil {
		test.Fatal(readError)
	}
	return run{stdout: answer}
}

func stringsOracleIdentity(test *testing.T, path string) string { return oracles.Identity(test, path) }
