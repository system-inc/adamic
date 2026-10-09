package gitignore

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// git's check-ignore tests, t/t0008-ignores.sh, read from the git source checkout named by
// COHERE_GIT_SOURCE, as the wildmatch corpus is. Nothing of git's is kept in this repository.
//
// The script's fixture is rebuilt from its setup test (its ignore files, directories and links), and every
// expectation it states in check-ignore's verbose form (`file:line:pattern<TAB>path`, or `::<TAB>path`
// for a path no rule matched) is asserted against the matcher. A path is resolved from the directory its
// test changes into, and an expectation kept in a heredoc from the directory of the tests that read it.
//
// Not asserted, each counted: expectations naming the global excludes file (not read, by design), a
// tracked file under the index (the matcher has no index, as `--no-index` has none), and the standalone
// tests that rewrite the fixture and state plain path lists (exact prefix matching, `**` with directories,
// whitespace through `ls-files -X`, links, oversized files). Those behaviors are held by this package's own
// tests and by the differential against git.
func TestGitCheckIgnoreCorpus(t *testing.T) {
	t.Parallel()
	source := gitSource(t)
	contents, err := os.ReadFile(filepath.Join(source, "t", "t0008-ignores.sh"))
	if err != nil {
		t.Fatal(err)
	}
	script := strings.Split(string(contents), "\n")

	root := t.TempDir()
	buildCheckIgnoreFixture(t, root, script)
	matcher, err := New(root)
	if err != nil {
		t.Fatal(err)
	}

	asserted, skippedGlobal, skippedIndex := 0, 0, 0
	seen := map[string]string{}
	for _, expectation := range checkIgnoreExpectations(t, script) {
		switch {
		case strings.Contains(expectation.file, "global_excludes"):
			skippedGlobal++
			continue
		case expectation.indexMode && strings.Contains(expectation.path, "ignored-but-in-index"):
			skippedIndex++
			continue
		}
		resolved := path.Clean(path.Join(expectation.directory, expectation.path))
		if resolved == "." {
			continue
		}
		want := ""
		if expectation.file != "" {
			want = expectation.file + ":" + expectation.line + ":" + expectation.pattern
		}
		if previous, present := seen[resolved]; present && previous != want {
			t.Fatalf("t0008 states two answers for %s: %q and %q", resolved, previous, want)
		}
		seen[resolved] = want

		info, statError := os.Lstat(filepath.Join(root, filepath.FromSlash(resolved)))
		isDirectory := statError == nil && info.IsDir()
		ignored, decided := decide(t, matcher, resolved, isDirectory)
		got := ""
		if !decided.IsZero() {
			got = decided.String()
		}
		wantIgnored := want != "" && !strings.HasPrefix(expectation.pattern, "!")
		if got != want || ignored != wantIgnored {
			t.Errorf("t0008 line %d: %s: got %v by %q, git expects %v by %q", expectation.scriptLine, resolved, ignored, got, wantIgnored, want)
		}
		asserted++
	}
	if len(seen) < 25 {
		t.Fatalf("read only %d distinct paths from t0008-ignores.sh; the reader has lost the corpus", len(seen))
	}
	t.Logf("t0008 at git %s: %d expectations asserted over %d distinct paths; not asserted: %d naming the global excludes file, %d about a tracked file under the index",
		gitCommit(t, source), asserted, len(seen), skippedGlobal, skippedIndex)
}

// checkIgnoreExpectation is one verbose line t0008 expects check-ignore to print.
type checkIgnoreExpectation struct {
	file, line, pattern, path string
	directory                 string
	indexMode                 bool
	scriptLine                int
}

var (
	testStart       = regexp.MustCompile(`^\s*test_expect_success(_multi|_no_index_multi)?\b`)
	changeDirectory = regexp.MustCompile(`^\s*\(?\s*cd ([^ &)]+) &&`)
	topLevelHeredoc = regexp.MustCompile(`^cat <<-(\\?)EOF >(\S+)$`)
	readsFile       = regexp.MustCompile(`<(?:\.\./)?(expected-[a-z]+)\b`)
	verboseLine     = regexp.MustCompile(`^(::|([^\s:]+):(\d+):([^\t]*))\t(.+)$`)
)

// checkIgnoreExpectations finds every verbose expectation in the script, with the directory its paths are
// relative to and whether the test runs against the index.
func checkIgnoreExpectations(t *testing.T, script []string) []checkIgnoreExpectation {
	t.Helper()
	// Each test's span, the directory it changes into, and whether it reads the index.
	type testSpan struct {
		start, end int
		directory  string
		indexMode  bool
	}
	var tests []testSpan
	for index, line := range script {
		if !testStart.MatchString(line) || strings.Contains(line, "()") {
			continue
		}
		if len(tests) > 0 {
			tests[len(tests)-1].end = index
		}
		tests = append(tests, testSpan{start: index, end: len(script), indexMode: !strings.Contains(line, "no_index")})
	}
	for index := range tests {
		for _, line := range script[tests[index].start:tests[index].end] {
			if match := changeDirectory.FindStringSubmatch(line); match != nil {
				tests[index].directory = match[1]
				break
			}
		}
	}
	enclosing := func(line int) *testSpan {
		for index := range tests {
			if line >= tests[index].start && line < tests[index].end {
				return &tests[index]
			}
		}
		return nil
	}

	// A heredoc written at the top level is read by tests after it, until it is written again; its lines
	// are relative to the directory those tests run in.
	heredocDirectory := func(name string, after int) (string, bool, bool) {
		directories := map[string]bool{}
		indexMode := false
		for index := after + 1; index < len(script); index++ {
			if match := topLevelHeredoc.FindStringSubmatch(script[index]); match != nil && match[2] == name {
				break
			}
			for _, match := range readsFile.FindAllStringSubmatch(script[index], -1) {
				if match[1] == name {
					if test := enclosing(index); test != nil {
						directories[test.directory] = true
						indexMode = indexMode || test.indexMode
					}
				}
			}
		}
		if len(directories) != 1 {
			return "", false, false
		}
		for directory := range directories {
			return directory, indexMode, true
		}
		return "", false, false
	}

	var expectations []checkIgnoreExpectation
	heredocEnd, heredocName, heredocUnquoted := -1, "", false
	for index, raw := range script {
		if match := topLevelHeredoc.FindStringSubmatch(raw); match != nil {
			heredocName, heredocUnquoted = match[2], match[1] == ""
			heredocEnd = index + 1
			for heredocEnd < len(script) && strings.TrimLeft(script[heredocEnd], "\t") != "EOF" {
				heredocEnd++
			}
			continue
		}
		insideHeredoc := index < heredocEnd
		line := strings.TrimSpace(raw)
		if insideHeredoc {
			// A heredoc line is the output itself, quotes and all; an unquoted heredoc undoes `\\` first.
			if heredocUnquoted {
				line = strings.ReplaceAll(line, `\\`, `\`)
			}
		} else {
			// Outside one, an expectation is part of a shell string: drop the quoting around it.
			line = strings.TrimSuffix(strings.TrimSuffix(line, `\`), " ")
			line = strings.TrimSuffix(strings.TrimSuffix(line, " '"), " &&")
			line = strings.TrimPrefix(line, "expect ")
			line = strings.Trim(line, `"'`)
		}
		match := verboseLine.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		var directory string
		var indexMode bool
		if insideHeredoc {
			var ok bool
			directory, indexMode, ok = heredocDirectory(heredocName, index)
			if !ok {
				continue
			}
		} else if test := enclosing(index); test != nil {
			directory, indexMode = test.directory, test.indexMode
		} else {
			continue
		}

		paths := []string{match[5]}
		if strings.Contains(match[5], "${subdir}") {
			paths = []string{strings.ReplaceAll(match[5], "${subdir}", ""), strings.ReplaceAll(match[5], "${subdir}", "a/")}
		}
		for _, expectedPath := range paths {
			if strings.HasPrefix(expectedPath, `"`) {
				unquoted, err := strconv.Unquote(expectedPath)
				if err != nil {
					t.Fatalf("t0008 line %d: path %s does not unquote: %v", index+1, expectedPath, err)
				}
				expectedPath = unquoted
			}
			expectation := checkIgnoreExpectation{path: expectedPath, directory: directory, indexMode: indexMode, scriptLine: index + 1}
			if match[1] != "::" {
				expectation.file, expectation.line, expectation.pattern = match[2], match[3], match[4]
				if strings.HasPrefix(expectation.file, "$") {
					expectation.file = "global_excludes"
				}
			}
			expectations = append(expectations, expectation)
		}
	}
	return expectations
}

// buildCheckIgnoreFixture rebuilds what t0008's setup test makes: its directories, its link, its nested
// repository and its ignore files. The files it creates only so they exist are not needed: a path's
// answer depends on whether it is a directory, not on whether it exists.
func buildCheckIgnoreFixture(t *testing.T, root string, script []string) {
	t.Helper()
	start, end := -1, -1
	for index, line := range script {
		if strings.HasPrefix(line, "test_expect_success 'setup'") {
			start = index
		} else if start >= 0 && line == "'" {
			end = index
			break
		}
	}
	if start < 0 || end < 0 {
		t.Fatal("t0008 has no setup test to rebuild the fixture from")
	}

	mustMkdir := func(relative string) {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(relative)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mustMkdir(".git/info")
	heredoc := regexp.MustCompile(`cat <<-\\EOF >(\S+)`)
	echo := regexp.MustCompile(`echo "([^"]*)" >(\S+)`)
	ignoreFiles := 0
	currentDirectory := ""
	for index := start; index < end; index++ {
		line := strings.TrimSpace(script[index])
		switch {
		case strings.HasPrefix(line, "mkdir -p "):
			for _, directory := range strings.Fields(strings.TrimSuffix(strings.TrimPrefix(line, "mkdir -p "), " &&")) {
				mustMkdir(directory)
			}
		case strings.HasPrefix(line, "ln -s "):
			fields := strings.Fields(line)
			if err := os.Symlink(fields[2], filepath.Join(root, filepath.FromSlash(fields[3]))); err != nil {
				t.Fatal(err)
			}
		case strings.HasPrefix(line, "cd "):
			currentDirectory = strings.TrimSuffix(strings.TrimPrefix(line, "cd "), " &&")
		case strings.HasPrefix(line, "git init"):
			mustMkdir(path.Join(currentDirectory, ".git"))
		case strings.HasPrefix(line, ")"):
			currentDirectory = ""
		case heredoc.MatchString(line):
			target := heredoc.FindStringSubmatch(line)[1]
			var body []string
			for index++; strings.TrimLeft(script[index], "\t") != "EOF"; index++ {
				body = append(body, strings.TrimLeft(script[index], "\t"))
			}
			if strings.Contains(target, "$") {
				continue
			}
			writeTree(t, root, map[string]string{target: strings.Join(body, "\n") + "\n"})
			ignoreFiles++
		case echo.MatchString(line):
			match := echo.FindStringSubmatch(line)
			writeTree(t, root, map[string]string{match[2]: match[1] + "\n"})
			ignoreFiles++
		}
	}
	if ignoreFiles < 4 {
		t.Fatalf("rebuilt only %d ignore files from t0008's setup", ignoreFiles)
	}
}
