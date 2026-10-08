package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// classDeclarations emits one table per class, with inherited slots retained in place. Function
// pointers are converted back to their original signature before calls, as C11 permits.
func (e *emitter) classDeclarations(builder *strings.Builder) {
	forward := map[string]bool{}
	for index := range e.program.Classes {
		name := e.className(index + 1)
		if !forward[name] {
			fmt.Fprintf(builder, "static const adamic_class %s;\n", name)
			forward[name] = true
		}
	}
	definitions := map[string]string{}
	for index, class := range e.program.Classes {
		var declaration strings.Builder
		methods := "NULL"
		if len(class.Methods) > 0 {
			entries := []string{}
			for _, method := range class.Methods {
				entries = append(entries, "(adamic_virtual_method)"+e.functionName(method))
			}
			methods = e.className(index+1) + "_methods"
			fmt.Fprintf(&declaration, "static const adamic_virtual_method %s[] = {%s};\n", methods, strings.Join(entries, ", "))
		}
		base := "NULL"
		if class.Base != 0 {
			base = "&" + e.className(class.Base)
		}
		flags := "NULL"
		if class.Static {
			flags = e.className(index+1) + "_flags"
			entries := []string{}
			for _, flag := range class.StaticFlags {
				entries = append(entries, fmt.Sprint(flag))
			}
			if len(entries) == 0 {
				entries = append(entries, "0")
			}
			fmt.Fprintf(&declaration, "static const size_t %s[] = {%s};\n", flags, strings.Join(entries, ", "))
		}
		fmt.Fprintf(&declaration, "static const adamic_class %s = {%s, %d, %d, %s, %d, &%s, %s, %d, %t, %d, %s};\n", e.className(index+1), base, class.OwnStart, len(class.Fields), methods, class.Definition, e.publicClassShape(class), e.accessorDeclarations(&declaration, index+1, class), len(class.Accessors), class.Static, class.StaticParent, flags)
		name, text := e.className(index+1), declaration.String()
		if previous, exists := definitions[name]; exists {
			if previous != text {
				panic("native: conflicting class definitions for stable symbol " + name)
			}
		} else {
			definitions[name] = text
			builder.WriteString(text)
		}

	}
}

func (e *emitter) callCode(call ir.Call, arguments []string) string {
	targets := e.program.CallTargets(call)
	if len(targets) == 1 {
		return fmt.Sprintf("%s(%s)", e.functionName(targets[0]), strings.Join(arguments, ", "))
	}
	if class := e.exactReceiverClass(call.Arguments[0]); class != 0 {
		return fmt.Sprintf("%s(%s)", e.functionName(e.program.Classes[class-1].Methods[call.Virtual-1]), strings.Join(arguments, ", "))
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
