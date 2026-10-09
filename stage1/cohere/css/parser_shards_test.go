package css

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

type cssParserCorpus struct{ inputs, answers []string }

func readCSSParserCorpus(t *testing.T, path, answers string) cssParserCorpus {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	lines := strings.Split(strings.TrimSuffix(answers, "\n"), "\n")
	if len(lines) != len(inputs)*2 {
		t.Fatalf("oracle has %d lines for %d cases", len(lines), len(inputs))
	}
	corpus := cssParserCorpus{inputs: inputs}
	for i := range inputs {
		if lines[i*2] != fmt.Sprintf("case %d", i) {
			t.Fatalf("missing or repeated unsplit case ID at %d: %q", i, lines[i*2])
		}
		corpus.answers = append(corpus.answers, lines[i*2+1])
	}
	return corpus
}

type cssParserShard struct {
	name, kind       string
	lo, hi, mutation int
}

func cssParserShards(cases int) []cssParserShard {
	var units []cssParserShard
	const casesPerShard = 256
	for lo := 0; lo < cases; lo += casesPerShard {
		units = append(units, cssParserShard{kind: "agreement", lo: lo, hi: min(lo+casesPerShard, cases), mutation: -1})
	}
	for i := range mutants {
		units = append(units, cssParserShard{kind: "mutant", lo: 0, hi: cases, mutation: i})
	}
	for ordinal := range units {
		units[ordinal].name = fmt.Sprintf("shard-%03d", ordinal)
	}
	return units
}

func cssParserUnionError(cases int, units []cssParserShard) error {
	expected := make(map[string]bool)
	for i := 0; i < cases; i++ {
		expected[fmt.Sprintf("case-%06d", i)] = true
	}
	for i := range mutants {
		expected[fmt.Sprintf("mutant-%d", i)] = true
	}
	seen := make(map[string]string)
	names := make(map[string]bool)
	total := 0
	for _, unit := range units {
		if names[unit.name] {
			return fmt.Errorf("repeated shard %s", unit.name)
		}
		names[unit.name] = true
		var ids []string
		switch unit.kind {
		case "agreement":
			if unit.lo < 0 || unit.hi > cases || unit.lo >= unit.hi {
				return fmt.Errorf("invalid range in %s", unit.name)
			}
			for i := unit.lo; i < unit.hi; i++ {
				ids = append(ids, fmt.Sprintf("case-%06d", i))
			}
		case "mutant":
			if unit.mutation < 0 || unit.mutation >= len(mutants) || unit.lo != 0 || unit.hi != cases {
				return fmt.Errorf("invalid full-corpus mutant in %s", unit.name)
			}
			ids = []string{fmt.Sprintf("mutant-%d", unit.mutation)}
		default:
			return fmt.Errorf("unknown kind %q", unit.kind)
		}
		for _, id := range ids {
			total++
			if !expected[id] {
				return fmt.Errorf("unexpected case %s", id)
			}
			if previous, ok := seen[id]; ok {
				return fmt.Errorf("repeated case %s in %s and %s", id, previous, unit.name)
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
func verifyCSSParserUnion(t *testing.T, cases int, units []cssParserShard) {
	t.Helper()
	if err := cssParserUnionError(cases, units); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d CSS/SCSS cases + %d full-corpus mutants = %d unique IDs across %d shards, equals unsplit enumeration", cases, len(mutants), cases+len(mutants), len(units))
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
func writeCSSParserRange(t *testing.T, corpus cssParserCorpus, lo, hi int) (string, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(strings.Join(corpus.inputs[lo:hi], "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	var answers strings.Builder
	for i := lo; i < hi; i++ {
		fmt.Fprintf(&answers, "case %d\n%s\n", i-lo, corpus.answers[i])
	}
	return path, answers.String()
}
func cssParserDifference(unit cssParserShard, got, want string) string {
	if difference := firstDifference(got, want); difference != "" {
		return fmt.Sprintf("%s cases [%d,%d): %s", unit.name, unit.lo, unit.hi, difference)
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

// Normal products use the shared cache. Overlay oracles and temporary mutants
// remain private builds, once per parent; published products are never modified.
type cssParserInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func cssParserProduct(t *testing.T, inputs cssParserInputs, build func(dir string) error) string {
	t.Helper()
	if keyed, ok := cssParserCacheInputs(t, inputs); ok {
		return buildcache.Product(t, keyed, build)
	}
	dir := t.TempDir()
	start := time.Now()
	if err := build(dir); err != nil {
		t.Fatalf("build %s: %v", inputs.Name, err)
	}
	t.Logf("private build %s %.6fs; inputs=%+v", inputs.Name, time.Since(start).Seconds(), inputs)
	return dir
}

// Generated C is an input by content, not its temporary filename. Every other
// cached file must be repository-relative; temporary source overlays stay local.
func cssParserCacheInputs(t *testing.T, inputs cssParserInputs) (buildcache.Inputs, bool) {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	keyed := buildcache.Inputs{Name: "css " + inputs.Name, Flags: append([]string{}, inputs.Flags...), Toolchain: []string{inputs.Toolchain, buildcache.Tool("go", "version"), runtime.GOOS + "/" + runtime.GOARCH}}
	keyed.Files = append(keyed.Files, "stage1/cohere/css/parser_shards_test.go", "stage1/cohere/css/css_test.go")
	keyed.Flags = append(keyed.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	for _, flag := range inputs.Flags {
		if flag == "overlay" {
			return keyed, false
		}
	}

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
			if filepath.Base(name) != "program.c" {
				return keyed, false
			}
			content, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			keyed.Flags = append(keyed.Flags, fmt.Sprintf("generated-C-sha256=%x", sha256.Sum256(content)))
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
	return keyed, true
}

func cssParserOracle(t *testing.T) string {
	t.Helper()
	repo, _ := filepath.Abs(repository)
	cohere := filepath.Join(repo, "cohere")
	side, _ := filepath.Abs("testdata/cohere_side_test.go")
	dir := cssParserProduct(t, cssParserInputs{Name: "Go parser oracle", Files: []string{side, filepath.Join(cohere, "internal/format/css/postcss"), filepath.Join(cohere, "go.mod"), filepath.Join(cohere, "go.sum")}, Flags: []string{"go test -c", "overlay"}, Toolchain: runtime.Version() + "; cohere " + corpusfiles.CohereCommit}, func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(cohere, "internal/format/css/postcss/adamic_port_side_test.go"): side}})
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		command := bounded(t, "go", "test", "-c", "-overlay="+path, "-o", filepath.Join(dir, "oracle"), "./internal/format/css/postcss")
		command.Dir = cohere
		if output, err := childguard.CombinedOutput(command, childguard.Options{Stall: childStall}); err != nil {
			return fmt.Errorf("Go oracle: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(dir, "oracle")
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
	version := execute(t, nil, "clang", "--version")
	if version.exitCode != 0 {
		t.Fatalf("clang: %s", version.stderr)
	}
	options := native.Options{Sanitize: true}
	type job struct {
		dir     string
		inputs  cssParserInputs
		elapsed time.Duration
		keyed   buildcache.Inputs
		cached  bool
		err     error
	}
	jobs := make([]job, len(programs))
	for i, program := range programs {
		jobs[i] = job{dir: t.TempDir(), inputs: cssParserInputs{Name: program.name + " sanitized", Files: []string{program.c, repository + "/internal/native", repository + "/go.mod"}, Flags: native.Flags(options), Toolchain: strings.TrimSpace(string(version.stdout))}}
		if program.name != "parser" {
			jobs[i].inputs.Flags = append(jobs[i].inputs.Flags, "overlay")
		}
		jobs[i].keyed, jobs[i].cached = cssParserCacheInputs(t, jobs[i].inputs)
	}
	var workers sync.WaitGroup
	for i, program := range programs {
		workers.Add(1)
		go func() {
			defer workers.Done()
			build := func(dir string) error {
				source, err := os.ReadFile(program.c)
				if err != nil {
					return err
				}
				return native.Build(string(source), filepath.Join(dir, "parser"), options)
			}
			start := time.Now()
			if jobs[i].cached {
				jobs[i].dir, jobs[i].err = buildcache.Get(jobs[i].keyed, build)
			} else {
				jobs[i].err = build(jobs[i].dir)
			}
			jobs[i].elapsed = time.Since(start)
		}()
	}
	workers.Wait()
	var binaries []string
	for _, job := range jobs {
		if job.err != nil {
			t.Fatalf("build %s: %v", job.inputs.Name, job.err)
		}
		t.Logf("build elapsed %s %.6fs; cached=%t; inputs=%+v", job.inputs.Name, job.elapsed.Seconds(), job.cached, job.inputs)
		binaries = append(binaries, filepath.Join(job.dir, "parser"))
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
		result := execute(t, cssParserASAN(true), sanitized, args...)
		if result.exitCode != 0 {
			t.Errorf("leaks: exit %d\n%s", result.exitCode, result.stderr)
		}
	case "darwin":
		result := execute(t, nil, "leaks", append([]string{"--atExit", "--", unsanitized}, args...)...)
		if result.exitCode != 0 {
			t.Errorf("leaks: %s", result.stdout)
		}
	default:
		t.Fatalf("no leak check for %s", runtime.GOOS)
	}
}

func TestCSSParserShardUnion(t *testing.T) {
	t.Parallel()
	// Synthetic IDs exercise the same ranges as the complete measured corpus.
	// The live oracle's complete case IDs are independently checked in the parent.
	const cases = 24594
	units := cssParserShards(cases)
	if len(units) != testThePortParsesAsGoCohereDoesShards {
		t.Fatalf("shard count %d", len(units))
	}
	verifyCSSParserUnion(t, cases, units)
	if err := cssParserUnionError(cases, units[1:]); err == nil {
		t.Fatal("missing cases survived")
	}
	duplicate := append(append([]cssParserShard{}, units...), units[0])
	duplicate[len(duplicate)-1].name = "repeated-range"
	if err := cssParserUnionError(cases, duplicate); err == nil || !strings.Contains(err.Error(), "repeated case") {
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
		if err := cssParserUnionError(cases, all); err != nil {
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
	const cases = 24594
	const planted = 2049
	units := cssParserShards(cases)
	owner := ""
	for _, unit := range units {
		if unit.kind == "agreement" && unit.lo <= planted && planted < unit.hi {
			owner = unit.name
		}
	}
	if owner == "" {
		t.Fatal("planted case absent")
	}
	if os.Getenv("ADAMIC_CSS_PARSER_DISAGREEMENT_PROBE") == "1" {
		for _, unit := range units {
			if unit.kind != "agreement" {
				continue
			}
			t.Run(unit.name, func(t *testing.T) {
				t.Parallel()
				var want, got strings.Builder
				for i := unit.lo; i < unit.hi; i++ {
					fmt.Fprintf(&want, "case %d\nanswer-%d\n", i-unit.lo, i)
					fmt.Fprintf(&got, "case %d\n", i-unit.lo)
					if i == planted {
						got.WriteString("planted disagreement\n")
					} else {
						fmt.Fprintf(&got, "answer-%d\n", i)
					}
				}
				if difference := cssParserDifference(unit, got.String(), want.String()); difference != "" {
					t.Fatal(difference)
				}
			})
		}
		return
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	result := execute(t, []string{"ADAMIC_CSS_PARSER_DISAGREEMENT_PROBE=1"}, binary, "-test.run=^TestCSSParserPlantedDisagreement$", "-test.v", "-test.parallel=4")
	prefix := "--- FAIL: TestCSSParserPlantedDisagreement/"
	if result.exitCode != 1 || len(result.stderr) != 0 || strings.Count(string(result.stdout), prefix) != 1 || !strings.Contains(string(result.stdout), prefix+owner+" (") || !strings.Contains(string(result.stdout), "planted disagreement") {
		t.Fatalf("expected only %s to catch case %d: exit %d\n%s\n%s", owner, planted, result.exitCode, result.stdout, result.stderr)
	}
	t.Logf("planted disagreement case %d caught by exactly %s", planted, owner)
}
