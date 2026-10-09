package prettier

import "testing"

// TestParserForMatchesPrettierInference pins the parser table against what Prettier's own
// getFileInfo reports for each name. Every expectation here was read out of Prettier rather than
// written from memory, because a hand-written mapping is exactly where a silent divergence enters:
// the probe that preceded this engine keyed package.json off its extension and rewrote npm's array
// formatting on every run, which was the single disagreement in a 1,556-file corpus.
func TestParserForMatchesPrettierInference(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"a.ts": "typescript", "a.tsx": "typescript",
		"a.js": "babel", "a.jsx": "babel", "a.mjs": "babel", "a.cjs": "babel",
		"a.json": "json", "package.json": "json-stringify",
		"a.css": "css", "a.scss": "scss", "a.less": "less",
		"a.md": "markdown", "a.graphql": "graphql", "a.gql": "graphql",
		"a.yaml": "yaml", "a.yml": "yaml",
	}
	for fileName, want := range expected {
		got, handled := parserFor(fileName)
		if !handled {
			t.Errorf("%s: not handled, want parser %q", fileName, want)
			continue
		}
		if got != want {
			t.Errorf("%s: parser %q, want %q", fileName, got, want)
		}
	}
}

// TestPackageJsonIsNotKeyedByExtension is the regression that the corpus caught. A nested
// package.json must still get json-stringify, and a differently-named .json must not.
func TestPackageJsonIsNotKeyedByExtension(t *testing.T) {
	t.Parallel()

	if parser, _ := parserFor("/a/b/package.json"); parser != "json-stringify" {
		t.Errorf("nested package.json got %q, want json-stringify", parser)
	}
	if parser, _ := parserFor("/a/b/tsconfig.json"); parser != "json" {
		t.Errorf("tsconfig.json got %q, want json", parser)
	}
}

// TestUnhandledTypesAreRefused is the control that proves Handles can say no. A table that answered
// yes to everything would be indistinguishable from one that works, and an engine that silently
// skips an unhandled type reports a tree as formatted when it never looked at it.
func TestUnhandledTypesAreRefused(t *testing.T) {
	t.Parallel()

	for _, fileName := range []string{"a.rb", "a.go", "a.py", "a.txt", "Makefile", "a"} {
		if parser, handled := parserFor(fileName); handled {
			t.Errorf("%s: handled with parser %q, want refused", fileName, parser)
		}
	}
}
