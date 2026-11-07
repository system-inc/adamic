package lower

import (
	"context"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// cycles.go's reachLinks reads a MapIterator or SetIterator by its type and follows only a collection
// iterator's links (libraryIteratorCaptures). A type can't tell a generator declared SetIterator from
// a Set's own iterator, so the origin test for that arm is that no slot it walks can hold a
// generator: the cycle finder walks the checker types of locals and what they reach, and in this
// slice no local holds a generator's call but the consumer's own, which has no checker type (Compiler
// and Adamic, 07:43 on #qk8rztp; the value roads that could put one in a slot stop at lowering, as
// TestGeneratorFormsThisSliceDoesntLowerAreNotYet pins). declared_iterators.a has no Map or Set, so
// any local typed MapIterator or SetIterator would be a generator's. The control, a Set's own
// iterator held in a local, shows the check can see one. From S2, when a slot can hold a generator,
// the arm must take the union of the frame's links and the iterator's (Lower, q6).
func TestNoSlotTheIteratorArmWalksHoldsAGenerator(test *testing.T) {
	test.Parallel()
	iteratorLocals := func(path string) int {
		test.Helper()
		loaded, err := load.Load([]string{path})
		if err != nil {
			test.Fatal(err)
		}
		entry := loaded.Files()[0]
		check, release := loaded.Checker(context.Background(), entry)
		defer release()
		lowered, err := lowerChecked(loaded, check, entry, nil, false, false)
		if err != nil {
			test.Fatal(err)
		}
		found := 0
		for _, proven := range lowered.localTypes {
			if proven != nil && lowered.isLibraryType(proven, "MapIterator", "SetIterator") {
				found++
			}
		}
		return found
	}
	if found := iteratorLocals("../oracle/testdata/generators/declared_iterators.a"); found != 0 {
		test.Errorf(
			"%d locals of declared_iterators.a are typed MapIterator or SetIterator: a generator's call reached a slot the cycle finder walks",
			found,
		)
	}
	control := writeAgreementSource(
		test,
		"main.a",
		"const values = new Set<number>([1, 2]);\nconst walked = values.values();\nfor (const value of walked) {\n    console.log(String(value));\n}\n",
	)
	if found := iteratorLocals(control); found == 0 {
		test.Error("the control's SetIterator local went unseen: the check can't fail")
	}
}
