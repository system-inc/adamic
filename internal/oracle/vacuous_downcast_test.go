package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// In .a a downcast is admitted where its tag tells the target from the source (#q97a1rr): the sound
// sibling of a vacuous cast, a discriminated union narrowed by ===, a cast whose test a source value
// can fail, and a cast whose first tag every source value passes but whose second a source value can
// fail, which .a tests instead. Each lowers and agrees with Node natively and through the JavaScript
// backend. Mutant: withinLiterals calling every tag vacuous refuses distinct_tag_cast and later_tag.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/vacuous_downcast/distinct_tag_union.a",
		"internal/oracle/testdata/vacuous_downcast/distinct_tag_cast.a",
		"internal/oracle/testdata/vacuous_downcast/later_tag.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
	// The .ts twin of vacuous_downcast_refused/language.a still compiles to a checked view: the kind
	// test passes, and the read of count, which this Root never had, stops at the view's check where
	// Node prints count NaN.
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/vacuous_downcast/language.ts", true, true})
}

// A downcast in .a whose tag test every source value passes is refused, naming the fix: the target
// declares no tag that tells it from the source, so the test proves nothing and the cast would be a
// type the compiler trusts. language is Language's case, View extends Root with no kind of its own;
// cross_enum_union's source kind is a union of two enums' members that are both 'Holder' at run time,
// all of which the target's tag allows. Each was never a View, and Node prints count NaN. Each
// program's whole diagnostic is pinned in its .refused file. Mutant: the vacuous test admitted again
// (interfaceCast checking the tag as it did before #q97a1rr) lowers both.
func TestVacuousDowncastsAreRefused(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/vacuous_downcast_refused"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".a") {
			continue
		}
		programs++
		path := directory + "/" + name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want a refusal, got %v", err)
			}
			expected, err := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 2 {
		t.Fatalf("%d programs in %s, want 2: a program added or removed needs this count changed with it", programs, directory)
	}
}
