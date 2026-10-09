package yaml

import (
	"reflect"
	"testing"

	"github.com/system-inc/cohere/internal/format/doc"
)

// TestPreservedFlowScalarMatchesGeneralPath checks that printPreservedFlowScalarContent builds the doc the
// general path builds under proseWrap preserve, node for node (#v6ksqg3): getFlowScalarLineContents, then
// fillWords per line, then doc.Join with hardlines. The fixtures cover what the trims and the split decide
// on: a no-break space, a tab and a run of spaces at a line's ends and inside it, empty and blank lines,
// a lone line, and a text that is only line feeds.
func TestPreservedFlowScalarMatchesGeneralPath(t *testing.T) {
	t.Parallel()
	contents := []string{
		"", "a", "a b", "a  b", " a ", "\ta\t", " a ", "a b",
		"\n", "\n\n", "a\n", "\na", "a\nb", "a \n b", " a\n\nb ", "a\n \n\tb",
		"first  \n  middle  \n\t last", "a\n \n b\n", "x y\nz", "  \n  ",
	}
	settings := &settings{proseWrap: "preserve"}
	for _, nodeType := range []string{"plain", "quoteDouble", "quoteSingle"} {
		for _, content := range contents {
			lineContents := getFlowScalarLineContents(nodeType, content, settings)
			lines := make([]doc.Doc, len(lineContents))
			for index, words := range lineContents {
				lines[index] = fillWords(words)
			}
			var expected doc.Doc = doc.Join(doc.Hardline, lines)
			actual := printPreservedFlowScalarContent(content)
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("%s %q: built %#v, the general path %#v", nodeType, content, actual, expected)
			}
		}
	}
}
