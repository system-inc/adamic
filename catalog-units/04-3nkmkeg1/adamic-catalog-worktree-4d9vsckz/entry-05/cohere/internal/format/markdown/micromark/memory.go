package micromark

import (
	"slices"

	"github.com/system-inc/cohere/internal/format/arena"
)

// Memory is what a parse takes its tokens, attempts and tokenizers from, kept for the next parse once the
// events it returned are done (#93dpede, #vbjv3d6). A Memory belongs to one parse at a time.
//
// Attempts and tokenizers are the tokenizer's small objects: an attempt's three states and a tokenizer's
// effects are closures over the struct, so a fresh struct costs a closure per state as well. Here each
// slot binds its closures once, when it is first made, and every later parse that takes the slot reuses
// them. Within one parse no slot is handed out twice, exactly as fresh structs were.
type Memory struct {
	Tokens arena.Arena[Token]

	attempts       slots[attempt]
	tokenizers     slots[TokenizeContext]
	spaceRuns      slots[spaceRun]
	emailAutolinks slots[emailAutolink]
	blankLines     slots[blankLineRun]
	gfmTables      slots[gfmTableRun]

	// stackCopies and eventCopies hold the copies an attempt saves of the token stack and a resolver is
	// given of the events. A released copy holds nil tokens, so reading one stops the parse.
	stackCopies arena.Slab[*Token]
	eventCopies arena.Slab[Event]
}

// slots hands out values of T, each made once by its fresh function and reused by later parses.
type slots[T any] struct {
	values []*T
	used   int
}

func (slots *slots[T]) take(fresh func() *T) *T {
	if slots.used == len(slots.values) {
		slots.values = append(slots.values, fresh())
	}
	value := slots.values[slots.used]
	slots.used++
	return value
}

// release hands every value taken back, each reset for its next parse.
func (slots *slots[T]) release(reset func(*T)) {
	for _, value := range slots.values[:slots.used] {
		reset(value)
	}
	slots.used = 0
}

// NewMemory is an empty Memory whose released tokens hold a poison no parse produces.
func NewMemory() *Memory {
	released := Point{Line: 1 << 30, Column: 1 << 30, Offset: 1 << 30}
	return &Memory{Tokens: arena.Arena[Token]{Poison: Token{Type: "released", Start: released, End: released}}}
}

// Reset releases everything the last parse took. Nothing may read its events, tokens or tokenizers after.
//
// A released attempt or tokenizer is zeroed but for its bound closures, so one read after its parse
// dereferences nil and stops the format rather than reading another file's state.
func (memory *Memory) Reset() {
	memory.Tokens.Reset()
	memory.attempts.release((*attempt).release)
	memory.tokenizers.release((*TokenizeContext).release)
	memory.spaceRuns.release((*spaceRun).release)
	memory.emailAutolinks.release((*emailAutolink).release)
	memory.blankLines.release((*blankLineRun).release)
	memory.gfmTables.release((*gfmTableRun).release)
	memory.stackCopies.Reset()
	memory.eventCopies.Reset()
}

// copyStack is a copy of the token stack, for an attempt to restore.
func (memory *Memory) copyStack(stack []*Token) []*Token {
	if memory == nil {
		return slices.Clone(stack)
	}
	return append(memory.stackCopies.Make(len(stack)), stack...)
}

// copyEvents is a copy of events, for a resolver to rewrite.
func (memory *Memory) copyEvents(events []Event) []Event {
	if memory == nil {
		return slices.Clone(events)
	}
	return append(memory.eventCopies.Make(len(events)), events...)
}

// tokens is where a parse's tokens come from, nil to allocate each.
func (memory *Memory) tokens() *arena.Arena[Token] {
	if memory == nil {
		return nil
	}
	return &memory.Tokens
}

// attempt is an attempt for one call of attempt, check or interrupt, its states bound.
func (memory *Memory) attempt() *attempt {
	if memory == nil {
		return newAttempt()
	}
	return memory.attempts.take(newAttempt)
}

// tokenizer is a tokenizer with its effects bound and every other field zero.
func (memory *Memory) tokenizer() *TokenizeContext {
	if memory == nil {
		return newTokenizeContext()
	}
	return memory.tokenizers.take(newTokenizeContext)
}

// spaceRun is a factorySpace with its states bound.
func (memory *Memory) spaceRun() *spaceRun {
	if memory == nil {
		return newSpaceRun()
	}
	return memory.spaceRuns.take(newSpaceRun)
}

// emailAutolink is an email autolink literal with its states bound.
func (memory *Memory) emailAutolink() *emailAutolink {
	if memory == nil {
		return newEmailAutolink()
	}
	return memory.emailAutolinks.take(newEmailAutolink)
}

// blankLine is a blank line with its states bound.
func (memory *Memory) blankLine() *blankLineRun {
	if memory == nil {
		return newBlankLineRun()
	}
	return memory.blankLines.take(newBlankLineRun)
}

// gfmTable is a GFM table with its states bound.
func (memory *Memory) gfmTable() *gfmTableRun {
	if memory == nil {
		return newGfmTableRun()
	}
	return memory.gfmTables.take(newGfmTableRun)
}
