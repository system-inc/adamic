package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// classDeclarations emits one table per class, with inherited slots retained in place. Function
// pointers are converted back to their original signature before calls, as C11 permits.
func (e *emitter) classDeclarations(builder *strings.Builder) {
	for index := range e.program.Classes {
		fmt.Fprintf(builder, "static const adamic_class adamic_class_%d;\n", index+1)
	}
	for index, class := range e.program.Classes {
		methods := "NULL"
		if len(class.Methods) > 0 {
			entries := []string{}
			for _, method := range class.Methods {
				entries = append(entries, "(adamic_method)"+e.functionName(method))
			}
			methods = fmt.Sprintf("adamic_methods_%d", index+1)
			fmt.Fprintf(builder, "static const adamic_method %s[] = {%s};\n", methods, strings.Join(entries, ", "))
		}
		base := "NULL"
		if class.Base != 0 {
			base = fmt.Sprintf("&adamic_class_%d", class.Base)
		}
		fmt.Fprintf(builder, "static const adamic_class adamic_class_%d = {%s, %d, %d, %s, %d, &%s, %s, %d};\n", index+1, base, class.OwnStart, len(class.Fields), methods, class.Definition, e.publicClassShape(class), e.accessorDeclarations(builder, index+1, class), len(class.Accessors))
	}
}

func (e *emitter) callCode(call ir.Call, arguments []string) string {
	if call.Virtual == 0 {
		return fmt.Sprintf("%s(%s)", e.functionName(call.Function), strings.Join(arguments, ", "))
	}
	signature := e.program.Functions[call.Function]
	parameters := []string{}
	for _, parameter := range signature.Parameters {
		parameters = append(parameters, cType(e.program.Locals[parameter].Type))
	}
	result := "void"
	if signature.Returns != 0 {
		result = cType(signature.Returns)
	}
	code := fmt.Sprintf("((%s (*)(%s))adamic_virtual(%s, %d))", result, strings.Join(parameters, ", "), arguments[0], call.Virtual-1)
	return code + "(" + strings.Join(arguments, ", ") + ")"
}
