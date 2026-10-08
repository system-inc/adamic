package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) viewArrayTupleLength(read ir.Length) string {
	if !ir.HasArrayTupleViews(e.program) {
		return cNumber(float64(read.TupleArity))
	}
	receiver := e.value(read.Array)
	e.declarations = append(e.declarations, "#include \"view_tuples.h\"")
	e.line("adamic_view_tuple((const adamic_object *)%s, %d, %s, %s);", receiver, read.TupleArity, cString(read.TupleType), cString(read.TupleView))
	return e.snapshot(ir.Number, fmt.Sprintf("%d.0", read.TupleArity))
}
