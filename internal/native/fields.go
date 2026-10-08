package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// uniformFieldSlot skips shape lookup when every layout containing name puts it in the same slot.
// The checker has already proved the field exists; unlike a class tag, that proof also holds for
// structural literals and generic instantiations. Conflicting layouts keep the existing checks.
func (e *emitter) uniformFieldSlot(object, name string) string {
	if !cName.MatchString(object) {
		return ""
	}
	if e.fieldOffsets == nil {
		e.fieldOffsets = uniformFieldOffsets(e.program)
	}
	if index, found := e.fieldOffsets[name]; found && index >= 0 {
		return fmt.Sprintf("(&%s->slots[%d])", object, index)
	}
	return ""
}

// ObjectLiteral and MapEntries produce generated layouts (shapeOf in emit.go).
// Copies, reuse and region allocation preserve those layouts. A spread preserves its source's
// layout too, except the undefined branch, whose emptyFields layout must be counted explicitly.
func uniformFieldOffsets(program *ir.Program) map[string]int {
	// Named regex groups create layouts outside ObjectLiteral. Until their names and
	// offsets participate in this proof, regex programs retain checked shape lookup.
	if len(program.Regexps) != 0 {
		return map[string]int{}
	}
	// Runtime-produced layouts are not ObjectLiterals: map entries (0, 1), Error (name, message, code),
	// and input/directory results (kind, text/names/message/type, size, symbolicLink). Including
	// them unconditionally is conservative when the program does not use that runtime API.
	// TestRuntimeFieldLayoutsAreIncluded checks this list against the embedded C declarations.
	offsets := map[string]int{
		"0": 0, "1": 1, "name": 0, "message": 1, "kind": 0, "text": 1,
		"names": 1, "type": 1, "size": 2, "symbolicLink": 3, "#compiler": 12,
	}
	for _, names := range [][]string{
		{"name", "message", "code"}, {"_fsFileTime"}, {"size", "mtimeMs", "mtime", "_fsFileMode", "atime"},
		// The Node directory host and the existing realPath result are created in C.
		{"kind", "path"}, {"name", "message", "code"}, {"name", "type"},
		{"__adamic_iterator_state"}, {"next"}, {"iterator", "part", "key", "value", "set"}, {"done", "value"},
		{"__program", "lastIndex", "source", "flags", "global", "ignoreCase", "multiline", "unicode", "sticky", "hasIndices", "unicodeSets", "dotAll"},
		{"index", "input", "groups", "indices"}, {"regex", "input", "done"},
		{"nodeKind", "symbolName", "type"},
		{"bytes", "finalized"},
		{"value", "writable", "enumerable", "configurable"},
		// Process host layouts are part of the same whole-program proof.
		{"name", "message", "code"}, {"heapUsed"},
		{"name", "entryType", "startTime", "duration"}, {"setBlocking"},
		{"timeOrigin", "now", "mark", "measure", "clearMarks", "clearMeasures"},
	} {
		for index, name := range names {
			if before, found := offsets[name]; found && before != index {
				offsets[name] = -1
			} else if !found {
				offsets[name] = index
			}
		}
	}
	record := func(fields []ir.Field) {
		for index, field := range fields {
			if before, found := offsets[field.Name]; found && before != index {
				offsets[field.Name] = -1 // A conflict stays a conflict, even if a later layout agrees.
			} else if !found {
				offsets[field.Name] = index
			}
		}
	}
	// Class spreads copy only public fields, compacting private and accessor storage.
	// These descriptor layouts are emitted separately from the constructor literal.
	for _, class := range program.Classes {
		if class.Literal {
			record(class.PublicFields)
			continue
		}
		public := []ir.Field{}
		for _, field := range class.Fields {
			if !field.Private {
				public = append(public, field)
			}
		}
		record(public)
	}
	walkExpressions(program, func(expression ir.Expression) {
		if literal, ok := expression.(ir.ObjectLiteral); ok {
			if literal.Spread == nil {
				record(literal.Fields)
			} else if literal.SpreadMaybeUndefined {
				record(emptyFields(literal))
			}
		}
	})
	return offsets
}
