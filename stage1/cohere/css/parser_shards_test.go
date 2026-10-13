package css

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func cssParserUnionError(keys []string, units []cssParserShard) error {
	expected := make(map[string]bool)
	for variant := -1; variant < len(mutants); variant++ {
		kind := "agreement"
		if variant >= 0 {
			kind = "mutant"
		}
		for _, key := range keys {
			id := cssParserCaseKey(kind, variant, key)
			if expected[id] {
				return fmt.Errorf("repeated unsplit case %s", id)
			}
			expected[id] = true
		}
	}
	seen := make(map[string]string)
	names := make(map[string]bool)
	for _, unit := range units {
		if names[unit.name] {
			return fmt.Errorf("repeated shard %s", unit.name)
		}
		names[unit.name] = true
		for _, group := range unit.groups {
			for _, i := range group.cases {
				if i < 0 || i >= len(keys) {
					return fmt.Errorf("invalid case %d", i)
				}
				id := cssParserCaseKey(group.kind, group.mutation, keys[i])
				if !expected[id] {
					return fmt.Errorf("unexpected case %s", id)
				}
				if previous, ok := seen[id]; ok {
					return fmt.Errorf("repeated case %s in %s and %s", id, previous, unit.name)
				}
				seen[id] = unit.name
			}
		}
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("union count %d, live enumeration %d", len(seen), len(expected))
	}
	for id := range expected {
		if _, ok := seen[id]; !ok {
			return fmt.Errorf("missing case %s", id)
		}
	}
	return nil
}
func verifyCSSParserUnion(t *testing.T, keys []string, units []cssParserShard) {
	t.Helper()
	if len(units) != testThePortParsesAsGoCohereDoesShards {
		t.Fatalf("enumerated %d shards, constant %d", len(units), testThePortParsesAsGoCohereDoesShards)
	}
	if err := cssParserUnionError(keys, units); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d live cases x %d original checks = %d unique case IDs across %d shards", len(keys), 1+len(mutants), len(keys)*(1+len(mutants)), len(units))
}
func selectCSSParserShards(value string, units []cssParserShard) ([]cssParserShard, error) {
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
	var selected []cssParserShard
	for ordinal, unit := range units {
		if ordinal%n == i {
			selected = append(selected, unit)
		}
	}
	return selected, nil
}
func selectedCSSParserShards(t *testing.T, units []cssParserShard) []cssParserShard {
	t.Helper()
	selected, err := selectCSSParserShards(os.Getenv("ADAMIC_TEST_SHARD"), units)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("selected %d/%d shards, ADAMIC_TEST_SHARD=%q", len(selected), len(units), os.Getenv("ADAMIC_TEST_SHARD"))
	return selected
}

func cssParserDifference(unit cssParserShard, got, want string) string {
	if difference := firstDifference(got, want); difference != "" {
		return fmt.Sprintf("%s: %s", unit.name, difference)
	}
	return ""
}

func checkCSSParserPostCSS(t *testing.T, unit cssParserShard, inputs []string, want, got string) {
	t.Helper()
	goLines := strings.Split(strings.TrimSuffix(want, "\n"), "\n")
	jsLines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(goLines) != len(jsLines) {
		t.Fatalf("library printed %d lines, Go %d", len(jsLines), len(goLines))
	}
	gaps := 0
	for i, line := range inputs {
		if goLines[i*2] != jsLines[i*2] {
			t.Fatalf("case marker differs at %d", unit.lo+i)
		}
		goAnswer, jsAnswer := goLines[i*2+1], jsLines[i*2+1]
		if goAnswer == jsAnswer {
			continue
		}
		// Preserve the original narrowly proved astral-surrogate gap, for both dialects.
		if line[2:] == `\\😀|a` &&
			goAnswer == `error {"column":1,"endColumn":4,"endLine":1,"endOffset":5,"line":1,"name":"CssSyntaxError","offset":0,"reason":"Unknown word \\😀"}` &&
			jsAnswer == `error {"column":1,"endColumn":3,"endLine":1,"endOffset":4,"line":1,"name":"CssSyntaxError","offset":0,"reason":"Unknown word \\\ud83d"}` {
			gaps++
			continue
		}
		t.Errorf("case %d %q: unrecorded library difference: %s", unit.lo+i, line, firstDifference(jsAnswer, goAnswer))
	}
	t.Logf("PostCSS: %d exact agreements, %d occurrences of proved surrogate gap", len(inputs)-gaps, gaps)
}

type cssParserProgram struct{ name, main, c, javascript string }

func buildCSSParserProgram(t *testing.T, name, directory string) cssParserProgram {
	t.Helper()
	main := filepath.Join(directory, "main.ts")
	var files []string
	for _, slice := range []string{"css", "selector", "values", "mediaquery", "cssstrings", "cssnumbers"} {
		paths, err := filepath.Glob(filepath.Join(filepath.Dir(directory), slice, "*.ts"))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, paths...)
	}
	files = append(files, repository+"/internal", repository+"/go.mod", repository+"/cohere/TypeScript", repository+"/cohere/TypeScript-shim")
	dir := cssParserProduct(t, cssParserInputs{Name: name + " lowering and backends", Files: files, Flags: []string{"lower", "C", "JavaScript"}, Toolchain: runtime.Version() + "; checker " + corpusfiles.TypeScriptGoCommit}, func(dir string) error {
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
	return cssParserProgram{name: name, main: main, c: filepath.Join(dir, "program.c"), javascript: filepath.Join(dir, "program.mjs")}
}
func buildCSSParserBinaries(t *testing.T, programs []cssParserProgram) []string {
	t.Helper()
	options := native.Options{Sanitize: true}
	var binaries []string
	for _, program := range programs {
		dir := cssParserProduct(t, cssParserInputs{Name: program.name + " sanitized", Files: []string{program.c, repository + "/internal/native", repository + "/go.mod"}, Flags: native.Flags(options), Toolchain: buildcache.Tool("clang", "--version")}, func(dir string) error {
			source, err := os.ReadFile(program.c)
			if err != nil {
				return err
			}
			return native.Build(string(source), filepath.Join(dir, "parser"), options)
		})
		binaries = append(binaries, filepath.Join(dir, "parser"))
	}
	return binaries
}

func buildCSSParserUnsanitized(t *testing.T, program cssParserProgram) string {
	t.Helper()
	options := native.Options{}
	dir := cssParserProduct(t, cssParserInputs{Name: "parser unsanitized for leaks", Files: []string{program.c, repository + "/internal/native", repository + "/go.mod"}, Flags: native.Flags(options), Toolchain: buildcache.Tool("clang", "--version")}, func(dir string) error {
		source, err := os.ReadFile(program.c)
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(dir, "parser"), options)
	})
	return filepath.Join(dir, "parser")
}
func cssParserASAN(leaks bool) []string {
	if runtime.GOOS != "linux" {
		return nil
	}
	if leaks {
		return []string{"ASAN_OPTIONS=detect_leaks=1"}
	}
	return []string{"ASAN_OPTIONS=detect_leaks=0"}
}
func checkCSSParserLeaks(t *testing.T, sanitized, unsanitized string, args ...string) {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		result := cssParserExecute(t, cssParserASAN(true), sanitized, args...)
		if result.exitCode != 0 {
			t.Errorf("leaks: exit %d\n%s", result.exitCode, result.stderr)
		}
	case "darwin":
		result := cssParserExecute(t, nil, "leaks", append([]string{"--atExit", "--", unsanitized}, args...)...)
		if result.exitCode != 0 {
			t.Errorf("leaks: %s", result.stdout)
		}
	default:
		t.Fatalf("no leak check for %s", runtime.GOOS)
	}
}

func TestThePortParsesAsGoCohereDoesUnion(t *testing.T) {
	t.Parallel()
	// Compare every check ID against the complete live oracle enumeration.
	keys := cssParserTopCorpus(t).keys
	units := cssParserShards(keys)
	if len(units) != testThePortParsesAsGoCohereDoesShards {
		t.Fatalf("shard count %d", len(units))
	}
	verifyCSSParserUnion(t, keys, units)

	// Adding a repository file preserves every existing check's owner.
	expanded := append(append([]string{}, keys...), "repository/new-fixture-for-union.css:0:C")
	grown := cssParserShards(expanded)
	if len(grown) != len(units) {
		t.Fatal("corpus growth changed the shard count")
	}
	owners := func(list []cssParserShard, caseKeys []string) map[string]string {
		result := make(map[string]string)
		for _, unit := range list {
			for _, group := range unit.groups {
				for _, i := range group.cases {
					result[cssParserCaseKey(group.kind, group.mutation, caseKeys[i])] = unit.name
				}
			}
		}
		return result
	}
	before, after := owners(units, keys), owners(grown, expanded)
	for id, owner := range before {
		if after[id] != owner {
			t.Fatalf("adding a file moved %s from %s to %s", id, owner, after[id])
		}
	}
	if err := cssParserUnionError(expanded, grown); err != nil {
		t.Fatal(err)
	}
	if err := cssParserUnionError(keys, units[1:]); err == nil {
		t.Fatal("missing cases survived")
	}
	duplicate := append(append([]cssParserShard{}, units...), units[0])
	duplicate[len(duplicate)-1].name = "repeated-range"
	if err := cssParserUnionError(keys, duplicate); err == nil || !strings.Contains(err.Error(), "repeated case") {
		t.Fatalf("duplicate: %v", err)
	}
	for _, n := range []int{1, 2, 7, len(units) + 1} {
		var all []cssParserShard
		for i := 0; i < n; i++ {
			selected, err := selectCSSParserShards(fmt.Sprintf("%d/%d", i, n), units)
			if err != nil {
				t.Fatal(err)
			}
			all = append(all, selected...)
		}
		if err := cssParserUnionError(keys, all); err != nil {
			t.Fatal(err)
		}
	}
	for _, bad := range []string{"1", "0/0", "-1/2", "2/2", "x/2", "0/1/2"} {
		if _, err := selectCSSParserShards(bad, units); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}
func TestCSSParserPlantedDisagreement(t *testing.T) {
	t.Parallel()
	corpus := cssParserTopCorpus(t)
	units := cssParserShards(corpus.keys)
	owner := ""
	for _, unit := range units {
		for _, group := range unit.groups {
			if group.kind == "agreement" {
				for _, i := range group.cases {
					if i == 0 {
						owner = unit.name
					}
				}
			}
		}
	}
	if owner == "" {
		t.Fatal("planted case absent")
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	result := cssParserExecute(t, []string{"ADAMIC_CSS_PARSER_DISAGREEMENT_PROBE=1"}, binary, "-test.run=^TestThePortParsesAsGoCohereDoes_[0-9]{3}$", "-test.timeout=90s", "-test.v", "-test.parallel=4")
	topOwner := "TestThePortParsesAsGoCohereDoes_" + strings.TrimPrefix(owner, "shard-")
	prefix := "--- FAIL: TestThePortParsesAsGoCohereDoes_"
	if result.exitCode != 1 || len(result.stderr) != 0 || strings.Count(string(result.stdout), prefix) != 1 || !strings.Contains(string(result.stdout), "--- FAIL: "+topOwner+" (") || !strings.Contains(string(result.stdout), "planted disagreement") {
		t.Fatalf("expected only %s to catch %s: exit %d\n%s\n%s", topOwner, corpus.keys[0], result.exitCode, result.stdout, result.stderr)
	}
	t.Logf("planted disagreement %s caught by exactly %s", corpus.keys[0], topOwner)
}

func cssParserTopProbe(t *testing.T, ordinal int) {
	t.Helper()
	corpus := cssParserTopCorpus(t)
	unit := cssParserTopCorpusSetup.units[ordinal]
	for _, group := range unit.groups {
		if group.kind == "agreement" {
			var want, got strings.Builder
			for local, i := range group.cases {
				fmt.Fprintf(&want, "case %d\n%s\n", local, corpus.answers[i])
				fmt.Fprintf(&got, "case %d\n", local)
				if i == 0 {
					got.WriteString("planted disagreement\n")
				} else {
					fmt.Fprintf(&got, "%s\n", corpus.answers[i])
				}
			}
			if difference := cssParserDifference(group, got.String(), want.String()); difference != "" {
				t.Fatal(difference)
			}
		}
	}
}

// These are the first disagreements observed in the complete unsplit corpus.
// Only their owning ranges require a disagreement; every other mutant range still
// executes both original sides and compares its output. A surviving mutant fails
// on both sides in its witness range even when the gate selects that range alone.
func cssParserMutantWitness(mutation int) int {
	switch mutation {
	case 0:
		return 80
	case 1:
		return 1574
	case 2:
		return 2
	}
	panic("parser mutant has no witness")
}
