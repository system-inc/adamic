// Package arena hands out values from chunks of memory that are kept and reused, for the trees a format
// builds and drops one file at a time (#93dpede).
//
// Formatting allocates a parse's worth of tokens and nodes for every file, and with the collector off (the
// policy cohere runs under) none of it is reclaimed: every file's tree is new memory. A tree whose values
// all die when its file is done can take them from an Arena instead, and Reset hands the same memory to
// the next file. An Arena belongs to one goroutine at a time; formatters keep theirs in a sync.Pool.
//
// Under the cohere_poison build tag, Reset fills every value it releases with the arena's Poison, so a
// value read after its file was done changes the output, or stops it, where the corpus tests see it.
package arena

// Arena hands out values of T. The zero Arena is ready to use, and a nil *Arena allocates each value.
type Arena[T any] struct {
	chunks [][]T
	chunk  int
	used   int

	// Poison is what a released value holds under cohere_poison: something no parse produces.
	Poison T
}

// chunkBytes is about how much memory one chunk takes, whatever the size of T.
const chunkBytes = 96 << 10

// New returns a value holding value, from the arena, or allocated when the arena is nil.
func (arena *Arena[T]) New(value T) *T {
	if arena == nil {
		// A new value rather than &value: taking value's address would move every caller's value to the
		// heap, arena or not.
		allocated := new(T)
		*allocated = value
		return allocated
	}
	if arena.chunk == len(arena.chunks) {
		arena.chunks = append(arena.chunks, make([]T, chunkLength[T]()))
	}
	current := arena.chunks[arena.chunk]
	slot := &current[arena.used]
	*slot = value
	arena.used++
	if arena.used == len(current) {
		arena.chunk++
		arena.used = 0
	}
	return slot
}

// Len is how many values the arena has handed out since it was made or Reset, zero for a nil arena. A tree
// built from one arena has as many nodes as it, less any it built and dropped, which is the cheap count
// printing.Options.NodeCount asks for (#4bq8vyn).
func (arena *Arena[T]) Len() int {
	if arena == nil {
		return 0
	}
	return arena.chunk*chunkLength[T]() + arena.used
}

// Reset releases every value the arena has handed out, for the next user to reuse. Nothing may read a
// value handed out before.
func (arena *Arena[T]) Reset() {
	for index := 0; index <= arena.chunk && index < len(arena.chunks); index++ {
		end := len(arena.chunks[index])
		if index == arena.chunk {
			end = arena.used
		}
		Release(arena.chunks[index][:end], arena.Poison)
	}
	arena.chunk, arena.used = 0, 0
}

// Release hands back values a caller keeps for its next use, as Reset does an arena's: cleared, so the
// strings and slices they held are not kept alive, or under cohere_poison filled with poison, so a value
// read after its release reads something no parse produces (#v6ksqg3).
func Release[T any](values []T, poison T) {
	if poisonReleased {
		for index := range values {
			values[index] = poison
		}
		return
	}
	clear(values)
}

// chunkLength is how many values of T fit in chunkBytes, at least 16.
func chunkLength[T any]() int {
	var sample [1]T
	size := int(sizeOf(sample[:]))
	if size == 0 {
		return 1024
	}
	return max(16, chunkBytes/size)
}
