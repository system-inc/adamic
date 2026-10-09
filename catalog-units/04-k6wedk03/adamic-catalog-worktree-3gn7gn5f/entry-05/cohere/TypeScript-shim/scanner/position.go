package scanner

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

// GetLineAndCharacterOfPosition turns a byte offset into a zero-based line and column.
//
// This is hand-written rather than generated because upstream no longer has a function of this
// shape. `microsoft/TypeScript` split the old one-call helper into two primitives — a line-start
// table for the file, and a binary search over it — so that a caller converting many positions in
// one file builds the table once instead of per lookup.
//
// We keep the old signature because our callers convert one position at a time, at the moment a
// finding is printed, and the table for a file is cheap next to the work that produced the finding.
// A caller in a loop over many positions in the same file should reach for the two primitives
// directly rather than paying for the table on every iteration.
//
// The returned column is a UTF-8 byte offset from the line start, which is what upstream's
// PositionToLineAndByteOffset produces and what the old function returned for ASCII source. It is
// not a UTF-16 code unit count, so a line containing characters outside the BMP will report a column
// that differs from what an editor shows.
func GetLineAndCharacterOfPosition(sourceFile ast.SourceFileLike, position int) (line int, character int) {
	return core.PositionToLineAndByteOffset(position, GetECMALineStarts(sourceFile))
}
