package css

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// cssParserUnionError holds a plan exact: every case once in agreement, and each mutant's witnesses exactly once, in
// one shard, none of them empty.
func cssParserUnionError(keys []string, units []oracles.ParserShard, witnesses [][]int) error {
	if len(witnesses) != len(mutants) {
		return fmt.Errorf("witnesses for %d parser mutants, want %d", len(witnesses), len(mutants))
	}
	expected := make(map[string]bool)
	for _, key := range keys {
		id := oracles.ParserCaseKey("agreement", -1, key)
		if expected[id] {
			return fmt.Errorf("repeated unsplit case %s", id)
		}
		expected[id] = true
	}
	for mutation, ids := range witnesses {
		if err := cssParserWitnessError(mutants[mutation].name, ids); err != nil {
			return err
		}
		for _, i := range ids {
			if i < 0 || i >= len(keys) {
				return fmt.Errorf("mutant %d's witness %d is not in the corpus of %d", mutation, i, len(keys))
			}
			id := oracles.ParserCaseKey("mutant", mutation, keys[i])
			if expected[id] {
				return fmt.Errorf("repeated witness %s", id)
			}
			expected[id] = true
		}
	}
	seen := make(map[string]string)
	names := make(map[string]bool)
	mutantShards := make(map[int]int)
	for _, unit := range units {
		if names[unit.Name] {
			return fmt.Errorf("repeated shard %s", unit.Name)
		}
		names[unit.Name] = true
		for _, group := range unit.Groups {
			if group.Kind == "mutant" {
				mutantShards[group.Mutation]++
			}
			for _, i := range group.Cases {
				if i < 0 || i >= len(keys) {
					return fmt.Errorf("invalid case %d", i)
				}
				id := oracles.ParserCaseKey(group.Kind, group.Mutation, keys[i])
				if !expected[id] {
					return fmt.Errorf("unexpected case %s", id)
				}
				if previous, ok := seen[id]; ok {
					return fmt.Errorf("repeated case %s in %s and %s", id, previous, unit.Name)
				}
				seen[id] = unit.Name
			}
		}
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("union count %d, live enumeration %d", len(seen), len(expected))
	}
	for mutation, count := range mutantShards {
		if count != 1 {
			return fmt.Errorf("mutant %d runs in %d shards, want one", mutation, count)
		}
	}
	for id := range expected {
		if _, ok := seen[id]; !ok {
			return fmt.Errorf("missing case %s", id)
		}
	}
	return nil
}
func verifyCSSParserUnion(test *testing.T, keys []string, units []oracles.ParserShard, witnesses [][]int) {
	test.Helper()
	if len(units) != oracles.ParserShardCount {
		test.Fatalf("enumerated %d shards, constant %d", len(units), oracles.ParserShardCount)
	}
	if err := cssParserUnionError(keys, units, witnesses); err != nil {
		test.Fatal(err)
	}
	witnessCases := 0
	for _, unit := range units {
		for _, group := range unit.Groups {
			if group.Kind == "mutant" {
				witnessCases += len(group.Cases)
			}
		}
	}
	test.Logf(
		"union: %d live cases in agreement plus %d in %d mutant witness groups = %d unique case IDs across %d shards",
		len(keys),
		witnessCases,
		len(mutants),
		len(keys)+witnessCases,
		len(units),
	)
}
func selectCSSParserShards(value string, units []oracles.ParserShard) ([]oracles.ParserShard, error) {
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
	var selected []oracles.ParserShard
	for ordinal, unit := range units {
		if ordinal%n == i {
			selected = append(selected, unit)
		}
	}
	return selected, nil
}
func selectedCSSParserShards(test *testing.T, units []oracles.ParserShard) []oracles.ParserShard {
	test.Helper()
	selected, err := selectCSSParserShards(os.Getenv("ADAMIC_TEST_SHARD"), units)
	if err != nil {
		test.Fatal(err)
	}
	test.Logf("selected %d/%d shards, ADAMIC_TEST_SHARD=%q", len(selected), len(units), os.Getenv("ADAMIC_TEST_SHARD"))
	return selected
}

func cssParserDifference(unit oracles.ParserShard, got, want string) string {
	if difference := firstDifference(got, want); difference != "" {
		return fmt.Sprintf("%s: %s", unit.Name, difference)
	}
	return ""
}

func checkCSSParserPostCSS(test *testing.T, unit oracles.ParserShard, inputs []string, want, got string) {
	test.Helper()
	goLines := strings.Split(strings.TrimSuffix(want, "\n"), "\n")
	jsLines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(goLines) != len(jsLines) {
		test.Fatalf("library printed %d lines, Go %d", len(jsLines), len(goLines))
	}
	gaps := 0
	for i, line := range inputs {
		if goLines[i*2] != jsLines[i*2] {
			test.Fatalf("case marker differs at %d", unit.LowerBound+i)
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
		test.Errorf(
			"case %d %q: unrecorded library difference: %s",
			unit.LowerBound+i,
			line,
			firstDifference(jsAnswer, goAnswer),
		)
	}
	test.Logf("PostCSS: %d exact agreements, %d occurrences of proved surrogate gap", len(inputs)-gaps, gaps)
}

type cssParserProgram struct{ name, main, c, javascript string }

func buildCSSParserProgram(test *testing.T, name, directory string) cssParserProgram {
	test.Helper()
	main := filepath.Join(directory, "main.ts")
	var files []string
	for _, slice := range []string{"css", "selector", "values", "mediaquery", "cssstrings", "cssnumbers"} {
		paths, err := filepath.Glob(filepath.Join(filepath.Dir(directory), slice, "*.ts"))
		if err != nil {
			test.Fatal(err)
		}
		files = append(files, paths...)
	}
	files = append(
		files,
		repository+"/internal",
		repository+"/go.mod",
		repository+"/cohere/TypeScript",
		repository+"/cohere/TypeScript-shim",
	)
	dir := oracles.ParserProduct(test, oracles.ParserProductInputs{
		Name:      name + " lowering and backends",
		Files:     files,
		Flags:     []string{"lower", "C", "JavaScript"},
		Toolchain: runtime.Version() + "; checker " + corpusfiles.TypeScriptGoCommit,
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
	return cssParserProgram{
		name:       name,
		main:       main,
		c:          filepath.Join(dir, "program.c"),
		javascript: filepath.Join(dir, "program.mjs"),
	}
}
func buildCSSParserBinaries(test *testing.T, programs []cssParserProgram) []string {
	test.Helper()
	options := native.Options{Sanitize: true}
	var binaries []string
	for _, program := range programs {
		dir := oracles.ParserProduct(test, oracles.ParserProductInputs{
			Name:      program.name + " sanitized",
			Files:     []string{program.c, repository + "/internal/native", repository + "/go.mod"},
			Flags:     native.Flags(options),
			Toolchain: buildcache.Tool("clang", "--version"),
		}, func(dir string) error {
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

func buildCSSParserUnsanitized(test *testing.T, program cssParserProgram) string {
	test.Helper()
	options := native.Options{}
	dir := oracles.ParserProduct(test, oracles.ParserProductInputs{
		Name:      "parser unsanitized for leaks",
		Files:     []string{program.c, repository + "/internal/native", repository + "/go.mod"},
		Flags:     native.Flags(options),
		Toolchain: buildcache.Tool("clang", "--version"),
	}, func(dir string) error {
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
func checkCSSParserLeaks(test *testing.T, sanitized, unsanitized string, args ...string) {
	test.Helper()
	switch runtime.GOOS {
	case "linux":
		// The agreement's sanitized run has leak detection on (cssParserSanitizedAgreement), so one run checks
		// output, ASan/UBSan and LeakSanitizer, as the printer's does. TestCSSParserSanitizedRunCatchesALeak holds it.
	case "darwin":
		result := cssParserExecute(test, nil, "leaks", append([]string{"--atExit", "--", unsanitized}, args...)...)
		if result.exitCode != 0 {
			test.Errorf("leaks: %s", result.stdout)
		}
	default:
		test.Fatalf("no leak check for %s", runtime.GOOS)
	}
}

func TestThePortParsesAsGoCohereDoesUnion(test *testing.T) {
	test.Parallel()
	// Compare every check ID against the complete live oracle enumeration.
	keys := oracles.LoadParserCorpus(test).CaseKeys
	witnesses := cssParserAllWitnesses(test)
	units := oracles.ParserWithMutants(oracles.ParserAgreementShards(keys), keys, witnesses)
	if len(units) != oracles.ParserShardCount {
		test.Fatalf("shard count %d", len(units))
	}
	verifyCSSParserUnion(test, keys, units, witnesses)
	for mutation, ids := range witnesses {
		test.Logf(
			"mutant %s: %d witnesses in shard-%03d",
			mutants[mutation].name,
			len(ids),
			oracles.ParserMutantBucket(keys, mutation, ids),
		)
	}

	// Adding a repository file preserves every existing check's owner.
	expanded := append(append([]string{}, keys...), "repository/new-fixture-for-union.css:0:C")
	grown := oracles.ParserWithMutants(oracles.ParserAgreementShards(expanded), expanded, witnesses)
	if len(grown) != len(units) {
		test.Fatal("corpus growth changed the shard count")
	}
	owners := func(list []oracles.ParserShard, caseKeys []string) map[string]string {
		result := make(map[string]string)
		for _, unit := range list {
			for _, group := range unit.Groups {
				for _, i := range group.Cases {
					result[oracles.ParserCaseKey(group.Kind, group.Mutation, caseKeys[i])] = unit.Name
				}
			}
		}
		return result
	}
	before, after := owners(units, keys), owners(grown, expanded)
	for id, owner := range before {
		if after[id] != owner {
			test.Fatalf("adding a file moved %s from %s to %s", id, owner, after[id])
		}
	}
	if err := cssParserUnionError(expanded, grown, witnesses); err != nil {
		test.Fatal(err)
	}
	if err := cssParserUnionError(keys, units[1:], witnesses); err == nil {
		test.Fatal("missing cases survived")
	}
	duplicate := append(append([]oracles.ParserShard{}, units...), units[0])
	duplicate[len(duplicate)-1].Name = "repeated-range"
	if err := cssParserUnionError(
		keys,
		duplicate,
		witnesses,
	); err == nil || !strings.Contains(err.Error(), "repeated case") {
		test.Fatalf("duplicate: %v", err)
	}
	// A mutant's witness group split across two shards holds the same cases but runs the mutant twice.
	split := oracles.ParserWithMutants(oracles.ParserAgreementShards(keys), keys, witnesses)
	owner := oracles.ParserMutantBucket(keys, 0, witnesses[0])
	for index, group := range split[owner].Groups {
		if group.Kind == "mutant" && group.Mutation == 0 {
			if len(group.Cases) < 2 {
				test.Fatalf("mutant 0's witness group holds %d cases, too few to split", len(group.Cases))
			}
			moved := group
			moved.Cases = group.Cases[1:]
			split[owner].Groups[index].Cases = group.Cases[:1]
			other := (owner + 1) % len(split)
			moved.Name = split[other].Name
			split[other].Groups = append(split[other].Groups, moved)
		}
	}
	if err := cssParserUnionError(
		keys,
		split,
		witnesses,
	); err == nil || !strings.Contains(err.Error(), "runs in 2 shards") {
		test.Fatalf("split mutant group: %v", err)
	}
	// A mutant group missing one witness, or holding a case that isn't one, is refused.
	for _, edit := range []struct {
		name  string
		cases func([]int) []int
	}{
		{"missing a witness", func(ids []int) []int { return ids[1:] }},
		{"holding a case that isn't a witness", func(ids []int) []int {
			foreign := 0
			for slices.Contains(ids, foreign) {
				foreign++
			}
			return append(append([]int{}, ids...), foreign)
		}},
	} {
		edited := oracles.ParserWithMutants(oracles.ParserAgreementShards(keys), keys, witnesses)
		for index, group := range edited[owner].Groups {
			if group.Kind == "mutant" && group.Mutation == 0 {
				edited[owner].Groups[index].Cases = edit.cases(group.Cases)
			}
		}
		if err := cssParserUnionError(keys, edited, witnesses); err == nil {
			test.Fatalf("a mutant group %s was accepted", edit.name)
		}
	}
	for _, n := range []int{1, 2, 7, len(units) + 1} {
		var all []oracles.ParserShard
		for i := 0; i < n; i++ {
			selected, err := selectCSSParserShards(fmt.Sprintf("%d/%d", i, n), units)
			if err != nil {
				test.Fatal(err)
			}
			all = append(all, selected...)
		}
		if err := cssParserUnionError(keys, all, witnesses); err != nil {
			test.Fatal(err)
		}
	}
	for _, bad := range []string{"1", "0/0", "-1/2", "2/2", "x/2", "0/1/2"} {
		if _, err := selectCSSParserShards(bad, units); err == nil {
			test.Fatalf("accepted %q", bad)
		}
	}
}
func TestCSSParserPlantedDisagreement(test *testing.T) {
	test.Parallel()
	corpus := oracles.LoadParserCorpus(test)
	units := oracles.ParserAgreementShards(corpus.CaseKeys)
	owner := ""
	for _, unit := range units {
		for _, group := range unit.Groups {
			if group.Kind == "agreement" {
				for _, i := range group.Cases {
					if i == 0 {
						owner = unit.Name
					}
				}
			}
		}
	}
	if owner == "" {
		test.Fatal("planted case absent")
	}
	binary, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	result := cssParserExecute(
		test,
		[]string{"ADAMIC_CSS_PARSER_DISAGREEMENT_PROBE=1"},
		binary,
		"-test.run=^TestThePortParsesAsGoCohereDoes_[0-9]{3}$",
		"-test.timeout=90s",
		"-test.v",
		"-test.parallel=4",
	)
	topOwner := "TestThePortParsesAsGoCohereDoes_" + strings.TrimPrefix(owner, "shard-")
	prefix := "--- FAIL: TestThePortParsesAsGoCohereDoes_"
	if result.exitCode != 1 ||
		len(result.stderr) != 0 ||
		strings.Count(string(result.stdout), prefix) != 1 ||
		!strings.Contains(string(result.stdout), "--- FAIL: "+topOwner+" (") ||
		!strings.Contains(string(result.stdout), "planted disagreement") {
		test.Fatalf(
			"expected only %s to catch %s: exit %d\n%s\n%s",
			topOwner,
			corpus.CaseKeys[0],
			result.exitCode,
			result.stdout,
			result.stderr,
		)
	}
	test.Logf("planted disagreement %s caught by exactly %s", corpus.CaseKeys[0], topOwner)
}

func cssParserTopProbe(test *testing.T, ordinal int) {
	test.Helper()
	corpus := oracles.LoadParserCorpus(test)
	unit := oracles.ParserAgreementShards(corpus.CaseKeys)[ordinal]
	for _, group := range unit.Groups {
		if group.Kind == "agreement" {
			var want, got strings.Builder
			for local, i := range group.Cases {
				fmt.Fprintf(&want, "case %d\n%s\n", local, corpus.Answers[i])
				fmt.Fprintf(&got, "case %d\n", local)
				if i == 0 {
					got.WriteString("planted disagreement\n")
				} else {
					fmt.Fprintf(&got, "%s\n", corpus.Answers[i])
				}
			}
			if difference := cssParserDifference(group, got.String(), want.String()); difference != "" {
				test.Fatal(difference)
			}
		}
	}
}
