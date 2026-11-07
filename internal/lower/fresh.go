package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
)

// The cycle finder's relaxation for fresh writes (docs/memory.md). The finder refuses every mutable
// slot whose type can reach back to its holder; a slot stands undeclared when every write into it in
// the whole program is proven not to close a cycle (fresh.ProveWrites). Lowering records, for each
// write it makes, the type of what it writes into, so a write the proof can't vouch for keeps every
// slot it could be writing refused, through any view of the holder, as the finder's related matches
// views.

// writeSite is one write lowering made: the type of what it writes into, as the instantiation being
// lowered sees it, and where.
type writeSite struct {
	holder *checker.Type
	node   *ast.Node
}

// writeSite records a write into receiver's slots, and is the IR node's Site.
func (lowering *lowering) writeSite(receiver *ast.Node) int {
	holder := lowering.concrete(lowering.checker.GetTypeAtLocation(receiver))
	if ast.SkipParentheses(receiver).Kind == ast.KindThisKeyword && lowering.classType != nil {
		// this is the polymorphic this type, a type parameter; the instantiation being lowered is what
		// it stands for here.
		holder = lowering.classType
	}
	lowering.writeSites = append(lowering.writeSites, writeSite{holder: holder, node: receiver})
	return len(lowering.writeSites)
}

// unproven is the first write in the program that may write a slot of kind in holder (a field by
// its name) and isn't proven not to close a cycle, or nil when every such write is.
func (finder *cycleFinder) unproven(kind fresh.WriteKind, holder *checker.Type, name string) *fresh.Write {
	if finder.writes == nil {
		finder.writes = fresh.ProveWrites(finder.l.result)
	}
	for index := range finder.writes {
		write := &finder.writes[index]
		if write.Proven {
			continue
		}
		if write.Kind == fresh.WriteUnknown {
			return write
		}
		if write.Kind != kind || (kind == fresh.WriteField && !finder.l.cycleFieldMatches(holder, name, write.Name)) {
			continue
		}
		if write.Site == 0 {
			// A write lowering didn't record: it may be into anything of its kind.
			return write
		}
		written := finder.l.writeSites[write.Site-1].holder
		if written == nil || finder.template(written) || finder.related(holder, written) {
			return write
		}
	}
	return nil
}

// writtenAt names an unproven write: where it is, or, for one lowering didn't record, what it is.
func (finder *cycleFinder) writtenAt(write *fresh.Write) string {
	if write.Site == 0 {
		return "a write in " + finder.l.functionName(write.Function)
	}
	return "the write at " + finder.l.program.Where(finder.l.writeSites[write.Site-1].node)
}

// functionName is an IR function's name, or the top level's.
func (lowering *lowering) functionName(function int) string {
	if function < 0 {
		return "the top level"
	}
	return lowering.result.Functions[function].Name
}
