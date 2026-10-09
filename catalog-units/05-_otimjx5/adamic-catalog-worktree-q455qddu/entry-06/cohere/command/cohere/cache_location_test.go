package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestTheCacheLivesInTheProjectAndNoCacheTouchesNone holds where cohere keeps a project's cache and what
// `--no-cache` promises about it.
//
// The table lands in `<root>/.cache/cohere/`, each root gets its own, and nothing is written to the user
// cache. With the table inside the project, an unchanged tree still replays: the table's own writes must
// not move anything the run records as an input. A `--no-cache` run, in every shape, leaves the table and
// the tsconfig's build info byte for byte and second for second as they were, reports the same findings
// as a cached run, and says it ran cold.
func TestTheCacheLivesInTheProjectAndNoCacheTouchesNone(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	home := t.TempDir()
	project := func(gitignore string) string {
		root := t.TempDir()
		files := map[string]string{
			"tsconfig.json":            incrementalFixtureConfig,
			"CohereSettings.json":      "{ \"extends\": \"./NexusCohereSettings.json\", \"rules\": { \"no-debugger\": \"error\" } }\n",
			"NexusCohereSettings.json": "{ \"format\": { \"ignore\": [] } }\n",
			"index.ts":                 "export function value(): number {\n  debugger;\n  return 1;\n}\n",
			"Ugly.ts":                  "export const ugly   =   1\n",
		}
		if gitignore != "" {
			files[".gitignore"] = gitignore
		}
		writeTree(t, root, files)
		return root
	}
	run := func(root string, arguments ...string) (string, string) {
		t.Helper()
		command := exec.Command(binary, verboseArguments(arguments)...)
		command.Dir = root
		command.Env = append(os.Environ(), "HOME="+home, "XDG_CACHE_HOME=")
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		_ = command.Run()
		return stdout.String(), stderr.String()
	}

	first := project("tsconfig.tsbuildinfo\n.cache/\n")
	if _, stderr := run(first, "--no-fix"); strings.Contains(stderr, "does not ignore .cache/") {
		t.Errorf("a project that ignores .cache/ was warned anyway:\n%s", stderr)
	}
	if len(cacheTableFiles(t, first)) == 0 {
		t.Fatalf("the table is not in the project %s", first)
	}
	if entries, _ := os.ReadDir(filepath.Join(home, "Library", "Caches", "cohere")); len(entries) > 0 {
		t.Fatalf("a project run wrote to the user cache: %v", entries)
	}
	second := project("tsconfig.tsbuildinfo\n")
	if _, stderr := run(second, "--no-fix"); !strings.Contains(stderr, "does not ignore .cache/, so cohere's cache") ||
		!strings.Contains(stderr, "add the line `.cache/`") {
		t.Errorf("a project that does not ignore .cache/ was not told:\n%s", stderr)
	}
	if len(cacheTableFiles(t, second)) == 0 {
		t.Fatalf("a second root %s did not get a table of its own", second)
	}

	// The table's own writes move nothing the run records: an unchanged tree replays, and keeps replaying.
	//
	// The run records every ancestor directory, so creating the second project above moved the test's
	// own temporary directory: a miss here would be the fixture's doing. So the first project is recorded
	// once more now that nothing around it moves, and only then must it replay. Its own table writes,
	// between the replays, are what is under test; checked by listing the recorded inputs, none of them is
	// under .cache.
	// Each failure carries the runs' stderr: a run that was not recorded says why there, and that line is what
	// tells a real miss from a slow machine (#q51f02a). Logged, so it shows only when the test fails.
	if _, stderr := run(first, "--no-fix"); stderr != "" {
		t.Logf("the recording run's stderr:\n%s", stderr)
	}
	for attempt := 1; attempt <= 2; attempt++ {
		if output, stderr := run(first, "--no-fix"); !strings.HasPrefix(output, "cached: ") {
			t.Fatalf("replay %d missed with the table inside the project:\n%s\n--- stderr\n%s", attempt, output, stderr)
		}
	}

	// Each shape of --no-cache leaves both caches exactly as they were. A writing types run makes the
	// build info first, so there is one to leave alone. Without the table, since its types section would let
	// that run replay every file and never open the build info; the run writes a new table too.
	removeCacheTable(t, first)
	run(first, "--types")
	// Then a new file, so the build info is behind the tree: an incremental check would rewrite it, and
	// only a run that does not use it leaves it alone.
	writeTree(t, first, map[string]string{"Extra.ts": "export const extra = 2;\n"})
	buildInfo := filepath.Join(first, "tsconfig.tsbuildinfo")
	snapshot := func() map[string]string {
		state := map[string]string{}
		for _, path := range append(cacheTableFiles(t, first), buildInfo) {
			information, err := os.Stat(path)
			if err != nil {
				t.Fatalf("%s: %v", path, err)
			}
			contents, _ := os.ReadFile(path)
			state[path] = information.ModTime().Format(time.RFC3339Nano) + " " + string(contents)
		}
		return state
	}
	cold := regexp.MustCompile(`(?m)^  cache: off, by --no-cache.*$`)
	durations := regexp.MustCompile(`\d+(\.\d+)?(ms|s|µs)\b`)
	// How a cached run's types and lint lines were paid for, which a cold run cannot share.
	replayedClause := regexp.MustCompile(`; \d+ of \d+ files(' semantic diagnostics)? replayed from cache`)
	findings := func(output string) string {
		kept := []string{}
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, " - ") || strings.HasPrefix(line, "lint: ") || strings.HasPrefix(line, "types: ") {
				kept = append(kept, foldNodeCounts(replayedClause.ReplaceAllString(durations.ReplaceAllString(line, "T"), "")))
			}
		}
		return strings.Join(kept, "\n")
	}
	// The warm side reads the record and the findings cache, and is not a whole-run replay, which leaves the
	// types and lint lines out: the same check in the other order is the same run under another run-cache key.
	run(first, "--no-fix", "--format")
	cached, _ := run(first, "--format", "--no-fix")
	// The writing shape goes last: it fixes and formats the tree, which the comparison before it must not
	// see.
	for _, arguments := range [][]string{{"--no-cache", "--no-fix", "--format"}, {"--no-cache", "--no-fix"}, {"--no-cache", "--lint"}, {"--no-cache", "--types"}, {"--no-cache", "--fix", "--format"}} {
		before := snapshot()
		output, stderr := run(first, arguments...)
		after := snapshot()
		// Both ways, so a table file the run added is caught as surely as one it rewrote.
		for path := range before {
			if before[path] != after[path] {
				t.Errorf("cohere %v changed %s\n--- output\n%s\n--- stderr\n%s", arguments, path, output, stderr)
			}
		}
		for path := range after {
			if _, existed := before[path]; !existed {
				t.Errorf("cohere %v created %s\n--- output\n%s\n--- stderr\n%s", arguments, path, output, stderr)
			}
		}
		if !cold.MatchString(output) {
			t.Errorf("cohere %v did not say it ran without the cache:\n%s", arguments, output)
		}
		if strings.HasPrefix(output, "cached: ") {
			t.Errorf("cohere %v replayed a cached run:\n%s", arguments, output)
		}
		if strings.Join(arguments, " ") == "--no-cache --no-fix --format" && findings(output) != findings(cached) {
			t.Errorf("--no-cache found something different from a cached run:\n--- cached\n%s\n--- --no-cache\n%s",
				findings(cached), findings(output))
		}
	}
	// The control: a run without --no-cache reads the record, so the comparison had a warm side.
	if !strings.Contains(cached, "not on record as formatted") {
		t.Fatalf("the warm side of the comparison never read the record:\n%s", cached)
	}
}
