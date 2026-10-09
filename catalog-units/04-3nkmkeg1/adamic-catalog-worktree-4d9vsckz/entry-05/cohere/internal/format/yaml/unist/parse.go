package unist

// Ported from yaml-unist-parser 3.2.0, dist/parse.mjs and dist/yaml-syntax-error.mjs.

import (
	"fmt"
	"runtime"
	"slices"
	"sort"
	"unicode/utf16"

	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/yaml/compose"
	"github.com/system-inc/cohere/internal/format/yaml/cst"
)

// Parse is yaml-unist-parser's parse(text, { uniqueKeys: false }), the call Prettier's parser-yaml.js
// makes: eemeli/yaml's parser and composer, the transforms into the unist tree, attachComments (which
// defines every Parent) and updatePositions. Offsets are converted to bytes once, at the end.
//
// A document with an error fails the parse with a *SyntaxError, upstream's YAMLSyntaxError. Where
// upstream throws anything else (its own invariant checks, or a TypeError on a shape it does not
// expect), Parse returns a *ThrownError with the same name and message. A panic in the port itself is
// returned as a plain error.
//
// The root keeps Comments, every comment in the file, as upstream's root does; Prettier's parser deletes
// root.comments before printing, and the printer never reads it.
//
// The nodes come from nodes, which may be nil; the tree is valid until it is Reset.
func Parse(text string, nodes *arena.Arena[Node]) (root *Node, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if thrown, isThrown := recovered.(*ThrownError); isThrown {
				root, err = nil, thrown
				return
			}
			if runtimeError, isRuntimeError := recovered.(runtime.Error); isRuntimeError {
				root, err = nil, fmt.Errorf("unist: %w", runtimeError)
				return
			}
			root, err = nil, fmt.Errorf("unist: %v", recovered)
		}
	}()

	memory := parseMemories.Get().(*parseMemory)
	defer func() {
		memory.reset()
		parseMemories.Put(memory)
	}()

	// A JavaScript string: UTF-16 units. An invalid UTF-8 byte is one U+FFFD, as newUnitOffsets counts.
	// Both tables are the memory's, kept across parses: nothing past Parse reads either, since node values
	// are copied out as strings and positions are converted into the nodes before Parse returns.
	memory.units = appendUnits(memory.units[:0], text)
	units := memory.units
	memory.offsets = newUnitOffsets(text, memory.offsets[:0])
	offsets := memory.offsets

	lineCounter := cst.NewLineCounter()
	context := newContext(units, lineCounter, nodes, memory)
	parser := cst.NewParser(lineCounter.AddNewLine)
	parser.Tokens = &memory.tokens
	composeOptions := compose.UnistParserOptions()
	composeOptions.Nodes = &memory.composeNodes
	composer := compose.NewComposer(composeOptions)
	parsedDocuments := []*compose.Document{}
	cstTokens := slices.Collect(parser.Parse(units, false))
	for parsedDocument := range composer.Compose(cstTokens, true, len(units)) {
		if len(parsedDocument.Errors) > 0 {
			return nil, context.newYAMLSyntaxError(parsedDocument.Errors[0], offsets)
		}
		parsedDocuments = append(parsedDocuments, parsedDocument)
	}

	documents := context.transformDocuments(parsedDocuments, cstTokens)
	// `context.comments.sort((commentA, commentB) => commentA.position.start.offset -
	// commentB.position.end.offset)`. Comments never overlap, and each ends after it starts, so for two
	// different comments the comparator says the earlier one goes first either way round: it orders them
	// by start offset, which any sort reproduces.
	comments := context.comments
	sort.SliceStable(comments, func(i int, j int) bool {
		return context.position(comments[i]).start.offset < context.position(comments[j]).start.offset
	})
	root = context.createRoot(context.transformRange(0, len(units)), documents, comments)
	context.attachComments(root)
	context.updatePositions(root)

	// Every node was made by a factory, which registered its position object.
	for node, position := range context.positions {
		node.Position = offsets.convert(position)
	}
	return root, nil
}

// unitOffsets maps a UTF-16 unit index to its byte offset in the text, the conversion the printer's tree
// loader makes.
type unitOffsets []int

// newUnitOffsets appends text's offsets to offsets, which may be a table kept from an earlier parse.
func newUnitOffsets(text string, offsets unitOffsets) unitOffsets {
	offsets = slices.Grow(offsets, len(text)+1)
	for index, character := range text {
		offsets = append(offsets, index)
		if character > 0xFFFF {
			// The low surrogate of a pair sits at the pair's start too; no position upstream records
			// falls between the two units of one character.
			offsets = append(offsets, index)
		}
	}
	return append(offsets, len(text))
}

// appendUnits appends text as UTF-16 units to units: utf16.Encode([]rune(text)), into a table kept from an
// earlier parse. Ranging a string decodes an invalid byte to U+FFFD, as the conversion to runes does.
func appendUnits(units []uint16, text string) []uint16 {
	units = slices.Grow(units, len(text))
	for _, character := range text {
		units = utf16.AppendRune(units, character)
	}
	return units
}

// byteOffset is the unit's byte offset. Past the end of the text, each unit is one byte further.
func (offsets unitOffsets) byteOffset(unit int) int {
	last := len(offsets) - 1
	if unit > last {
		return offsets[last] + unit - last
	}
	if unit < 0 {
		return unit
	}
	return offsets[unit]
}

// convert is the position as Node.Position holds it.
func (offsets unitOffsets) convert(position *position) Position {
	return Position{
		Start: Point{Line: position.start.line, Column: position.start.column, Offset: offsets.byteOffset(position.start.offset)},
		End:   Point{Line: position.end.line, Column: position.end.column, Offset: offsets.byteOffset(position.end.offset)},
	}
}
