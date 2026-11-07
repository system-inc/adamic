package javascript

import "testing"

func TestViewUntaggedFailureCategories(test *testing.T) {
	test.Parallel()
	for _, sample := range []struct{ name, source, stored, found string }{
		{"Map", "new Map()", "new Map()", "Map"},
		{"closure", "value => value", "new AdamicClosure(value => value)", "function"},
		{"null", "null", "null", "null"},
		{"undefined", "undefined", "undefined", "undefined"},
	} {
		test.Run(sample.name, func(test *testing.T) {
			test.Parallel()
			// Node classifies the source value; the backend closure has a represented wrapper.
			runViewNode(
				test,
				"const value = "+
					sample.source+
					"; console.log(value === null ? 'null' : value instanceof Map ? 'Map' : typeof value);",
				sample.found+"\n",
				"",
				0,
			)
			const declared = "Left | Right"
			want := "adamic: panic: field read failed: view.value matches no member of " +
				declared +
				"; expected " +
				declared +
				", found " +
				sample.found +
				"\n"
			source := viewTestRuntime +
				viewUntaggedPlainRuntime +
				"adamicUntaggedPlainSelect(" +
				sample.stored +
				", [{Members: []}], 1, 'view.value', '" +
				declared +
				"');"
			runViewNode(test, source, "", want, 70)
		})
	}
}
