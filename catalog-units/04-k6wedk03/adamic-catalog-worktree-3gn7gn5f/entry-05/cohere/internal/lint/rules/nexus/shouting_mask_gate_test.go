package nexus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
)

// The gates must not change what masking produces, only how fast it gets there.
func TestShoutingMaskGatesPreserveOutput(t *testing.T) {
	t.Parallel()

	var texts []string
	root := corpus.Structure.Path(t, "source")
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || len(texts) >= 800 {
			return nil
		}
		if !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".tsx") {
			return nil
		}
		if data, readError := os.ReadFile(path); readError == nil {
			texts = append(texts, string(data))
		}
		return nil
	})
	if len(texts) < 100 {
		t.Fatal("the structure corpus gave no comment texts, so there is nothing to compare the gates over")
	}

	// Ungated reference implementations, copied from before the gates.
	referenceCommand := func(text string) string {
		lines := strings.Split(text, "\n")
		for index, line := range lines {
			if commandLine.MatchString(jsDocGutter.ReplaceAllString(line, "")) {
				lines[index] = strings.Repeat(" ", len(line))
			}
		}
		return strings.Join(lines, "\n")
	}
	referenceExamples := func(text string) string {
		lines := strings.Split(text, "\n")
		insideExample := false
		for index, line := range lines {
			withoutGutter := jsDocGutter.ReplaceAllString(line, "")
			if exampleTag.MatchString(withoutGutter) {
				insideExample = true
				lines[index] = strings.Repeat(" ", len(line))
				continue
			}
			if insideExample && anyJsDocTag.MatchString(withoutGutter) {
				insideExample = false
			}
			if insideExample {
				lines[index] = strings.Repeat(" ", len(line))
			}
		}
		return strings.Join(lines, "\n")
	}

	commandDifferences, exampleDifferences := 0, 0
	for _, text := range texts {
		if maskCommandLines(text) != referenceCommand(text) {
			commandDifferences++
		}
		if maskJsDocExamples(text) != referenceExamples(text) {
			exampleDifferences++
		}
	}
	if commandDifferences != 0 || exampleDifferences != 0 {
		t.Fatalf("gates changed output: %d command diffs, %d example diffs across %d files",
			commandDifferences, exampleDifferences, len(texts))
	}
	t.Logf("gates preserve output across %d files", len(texts))
}

// The gates must actually skip work, or they are pure overhead.
func TestShoutingMaskGatesActuallySkip(t *testing.T) {
	t.Parallel()

	plain := "// just some ordinary prose about the thing\n// with no commands and no tags\n"
	if containsAnyCommandStarter(plain) {
		t.Fatalf("command gate admits plain prose")
	}
	if strings.ContainsRune(plain, '@') {
		t.Fatalf("example gate admits plain prose")
	}

	// And they must admit the real cases, or the rule goes quiet for them.
	for _, admitted := range []string{"$ ls", "git commit", "pnpm install", "curl x", "s c", "ahra os", "sqlite3 db", "npm run"} {
		if !containsAnyCommandStarter(admitted) {
			t.Fatalf("command gate rejects %q, which the pattern matches", admitted)
		}
	}
	// The whitespace arm is the one a literal cannot express, and a fixture is the only thing that
	// catches it: 800 real files contained no multi-space form, so the corpus diff passed while the
	// gate was wrong.
	for _, spacing := range []string{"s c", "s  c", "s\tc"} {
		if !containsAnyCommandStarter(spacing) {
			t.Fatalf("command gate rejects %q, which the pattern matches", spacing)
		}
	}
}

// The uppercase gate must not change which tokens are reported, only how fast the answer arrives.
//
// Compared token by token rather than by count, for the same reason the comment guard is: a gate
// that lost one shout and gained another would hold a count stable while silencing a real finding.
func TestShoutingUppercaseGatePreservesTokens(t *testing.T) {
	t.Parallel()

	comments := realCommentCorpus(t)

	ungated := func(body string) []string {
		masked := maskCodeAndCommands(body)
		var tokens []string
		seen := map[string]bool{}
		for _, token := range uppercaseToken.FindAllString(masked, -1) {
			if seen[token] || !isShoutedToken(token) {
				continue
			}
			seen[token] = true
			tokens = append(tokens, token)
		}
		return tokens
	}

	for _, comment := range comments {
		gated := shoutedTokensIn(comment)
		reference := ungated(comment)
		if len(gated) != len(reference) {
			t.Fatalf("gate changed the token count for %q: %v against %v", comment, gated, reference)
		}
		for index := range gated {
			if gated[index] != reference[index] {
				t.Fatalf("gate changed token %d for %q: %q against %q",
					index, comment, gated[index], reference[index])
			}
		}
	}
	t.Logf("uppercase gate preserves every token across %d real comments", len(comments))
}

// The cases a corpus cannot be relied on to hold, written out because that lesson has now cost twice.
func TestShoutingUppercaseGateAdmitsRealShouts(t *testing.T) {
	t.Parallel()

	// Every one of these must reach the expensive path.
	admitted := []string{
		"NEVER do this", "this is REALLY bad", "DO NOT EDIT",
		"a TODO marker", "the API contract", // acronyms: admitted, then declined downstream
		"MAX_RETRY_COUNT is an identifier", // underscore token
		"shouting at the very ENDOFLINE",
	}
	for _, body := range admitted {
		if !hasAdjacentUppercaseLetters(body) {
			t.Fatalf("gate rejects %q, which reaches the token scan", body)
		}
	}

	// And these must be skipped, or the gate is not a gate.
	skipped := []string{
		"ordinary prose about a thing",
		"a sentence with One capital",
		"CamelCase identifiers do not shout",
		"",
	}
	for _, body := range skipped {
		if hasAdjacentUppercaseLetters(body) {
			t.Fatalf("gate admits %q, which holds no adjacent uppercase pair", body)
		}
	}

	// The boundary: exactly two adjacent uppercase letters is the shortest reportable shape.
	if !hasAdjacentUppercaseLetters("it IS wrong") {
		t.Fatalf("gate rejects a two-letter shout")
	}
}
