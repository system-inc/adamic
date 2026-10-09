package nexus

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// The abbreviation vocabulary, read from cohere's policy/Abbreviations.json rather than written here as
// code.
//
// It is data because two engines read it. The Swift engine's naming rules must judge the same words
// this rule does, so the list lives in one file under policy/, which both engines compile in (#2cqemcd).
// Each entry carries its reason with it, so the file can be pruned by someone who was not there when it
// was written.
//
// What stays in Go is policy: the framework exemptions in the rule's listener, and the millisecond
// message, which cites a measurement about this tree rather than a fact about the word.
//
// # Order is part of the data
//
// Within each form the first matching entry reports, and that is observable: `timeoutMsRef` reports
// the millisecond form because `ms` precedes `ref` among the suffixes. The phases run in a fixed
// order in `vocabularyFinding`, and within a phase the file's order is the order. The move from Go
// literals to this file kept every phase's order and was checked by linting 82,704 names, every
// distinct identifier in the ahra tree plus every abbreviation in every form, with the binary
// before and after: the findings were byte-identical.
//
// # Two guards the literals carried, and why they did not come along
//
// Each prefix had an "already spelled" guard and each suffix an "unless it ends with" list, so that
// `database` would not read as `db` and `parsedValue` not as `Val`. Neither could ever fire. A
// prefix matches only when an uppercase letter follows the abbreviation (`^db[A-Z]`), and every
// expansion continues in lowercase (`database`), so no name both matched and was already spelled.
// A suffix guard needs a name ending in both `Val` and `Value`, which no string does. Checked over
// every entry rather than reasoned per word, then measured: dropping all of them left the
// 82,704-name run byte-identical. They were not moved into the data, because Swift would have copied them and
// someone would have tried to tune a guard that changes nothing.

// abbreviationVocabularyFile is the file's shape. Unknown keys are refused when it is read, so a
// misspelled form is a startup failure rather than an entry that silently never matches.
type abbreviationVocabularyFile struct {
	About           string                       `json:"about"`
	Abbreviations   []abbreviationEntry          `json:"abbreviations"`
	AllowedNames    []allowedAbbreviatedName     `json:"allowedNames"`
	AllowedSegments []allowedAbbreviationSegment `json:"allowedSegments"`
}

// abbreviationEntry is one abbreviated word and the forms it is judged in.
type abbreviationEntry struct {
	Abbreviation string `json:"abbreviation"`
	// Expansion is the word the abbreviation stands in for. Absent for a word that has no single
	// expansion, where `Advice` says what to write instead.
	Expansion string `json:"expansion"`
	// Advice replaces "Use <expansion>" in the whole-word message, and in the prefix message when
	// the prefix form's style is advice.
	Advice  string                   `json:"advice"`
	Reason  string                   `json:"reason"`
	Whole   *abbreviationWholeForm   `json:"whole"`
	Prefix  *abbreviationPrefixForm  `json:"prefix"`
	Suffix  *abbreviationSuffixForm  `json:"suffix"`
	Segment *abbreviationSegmentForm `json:"segment"`
}

// abbreviationWholeForm judges a name that is the abbreviation entire.
type abbreviationWholeForm struct {
	// Style is "orDescriptive" (use the expansion or a more descriptive name), "plain" (use the
	// expansion), or "advice" (the entry's advice).
	Style string `json:"style"`
}

// abbreviationPrefixForm judges a camelCase name that begins with the abbreviation.
type abbreviationPrefixForm struct {
	// Phase is "early", judged before any suffix, or "late", judged after them.
	Phase string `json:"phase"`
	// Style is empty for a rename to the expansion, or "advice" for the entry's advice.
	Style string `json:"style"`
}

// abbreviationSuffixForm judges a name that ends with the abbreviation capitalized.
type abbreviationSuffixForm struct {
	// Advice replaces the rename suggestion.
	Advice string `json:"advice"`
	// Matcher replaces the plain suffix test. The only one is "millisecondWord".
	Matcher string `json:"matcher"`
	// Replacement overrides the capitalized expansion as the text the suggestion writes.
	Replacement string `json:"replacement"`
}

// abbreviationSegmentForm judges the abbreviation capitalized as a camelCase word anywhere in a
// name.
type abbreviationSegmentForm struct{}

type allowedAbbreviatedName struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type allowedAbbreviationSegment struct {
	Segment string `json:"segment"`
	Reason  string `json:"reason"`
}

// abbreviationVocabulary is the file turned into the tables each phase walks.
type abbreviationVocabulary struct {
	wholeByName     map[string]*abbreviationEntry
	earlyPrefixes   []*abbreviationEntry
	latePrefixes    []*abbreviationEntry
	suffixes        []*abbreviationEntry
	segments        []*abbreviationEntry
	allowedNames    map[string]bool
	allowedSegments []string

	// prefixPatterns is `^<abbreviation>[A-Z]`, built once per entry rather than per name.
	prefixPatterns map[string]*regexp.Regexp
	// segmentPatterns are the two shapes a mid-name segment can take, plus the replacement form.
	segmentPatterns map[string]wordSegmentPatterns
}

// vocabulary is policy's file, read once. An invalid file panics at startup, which is the loud
// failure: a vocabulary read wrongly would make the rule quietly judge less than it appears to.
var vocabulary = mustLoadAbbreviationVocabulary(policy.AbbreviationsFile())

func mustLoadAbbreviationVocabulary(data []byte) *abbreviationVocabulary {
	loaded, err := loadAbbreviationVocabulary(data)
	if err != nil {
		panic(fmt.Sprintf("policy/Abbreviations.json: %v", err))
	}
	return loaded
}

func loadAbbreviationVocabulary(data []byte) (*abbreviationVocabulary, error) {
	var file abbreviationVocabularyFile
	if err := rule.UnmarshalOptions(data, &file); err != nil {
		return nil, err
	}

	loaded := &abbreviationVocabulary{
		wholeByName:     map[string]*abbreviationEntry{},
		allowedNames:    map[string]bool{},
		prefixPatterns:  map[string]*regexp.Regexp{},
		segmentPatterns: map[string]wordSegmentPatterns{},
	}
	seen := map[string]bool{}
	for index := range file.Abbreviations {
		entry := &file.Abbreviations[index]
		if err := validateAbbreviationEntry(entry); err != nil {
			return nil, fmt.Errorf("entry %d (%q): %w", index, entry.Abbreviation, err)
		}
		if seen[entry.Abbreviation] {
			return nil, fmt.Errorf("entry %q appears twice", entry.Abbreviation)
		}
		seen[entry.Abbreviation] = true

		if entry.Whole != nil {
			loaded.wholeByName[entry.Abbreviation] = entry
		}
		if entry.Prefix != nil {
			loaded.prefixPatterns[entry.Abbreviation] = regexp.MustCompile(`^` + entry.Abbreviation + `[A-Z]`)
			if entry.Prefix.Phase == "early" {
				loaded.earlyPrefixes = append(loaded.earlyPrefixes, entry)
			} else {
				loaded.latePrefixes = append(loaded.latePrefixes, entry)
			}
		}
		if entry.Suffix != nil {
			loaded.suffixes = append(loaded.suffixes, entry)
		}
		if entry.Segment != nil {
			word := capitalizeAbbreviation(entry.Abbreviation)
			loaded.segments = append(loaded.segments, entry)
			loaded.segmentPatterns[entry.Abbreviation] = wordSegmentPatterns{
				boundary: regexp.MustCompile(`(^|[^a-zA-Z])` + word + `($|[A-Z0-9])`),
				camel:    regexp.MustCompile(`[a-z0-9]` + word + `($|[A-Z0-9])`),
				replace:  regexp.MustCompile(word + `($|[A-Z0-9])`),
			}
		}
	}
	for _, allowed := range file.AllowedNames {
		loaded.allowedNames[allowed.Name] = true
	}
	for _, allowed := range file.AllowedSegments {
		loaded.allowedSegments = append(loaded.allowedSegments, allowed.Segment)
	}
	return loaded, nil
}

// validateAbbreviationEntry refuses an entry the engine would read as something other than meant.
func validateAbbreviationEntry(entry *abbreviationEntry) error {
	if entry.Abbreviation == "" || strings.ToLower(entry.Abbreviation) != entry.Abbreviation {
		return fmt.Errorf("the abbreviation must be lowercase and present")
	}
	if entry.Whole == nil && entry.Prefix == nil && entry.Suffix == nil && entry.Segment == nil {
		return fmt.Errorf("no form, so the entry judges nothing")
	}
	needsExpansion := false
	if entry.Whole != nil {
		switch entry.Whole.Style {
		case "orDescriptive", "plain":
			needsExpansion = true
		case "advice":
			if entry.Advice == "" {
				return fmt.Errorf("whole style is advice and the entry has none")
			}
		default:
			return fmt.Errorf("unknown whole style %q", entry.Whole.Style)
		}
	}
	if entry.Prefix != nil {
		if entry.Prefix.Phase != "early" && entry.Prefix.Phase != "late" {
			return fmt.Errorf("unknown prefix phase %q", entry.Prefix.Phase)
		}
		switch entry.Prefix.Style {
		case "":
			needsExpansion = true
		case "advice":
			if entry.Advice == "" {
				return fmt.Errorf("prefix style is advice and the entry has none")
			}
		default:
			return fmt.Errorf("unknown prefix style %q", entry.Prefix.Style)
		}
	}
	if entry.Suffix != nil {
		if entry.Suffix.Matcher != "" && entry.Suffix.Matcher != "millisecondWord" {
			return fmt.Errorf("unknown suffix matcher %q", entry.Suffix.Matcher)
		}
		if entry.Suffix.Advice == "" && entry.Suffix.Replacement == "" {
			needsExpansion = true
		}
	}
	if entry.Segment != nil {
		needsExpansion = true
	}
	if needsExpansion && entry.Expansion == "" {
		return fmt.Errorf("a form suggests the expansion and the entry has none")
	}
	return nil
}

// capitalizeAbbreviation is the spelling a suffix or segment form matches: `prop` becomes `Prop`.
func capitalizeAbbreviation(word string) string {
	if word == "" {
		return word
	}
	return strings.ToUpper(word[:1]) + word[1:]
}

// abbreviationFinding is what the vocabulary says about a name before any framework exemption.
type abbreviationFinding struct {
	// form is "whole", "prefix", "suffix" or "segment", and entry is the word that matched. The
	// listener's framework exemptions are keyed on the two together.
	form    string
	entry   *abbreviationEntry
	message rule.Message
}

// containsAllowedSegment reports whether a name contains a word that merely holds an
// abbreviation's letters, like `InnoDb`.
func (vocabulary *abbreviationVocabulary) containsAllowedSegment(name string) bool {
	for _, segment := range vocabulary.allowedSegments {
		if strings.Contains(name, segment) {
			return true
		}
	}
	return false
}

// allowedSegmentHolds reports whether an allowed segment in the name contains this abbreviation's
// letters, which is what exempts `db` from an `InnoDb` name's early prefix check.
func (vocabulary *abbreviationVocabulary) allowedSegmentHolds(name string, abbreviation string) bool {
	for _, segment := range vocabulary.allowedSegments {
		if strings.Contains(name, segment) && strings.Contains(strings.ToLower(segment), abbreviation) {
			return true
		}
	}
	return false
}

// find answers what the vocabulary reports for a name, if anything, in the phase order the rule has
// always used: whole word, early prefix, suffix, late prefix, segment.
//
// It knows nothing about where the name sits. The listener asks it only after the cheap gate and
// the foreign-name skips, and then silences the framework exemptions by the form and word it
// returns, which is exactly what each exemption did when it was a `return` at that branch.
func (vocabulary *abbreviationVocabulary) find(name string) (abbreviationFinding, bool) {
	if vocabulary.allowedNames[name] {
		return abbreviationFinding{}, false
	}

	if entry, found := vocabulary.wholeByName[name]; found {
		return abbreviationFinding{form: "whole", entry: entry, message: wholeAbbreviationMessage(entry, name)}, true
	}

	for _, entry := range vocabulary.earlyPrefixes {
		if vocabulary.allowedSegmentHolds(name, entry.Abbreviation) {
			continue
		}
		if vocabulary.prefixPatterns[entry.Abbreviation].MatchString(name) {
			return abbreviationFinding{form: "prefix", entry: entry, message: prefixAbbreviationMessage(entry, name)}, true
		}
	}

	for _, entry := range vocabulary.suffixes {
		if entry.Suffix.Matcher == "millisecondWord" {
			if millisecondSegmentPattern.MatchString(name) {
				replacement := "${1}" + entry.Suffix.Replacement + "${2}"
				suggestion := replaceFirst(millisecondSegmentPattern, name, replacement)
				return abbreviationFinding{form: "suffix", entry: entry, message: messageNoMsSuffix(name, suggestion)}, true
			}
			continue
		}
		suffix := capitalizeAbbreviation(entry.Abbreviation)
		if !strings.HasSuffix(name, suffix) {
			continue
		}
		return abbreviationFinding{form: "suffix", entry: entry, message: suffixAbbreviationMessage(entry, name, suffix)}, true
	}

	for _, entry := range vocabulary.latePrefixes {
		if !vocabulary.prefixPatterns[entry.Abbreviation].MatchString(name) {
			continue
		}
		return abbreviationFinding{form: "prefix", entry: entry, message: prefixAbbreviationMessage(entry, name)}, true
	}

	if vocabulary.containsAllowedSegment(name) {
		return abbreviationFinding{}, false
	}

	for _, entry := range vocabulary.segments {
		patterns := vocabulary.segmentPatterns[entry.Abbreviation]
		if !patterns.boundary.MatchString(name) && !patterns.camel.MatchString(name) {
			continue
		}
		suggestion := replaceFirst(patterns.replace, name, capitalizeAbbreviation(entry.Expansion)+"${1}")
		if suggestion == name {
			continue
		}
		word := capitalizeAbbreviation(entry.Abbreviation)
		return abbreviationFinding{form: "segment", entry: entry, message: messageNoWordSegment(name, word, suggestion)}, true
	}

	return abbreviationFinding{}, false
}

// wholeAbbreviationMessage is the finding for a name that is the abbreviation entire.
func wholeAbbreviationMessage(entry *abbreviationEntry, name string) rule.Message {
	switch entry.Whole.Style {
	case "advice":
		return abbreviatedIdentifierMessage(name, entry.Advice)
	case "plain":
		return abbreviatedIdentifierMessage(name, `Use "`+entry.Expansion+`".`)
	}
	return abbreviatedIdentifierMessage(name,
		`Use "`+entry.Expansion+`" or a more descriptive name.`)
}

// prefixAbbreviationMessage is the finding for a camelCase name that begins with the abbreviation.
//
// An advice-style prefix names the abbreviation rather than the name, so `argCount` reads
// `Identifier "arg" should not be abbreviated`. That is how the role words have always reported.
func prefixAbbreviationMessage(entry *abbreviationEntry, name string) rule.Message {
	if entry.Prefix.Style == "advice" {
		return abbreviatedIdentifierMessage(entry.Abbreviation, entry.Advice)
	}
	suggestion := entry.Expansion + strings.TrimPrefix(name, entry.Abbreviation)
	return abbreviatedIdentifierMessage(name, `Use "`+suggestion+`".`)
}

// suffixAbbreviationMessage is the finding for a name that ends with the abbreviation capitalized.
func suffixAbbreviationMessage(entry *abbreviationEntry, name string, suffix string) rule.Message {
	advice := entry.Suffix.Advice
	if advice == "" {
		replacement := entry.Suffix.Replacement
		if replacement == "" {
			replacement = capitalizeAbbreviation(entry.Expansion)
		}
		advice = `Use "` + strings.TrimSuffix(name, suffix) + replacement + `".`
	}
	return rule.Message{
		Id:          abbreviatedSuffixText.Id,
		Description: abbreviatedSuffixText.Render(map[string]string{"name": name, "suffix": suffix, "advice": advice}),
	}
}

// abbreviatedIdentifierMessage is the shape every whole-word and prefix finding takes.
func abbreviatedIdentifierMessage(name string, advice string) rule.Message {
	return rule.Message{
		Id:          abbreviatedIdentifierText.Id,
		Description: abbreviatedIdentifierText.Render(map[string]string{"name": name, "advice": advice}),
	}
}
