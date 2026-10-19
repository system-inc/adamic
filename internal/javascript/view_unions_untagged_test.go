package javascript

import "testing"

func TestViewUntaggedFailureCategories(t *testing.T) {
	t.Parallel()
	for _, sample := range []struct{ name, source, stored, found string }{
		{"Map", "new Map()", "new Map()", "Map"},
		{"closure", "value => value", "new AdamicClosure(value => value)", "function"},
		{"null", "null", "null", "null"},
		{"undefined", "undefined", "undefined", "undefined"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			t.Parallel()
			// Node classifies the source value; the backend closure has a represented wrapper.
			runViewNode(t, "const value = "+sample.source+"; console.log(value === null ? 'null' : value instanceof Map ? 'Map' : typeof value);", sample.found+"\n", "", 0)
			const declared = "Left | Right"
			want := "adamic: panic: field read failed: view.value matches no member of " + declared + "; expected " + declared + ", found " + sample.found + "\n"
			source := viewTestRuntime + viewUntaggedPlainRuntime + "adamicUntaggedPlainSelect(" + sample.stored + ", [{Members: []}], 1, 'view.value', '" + declared + "');"
			runViewNode(t, source, "", want, 70)
		})
	}
}
