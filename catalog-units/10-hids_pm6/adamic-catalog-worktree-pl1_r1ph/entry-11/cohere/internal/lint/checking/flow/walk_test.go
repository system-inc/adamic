package flow

import "testing"

// TestPathTextReadsAsTheReaderReachesThePart: the slot a finding names is the one a reader would write
// through, so the path reads from the whole value inward. The walks themselves are exercised through the
// rules that judge them, in internal/lint/rules/adamic, against the probes on #drbrp8c.
func TestPathTextReadsAsTheReaderReachesThePart(t *testing.T) {
	t.Parallel()
	for want, path := range map[string][]Step{
		"":                {},
		"[]":              {{Kind: StepElement, Index: -1}},
		"[1]":             {{Kind: StepElement, Index: 1}},
		".pets[]":         {{Kind: StepProperty, Name: "pets"}, {Kind: StepElement, Index: -1}},
		"<Map value>":     {{Kind: StepTypeArgument, Name: "Map", Index: 1}},
		"<WeakMap key>":   {{Kind: StepTypeArgument, Name: "WeakMap", Index: 0}},
		"<Set member>":    {{Kind: StepTypeArgument, Name: "Set", Index: 0}},
		".onDone(animal)": {{Kind: StepProperty, Name: "onDone"}, {Kind: StepParameter, Name: "animal", Index: 0}},
		".make()":         {{Kind: StepProperty, Name: "make"}, {Kind: StepReturn, Index: -1}},
		// A well-known symbol's internal name carries a creation id, which no message may print (#y0ejf6a).
		"[iterator]().next()": {{Kind: StepProperty, Name: "\xfe@iterator@224"}, {Kind: StepReturn, Index: -1}, {Kind: StepProperty, Name: "next"}, {Kind: StepReturn, Index: -1}},
		`["content-type"]`:    {{Kind: StepProperty, Name: "content-type"}},
	} {
		if got := PathText(path); got != want {
			t.Errorf("%+v reads %q, want %q", path, got, want)
		}
	}
}
