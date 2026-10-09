package main

import (
	"strings"
	"testing"
)

/*
 * A finding prints as one line, whatever its message does internally.
 *
 * Two rules write a second paragraph into their description with an embedded newline:
 * `no-misused-spread` and `no-unsafe-assignment`. The printer emitted that newline verbatim, so the
 * rule tag landed on a continuation line carrying no position.
 *
 * The cost is silent and it is biased. Any reader keyed on `position ... [rule]` drops those
 * findings, and it drops the LONGEST messages, which are the ones carrying advice. Four of 5,312
 * findings on the ahra tree wrapped that way. An extractor of mine missed them and I spent four
 * hours instrumenting a rule that was working correctly, chasing a gap that existed only in the
 * tool reading the output.
 *
 * This test asserts the property rather than the two rules, because the next rule to want a second
 * paragraph should not have to know about it. The message is the rule's to write; the line
 * discipline is the printer's to keep.
 */
func TestDiagnosticDescriptionIsASingleLine(t *testing.T) {
	t.Parallel()

	for name, testCase := range map[string]struct {
		description string
		want        string
	}{
		"plain": {
			description: "Unsafe member access .value on an `any` value.",
			want:        "Unsafe member access .value on an `any` value.",
		},
		"twoParagraphs": {
			description: "Using the spread operator on a string can mishandle characters.\n" +
				"Consider `Intl.Segmenter` instead.",
			want: "Using the spread operator on a string can mishandle characters. " +
				"Consider `Intl.Segmenter` instead.",
		},
		"severalNewlines": {
			description: "first\nsecond\nthird",
			want:        "first second third",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := singleLineDescription(testCase.description)
			if got != testCase.want {
				t.Errorf("description = %q, want %q", got, testCase.want)
			}
			if strings.Contains(got, "\n") {
				t.Errorf("description still contains a newline: %q", got)
			}
		})
	}
}
