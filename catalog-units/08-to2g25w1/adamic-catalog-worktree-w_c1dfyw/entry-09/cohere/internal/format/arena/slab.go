package arena

// Slab hands out slices of T from chunks kept and reused, as Arena hands out single values, for the lists
// a tree holds: each slice is cut from a chunk at the length asked for and capped there, so appending past
// it moves the slice to new memory rather than into its neighbour's. The zero Slab is ready to use, and a
// nil *Slab allocates each slice.
type Slab[T any] struct {
	chunks [][]T
	chunk  int
	used   int

	// Poison is what a released element holds under cohere_poison: something no parse produces.
	Poison T
}

// Make returns an empty slice with room for capacity elements, from the slab, or allocated when the slab
// is nil or capacity is more than a chunk holds.
func (slab *Slab[T]) Make(capacity int) []T {
	if slab == nil || capacity > chunkLength[T]() {
		return make([]T, 0, capacity)
	}
	for {
		if slab.chunk == len(slab.chunks) {
			slab.chunks = append(slab.chunks, make([]T, chunkLength[T]()))
		}
		current := slab.chunks[slab.chunk]
		if slab.used+capacity <= len(current) {
			taken := current[slab.used : slab.used : slab.used+capacity]
			slab.used += capacity
			return taken
		}
		slab.chunk++
		slab.used = 0
	}
}

// Reset releases every slice the slab has handed out, for the next user to reuse. Nothing may read a slice
// handed out before.
func (slab *Slab[T]) Reset() {
	for index := 0; index <= slab.chunk && index < len(slab.chunks); index++ {
		end := len(slab.chunks[index])
		if index == slab.chunk {
			end = slab.used
		}
		released := slab.chunks[index][:end]
		if poisonReleased {
			for element := range released {
				released[element] = slab.Poison
			}
			continue
		}
		// Cleared, so the strings and slices a released element held are not kept alive by the slab.
		clear(released)
	}
	slab.chunk, slab.used = 0, 0
}
