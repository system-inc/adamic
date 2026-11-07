package native

import "github.com/system-inc/adamic/internal/ir"

func (nativeEmitter *emitter) publicClassShape(class ir.Class) string {
	if class.Literal {
		return nativeEmitter.shape(class.PublicFields)
	}
	fields := []ir.Field{}
	for _, field := range class.Fields {
		if !field.Private {
			fields = append(fields, field)
		}
	}
	return nativeEmitter.shape(fields)
}

// Register shapes before the declarations are written; descriptors refer to these static layouts.
func (nativeEmitter *emitter) prepareClassShapes() {
	for _, class := range nativeEmitter.program.Classes {
		nativeEmitter.publicClassShape(class)
	}
}
