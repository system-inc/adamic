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
				entries = append(entries, "(adamic_virtual_method)"+e.virtualAdapter(builder, method))
			}
			methods = fmt.Sprintf("adamic_methods_%d", index+1)
			fmt.Fprintf(builder, "static const adamic_virtual_method %s[] = {%s};\n", methods, strings.Join(entries, ", "))
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
		return e.directVirtualCode(call, targets[0], arguments)
	}
	if class := e.exactReceiverClass(call.Arguments[0]); class != 0 {
		return e.directVirtualCode(call, e.program.Classes[class-1].Methods[call.Virtual-1], arguments)
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

// virtualAdapter gives the table a borrowing ABI. A consuming implementation alone
// takes a count; direct callers keep the implementation's own convention.
func (e *emitter) virtualAdapter(builder *strings.Builder, function int) string {
	declared := e.program.Functions[function]
	consumes := false
	for _, parameter := range declared.Parameters {
		consumes = consumes || e.reuse.consumed[parameter]
	}
	if !consumes {
		return e.functionName(function)
	}
	name := e.functionName(function) + "_virtual"
	// Tables may share a function. Each class gets its own adapter name.
	name += fmt.Sprintf("_%d", builder.Len())
	parameters, arguments := []string{}, []string{}
	for position, parameter := range declared.Parameters {
		argument := fmt.Sprintf("argument_%d", position)
		parameters = append(parameters, cType(e.program.Locals[parameter].Type)+" "+argument)
		if e.reuse.consumed[parameter] {
			argument = retained(argument)
		}
		arguments = append(arguments, argument)
	}
	result := "void"
	if declared.Returns != 0 {
		result = cType(declared.Returns)
	}
	fmt.Fprintf(builder, "static %s %s(%s) {\n", result, name, strings.Join(parameters, ", "))
	prefix := ""
	if declared.Returns != 0 {
		prefix = "return "
	}
	fmt.Fprintf(builder, "\t%s%s(%s);\n}\n", prefix, e.functionName(function), strings.Join(arguments, ", "))
	return name
}

// A devirtualized call still enters the virtual borrowing ABI. Replacing the
// table lookup must not hand a caller's count to a consuming implementation.
func (e *emitter) directVirtualCode(call ir.Call, function int, arguments []string) string {
	if call.Virtual != 0 && e.reuse != nil {
		arguments = append([]string(nil), arguments...)
		for position, parameter := range e.program.Functions[function].Parameters {
			if position < len(arguments) && e.reuse.consumed[parameter] {
				arguments[position] = retained(arguments[position])
			}
		}
	}
	return fmt.Sprintf("%s(%s)", e.functionName(function), strings.Join(arguments, ", "))
}
