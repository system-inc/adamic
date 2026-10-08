package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) viewArrayTupleLength(read ir.Length) string {
	if !ir.HasArrayTupleViews(e.program) {
		return fmt.Sprint(read.TupleArity)
	}
	message := "field read failed: " + read.TupleView + " is not a " + read.TupleType + "; expected " + read.TupleType + ", found "
	return fmt.Sprintf("((value)=>{if (!(Array.isArray(value) && value.length===%d)) panic(%s+(value===undefined?'undefined':Array.isArray(value)?'array':typeof value));return value.length;})(%s)", read.TupleArity, quote(message), e.value(read.Array))
}
