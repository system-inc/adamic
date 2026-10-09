package text

import "testing"

// TestGraphemeCountAgreesWithIntlSegmenter scores GraphemeCount against Node's Intl.Segmenter on the
// shapes it handles and on random strings built from the code points it treats specially
// (grapheme_corpus_data_test.go).
func TestGraphemeCountAgreesWithIntlSegmenter(t *testing.T) {
	t.Parallel()
	if len(graphemeCorpus) < 400 {
		t.Fatalf("the corpus holds %d rows, fewer than it was generated with", len(graphemeCorpus))
	}
	for _, row := range graphemeCorpus {
		if got := GraphemeCount(row.text); got != row.count {
			t.Errorf("%s %q: GraphemeCount is %d, Intl.Segmenter counts %d", row.name, row.text, got, row.count)
		}
	}
}
