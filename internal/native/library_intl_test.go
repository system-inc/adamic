package native

import (
	"regexp"
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

// A collator's shape names no field compare (#zjndrx5 head 1-4, Runtime's ruling at 16:11Z): Node's
// collator has no own property, compare being a getter on Intl.Collator.prototype, so native's holds it
// in the private #compare slot and a view naming compare finds no field to read or write by that name.
// x.compare(...) reaches #compare through adamic_object_callee's arm for this shape (object.c). Mutant:
// compare put back as a plain field named compare, which n1 to n3 then write natively, printing wrote and
// 0 where Node throws and prints -1.
func TestIntlCollatorShapeNamesNoCompare(t *testing.T) {
	t.Parallel()
	source, err := runtime.ReadFile("runtime/library_intl.c")
	if err != nil {
		t.Fatal(err)
	}
	names := regexp.MustCompile(`static const char \*const collator_names\[\] = \{([^}]*)\};`).FindSubmatch(source)
	if names == nil {
		t.Fatal("library_intl.c declares no collator_names")
	}
	if got, want := string(names[1]), `"#compare", "#locale"`; got != want {
		t.Errorf("collator_names = {%s}, want {%s}", got, want)
	}
}
