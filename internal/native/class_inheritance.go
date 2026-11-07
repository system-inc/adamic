package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// classDeclarations emits one table per class, with inherited slots retained in place. Function
// identities select a typed direct call; no function pointer loses its signature.
func (nativeEmitter *emitter) classDeclarations(builder *strings.Builder) {
	forward := map[string]bool{}
	for index := range nativeEmitter.program.Classes {
		name := nativeEmitter.className(index + 1)
		if !forward[name] {
			fmt.Fprintf(builder, "static const adamic_class %s;\n", name)
			forward[name] = true
		}
	}
	definitions := map[string]string{}
	for index, class := range nativeEmitter.program.Classes {
		var declaration strings.Builder
		methods := "NULL"
		if len(class.Methods) > 0 {
			entries := []string{}
			for _, method := range class.Methods {
				entries = append(entries, fmt.Sprint(method))
			}
			methods = nativeEmitter.className(index+1) + "_methods"
			fmt.Fprintf(&declaration, "static const size_t %s[] = {%s};\n", methods, strings.Join(entries, ", "))
		}
		base := "NULL"
		if class.Base != 0 {
			base = "&" + nativeEmitter.className(class.Base)
		}
		flags := "NULL"
		if class.Static {
			flags = nativeEmitter.className(index+1) + "_flags"
			entries := []string{}
			for _, flag := range class.StaticFlags {
				entries = append(entries, fmt.Sprint(flag))
			}
			if len(entries) == 0 {
				entries = append(entries, "0")
			}
			fmt.Fprintf(&declaration, "static const size_t %s[] = {%s};\n", flags, strings.Join(entries, ", "))
		}
		fmt.Fprintf(
			&declaration,
			"static const adamic_class %s = {%s, %d, %d, %s, %d, &%s, %s, %d, %t, %d, %s};\n",
			nativeEmitter.className(index+1),
			base,
			class.OwnStart,
			len(class.Fields),
			methods,
			class.Definition,
			nativeEmitter.publicClassShape(class),
			nativeEmitter.accessorDeclarations(&declaration, index+1, class),
			len(class.Accessors),
			class.Static,
			class.StaticParent,
			flags,
		)
		name, text := nativeEmitter.className(index+1), declaration.String()
		if previous, exists := definitions[name]; exists {
			if previous != text {
				panic("native: conflicting class definitions for stable symbol " + name)
			}
		} else {
			definitions[name] = text
			builder.WriteString(text)
		}

	}
	nativeEmitter.keyedClassDeclarations(builder)
}

// keyedClassRow is one class's row in the table a run-time key read takes (adamic_keyed_class): its
// emitted name, whether an object literal made it, and its method table's names.
type keyedClassRow struct {
	name    string
	literal bool
	methods []string
}

// keyedClassRows are the program's classes as a run-time key read judges them, one row for each
// emitted class, taken from the IR class the JavaScript backend's helper reads (literal, and
// ir.Program's DynamicKeyedMethodNames), or none in a program without the read.
func (nativeEmitter *emitter) keyedClassRows() []keyedClassRow {
	if !nativeEmitter.keyedReads() {
		return nil
	}
	rows := []keyedClassRow{}
	seen := map[string]bool{}
	for index, class := range nativeEmitter.program.Classes {
		name := nativeEmitter.className(index + 1)
		if seen[name] {
			// One stable symbol, one definition (classDeclarations), so one row.
			continue
		}
		seen[name] = true
		rows = append(
			rows,
			keyedClassRow{name, class.Literal, nativeEmitter.program.DynamicKeyedMethodNames(index + 1)},
		)
	}
	return rows
}

// keyedClassDeclarations writes the table adamic_dynamic_property_keyed takes, only into a program with
// a run-time key read, so every other program's C is what it was.
func (nativeEmitter *emitter) keyedClassDeclarations(builder *strings.Builder) {
	rows := nativeEmitter.keyedClassRows()
	if len(rows) == 0 {
		return
	}
	entries := []string{}
	for _, row := range rows {
		names := "NULL"
		if len(row.methods) > 0 {
			names = row.name + "_keyed_method_names"
			quoted := []string{}
			for _, method := range row.methods {
				quoted = append(quoted, cString(method))
			}
			fmt.Fprintf(builder, "static const char *const %s[] = {%s};\n", names, strings.Join(quoted, ", "))
		}
		entries = append(entries, fmt.Sprintf("\t{&%s, %t, %s, %d},\n", row.name, row.literal, names, len(row.methods)))
	}
	fmt.Fprintf(
		builder,
		"static const adamic_keyed_class adamic_keyed_classes[] = {\n%s};\n",
		strings.Join(entries, ""),
	)
}

func (nativeEmitter *emitter) callCode(call ir.Call, arguments []string) string {
	targets := nativeEmitter.program.CallTargets(call)
	if len(targets) == 1 {
		return fmt.Sprintf("%s(%s)", nativeEmitter.functionName(targets[0]), strings.Join(arguments, ", "))
	}
	if class := nativeEmitter.exactReceiverClass(call.Arguments[0]); class != 0 {
		return fmt.Sprintf(
			"%s(%s)",
			nativeEmitter.functionName(nativeEmitter.program.Classes[class-1].Methods[call.Virtual-1]),
			strings.Join(arguments, ", "),
		)
	}
	signature := nativeEmitter.program.Functions[call.Function]
	name := stableName(
		"adamic_virtual_dispatch",
		nativeEmitter.sourceReadable(call.Function),
		nativeEmitter.functionKey(call.Function, map[int]bool{}),
	)
	parameters, values := []string{}, []string{}
	for index, parameter := range signature.Parameters {
		value := fmt.Sprintf("argument_%d", index)
		parameters = append(parameters, cType(nativeEmitter.program.Locals[parameter].Type)+" "+value)
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
	for _, declaration := range nativeEmitter.declarations {
		if strings.HasPrefix(declaration, prefix) {
			declared = true
			break
		}
	}
	if !declared {
		var body strings.Builder
		fmt.Fprintf(
			&body,
			"%s%s) {\n\tswitch (adamic_virtual(%s, %d)) {\n",
			prefix,
			strings.Join(parameters, ", "),
			values[0],
			call.Virtual-1,
		)
		for _, target := range targets {
			invocation := fmt.Sprintf("%s(%s)", nativeEmitter.functionName(target), strings.Join(values, ", "))
			if signature.Returns == 0 {
				fmt.Fprintf(&body, "\tcase %d: %s; return;\n", target, invocation)
			} else {
				fmt.Fprintf(&body, "\tcase %d: return %s;\n", target, invocation)
			}
		}
		body.WriteString(
			"\tdefault: adamic_panic(\"compiler bug: virtual target outside the closed world\", sizeof \"compiler bug: virtual target outside the closed world\" - 1);\n\t}\n}",
		)
		nativeEmitter.declarations = append(nativeEmitter.declarations, body.String())
	}
	return fmt.Sprintf("%s(%s)", name, strings.Join(arguments, ", "))
}
