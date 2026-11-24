package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// An odd count's median is its middle, and an even count's the lower of the middle two, so the median is always a round
// that ran; minimum and maximum are the ends, whatever order the rounds came in. Mutant: the upper middle taken (index
// len/2), and the even case reads 40.
func TestPerformanceMedianTakesTheLowerMiddle(t *testing.T) {
	t.Parallel()
	for _, each := range []struct {
		values []int64
		want   spread
	}{
		{values: []int64{7}, want: spread{Median: 7, Minimum: 7, Maximum: 7}},
		{values: []int64{50, 10, 30}, want: spread{Median: 30, Minimum: 10, Maximum: 50}},
		{values: []int64{40, 10, 30, 20}, want: spread{Median: 20, Minimum: 10, Maximum: 40}},
		{values: []int64{60, 10, 40, 30, 20, 50}, want: spread{Median: 30, Minimum: 10, Maximum: 60}},
	} {
		values := slices.Clone(each.values)
		if got := spreadOf(values); got != each.want {
			t.Errorf("spreadOf(%v) is %+v, want %+v", each.values, got, each.want)
		}
		if !slices.Equal(values, each.values) {
			t.Errorf("spreadOf reordered its rounds to %v", values)
		}
	}
}

// A fakeAdamic stands in for the adamic binary: it records every child it's handed, writes a build log a warm build
// would, and prints C for adamic c, every time a fixed function of the program and the run.
type fakeAdamic struct {
	work     string
	children []string
	// warmBuilds counts each program's warm builds so far, by base name.
	warmBuilds map[string]int
	// missingHit is the program and warm build whose log has no cArchive line, or empty for none.
	missingHit string
	// archiveMisses are the programs and warm builds whose cArchive line is a miss.
	archiveMisses []string
}

func (fake *fakeAdamic) run(child child) (childResult, error) {
	command, base := child.arguments[0], child.arguments[1]
	logPath, caches := "", 0
	for _, setting := range child.environment {
		if path, found := strings.CutPrefix(setting, "ADAMIC_BUILD_LOG="); found {
			logPath = path
			continue
		}
		if setting != "XDG_CACHE_HOME="+filepath.Join(fake.work, "cache") && setting != "ADAMIC_BUILD_CACHE_DIR="+filepath.Join(fake.work, "build-cache") {
			return childResult{}, fmt.Errorf("adamic %s was handed %q", command, setting)
		}
		caches++
	}
	if caches != 2 {
		return childResult{}, fmt.Errorf("adamic %s %s was handed %d of the two caches", command, base, caches)
	}
	if _, err := os.Stat(filepath.Join(child.directory, base)); err != nil {
		return childResult{}, fmt.Errorf("adamic %s %s in %s: %w", command, base, child.directory, err)
	}
	if command == "c" {
		fake.children = append(fake.children, "c "+base)
		var stdout []byte
		if child.keepOutput {
			stdout = []byte(strings.Repeat("int x;\n", 10+len(base)))
		}
		return childResult{elapsed: time.Duration(20+len(base)+len(fake.children)%3) * time.Millisecond, stdout: stdout}, nil
	}
	if logPath == "" {
		fake.children = append(fake.children, "build "+base)
		return childResult{elapsed: 3 * time.Second}, nil
	}
	fake.warmBuilds[base]++
	round := fake.warmBuilds[base]
	fake.children = append(fake.children, fmt.Sprintf("warm %s %d", base, round))
	lines := []string{
		fmt.Sprintf("phase load - off 0.%03d", 40+round),
		"phase lower - off 0.010",
		"phase emit - off 0.005",
		"cPreprocess main.c 0123456789ab off 0.002",
		fmt.Sprintf("cCompile main.c 0123456789ab hit 0.%03d", 20+round),
		"cArchive runtime.a 0123456789ab hit 0.000",
		fmt.Sprintf("cLink %s - off 0.%03d", base, 100+round),
	}
	if fmt.Sprintf("%s %d", base, round) == fake.missingHit {
		lines = slices.DeleteFunc(lines, func(line string) bool { return strings.HasPrefix(line, "cArchive") })
	}
	if slices.Contains(fake.archiveMisses, fmt.Sprintf("%s %d", base, round)) {
		for index, line := range lines {
			if strings.HasPrefix(line, "cArchive") {
				lines[index] = "cArchive runtime.a 0123456789ab miss 1.500"
			}
		}
	}
	if err := os.WriteFile(logPath, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		return childResult{}, err
	}
	return childResult{elapsed: time.Duration(200+10*round+len(base)) * time.Millisecond}, nil
}

// performanceTree is a stand-in repository holding the three programs the measure reads from the tree.
func performanceTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, source := range sources[1:] {
		path := filepath.Join(root, filepath.FromSlash(source.path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(strings.Repeat("const x = 1;\n", len(source.name))), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// The whole measure through a fake runner: hello's cold build first and alone, one untimed build of every other
// program, then the warm rounds and the front end's interleaved a program at a time, every child in its program's own
// directory with both caches in the temporary directory. Each number is the rounds' median with their ends, clang's
// and the phases' from the build logs. Mutants: the rounds run a program at a time instead of interleaved; the cold
// build timed on a warm cache (after the primes); cCompile's seconds left out of the compile.
func TestPerformanceMeasureRunsEveryRoundThroughItsRunner(t *testing.T) {
	t.Parallel()
	root, work := performanceTree(t), t.TempDir()
	fake := &fakeAdamic{work: work, warmBuilds: map[string]int{}}
	measured, err := measureCompile(compileRequest{root: root, work: work, rounds: 3}, fake.run)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"build hello.a", "build Parser.a", "build tokenizer.ts", "build trees.ts"}
	for round := 1; round <= 3; round++ {
		for _, base := range []string{"hello.a", "Parser.a", "tokenizer.ts", "trees.ts"} {
			want = append(want, fmt.Sprintf("warm %s %d", base, round))
		}
	}
	for range 3 {
		want = append(want, "c hello.a", "c Parser.a", "c tokenizer.ts", "c trees.ts")
	}
	if !slices.Equal(fake.children, want) {
		t.Errorf("children ran as\n%s\nwant\n%s", strings.Join(fake.children, "\n"), strings.Join(want, "\n"))
	}
	if measured.Rounds != 3 {
		t.Errorf("rounds %d, want 3", measured.Rounds)
	}
	var names []string
	for _, entry := range measured.Programs {
		names = append(names, entry.Name)
	}
	if !slices.Equal(names, []string{"hello", "Parser", "tokenizer", "trees"}) {
		t.Fatalf("programs %v", names)
	}
	hello, _ := measured.Programs.lookup("hello")
	if hello.Path != generatedPath || hello.Lines != 1 || hello.ColdBuild == nil || *hello.ColdBuild != 3000 {
		t.Errorf("hello is %+v, cold build %v", hello, hello.ColdBuild)
	}
	parser, _ := measured.Programs.lookup("Parser")
	wantParser := program{
		Path:      "stage3/namespaces/shapes/Parser.a",
		Lines:     len("Parser"),
		CLines:    10 + len("Parser.a"),
		WarmBuild: spread{Median: 228, Minimum: 218, Maximum: 238},
		Clang:     clangTimes{Compile: 22, Link: 102},
		Phases:    phaseTimes{Load: 42, Lower: 10, Emit: 5},
	}
	if parser.ColdBuild != nil {
		t.Errorf("Parser has a cold build, %d", *parser.ColdBuild)
	}
	parser.FrontEnd, parser.ColdBuild = spread{}, nil
	if parser != wantParser {
		t.Errorf("Parser is %+v, want %+v", parser, wantParser)
	}
	for _, entry := range measured.Programs {
		frontEnd := entry.Program.FrontEnd
		if frontEnd.Minimum > frontEnd.Median || frontEnd.Median > frontEnd.Maximum || frontEnd.Minimum < 20 {
			t.Errorf("%s's front end is %+v", entry.Name, frontEnd)
		}
	}
}

// A warm build whose log has no cArchive hit is an error naming the program and its round, and nothing is measured:
// the runtime archive's cache didn't hold, and its time would be a cold build's read as warm. Mutant: that check
// dropped, and the measure returns numbers.
func TestPerformanceWarmBuildWithoutAnArchiveHitIsAnError(t *testing.T) {
	t.Parallel()
	root, work := performanceTree(t), t.TempDir()
	fake := &fakeAdamic{work: work, warmBuilds: map[string]int{}, missingHit: "tokenizer.ts 2"}
	_, err := measureCompile(compileRequest{root: root, work: work, rounds: 3}, fake.run)
	if err == nil || !strings.Contains(err.Error(), "tokenizer: warm build 2 logged no cArchive hit") {
		t.Fatalf("error %v, want tokenizer's warm build 2 named", err)
	}
}

// The build log's own lines are read and every other passes by, a cBuild or a census line among them; a line of a kind
// read here that doesn't end in seconds is an error, never a zero. Mutant: a malformed cLink taken as 0.
func TestPerformanceBuildLogReadsItsOwnLines(t *testing.T) {
	t.Parallel()
	log, err := parseBuildLog("phase load - off 0.250\ncCompile runtime/string.c 0123456789ab hit 0.010\ncCompile main.c - off 0.120\n" +
		"cArchive runtime.a 0123456789ab hit 0.000\ncLink hello - off 0.300\nhit sources 0123456789ab 12\n\n")
	if err != nil {
		t.Fatal(err)
	}
	if milliseconds(log.compile) != 130 || milliseconds(log.link) != 300 || log.compiles != 2 || log.links != 1 || log.archives != 1 || !log.archiveHit || log.phases["load"] != 0.25 {
		t.Errorf("read %+v", log)
	}
	if log, err := parseBuildLog("cArchive runtime.a 0123456789ab miss 1.500\n"); err != nil || log.archiveHit || log.archives != 1 {
		t.Errorf("an archive miss read as %+v, %v", log, err)
	}
	for _, malformed := range []string{"cLink hello - off\n", "cLink hello - off soon\n", "phase load - off -1\n", "cLink hello - off NaN\n", "cLink hello - off +Inf\n"} {
		if _, err := parseBuildLog(malformed); err == nil {
			t.Errorf("%q read without an error", malformed)
		}
	}
}
