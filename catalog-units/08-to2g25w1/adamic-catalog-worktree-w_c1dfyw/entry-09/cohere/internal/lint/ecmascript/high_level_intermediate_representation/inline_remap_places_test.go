package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

// TestCopyNestedBodyLeavesTheNestedFunctionsPlacesUntouched is the sibling of
// `TestCopyNestedBodyLeavesTheNestedFunctionUntouched`, and it exists because that test's shape was
// not enough.
//
// That test asserts over block REFERENCES. It caught a shared terminal pointer, which is real. But
// the same aliasing existed one field over and it was invisible to a block-reference assertion:
// `Instruction.Value` is an interface holding a pointer, the copy assigned it directly, and
// `EachInstructionPlacePointer` then rewrote the operands of the NESTED function's instruction
// rather than the copy's.
//
// Measured before the fix, on the fixture below: ten of nineteen places in the nested body had been
// renamed to the parent's identifiers, and printing the nested function panicked with an index past
// the end of its own identifier table. Nothing else in the package saw it, because the parent -- the
// thing every caller looks at -- was correct.
//
// The assertion is therefore over places, and over the nested function rather than the parent. The
// nested function is the control: `CopyNestedBodyInto` promises to copy, and a copy that renames its
// source has not copied.
func TestCopyNestedBodyLeavesTheNestedFunctionsPlacesUntouched(t *testing.T) {
	t.Parallel()

	// The body must hold a value with a SLICE of places -- a method call's arguments here. A
	// one-level struct copy duplicates a slice header while both headers point at one backing
	// array, so a body of only single-place values would be copied correctly by a shallow copier
	// and this test would pass against the defect it exists to catch.
	const source = `
		function Component(properties: {items: Array<number>; flag: boolean}) {
			const compute = () => {
				const out = [];
				out.push(properties.items, properties.flag);
				if (properties.flag) {
					out.push(1);
				}
				return out;
			};
			return compute();
		}
	`
	parent, nested, captures := loweredParentAndNested(t, source)
	if parent == nil || nested == nil {
		t.Fatal("the fixture lowered no nested function")
	}

	type key struct {
		instruction InstructionId
		index       int
	}
	before := map[key]static_single_assignment.IdentifierId{}
	for id, instruction := range nested.Instructions {
		if instruction == nil {
			continue
		}
		index := 0
		EachInstructionPlacePointer(instruction, func(place *Place, role PlaceRole) {
			before[key{InstructionId(id), index}] = place.Identifier
			index++
		})
	}
	if len(before) == 0 {
		t.Fatal("the nested body holds no places, so this test asserts nothing")
	}

	if _, ok := CopyNestedBodyInto(parent, nested, captures); !ok {
		t.Fatal("the copy declined a fixture it should accept")
	}

	for id, instruction := range nested.Instructions {
		if instruction == nil {
			continue
		}
		index := 0
		EachInstructionPlacePointer(instruction, func(place *Place, role PlaceRole) {
			position := key{InstructionId(id), index}
			index++
			was, known := before[position]
			if known && was != place.Identifier {
				t.Errorf("nested instruction %d place %d named identifier %d before the copy and "+
					"names %d after it; the copy shares an instruction value with the function it "+
					"was copying and renamed the source instead of the copy",
					position.instruction, position.index, was, place.Identifier)
			}
		})
	}

	// Printing is the symptom the shared pointer produced, and it is worth asserting directly: a
	// renamed place whose id lands past the nested function's own table panics rather than reading
	// wrong, so this is the difference between a wrong answer and a crashed process.
	for _, instruction := range nested.Instructions {
		if instruction == nil {
			continue
		}
		EachInstructionPlacePointer(instruction, func(place *Place, role PlaceRole) {
			if int(place.Identifier) >= len(nested.Identifiers) {
				t.Errorf("the nested body names identifier %d, past its own table of %d; the copy "+
					"wrote the parent's ids into the function it was copying",
					place.Identifier, len(nested.Identifiers))
			}
		})
	}
}
