package unicodeproperties

import (
	"slices"
	"testing"

	cohere "github.com/system-inc/cohere/unicodeproperties"
)

// TestCohereTablesAgree holds this package's code-point tables against cohere's,
// which #cz54235 makes the one home for them: every spelling Lookup accepts without
// the v flag is one cohere's Names lists and none other, each has the same ranges,
// and two spellings share a set here exactly when they share one there. It is
// temporary: it goes with the copy it compares.
func TestCohereTablesAgree(t *testing.T) {
	t.Parallel()
	if Version != cohere.Version {
		t.Fatalf("tables are Unicode %s, cohere's are %s", Version, cohere.Version)
	}
	ours := codePointExpressions()
	theirs := cohere.Names()
	for _, name := range ours {
		if _, ok := slices.BinarySearch(theirs, name); !ok {
			t.Errorf("Lookup accepts %q, cohere's Names does not list it", name)
		}
	}
	for _, name := range theirs {
		if _, ok := slices.BinarySearch(ours, name); !ok {
			t.Errorf("cohere's Names lists %q, Lookup does not accept it", name)
		}
	}
	for name := range stringProperties {
		if _, ok := cohere.Lookup(name); ok {
			t.Errorf("string property %s is a code-point property in cohere", name)
		}
	}
	// A set here and cohere's ranges for it are each other's only partner.
	partner := map[*Set]*cohere.Range{}
	partnerOf := map[*cohere.Range]*Set{}
	ranges := 0
	for _, name := range theirs {
		property, ok := Lookup(name, false)
		if !ok || property.Kind != KindCodePoints {
			continue
		}
		reference, _ := cohere.Lookup(name)
		if len(reference.Ranges) == 0 {
			t.Errorf("cohere's %s is empty", name)
			continue
		}
		if len(property.Set.Ranges) != len(reference.Ranges) {
			t.Errorf("%s: %d ranges here, %d in cohere", name, len(property.Set.Ranges), len(reference.Ranges))
			continue
		}
		for index, interval := range reference.Ranges {
			if property.Set.Ranges[index] != (Range{Start: uint32(interval.Lo), End: uint32(interval.Hi)}) {
				t.Errorf("%s: range %d is %#v here, %#v in cohere", name, index, property.Set.Ranges[index], interval)
				break
			}
		}
		first := &reference.Ranges[0]
		if known, seen := partner[property.Set]; seen && known != first {
			t.Errorf("%s shares a set here with a spelling cohere gives other ranges", name)
		}
		if known, seen := partnerOf[first]; seen && known != property.Set {
			t.Errorf("%s shares ranges in cohere with a spelling that has another set here", name)
		}
		if _, seen := partner[property.Set]; !seen {
			ranges += len(reference.Ranges)
		}
		partner[property.Set] = first
		partnerOf[first] = property.Set
	}
	t.Logf("%d spellings here, %d in cohere; %d sets, %d ranges, all equal", len(ours), len(theirs), len(partner), ranges)
}
