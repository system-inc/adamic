package ir

import (
	"fmt"
	"testing"
)

func distinctSyntaxKindTags(tags map[string]Type) error {
	seen := map[Type]string{}
	for name, tag := range tags {
		if previous, duplicate := seen[tag]; duplicate {
			return fmt.Errorf("%s and %s share representation tag %d", previous, name, tag)
		}
		seen[tag] = name
	}
	return nil
}

// Scalar absence tags and dictionary/typed-array tags share runtime value slots.
// The merger must not assign two meanings to any one of those tags.
func TestSyntaxKindRepresentationTags(t *testing.T) {
	t.Parallel()
	tags := map[string]Type{
		"number": Number, "boolean": Boolean, "string": String, "object": Object,
		"array": Array, "map": Map, "closure": Closure, "union": Union, "weak": Weak,
		"null": 12, "undefined": 13, "record": Record,
		"Uint8Array": Uint8Array, "Int32Array": Int32Array, "Float64Array": Float64Array,
	}
	if err := distinctSyntaxKindTags(tags); err != nil {
		t.Fatal(err)
	}
	for _, tag := range []Type{Record, Uint8Array, Int32Array, Float64Array} {
		if !tag.IsReference() {
			t.Fatalf("heap representation %d is not counted", tag)
		}
	}
	// Valid Go input, rather than a duplicate switch case rejected by the compiler.
	tags["Uint8Array"] = Record
	if err := distinctSyntaxKindTags(tags); err == nil {
		t.Fatal("record/typed-array collision mutant escaped the tag invariant")
	} else {
		t.Logf("collision mutant caught: %v", err)
	}
}
