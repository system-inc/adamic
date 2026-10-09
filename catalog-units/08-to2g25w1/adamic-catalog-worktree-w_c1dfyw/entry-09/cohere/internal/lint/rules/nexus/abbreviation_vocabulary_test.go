package nexus

import (
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/cohere/policy"
)

// TestAbbreviationVocabularyCarriesEveryWordInOrder pins what the file must hold, form by form.
//
// These lists are the Go literals the vocabulary was moved out of, read off the rule at cohere
// d53901e, so the move is checked as count in and count out with the order kept. They are a second
// copy on purpose: an expectation derived from the file could not notice an entry dropped from it.
// Adding a word means adding it here too, which is the moment to decide where in its phase it runs.
//
// Order matters only where entries can both match one name, and the test pins all of it anyway:
// `ms` between `params` and `ref` is observable (`timeoutMsRef` reports the millisecond form), and
// `vars` before `var` decides which replacement a name gets.
func TestAbbreviationVocabularyCarriesEveryWordInOrder(t *testing.T) {
	t.Parallel()

	abbreviationsOf := func(entries []*abbreviationEntry) []string {
		words := make([]string, 0, len(entries))
		for _, entry := range entries {
			words = append(words, entry.Abbreviation)
		}
		return words
	}
	wholeWords := make([]string, 0, len(vocabulary.wholeByName))
	for word := range vocabulary.wholeByName {
		wholeWords = append(wholeWords, word)
	}
	slices.Sort(wholeWords)

	cases := []struct {
		form string
		got  []string
		want []string
	}{
		// The switch (13), the early group (7) and exactAbbreviations (10). Whole-word matches cannot
		// overlap, so this one is compared as a set.
		{"whole", wholeWords, sortedCopy([]string{
			"prop", "props", "param", "params", "ref", "config", "idx", "arg", "args", "acc", "char", "fn", "str",
			"ctx", "db", "tx", "opts", "cur", "pct", "prev",
			"val", "arr", "obj", "num", "res", "err", "req", "msg", "min", "max",
		})},
		// The early group's loop, in its order.
		{"early prefix", abbreviationsOf(vocabulary.earlyPrefixes), []string{
			"ctx", "db", "tx", "opts", "cur", "pct", "prev",
		}},
		// The late group (9), the role prefixes (3) and prefixAbbreviations (10). No two of these can
		// match one name, so their relative order was never observable; the file's order is pinned.
		{"late prefix", abbreviationsOf(vocabulary.latePrefixes), []string{
			"prop", "props", "param", "params", "ref", "config", "idx", "arg", "args", "char", "fn", "str",
			"val", "arr", "obj", "num", "res", "err", "req", "msg", "min", "max",
		}},
		// The hand-written suffix chain (12), the millisecond check where it ran, and
		// suffixAbbreviations (10), in the chain's order.
		{"suffix", abbreviationsOf(vocabulary.suffixes), []string{
			"prop", "props", "param", "params", "ms", "ref", "config", "idx", "arg", "args", "char", "fn", "str",
			"val", "arr", "obj", "num", "res", "err", "req", "msg", "min", "max",
		}},
		// wordSegmentAbbreviations (10), in order.
		{"segment", abbreviationsOf(vocabulary.segments), []string{
			"cwd", "dir", "env", "cli", "len", "seq", "db", "tx", "vars", "var",
		}},
		{"allowed segment", vocabulary.allowedSegments, []string{"InnoDb"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.form, func(t *testing.T) {
			t.Parallel()
			if !slices.Equal(testCase.got, testCase.want) {
				t.Errorf("%d entries %v, want %d %v", len(testCase.got), testCase.got, len(testCase.want), testCase.want)
			}
		})
	}
	if len(vocabulary.allowedNames) != 1 || !vocabulary.allowedNames["URLSearchParams"] {
		t.Errorf("allowed names %v, want only URLSearchParams", vocabulary.allowedNames)
	}
}

// TestAbbreviationVocabularyJudgesEveryForm runs one generated name through every form of every
// entry and checks the message id and the suggestion each produces.
//
// The table above proves the entries are present; this proves each one is wired to the branch it
// names, which a form read into the wrong table would fail.
func TestAbbreviationVocabularyJudgesEveryForm(t *testing.T) {
	t.Parallel()

	judged := 0
	for word := range vocabulary.wholeByName {
		finding, found := vocabulary.find(word)
		if !found || finding.form != "whole" || finding.message.Id != abbreviatedIdentifierText.Id {
			t.Errorf("whole %q: got %+v, want %s", word, finding, abbreviatedIdentifierText.Id)
		}
		judged++
	}
	for _, entry := range append(append([]*abbreviationEntry{}, vocabulary.earlyPrefixes...), vocabulary.latePrefixes...) {
		name := entry.Abbreviation + "Widget"
		finding, found := vocabulary.find(name)
		if !found || finding.form != "prefix" || finding.message.Id != abbreviatedIdentifierText.Id {
			t.Errorf("prefix %q: got %+v, want %s", name, finding, abbreviatedIdentifierText.Id)
		}
		judged++
	}
	for _, entry := range vocabulary.suffixes {
		name := "widget" + capitalizeAbbreviation(entry.Abbreviation)
		wantId := abbreviatedSuffixText.Id
		if entry.Suffix.Matcher == "millisecondWord" {
			wantId = millisecondSuffixText.Id
		}
		finding, found := vocabulary.find(name)
		if !found || finding.form != "suffix" || finding.message.Id != wantId {
			t.Errorf("suffix %q: got %+v, want %s", name, finding, wantId)
		}
		judged++
	}
	for _, entry := range vocabulary.segments {
		name := "widget" + capitalizeAbbreviation(entry.Abbreviation) + "Name"
		finding, found := vocabulary.find(name)
		if !found || finding.form != "segment" || finding.entry != entry {
			t.Errorf("segment %q: got %+v, want %s", name, finding, entry.Abbreviation)
		}
		judged++
	}
	// 30 whole, 29 prefix, 23 suffix, 10 segment: the counts the table above pins.
	if judged != 92 {
		t.Errorf("judged %d forms, want 92", judged)
	}
}

// TestAbbreviationVocabularyRefusesAMalformedFile checks the loader fails loudly rather than
// reading a file into a vocabulary that quietly judges less.
func TestAbbreviationVocabularyRefusesAMalformedFile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		file string
	}{
		{"a misspelled form", `{"abbreviations": [{"abbreviation": "val", "expansion": "value", "suffx": {}}]}`},
		{"an entry with no form", `{"abbreviations": [{"abbreviation": "val", "expansion": "value"}]}`},
		{"an unknown whole style", `{"abbreviations": [{"abbreviation": "val", "expansion": "value", "whole": {"style": "loud"}}]}`},
		{"a rename with no expansion", `{"abbreviations": [{"abbreviation": "val", "whole": {"style": "plain"}}]}`},
		{"an unknown prefix phase", `{"abbreviations": [{"abbreviation": "val", "expansion": "value", "prefix": {"phase": "middle"}}]}`},
		{"an uppercase abbreviation", `{"abbreviations": [{"abbreviation": "Val", "expansion": "value", "segment": {}}]}`},
		{"a duplicate entry", `{"abbreviations": [{"abbreviation": "val", "expansion": "value", "segment": {}}, {"abbreviation": "val", "expansion": "value", "segment": {}}]}`},
		{"an unknown matcher", `{"abbreviations": [{"abbreviation": "ms", "expansion": "m", "suffix": {"matcher": "seconds"}}]}`},
		{"a retired per-word messageId", `{"abbreviations": [{"abbreviation": "val", "expansion": "value", "whole": {"messageId": "noVal", "style": "plain"}}]}`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if _, err := loadAbbreviationVocabulary([]byte(testCase.file)); err == nil {
				t.Errorf("the loader accepted it")
			}
		})
	}
	if _, err := loadAbbreviationVocabulary(policy.AbbreviationsFile()); err != nil {
		t.Fatalf("policy's file itself is refused: %v", err)
	}
}

func sortedCopy(words []string) []string {
	copied := slices.Clone(words)
	slices.Sort(copied)
	return copied
}

// TestAbbreviationVocabularyOrderIsObservableWhereItMatters pins the two orderings a name can see.
//
// The order table above would also fail on a harmless reorder, so these say which reorders change a
// finding. `var` cannot match inside `Vars`, since the lowercase `s` is no word boundary, so `vars`
// before `var` is visible only in a name holding both words. No name in the ahra corpus does, which
// is why the 82,704-name run could not see a swap and this case exists.
func TestAbbreviationVocabularyOrderIsObservableWhereItMatters(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		wantId      string
		wantMessage string
	}{
		{"themeVarsVar", "abbreviatedWordSegment", `Use "themeVariablesVar".`},
		{"timeoutMsRef", "millisecondSuffix", `Use "timeoutInMillisecondsRef"`},
		{"dbVal", "abbreviatedIdentifier", `Use "databaseVal".`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			finding, found := vocabulary.find(testCase.name)
			if !found || finding.message.Id != testCase.wantId {
				t.Fatalf("got %+v, want %s", finding, testCase.wantId)
			}
			if !strings.Contains(finding.message.Description, testCase.wantMessage) {
				t.Errorf("message %q does not contain %q", finding.message.Description, testCase.wantMessage)
			}
		})
	}
}
