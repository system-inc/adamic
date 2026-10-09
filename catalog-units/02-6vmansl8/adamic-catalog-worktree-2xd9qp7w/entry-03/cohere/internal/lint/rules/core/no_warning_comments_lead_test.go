package core

import (
	"testing"
)

// TestNoWarningCommentsLeadAnswersAsThePatternDoes holds the first-character check to the pattern it
// stands in front of: for every configuration and comment here, a matcher is skipped only where its
// pattern would not have matched. The comments are built to sit on each edge the check reasons about,
// and the test counts the matchers the check skipped, so a check that never skips cannot pass for
// being harmless.
func TestNoWarningCommentsLeadAnswersAsThePatternDoes(t *testing.T) {
	t.Parallel()

	configurations := []struct {
		name       string
		terms      []string
		location   NoWarningCommentsLocation
		decoration []string
	}{
		{"default", noWarningCommentsDefaultTerms, NoWarningCommentsStart, nil},
		{"decorated", noWarningCommentsDefaultTerms, NoWarningCommentsStart, []string{"*", "/"}},
		// `-` between two decoration characters spells a range in the class: `*` to `/` takes in `+`.
		{"a range in the decoration", noWarningCommentsDefaultTerms, NoWarningCommentsStart, []string{"*", "-", "/"}},
		// `k` folds to the Kelvin sign and `s` to the long s, so a decoration of either skips both cases.
		{"a decoration that folds", noWarningCommentsDefaultTerms, NoWarningCommentsStart, []string{"k"}},
		{"a term that opens with a skipped character", []string{"*todo", " fixme", "kelvin", "ſtop", "日本"}, NoWarningCommentsStart, []string{"*"}},
		{"an empty term", []string{""}, NoWarningCommentsStart, nil},
		{"anywhere", noWarningCommentsDefaultTerms, NoWarningCommentsAnywhere, nil},
	}
	comments := []string{
		"", " ", "\t\n", " TODO: x", "todo", "Todo later", "  fixme", "xxx", "XXX!", "something todo",
		"\v todo", " todo", "* todo", "** todo", "*/ todo", "+ todo", ", todo", "k todo", "K todo",
		"K todo", "Kelvin", "Kelvin", "ſtop", "Stop", "stop", "日本", "\xff todo", "todos",
		"*todo", "* *todo", " fixme", "  fixme",
	}

	skipped := 0
	for _, configuration := range configurations {
		compiled := noWarningCommentsMatchers(configuration.terms, configuration.location, configuration.decoration)
		for _, value := range comments {
			first, hasFirst := noWarningCommentsFirstUnskipped(value, compiled.skip)
			for index, matcher := range compiled.matchers {
				lead := compiled.leads[index]
				if lead == noWarningCommentsNoLead || (hasFirst && noWarningCommentsFoldEqual(lead, first)) {
					continue
				}
				skipped++
				if matcher.MatchString(value) {
					t.Errorf("%s: the check skips %q on %q, which its pattern %s matches",
						configuration.name, compiled.terms[index], value, matcher)
				}
			}
		}
	}
	if skipped == 0 {
		t.Fatal("the check skipped no matcher anywhere, so this test could not have failed")
	}
}
