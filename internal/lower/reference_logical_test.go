package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestReferenceLogicalKeepsMutableViews(t *testing.T) {
	t.Parallel()
	for _, expression := range []string{"dogs || empty", "dogs && dogs"} {
		_, err := lowerSource(t, `interface Animal { name: string } interface Dog extends Animal { bark: string }
const empty: Dog[] = [];
function view(dogs: Dog[] | undefined): Animal[] | undefined { return `+expression+`; }
const animals = view([{name: 'Kirk', bark: 'yes'}]);
if (animals !== undefined) { animals.push({name: 'Ahra'}); }`)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.Fix, "adamic/invariant-mutable") {
			t.Fatalf("%s: want mutable view refusal, got %v", expression, err)
		}
	}
}

func TestReferenceLogicalPrimitivesStayOnExistingPaths(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function or(text: string | undefined): string { return text || 'fallback'; } console.log(or(''));`)
	var notYet *NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("want primitive truthiness left to existing paths, got %v", err)
	}
}

func TestReferenceLogicalBothAbsencesNeedTag(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function and(arr: number[] | undefined | null): number[] | undefined | null { return arr && arr; } console.log(and(null) === null ? 'null' : 'wrong');`)
	var notYet *NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("want untagged null plus undefined refused, got %v", err)
	}
}

func TestNullableReferenceLookupNeedsTag(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `const slots: (number[] | null)[] = []; console.log(slots.pop() === null ? 'null' : 'missing');`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "distinct absence tag") {
		t.Fatalf("want untagged nullable lookup refused, got %v", err)
	}
}
