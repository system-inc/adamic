package program_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/types/program"
)

// A cached finding has to be printable, and today it is not.
//
// LintCacheFinding stores MessageId and not the description. That is only safe if an id determines
// its text, because printRuleDiagnostic (cmd/cohere/main.go) prints Message.Description directly and
// nothing in the tree derives text from an id. Where an id does not determine its text, replaying a
// cached finding prints an empty or wrong message beside a correct file, range and rule — a cache
// serving a confidently wrong answer, which is the one failure mode here that looks like success.
//
// This test measures the property rather than asserting the conclusion, so it keeps being useful
// after the format is fixed: it reports which ids are ambiguous and why. Whichever fix lands
// (storing text, or refusing the dynamic ids), the numbers here are what that fix has to cover.
//
// The check is textual, like the ProgramReads guard, and for the same reason: it covers every rule
// in the tree rather than only the ones a program happened to exercise.

var (
	staticMessagePattern  = regexp.MustCompile(`Id:\s*"([A-Za-z0-9_]+)"\s*,\s*Description:\s*("(?:[^"\\]|\\.)*")`)
	dynamicMessagePattern = regexp.MustCompile(`Id:\s*"([A-Za-z0-9_]+)"\s*,\s*Description:\s*fmt\.Sprintf`)
)

// messageIdTexts reports, per message id, the distinct static texts seen for it, and separately the
// ids whose description is built at report time.
func messageIdTexts(t *testing.T) (staticTexts map[string]map[string]struct{}, dynamicIds map[string]struct{}) {
	t.Helper()

	staticTexts = map[string]map[string]struct{}{}
	dynamicIds = map[string]struct{}{}

	root := filepath.Join("..", "..", "lint", "rules")
	walked := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		walked++
		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		source := string(contents)
		for _, match := range staticMessagePattern.FindAllStringSubmatch(source, -1) {
			id, text := match[1], match[2]
			if staticTexts[id] == nil {
				staticTexts[id] = map[string]struct{}{}
			}
			staticTexts[id][text] = struct{}{}
		}
		for _, match := range dynamicMessagePattern.FindAllStringSubmatch(source, -1) {
			dynamicIds[match[1]] = struct{}{}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}

	// A scan that read nothing passes every assertion below for the wrong reason, which is the same
	// shape as the defect being measured.
	if walked == 0 {
		t.Fatal("found no rule source files, so this test proved nothing")
	}
	if len(staticTexts) == 0 {
		t.Fatalf("matched no Id/Description pairs across %d rule files, so the pattern is looking "+
			"for the wrong thing rather than the tree having no messages", walked)
	}
	return staticTexts, dynamicIds
}

// TestMessageIdAloneCannotIdentifyText is why LintCacheFinding stores the description.
//
// This measures the property the format depends on rather than asserting a conclusion, so it keeps
// earning its place: it names exactly which ids are ambiguous, and it fails if that population ever
// becomes empty. An empty population would mean the scan broke, not that the tree got cleaner —
// rules interpolating types into messages is normal and permanent.
//
// The stakes, restated because this is the comment a future reader checks before deciding the text
// field is redundant: nothing in this tree derives message text from an id. printRuleDiagnostic
// prints Message.Description directly and there is no id-to-text table. Drop the stored text and a
// warm run replays a finding with the right file, range, rule and id, and the wrong sentence.
func TestMessageIdAloneCannotIdentifyText(t *testing.T) {
	t.Parallel()
	staticTexts, dynamicIds := messageIdTexts(t)

	// Ids whose text is interpolated at report time. One id maps to many strings, so no table could
	// recover them even if someone built one.
	if len(dynamicIds) == 0 {
		t.Error("no message id builds its Description with fmt.Sprintf, which is implausible for " +
			"this tree: the dynamic pattern is not matching, and this test is measuring nothing")
	}

	// Ids reused across rules with different text. Recoverable from the pair (RuleName, MessageId),
	// which is why LintCacheFinding.MessageKey exists, but never from the id alone.
	ambiguous := map[string]int{}
	for id, texts := range staticTexts {
		if len(texts) > 1 {
			ambiguous[id] = len(texts)
		}
	}
	if len(ambiguous) == 0 {
		t.Error("no message id carries more than one description, which contradicts the " +
			"collision this format is built around: the static pattern is probably not matching")
	}

	names := make([]string, 0, len(dynamicIds))
	for id := range dynamicIds {
		names = append(names, id)
	}
	t.Logf("message ids that the id alone cannot render: %d interpolated (%s), %d reused across "+
		"rules with different text (%v)",
		len(dynamicIds), strings.Join(sortedNames(names), ", "), len(ambiguous), ambiguous)
}

// TestStoredDescriptionSurvivesForAnInterpolatedMessage is the round trip that matters most.
//
// The 11 interpolated ids are the population no table could ever recover, so a cached finding for
// one of them is the sharpest test of whether the format actually carries text: the stored sentence
// has to come back with its runtime value in it, not a template or a placeholder.
func TestStoredDescriptionSurvivesForAnInterpolatedMessage(t *testing.T) {
	t.Parallel()
	// The shape no_unsafe_unary_minus produces: the resolved type is interpolated, so this exact
	// sentence exists nowhere in the source and cannot be rebuilt from "unaryMinus".
	interpolated := "Argument of unary negation should be assignable to number | bigint but is string instead."

	cache := &program.LintCache{
		Key: program.HashRuleSet([]string{"no-unsafe-unary-minus"}),
		Entries: []program.LintCacheEntry{{
			Path:        "/project/source/negate.ts",
			ContentHash: program.HashContent("-\"text\";\n"),
			Rules:       []string{"no-unsafe-unary-minus"},
			Findings: []program.LintCacheFinding{{
				RuleName:           "no-unsafe-unary-minus",
				Start:              0,
				End:                8,
				MessageId:          "unaryMinus",
				MessageDescription: interpolated,
			}},
		}},
	}

	decoded := roundTripLintCache(t, cache)
	entry, hit := decoded.Lookup(
		"/project/source/negate.ts",
		program.HashContent("-\"text\";\n"),
		program.HashRuleSet([]string{"no-unsafe-unary-minus"}),
		[]string{"no-unsafe-unary-minus"},
	)
	if !hit {
		t.Fatal("the entry just encoded did not come back as a hit")
	}
	findings := entry.Findings
	if len(findings) != 1 {
		t.Fatalf("findings: %d back from 1", len(findings))
	}
	if findings[0].MessageDescription != interpolated {
		t.Errorf("the interpolated description did not survive the round trip:\n  got  %q\n  want %q",
			findings[0].MessageDescription, interpolated)
	}

	// The pair is the identity of a message, and the id alone is not. Pinned so a later refactor
	// keying on the id alone fails here rather than in a plausible wrong message.
	ruleName, messageId := findings[0].MessageKey()
	if ruleName != "no-unsafe-unary-minus" || messageId != "unaryMinus" {
		t.Errorf("MessageKey: (%q, %q)", ruleName, messageId)
	}
}

// TestMessageIdScanSeesBothKinds is the known-dirty control for the scan above.
//
// The two patterns are the whole instrument, and a pattern that silently matches nothing turns
// every assertion into a pass. This asserts the scan finds both kinds in a tree known to contain
// both, so a broken pattern fails here rather than reporting a clean format.
func TestMessageIdScanSeesBothKinds(t *testing.T) {
	t.Parallel()
	staticTexts, dynamicIds := messageIdTexts(t)

	if len(staticTexts) < 100 {
		t.Errorf("only %d static message ids matched, which is implausible for this tree: the "+
			"static pattern is probably not matching", len(staticTexts))
	}
	if len(dynamicIds) == 0 {
		t.Error("no message id matched the fmt.Sprintf pattern, which is implausible: rules do " +
			"interpolate types and names into descriptions, so the dynamic pattern is not matching")
	}
	t.Logf("scanned rule sources: %d static message ids, %d dynamic", len(staticTexts), len(dynamicIds))
}

// sortedNames keeps the failure message stable across runs, since map iteration is not ordered.
func sortedNames(names []string) []string {
	for outer := 1; outer < len(names); outer++ {
		for inner := outer; inner > 0 && names[inner] < names[inner-1]; inner-- {
			names[inner], names[inner-1] = names[inner-1], names[inner]
		}
	}
	return names
}
