package reference

import (
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// IdentifiersNamed returns every identifier in the rule's file whose text is name, in the order a
// ForEachChild walk from the source file reaches them, which is source order.
//
// A rule asking "where else is this binding written, or read" has to look at the whole file: a write
// can sit before the declaration, after it, or inside a callback several scopes down. Walking the
// file once per binding visits every identifier for every binding, so a file with a hundred `let`s
// was walked two or three hundred times by prefer-const alone (#hekjpw3). This walks it once, the
// first time any rule asks, and indexes the identifiers by text, so the per-binding question reads
// the few nodes that could answer it. The order is the walk's own, so a caller that took the first
// match of a whole-file walk takes the same node from this list.
//
// Text equality is a pre-filter, not resolution: two bindings of one name in different scopes share
// a list, and the caller still asks the checker which of them a node resolves to.
//
// The returned slice is the index's own storage, capped at its length, so appending to it copies
// rather than writing over the next name's identifiers.
func IdentifiersNamed(ctx rule.Context, name string) []*ast.Node {
	if ctx.SourceFile == nil {
		return nil
	}
	index := rule.Cached(ctx.FileCache, "reference.identifiersByName", func() *identifierIndex {
		return buildIdentifierIndex(ctx.SourceFile)
	})
	position, found := index.positions[name]
	if !found {
		return nil
	}
	span := index.spans[position]
	return index.ordered[span.start:span.end:span.end]
}

// identifierIndex is a file's identifiers grouped by text: one slice holding every identifier, each
// name's run contiguous and in walk order, and the span of each name's run.
//
// One slice rather than a slice per name, because a file has hundreds of distinct names and each
// slice was its own allocation, grown by doubling. With four rules asking (#fcac58b) the index is
// built on nearly every file, and the per-name shape cost 21 MB and 220K objects on a cold ahra run
// over the files prefer-const alone had asked about. This shape is a few objects per file.
type identifierIndex struct {
	ordered []*ast.Node
	// positions is each name's place in spans, so a name is hashed once while the index is built.
	positions map[string]int
	spans     []identifierSpan
}

// identifierSpan is one name's run in identifierIndex.ordered, start inclusive and end exclusive.
type identifierSpan struct {
	start int
	end   int
}

// buildIdentifierIndex walks the file once, counting each name, then places each identifier in its
// name's run.
//
// Each identifier is hashed once, on the walk, which also records the name's place beside the
// identifier, so the placing pass reads that place rather than hashing again. While counting, a span's
// end holds its count; each span is then given its start, with end set to it as a cursor, and placing
// an identifier advances its name's end, which finishes at the run's end.
//
// The walk's two buffers are scratch, dropped once the identifiers are placed, so they come from a pool
// rather than being made per file. Only ordered, spans and positions outlive the build.
func buildIdentifierIndex(sourceFile *ast.SourceFile) *identifierIndex {
	scratch := identifierScratchPool.Get().(*identifierScratch)
	defer scratch.release()
	walked := scratch.walked[:0]
	walkedPositions := scratch.positions[:0]
	positions := map[string]int{}
	var spans []identifierSpan
	var visit func(*ast.Node)
	visit = func(current *ast.Node) {
		if current.Kind == ast.KindIdentifier {
			position, found := positions[current.Text()]
			if !found {
				position = len(spans)
				positions[current.Text()] = position
				spans = append(spans, identifierSpan{})
			}
			spans[position].end++
			walked = append(walked, current)
			walkedPositions = append(walkedPositions, int32(position))
		}
		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile.AsNode())

	offset := 0
	for position := range spans {
		count := spans[position].end
		spans[position] = identifierSpan{start: offset, end: offset}
		offset += count
	}
	ordered := make([]*ast.Node, len(walked))
	for index, identifier := range walked {
		span := &spans[walkedPositions[index]]
		ordered[span.end] = identifier
		span.end++
	}
	// Handed back grown, so the next file starts from this one's capacity.
	scratch.walked = walked
	scratch.positions = walkedPositions
	return &identifierIndex{ordered: ordered, positions: positions, spans: spans}
}

// identifierScratch is buildIdentifierIndex's walk buffers: every identifier in walk order, and each one's
// name's place in spans.
type identifierScratch struct {
	walked    []*ast.Node
	positions []int32
}

// identifierScratchPool lends the walk buffers, one per build in flight.
var identifierScratchPool = sync.Pool{New: func() any { return &identifierScratch{} }}

// release clears the node pointers, so a pooled buffer never keeps a file's tree alive, and returns the
// buffers to the pool.
func (scratch *identifierScratch) release() {
	clear(scratch.walked)
	identifierScratchPool.Put(scratch)
}
