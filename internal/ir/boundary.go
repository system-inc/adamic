package ir

// SourcePosition is the declaration token's position, with one-based line and column.
type SourcePosition struct {
	File         string
	Line, Column int
}

// FunctionBoundary preserves proven data types before runtime representation erases
// them. Parameters align with Function.Parameters, including an implicit receiver.
// Nil means the type cannot be described as boundary data.
type FunctionBoundary struct {
	Parameters []*JSONDecodeSchema
	Return     *JSONDecodeSchema
}
