package ir

// SourcePosition is the declaration name token's position, with one-based line and column.
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

// BoundaryStringTable interns metadata literals once per program, independently
// of the emitted program's constants.
type BoundaryStringTable struct {
	Values  []string
	indexes map[string]int
}

func (table *BoundaryStringTable) Intern(text string) int {
	if table.indexes == nil {
		table.indexes = map[string]int{}
	}
	if index, ok := table.indexes[text]; ok {
		return index
	}
	index := len(table.Values)
	table.indexes[text] = index
	table.Values = append(table.Values, text)
	return index
}
