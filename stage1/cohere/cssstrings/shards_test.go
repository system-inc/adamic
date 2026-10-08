package cssstrings

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var stringsEncode = strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

func stringsInput(texts []string) []byte {
	var input strings.Builder
	for _, text := range texts {
		for _, preference := range []string{"d", "s"} {
			input.WriteString(preference + stringsEncode.Replace(text) + "\n")
		}
	}
	return []byte(input.String())
}

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
func stringsCaseID(kind string, index int, preference string) string {
	return fmt.Sprintf("%s-%06d-%s", kind, index, preference)
}

func stringsUnits(corpus stringsCorpus) []stringsUnit {
	var units []stringsUnit
	const batchTexts = 512
	for lo := 0; lo < len(corpus.texts); lo += batchTexts {
		hi := min(lo+batchTexts, len(corpus.texts))
		unit := stringsUnit{name: fmt.Sprintf("batch-%06d-%06d", lo, hi-1), kind: "batch", lo: lo, hi: hi}
		for i := lo; i < hi; i++ {
			for _, pref := range []string{"d", "s"} {
				unit.ids = append(unit.ids, stringsCaseID("batch", i, pref))
			}
		}
		units = append(units, unit)
	}
	for i := range corpus.raw {
		units = append(units, stringsUnit{name: fmt.Sprintf("raw-%06d", i), kind: "raw", lo: i, hi: i + 1,
			ids: []string{stringsCaseID("raw", i, "d"), stringsCaseID("raw", i, "s")}})
	}
	for i, mutation := range stringsMutations {
		units = append(units, stringsUnit{name: "mutant-" + mutation.name, kind: "mutant", lo: i, hi: i + 1, ids: []string{"mutant-" + mutation.name}})
	}
	for _, side := range []string{"Go", "native", "Node", "Prettier"} {
		units = append(units, stringsUnit{name: "throughput-" + side, kind: "throughput", side: side, ids: []string{"throughput-" + side}})
	}
	return units
}

// Build the unsplit census independently of stringsUnits. Mutant and throughput cases
// each retain their full-corpus checks; the batch and raw preferences are counted individually.
func stringsUnionError(corpus stringsCorpus, units []stringsUnit) error {
	expected := make(map[string]bool)
	for kind, count := range map[string]int{"batch": len(corpus.texts), "raw": len(corpus.raw)} {
		for i := 0; i < count; i++ {
			for _, pref := range []string{"d", "s"} {
				expected[stringsCaseID(kind, i, pref)] = true
			}
		}
	}
	for _, m := range stringsMutations {
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
					actualIDs = append(actualIDs, stringsCaseID(unit.kind, i, pref))
				}
			}
		case "mutant":
			if unit.lo < 0 || unit.lo >= len(stringsMutations) {
				return fmt.Errorf("invalid mutant in %s", unit.name)
			}
			actualIDs = []string{"mutant-" + stringsMutations[unit.lo].name}
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

func verifyStringsUnion(t *testing.T, corpus stringsCorpus, units []stringsUnit) {
	t.Helper()
	if err := stringsUnionError(corpus, units); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d batch preferences + %d raw preferences + %d mutants + 4 throughput sides = %d unique case IDs across %d units; equals unsplit enumeration",
		len(corpus.texts)*2, len(corpus.raw)*2, len(stringsMutations), len(corpus.texts)*2+len(corpus.raw)*2+len(stringsMutations)+4, len(units))
}

func stringsSelection(value string, units []stringsUnit) ([]stringsUnit, error) {
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
	var selected []stringsUnit
	for ordinal, unit := range units {
		if ordinal%n == i {
			selected = append(selected, unit)
		}
	}
	return selected, nil
}

func selectedStringsUnits(t *testing.T, units []stringsUnit) []stringsUnit {
	t.Helper()
	selected, err := stringsSelection(os.Getenv("ADAMIC_TEST_SHARD"), units)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("selection: %d/%d units, ADAMIC_TEST_SHARD=%q", len(selected), len(units), os.Getenv("ADAMIC_TEST_SHARD"))
	return selected
}

// Keep inputs beside each builder. This base has no internal/buildcache: there is
// no package cache; the parent calls each builder once and shares its directory.
type stringsInputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain string
}

func stringsProduct(t *testing.T, inputs stringsInputs, build func(dir string) error) string {
	t.Helper()
	dir := t.TempDir()
	start := time.Now()
	if err := build(dir); err != nil {
		t.Fatalf("build %s: %v", inputs.Name, err)
	}
	t.Logf("build cold %s %.6fs; inputs=%+v", inputs.Name, time.Since(start).Seconds(), inputs)
	return dir
}

type stringsProgram struct{ dir, main, c, javascript string }

func buildStringsProgram(t *testing.T, name, main string) stringsProgram {
	t.Helper()
	dir := stringsProduct(t, stringsInputs{
		Name: name + " lowering and backends", Files: []string{main, filepath.Join(filepath.Dir(main), "strings.ts"), repository + "/internal/load", repository + "/internal/lower", repository + "/internal/native", repository + "/internal/javascript", repository + "/cohere/TypeScript", repository + "/cohere/TypeScript-shim"},
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
	return stringsProgram{dir: dir, main: main, c: filepath.Join(dir, "program.c"), javascript: filepath.Join(dir, "program.mjs")}
}

func stringsClangToolchain(t *testing.T) string {
	t.Helper()
	version := execute(t, nil, "clang", "--version")
	clean(t, "clang version", version)
	return strings.TrimSpace(string(version.stdout))
}

func buildStringsNative(t *testing.T, name string, program stringsProgram, sanitize bool) string {
	t.Helper()
	options := native.Options{Sanitize: sanitize}
	dir := stringsProduct(t, stringsInputs{Name: name, Files: []string{program.c, repository + "/internal/native/runtime"}, Flags: native.Flags(options), Toolchain: stringsClangToolchain(t)}, func(dir string) error {
		source, err := os.ReadFile(program.c)
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(dir, "port"), options)
	})
	return filepath.Join(dir, "port")
}

func prepareStringsMutant(t *testing.T, mutation stringsMutation) string {
	t.Helper()
	scratch := t.TempDir()
	source, err := os.ReadFile("strings.ts")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(source), mutation.from) != 1 {
		t.Fatal("mutant must change one site")
	}
	write(t, filepath.Join(scratch, "strings.ts"), []byte(strings.Replace(string(source), mutation.from, mutation.to, 1)))
	source, err = os.ReadFile("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(scratch, "main.ts")
	write(t, main, source)
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

func stringsLeaks(t *testing.T, sanitized, unsanitized string, args ...string) {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		r := execute(t, stringsSanitizerEnvironment(true), sanitized, args...)
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

func stringsOutputError(unit stringsUnit, side string, got, want []byte) error {
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

func checkStringsOutput(t *testing.T, unit stringsUnit, side string, got, want []byte) {
	t.Helper()
	if err := stringsOutputError(unit, side, got, want); err != nil {
		t.Fatal(err)
	}
}

func TestCSSStringsShardUnion(t *testing.T) {
	t.Parallel()
	corpus := enumerateStrings(t)
	units := stringsUnits(corpus)
	verifyStringsUnion(t, corpus, units)
	if err := stringsUnionError(corpus, units[1:]); err == nil {
		t.Fatal("missing range survived")
	}
	repeated := append(append([]stringsUnit{}, units...), units[0])
	repeated[len(repeated)-1].name = "repeated-range"
	if err := stringsUnionError(corpus, repeated); err == nil || !strings.Contains(err.Error(), "repeated case") {
		t.Fatalf("repeated range: %v", err)
	}
	// Gate selections partition every unit once, even when n exceeds the unit count.
	for _, n := range []int{1, 2, 7, len(units) + 1} {
		var union []stringsUnit
		for i := 0; i < n; i++ {
			selected, err := stringsSelection(fmt.Sprintf("%d/%d", i, n), units)
			if err != nil {
				t.Fatal(err)
			}
			union = append(union, selected...)
		}
		if err := stringsUnionError(corpus, union); err != nil {
			t.Fatal(err)
		}
	}
	for _, bad := range []string{"0/0", "-1/2", "2/2", "1", "a/2", "1/2/3"} {
		if _, err := stringsSelection(bad, units); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}

// The child uses the same comparator and batch-unit IDs as the real oracle checks.
// A disagreement is planted in exactly one preference case. All units execute;
// the parent requires exactly one failed unit, with its case ID and shard name.
func TestCSSStringsPlantedDisagreement(t *testing.T) {
	t.Parallel()
	corpus := enumerateStrings(t)
	units := stringsUnits(corpus)
	verifyStringsUnion(t, corpus, units)
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
	if os.Getenv("ADAMIC_CSSSTRINGS_DISAGREEMENT_PROBE") == "1" {
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
				checkStringsOutput(t, unit, "planted", []byte(got.String()), []byte(want.String()))
			})
		}
		return
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	r := execute(t, []string{"ADAMIC_CSSSTRINGS_DISAGREEMENT_PROBE=1"}, binary, "-test.run=^TestCSSStringsPlantedDisagreement$", "-test.v", "-test.parallel=4")
	if r.exitCode != 1 || len(r.stderr) != 0 {
		t.Fatalf("probe exit %d stderr %s", r.exitCode, r.stderr)
	}
	failedPrefix := "--- FAIL: TestCSSStringsPlantedDisagreement/"
	if strings.Count(string(r.stdout), failedPrefix) != 1 || !strings.Contains(string(r.stdout), failedPrefix+owner+" (") || !strings.Contains(string(r.stdout), "disagreement in case "+planted) {
		t.Fatalf("expected only shard %s to catch %s:\n%s", owner, planted, r.stdout)
	}
	t.Logf("planted disagreement %s caught by exactly shard %s", planted, owner)
}
