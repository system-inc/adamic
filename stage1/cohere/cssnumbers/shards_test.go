package cssnumbers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var numbersEncode = strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

func numbersInput(texts []string) []byte {
	var input strings.Builder
	for _, text := range texts {
		for _, preference := range []string{"d", "s"} {
			input.WriteString(preference + numbersEncode.Replace(text) + "\n")
		}
	}
	return []byte(input.String())
}

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
	const batchTexts = 512
	for lo := 0; lo < len(corpus.texts); lo += batchTexts {
		hi := min(lo+batchTexts, len(corpus.texts))
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
		units = append(units, numbersUnit{name: "mutant-" + mutation.name, kind: "mutant", lo: i, hi: i + 1, ids: []string{"mutant-" + mutation.name}})
	}
	for _, side := range []string{"Go", "native", "Node", "Prettier"} {
		units = append(units, numbersUnit{name: "throughput-" + side, kind: "throughput", side: side, ids: []string{"throughput-" + side}})
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

func verifyNumbersUnion(t *testing.T, corpus numbersCorpus, units []numbersUnit) {
	t.Helper()
	if err := numbersUnionError(corpus, units); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d batch preferences + %d raw preferences + %d mutants + 4 throughput sides = %d unique case IDs across %d units; equals unsplit enumeration",
		len(corpus.texts)*2, len(corpus.raw)*2, len(numbersMutations), len(corpus.texts)*2+len(corpus.raw)*2+len(numbersMutations)+4, len(units))
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

func selectedNumbersUnits(t *testing.T, units []numbersUnit) []numbersUnit {
	t.Helper()
	selected, err := numbersSelection(os.Getenv("ADAMIC_TEST_SHARD"), units)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("selection: %d/%d units, ADAMIC_TEST_SHARD=%q", len(selected), len(units), os.Getenv("ADAMIC_TEST_SHARD"))
	return selected
}

// Every non-Go product uses the shared cache, including temporary mutants.
// Go overlay builds stay private until GoBuild supports them; products are immutable.
type numbersInputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain string
}

func numbersProduct(t *testing.T, inputs numbersInputs, build func(dir string) error) string {
	t.Helper()
	keyed := numbersCacheInputs(t, inputs)
	return buildcache.Product(t, keyed, build)
}

// Go builds have no hand-listed key. Keep their private callback until GoBuild
// supports these oracle overlays; it is invoked once per parent.
func numbersGoProduct(t *testing.T, name string, build func(dir string) error) string {
	t.Helper()
	dir := t.TempDir()
	start := time.Now()
	if err := build(dir); err != nil {
		t.Fatalf("Go build %s: %v", name, err)
	}
	t.Logf("Go build %s %.6fs (private, awaiting GoBuild)", name, time.Since(start).Seconds())
	return dir
}

// Temporary generated inputs are keyed by content and logical filename. Their
// directory names never enter a key; repository inputs stay repository-relative.
func numbersCacheInputs(t *testing.T, inputs numbersInputs) buildcache.Inputs {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	keyed := buildcache.Inputs{Name: "cssnumbers " + inputs.Name, Flags: append([]string{}, inputs.Flags...), Toolchain: []string{inputs.Toolchain, buildcache.Tool("go", "version"), runtime.GOOS + "/" + runtime.GOARCH}}
	keyed.Files = append(keyed.Files, "stage1/cohere/cssnumbers/shards_test.go", "stage1/cohere/cssnumbers/port_test.go")
	keyed.Flags = append(keyed.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))

	for _, name := range inputs.Files {
		absolute, err := filepath.Abs(name)
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			t.Fatal(err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			content, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			logical := filepath.Base(name)
			if filepath.Ext(name) == ".ts" {
				logical = filepath.Base(filepath.Dir(name)) + "/" + logical
			}
			keyed.Flags = append(keyed.Flags, fmt.Sprintf("generated-input=%s:%x", logical, sha256.Sum256(content)))
		} else {
			keyed.Files = append(keyed.Files, filepath.ToSlash(relative))
		}
	}
	if strings.Contains(inputs.Name, "lowering and backends") {
		// The compiler's transitive Go dependencies include cohere's rule runner.
		// GoInputs resolves their files, module pins, build environment and tools.
		compiler, err := buildcache.GoInputs("css lowering compiler", "./cmd/adamic", nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		keyed.Files = append(keyed.Files, compiler.Files...)
		keyed.Flags = append(keyed.Flags, compiler.Flags...)
		keyed.Toolchain = append(keyed.Toolchain, compiler.Toolchain...)
	}
	return keyed
}

type numbersProgram struct{ dir, main, c, javascript string }

func buildNumbersProgram(t *testing.T, name, main string) numbersProgram {
	t.Helper()
	dir := numbersProduct(t, numbersInputs{
		Name: name + " lowering and backends", Files: []string{main, filepath.Join(filepath.Dir(main), "numbers.ts"), filepath.Join(filepath.Dir(main), "../cssstrings/strings.ts"), repository + "/internal", repository + "/go.mod", repository + "/cohere/TypeScript", repository + "/cohere/TypeScript-shim"},
		Toolchain: runtime.Version() + "; checker " + "d92d9bfee114c80be2c375d72edae966176e3a4f",
	}, func(dir string) error {
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
	return numbersProgram{dir: dir, main: main, c: filepath.Join(dir, "program.c"), javascript: filepath.Join(dir, "program.mjs")}
}

func numbersClangToolchain(t *testing.T) string {
	t.Helper()
	version := execute(t, nil, "clang", "--version")
	clean(t, "clang version", version)
	return strings.TrimSpace(string(version.stdout))
}

func buildNumbersNative(t *testing.T, name string, program numbersProgram, sanitize bool) string {
	t.Helper()
	options := native.Options{Sanitize: sanitize}
	flags := native.Flags(options)
	dir := numbersProduct(t, numbersInputs{Name: name, Files: []string{program.c, repository + "/internal/native", repository + "/go.mod"}, Flags: flags, Toolchain: numbersClangToolchain(t)}, func(dir string) error {
		source, err := os.ReadFile(program.c)
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(dir, "port"), options)
	})
	return filepath.Join(dir, "port")
}

func prepareNumbersMutant(t *testing.T, mutation numbersMutation) string {
	t.Helper()
	parent := t.TempDir()
	scratch := filepath.Join(parent, "cssnumbers")
	if err := os.Mkdir(scratch, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(parent, "cssstrings"), 0755); err != nil {
		t.Fatal(err)
	}
	dependency, err := os.ReadFile("../cssstrings/strings.ts")
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(parent, "cssstrings/strings.ts"), dependency)
	source, err := os.ReadFile("numbers.ts")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(source), mutation.from) != 1 {
		t.Fatal("mutant must change one site")
	}
	write(t, filepath.Join(scratch, "numbers.ts"), []byte(strings.Replace(string(source), mutation.from, mutation.to, 1)))
	source, err = os.ReadFile("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(scratch, "main.ts")
	write(t, main, source)
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

func numbersLeaks(t *testing.T, sanitized, unsanitized string, args ...string) {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		r := execute(t, numbersSanitizerEnvironment(true), sanitized, args...)
		if r.exitCode != 0 {
			t.Fatalf("leaks: exit %d\n%s", r.exitCode, r.stderr)
		}
	case "darwin":
		r := execute(t, nil, "leaks", append([]string{"--atExit", "--", unsanitized}, args...)...)
		if r.exitCode != 0 {
			t.Fatalf("leaks: %s", r.stdout)
		}
	default:
		t.Fatalf("no leak check for %s", runtime.GOOS)
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
		return fmt.Errorf("%s %s: output count got %d want %d census %d", unit.name, side, len(a), len(b), len(unit.ids))
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

func checkNumbersOutput(t *testing.T, unit numbersUnit, side string, got, want []byte) {
	t.Helper()
	if err := numbersOutputError(unit, side, got, want); err != nil {
		t.Fatal(err)
	}
}

func TestCSSNumbersShardUnion(t *testing.T) {
	t.Parallel()
	corpus := enumerateNumbers(t)
	units := numbersUnits(corpus)
	verifyNumbersUnion(t, corpus, units)
	if err := numbersUnionError(corpus, units[1:]); err == nil {
		t.Fatal("missing range survived")
	}
	repeated := append(append([]numbersUnit{}, units...), units[0])
	repeated[len(repeated)-1].name = "repeated-range"
	if err := numbersUnionError(corpus, repeated); err == nil || !strings.Contains(err.Error(), "repeated case") {
		t.Fatalf("repeated range: %v", err)
	}
	// Gate selections partition every unit once, even when n exceeds the unit count.
	for _, n := range []int{1, 2, 7, len(units) + 1} {
		var union []numbersUnit
		for i := 0; i < n; i++ {
			selected, err := numbersSelection(fmt.Sprintf("%d/%d", i, n), units)
			if err != nil {
				t.Fatal(err)
			}
			union = append(union, selected...)
		}
		if err := numbersUnionError(corpus, union); err != nil {
			t.Fatal(err)
		}
	}
	for _, bad := range []string{"0/0", "-1/2", "2/2", "1", "a/2", "1/2/3"} {
		if _, err := numbersSelection(bad, units); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}

// The child uses the same comparator and batch-unit IDs as the real oracle checks.
// A disagreement is planted in exactly one preference case. All units execute;
// the parent requires exactly one failed unit, with its case ID and shard name.
func TestCSSNumbersPlantedDisagreement(t *testing.T) {
	t.Parallel()
	corpus := enumerateNumbers(t)
	units := numbersUnits(corpus)
	verifyNumbersUnion(t, corpus, units)
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
		t.Fatal("planted case absent")
	}
	if os.Getenv("ADAMIC_CSSNUMBERS_DISAGREEMENT_PROBE") == "1" {
		for _, unit := range units {
			if unit.kind != "batch" {
				continue
			}
			t.Run(unit.name, func(t *testing.T) {
				t.Parallel()
				var want, got strings.Builder
				for _, id := range unit.ids {
					want.WriteString(id + "\n")
					if id == planted {
						got.WriteString("planted disagreement\n")
					} else {
						got.WriteString(id + "\n")
					}
				}
				checkNumbersOutput(t, unit, "planted", []byte(got.String()), []byte(want.String()))
			})
		}
		return
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	r := execute(t, []string{"ADAMIC_CSSNUMBERS_DISAGREEMENT_PROBE=1"}, binary, "-test.run=^TestCSSNumbersPlantedDisagreement$", "-test.v", "-test.parallel=4")
	if r.exitCode != 1 || len(r.stderr) != 0 {
		t.Fatalf("probe exit %d stderr %s", r.exitCode, r.stderr)
	}
	failedPrefix := "--- FAIL: TestCSSNumbersPlantedDisagreement/"
	if strings.Count(string(r.stdout), failedPrefix) != 1 || !strings.Contains(string(r.stdout), failedPrefix+owner+" (") || !strings.Contains(string(r.stdout), "disagreement in case "+planted) {
		t.Fatalf("expected only shard %s to catch %s:\n%s", owner, planted, r.stdout)
	}
	t.Logf("planted disagreement %s caught by exactly shard %s", planted, owner)
}

// Oracle answers are immutable inputs; throughput still executes every original round.
func numbersOracleAnswers(t *testing.T, name, command string, arguments []string, data []byte, identity string) run {
	t.Helper()
	dir := numbersProduct(t, numbersInputs{Name: name + " oracle answers", Files: []string{"testdata/library.mjs"}, Flags: []string{identity, fmt.Sprintf("input-sha256=%x", sha256.Sum256(data))}, Toolchain: buildcache.Tool("node", "--version")}, func(dir string) error {
		path := filepath.Join(dir, "input.txt")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
		cmd := bounded(t, command, append(append([]string{}, arguments...), path)...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := childguard.Run(cmd, childguard.Options{}); err != nil || stderr.Len() != 0 {
			return fmt.Errorf("%s oracle: %v; %s", name, err, &stderr)
		}
		return os.WriteFile(filepath.Join(dir, "answers.txt"), stdout.Bytes(), 0644)
	})
	answer, err := os.ReadFile(filepath.Join(dir, "answers.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return run{stdout: answer}
}

func numbersOracleIdentity(t *testing.T, path string) string {
	t.Helper()
	hash := sha256.New()
	err := filepath.WalkDir(path, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		content, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(path, name)
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%s\x00%x\n", filepath.ToSlash(relative), sha256.Sum256(content))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("oracle-content-sha256=%x", hash.Sum(nil))
}
