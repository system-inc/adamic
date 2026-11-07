package yaml

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestLexerGaps(test *testing.T) {
	test.Parallel()
	closedGap(test, "gaps/multiplePush.ts", "2\n")
}

func TestStructuralPositionMatchesNode(test *testing.T) {
	test.Parallel()
	closedGap(test, "gaps/structuralPosition.ts", "1\n")
}

func TestClosedStringPresenceGap(test *testing.T) {
	test.Parallel()
	closedGap(test, "gaps/stringPresence.ts", "false\n")
}

// TestClosedLexerGaps keeps the lexer gaps compiler/area-next closed: each proving program now lowers, and
// must print what Node prints from native under the sanitizers and from emitted JavaScript. The port's
// workarounds for them still stand; retiring each one is its own change against the corpus.
func TestClosedLexerGaps(test *testing.T) {
	test.Parallel()
	for _, gap := range []struct{ file, output string }{
		{"assignmentValue.ts", "1\n"},
		{"dynamicCase.ts", "1\n"},
		{"negativeCase.ts", "1\n"},
		{"stringFallback.ts", " \n"},
		{"valueConjunction.ts", "true\n"},
		{"prefixIncrement.ts", "1\n"},
		{"emptyAlternative.ts", "1\n"},
	} {
		// Not parallel: one gap at a time keeps this test to one child at a time, its share.
		test.Run(gap.file, func(test *testing.T) {
			closedGap(test, filepath.Join("gaps", gap.file), gap.output)
		})
	}
}

// closedGap compares file's program, read from the closed gaps' build (port_products_test.go), with what Node prints.
func closedGap(test *testing.T, file, output string) {
	test.Helper()
	binary := yamlClosedGapsBuild.binary(test, file)
	emitted := yamlClosedGapsBuild.emitted(test, file)
	path, err := filepath.Abs(file)
	if err != nil {
		test.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	expected := run(test, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(expected) != output {
		test.Fatalf("Node printed %q, want %q", expected, output)
	}
	actual := run(test, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	backend := run(test, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted)
	for _, side := range []struct {
		name   string
		output []byte
	}{{"native ASan/UBSan/LSan", actual}, {"JavaScript backend", backend}} {
		if !bytes.Equal(side.output, expected) {
			test.Fatalf("%s: %q, Node %q", side.name, side.output, expected)
		}
	}
}

func TestClosedValuePresenceGap(test *testing.T) {
	test.Parallel()
	closedGap(test, "gaps/valuePresence.ts", "false\n")
}
