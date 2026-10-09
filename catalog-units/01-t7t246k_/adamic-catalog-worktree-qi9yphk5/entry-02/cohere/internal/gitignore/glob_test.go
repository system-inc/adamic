package gitignore

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The glob cases gitignore(5) and fnmatch(3) describe, one per rule, so a regression names the rule it
// broke.
func TestGlobFollowsTheDocumentedRules(t *testing.T) {
	t.Parallel()
	cases := []struct {
		pattern string
		text    string
		path    bool
		want    bool
	}{
		{"foo", "foo", true, true},
		{"foo", "fo", true, false},
		{"f?o", "fxo", true, true},
		{"f?o", "f/o", true, false},
		{"f?o", "f/o", false, true},
		{"*.go", "main.go", true, true},
		{"*.go", "a/main.go", true, false},
		{"*.go", "a/main.go", false, true},
		{"a/*/c", "a/b/c", true, true},
		{"a/*/c", "a/b/x/c", true, false},
		{"**/foo", "foo", true, true},
		{"**/foo", "x/y/foo", true, true},
		{"**/foo/bar", "x/foo/bar", true, true},
		{"abc/**", "abc/x", true, true},
		{"abc/**", "abc/x/y", true, true},
		{"abc/**", "abc", true, false},
		{"a/**/b", "a/b", true, true},
		{"a/**/b", "a/x/b", true, true},
		{"a/**/b", "a/x/y/b", true, true},
		{"a/**/b", "a/xb", true, false},
		{"foo**/bar", "foo/bar", true, true},
		{"foo**/bar", "foox/bar", true, true},
		{"foo**/bar", "foo/x/bar", true, false},
		{"**foo", "x/foo", true, false},
		{"[abc]", "b", true, true},
		{"[!abc]", "b", true, false},
		{"[^abc]", "d", true, true},
		{"[a-c]x", "bx", true, true},
		{"[]]", "]", true, true},
		{"[a-]", "-", true, true},
		{"[[:digit:]]", "7", true, true},
		{"[[:digit:]]", "x", true, false},
		{"[[:nosuchclass:]]", "x", true, false},
		{"[/]", "/", true, false},
		{"[/]", "/", false, true},
		{"[abc", "[abc", true, false},
		{`\*`, "*", true, true},
		{`\*`, "x", true, false},
		{`\a`, "a", true, true},
		{`foo\`, "foo", true, false},
		{"*a*a*a*a*a*a*a*a*b", strings.Repeat("a", 200), true, false},
	}
	for _, testCase := range cases {
		compiled := compileGlob(testCase.pattern, testCase.path)
		if got := compiled.matches(testCase.text, testCase.path); got != testCase.want {
			t.Errorf("%q against %q (path %v): got %v, want %v", testCase.pattern, testCase.text, testCase.path, got, testCase.want)
		}
	}
}

// git's own wildmatch corpus, t/t3070-wildmatch.sh, read from a git source checkout named by
// COHERE_GIT_SOURCE. Nothing of git's is kept in this repository: the cases are read where they live.
//
// Each case gives four expectations. The two case-sensitive ones are asserted: wildmatch with
// WM_PATHNAME is how an anchored pattern matches a path, and without it how a base-name pattern
// matches. The two case-folding ones do not apply, since this matcher is case sensitive by design.
func TestGitWildmatchCorpus(t *testing.T) {
	t.Parallel()
	source := gitSource(t)
	contents, err := os.ReadFile(filepath.Join(source, "t", "t3070-wildmatch.sh"))
	if err != nil {
		t.Fatal(err)
	}

	asserted, cases := 0, 0
	for _, arguments := range matchLines(t, string(contents)) {
		expectations := arguments[:4]
		text, pattern := arguments[len(arguments)-2], arguments[len(arguments)-1]
		cases++
		for column, path := range map[int]bool{0: true, 2: false} {
			want := expectations[column]
			if want != "0" && want != "1" {
				t.Fatalf("case %q %q has expectation %q in column %d", text, pattern, want, column+1)
			}
			compiled := compileGlob(pattern, path)
			if got := compiled.matches(text, path); got != (want == "1") {
				t.Errorf("t3070: %q against %q with path %v: got %v, git expects %v", pattern, text, path, got, want == "1")
			}
			asserted++
		}
	}
	// A corpus that parsed to nothing would pass every assertion it never made.
	if cases < 150 {
		t.Fatalf("read only %d cases from t3070-wildmatch.sh; the parser has lost the corpus", cases)
	}
	t.Logf("t3070 at git %s: %d cases, %d expectations asserted (the case-folding columns do not apply)",
		gitCommit(t, source), cases, asserted)
}

// gitSource is the git checkout the corpus tests read, or a skip when none is named.
func gitSource(t *testing.T) string {
	t.Helper()
	source := os.Getenv("COHERE_GIT_SOURCE")
	if source == "" {
		t.Skip("COHERE_GIT_SOURCE does not name a git source checkout")
	}
	return source
}

// gitCommit names the commit a git checkout is at, so a run says which corpus it measured.
func gitCommit(t *testing.T, repository string) string {
	t.Helper()
	output, err := exec.Command("git", "-C", repository, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("reading the commit of %s: %v", repository, err)
	}
	return strings.TrimSpace(string(output))
}

// matchLines returns the arguments of every `match` call in t3070 after its function definitions, shell
// quoting undone and continued lines joined.
func matchLines(t *testing.T, script string) [][]string {
	t.Helper()
	var calls [][]string
	lines := strings.Split(script, "\n")
	for index := 0; index < len(lines); index++ {
		line := lines[index]
		if !strings.HasPrefix(line, "match ") {
			continue
		}
		for strings.HasSuffix(line, `\`) && index+1 < len(lines) {
			index++
			line = line[:len(line)-1] + " " + lines[index]
		}
		words := shellWords(t, strings.TrimPrefix(line, "match "))
		if len(words) != 6 && len(words) != 10 {
			t.Fatalf("t3070 line %d has %d arguments: %q", index+1, len(words), line)
		}
		calls = append(calls, words)
	}
	return calls
}

// shellWords splits a line the way sh does for the quoting t3070 uses: single quotes, double quotes with
// backslash escapes, and unquoted backslashes.
func shellWords(t *testing.T, line string) []string {
	t.Helper()
	var words []string
	var word strings.Builder
	inWord := false
	for index := 0; index < len(line); index++ {
		character := line[index]
		switch {
		case character == ' ' || character == '\t':
			if inWord {
				words = append(words, word.String())
				word.Reset()
				inWord = false
			}
		case character == '\'':
			closing := strings.IndexByte(line[index+1:], '\'')
			if closing < 0 {
				t.Fatalf("unclosed single quote in %q", line)
			}
			word.WriteString(line[index+1 : index+1+closing])
			index += closing + 1
			inWord = true
		case character == '"':
			index++
			for ; index < len(line) && line[index] != '"'; index++ {
				if line[index] == '\\' && index+1 < len(line) && strings.IndexByte("$`\"\\", line[index+1]) >= 0 {
					index++
				}
				word.WriteByte(line[index])
			}
			if index >= len(line) {
				t.Fatalf("unclosed double quote in %q", line)
			}
			inWord = true
		case character == '\\' && index+1 < len(line):
			index++
			word.WriteByte(line[index])
			inWord = true
		default:
			word.WriteByte(character)
			inWord = true
		}
	}
	if inWord {
		words = append(words, word.String())
	}
	return words
}
