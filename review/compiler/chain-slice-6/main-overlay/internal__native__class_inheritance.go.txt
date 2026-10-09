package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// classDeclarations emits one table per class, with inherited slots retained in place. Function
// identities select a typed direct call; no function pointer loses its signature.
func (e *emitter) classDeclarations(builder *strings.Builder) {
	for index := range e.program.Classes {
		fmt.Fprintf(builder, "static const adamic_class adamic_class_%d;\n", index+1)
	}
	for index, class := range e.program.Classes {
		methods := "NULL"
		if len(class.Methods) > 0 {
			entries := []string{}
			for _, method := range class.Methods {
				entries = append(entries, fmt.Sprint(method))
			}
			methods = fmt.Sprintf("adamic_methods_%d", index+1)
			fmt.Fprintf(builder, "static const size_t %s[] = {%s};\n", methods, strings.Join(entries, ", "))
		}
		base := "NULL"
		if class.Base != 0 {
			base = fmt.Sprintf("&adamic_class_%d", class.Base)
		}
		flags := "NULL"
		if class.Static {
			flags = fmt.Sprintf("adamic_static_flags_%d", index+1)
			entries := []string{}
			for _, flag := range class.StaticFlags {
				entries = append(entries, fmt.Sprint(flag))
			}
			if len(entries) == 0 {
				entries = append(entries, "0")
			}
			fmt.Fprintf(builder, "static const size_t %s[] = {%s};\n", flags, strings.Join(entries, ", "))
		}
		fmt.Fprintf(builder, "static const adamic_class adamic_class_%d = {%s, %d, %d, %s, %d, &%s, %s, %d, %t, %d, %s};\n", index+1, base, class.OwnStart, len(class.Fields), methods, class.Definition, e.publicClassShape(class), e.accessorDeclarations(builder, index+1, class), len(class.Accessors), class.Static, class.StaticParent, flags)
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
	name := fmt.Sprintf("adamic_virtual_dispatch_%d", call.Function)
	parameters, values := []string{}, []string{}
	for index, parameter := range signature.Parameters {
		value := fmt.Sprintf("argument_%d", index)
		parameters = append(parameters, cType(e.program.Locals[parameter].Type)+" "+value)
		values = append(values, value)
	}
	if signature.ArgumentsCount != 0 {
		parameters = append(parameters, "double argument_count")
		values = append(values, "argument_count")
	}
	result := "void"
	if signature.Returns != 0 {
		result = cType(signature.Returns)
	}
	prefix := fmt.Sprintf("static %s %s(", result, name)
	declared := false
	for _, declaration := range e.declarations {
		if strings.HasPrefix(declaration, prefix) {
			declared = true
			break
		}
	}
	if !declared {
		var body strings.Builder
		fmt.Fprintf(&body, "%s%s) {\n\tswitch (adamic_virtual(%s, %d)) {\n", prefix, strings.Join(parameters, ", "), values[0], call.Virtual-1)
		for _, target := range targets {
			invocation := fmt.Sprintf("%s(%s)", e.functionName(target), strings.Join(values, ", "))
			if signature.Returns == 0 {
				fmt.Fprintf(&body, "\tcase %d: %s; return;\n", target, invocation)
			} else {
				fmt.Fprintf(&body, "\tcase %d: return %s;\n", target, invocation)
			}
		}
		body.WriteString("\tdefault: adamic_panic(\"compiler bug: virtual target outside the closed world\", sizeof \"compiler bug: virtual target outside the closed world\" - 1);\n\t}\n}")
		e.declarations = append(e.declarations, body.String())
	}
	return fmt.Sprintf("%s(%s)", name, strings.Join(arguments, ", "))
}
