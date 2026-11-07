package native

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A collator's two run-time NotYets (#zjndrx5) are said natively in internal/ir's words, which the
// JavaScript backend writes from ir itself (library_intl.go): library_intl.c holds each whole, so
// changing one side's words fails here rather than only in the oracle's stop fixtures. Mutant: a word
// of library_intl.c's compare message changed.
func TestIntlCollatorStopsAreIRsWords(t *testing.T) {
	t.Parallel()
	source, err := runtime.ReadFile("runtime/library_intl.c")
	if err != nil {
		t.Fatal(err)
	}
	for _, message := range []string{ir.IntlCollatorCompareStop, ir.IntlCollatorLocaleStop} {
		if !strings.Contains(string(source), "static const char message[] = "+strconv.Quote(message)+";") {
			t.Errorf("library_intl.c doesn't stop in ir's words %q", message)
		}
	}
}
