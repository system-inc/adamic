package arena

import "testing"

// TestASlabSliceKeepsToItsOwnElements: each slice is capped at the length asked for, so filling one, or
// appending past it, never writes into the slice cut after it; a nil slab allocates; and Reset hands the
// same memory to the next user, cleared.
func TestASlabSliceKeepsToItsOwnElements(t *testing.T) {
	t.Parallel()
	var slab Slab[string]
	first := slab.Make(2)
	second := slab.Make(2)
	if len(first) != 0 || cap(first) != 2 || cap(second) != 2 {
		t.Fatalf("Make(2) gave length %d and capacities %d and %d, want 0, 2 and 2", len(first), cap(first), cap(second))
	}
	first = append(first, "a", "b")
	second = append(second, "c")
	first = append(first, "past the cap")
	if second[0] != "c" || len(first) != 3 || first[1] != "b" {
		t.Fatalf("appending past one slice's capacity wrote into the next: first %v, second %v", first, second)
	}

	var none *Slab[string]
	if allocated := none.Make(3); len(allocated) != 0 || cap(allocated) != 3 {
		t.Fatalf("a nil slab's Make(3) gave length %d capacity %d", len(allocated), cap(allocated))
	}
	if huge := slab.Make(chunkLength[string]() + 1); cap(huge) != chunkLength[string]()+1 {
		t.Fatalf("a slice larger than a chunk got capacity %d", cap(huge))
	}

	reused := &slab.chunks[0][0]
	slab.Reset()
	again := slab.Make(1)
	again = append(again, "d")
	if &again[0] != reused {
		t.Fatal("Reset did not hand the slab's memory to the next user")
	}
	if !poisonReleased && slab.chunks[0][1] != "" {
		t.Fatalf("Reset left a released element holding %q", slab.chunks[0][1])
	}
}
