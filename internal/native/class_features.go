package native

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) publicClassShape(class ir.Class) string {
	if class.Literal {
		return e.shape(class.PublicFields)
	}
	fields := []ir.Field{}
	for _, field := range class.Fields {
		if !field.Private {
			fields = append(fields, field)
		}
	}
	return e.shape(fields)
}

// Register shapes before the declarations are written; descriptors refer to these static layouts.
func (e *emitter) prepareClassShapes() {
	for _, class := range e.program.Classes {
		e.publicClassShape(class)
	}
}
