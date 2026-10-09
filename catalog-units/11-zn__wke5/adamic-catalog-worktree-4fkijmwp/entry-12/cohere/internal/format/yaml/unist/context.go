package unist

// Ported from yaml-unist-parser 3.2.0, dist/transforms/context.mjs, dist/transforms/comment.mjs and
// dist/cst.mjs, with the points and positions of dist/factories/position.mjs.

import (
	"fmt"
	"sync"
	"unicode/utf16"

	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/yaml/compose"
	"github.com/system-inc/cohere/internal/format/yaml/cst"
)

// point is upstream's point object, { line, column, offset }, with the offset in UTF-16 units as
// upstream counts it. Upstream never changes a point once made, only which point a position holds, and
// it compares points by identity (isExplicitMappingKey's `start !== end`), so a point is a pointer to a
// value nothing writes.
type point struct {
	line   int
	column int
	offset int
}

// position is upstream's position object, { start, end }. Upstream shares position objects between
// nodes (transformItemValue gives a mapping its one item's position object) and reassigns their start
// and end (updatePositions, a directive taking its trailing comment's end), so a node's position is a
// pointer, kept in context.positions until Parse converts it into Node.Position.
type position struct {
	start *point
	end   *point
}

// context is upstream's Context: the text, the comments in the order they were transformed, and the line
// counter. text is context.text as UTF-16 units, which upstream indexes (text[i], text.length).
type context struct {
	text        []uint16
	comments    []*Node
	lineCounter *cst.LineCounter

	// positions is every node's position object, upstream's node.position.
	positions map[*Node]*position

	// nodes is where the tree's nodes come from, nil to allocate each (#93dpede).
	nodes *arena.Arena[Node]
	// memory is where the parse's points and positions come from (#vbjv3d6).
	memory *parseMemory
}

func newContext(text []uint16, lineCounter *cst.LineCounter, nodes *arena.Arena[Node], memory *parseMemory) *context {
	if memory.positionsByNode == nil {
		// Sized for about a node per eight units of text, so the map is not rebuilt as it grows. Later
		// parses reuse it cleared, with the room it grew to.
		memory.positionsByNode = make(map[*Node]*position, len(text)/8)
	}
	return &context{text: text, comments: []*Node{}, lineCounter: lineCounter, positions: memory.positionsByNode, nodes: nodes, memory: memory}
}

// parseMemory is what one Parse uses only while it runs, kept for the next (#vbjv3d6): its point and
// position objects, which Parse copies into each node's Position before it returns, and its tables: the
// text as UTF-16 units, each unit's byte offset, and the nodes' positions (#v6ksqg3).
type parseMemory struct {
	points    arena.Arena[point]
	positions arena.Arena[position]

	units           []uint16
	offsets         unitOffsets
	positionsByNode map[*Node]*position

	// tokens and composeNodes are the CST's tokens and the composer's scalars, pairs and collections.
	// Both die with the parse: a unist node keeps neither, only strings and positions copied out of them
	// (#v6ksqg3).
	tokens       arena.Arena[cst.Token]
	composeNodes arena.Arena[compose.Node]
}

// parseMemories hold the memory of parses that have finished, one per caller at a time.
var parseMemories = sync.Pool{New: func() any {
	released := point{line: 1 << 30, column: 1 << 30, offset: 1 << 30}
	// A released position holds no points, so reading one through it stops the parse. A released token or
	// compose node reads a type and class no parse makes.
	return &parseMemory{
		points:       arena.Arena[point]{Poison: released},
		tokens:       arena.Arena[cst.Token]{Poison: cst.Token{Type: "released"}},
		composeNodes: arena.Arena[compose.Node]{Poison: compose.Node{Kind: "released", Class: "released"}},
	}
}}

func (memory *parseMemory) reset() {
	memory.points.Reset()
	memory.positions.Reset()
	memory.tokens.Reset()
	memory.composeNodes.Reset()
	// Under cohere_poison a released table reads a noncharacter and an offset past any text, so a parse that
	// read one after its release prints wrong or stops. A cleared map answers nil, which stops it too.
	arena.Release(memory.units, 0xFFFF)
	arena.Release(memory.offsets, 1<<40)
	clear(memory.positionsByNode)
}

// position is node.position.
func (context *context) position(node *Node) *position {
	return context.positions[node]
}

// transformOffset is upstream's transformOffset: a new point.
func (context *context) transformOffset(offset int) *point {
	line, col := context.lineCounter.LinePos(offset)
	return context.memory.points.New(point{line: line, column: col, offset: offset})
}

// transformRange is upstream's transformRange: a new position over two new points.
func (context *context) transformRange(start int, end int) *position {
	return context.createPosition(context.transformOffset(start), context.transformOffset(end))
}

// transformComment is upstream's Context.transformComment and transformComment (comment.mjs).
func (context *context) transformComment(comment *cst.Token) *Node {
	node := context.createComment(context.transformRange(comment.Offset, comment.Offset+len(comment.Source)),
		unitsToString(comment.Source[1:]))
	context.comments = append(context.comments, node)
	return node
}

// tokens is cst.mjs's generator over token lists, skipping absent lists and space and newline tokens.
// Every caller consumes it whole before acting on what it yielded, or acts on each token without
// touching the lists, so it is a slice here.
func tokens(lists ...[]*cst.Token) []*cst.Token {
	var result []*cst.Token
	for _, list := range lists {
		for _, token := range list {
			if isSpace(token) {
				continue
			}
			result = append(result, token)
		}
	}
	return result
}

// isSpace is upstream's isSpace.
func isSpace(token *cst.Token) bool {
	return token.Type == "space" || token.Type == "newline"
}

// maybeContentPropertyToken is upstream's maybeContentPropertyToken.
func maybeContentPropertyToken(token *cst.Token) bool {
	return token.Type == "comment" || token.Type == "tag" || token.Type == "anchor"
}

// extractComments is utils/extract-comments.mjs: transforms the comments and returns the rest.
func (context *context) extractComments(list []*cst.Token) []*cst.Token {
	restNodes := []*cst.Token{}
	for _, token := range tokens(list) {
		if token.Type == "comment" {
			context.transformComment(token)
		} else {
			restNodes = append(restNodes, token)
		}
	}
	return restNodes
}

// unitsToString is a slice of the source as a Go string. The source came from valid runes, and every
// token boundary falls between characters, so no slice holds half a surrogate pair.
func unitsToString(units []uint16) string {
	return string(utf16.Decode(units))
}

// ThrownError is an Error yaml-unist-parser throws that is not a YAMLSyntaxError: one of its own
// "Unexpected ..." checks (Name "Error"), or a TypeError the JavaScript engine raises where upstream
// reads a property of undefined or null (on a !!pairs or !!omap sequence, say), whose message is V8's.
type ThrownError struct {
	Name    string
	Message string
}

func (thrown *ThrownError) Error() string { return thrown.Name + ": " + thrown.Message }

// throwError is `throw new Error(message)`.
func throwError(format string, arguments ...any) {
	panic(&ThrownError{Name: "Error", Message: fmt.Sprintf(format, arguments...)})
}

// throwTypeError is the TypeError V8 raises reading property of undefined or null (missing names
// which).
func throwTypeError(missing string, property string) {
	panic(&ThrownError{Name: "TypeError", Message: fmt.Sprintf("Cannot read properties of %s (reading '%s')", missing, property)})
}

// SyntaxError is upstream's YAMLSyntaxError: the first error of the first document that has one, with
// the message and code eemeli/yaml gave it and its position in the text. Position is in Node.Position's
// units: byte offsets, 1-based lines, columns in UTF-16 units.
type SyntaxError struct {
	Message  string
	Code     string
	Position Position
}

func (syntaxError *SyntaxError) Error() string { return syntaxError.Message }

// newYAMLSyntaxError is `new YAMLSyntaxError(context, error)`.
func (context *context) newYAMLSyntaxError(yamlError *compose.YAMLError, offsets unitOffsets) *SyntaxError {
	return &SyntaxError{
		Message:  yamlError.Message,
		Code:     yamlError.Code,
		Position: offsets.convert(context.transformRange(yamlError.Pos[0], yamlError.Pos[1])),
	}
}
