package native

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// A foreign function (ir.Foreign) is one call into Apple's frameworks: its body is a typed
// objc_msgSend, its arguments converted on the way in and its result on the way out, by
// apple/apple.c (docs/apple.md).

//go:embed apple/*.c apple/*.h apple/*.swift
var apple embed.FS

// UsesApple reports whether a program calls into Apple's frameworks, and so builds with them.
func UsesApple(program *ir.Program) bool {
	for _, function := range program.Functions {
		if function.Foreign != nil {
			return true
		}
	}
	return false
}

// appleInclude is the header a program that calls Apple's frameworks needs, or nothing.
func appleInclude(program *ir.Program) string {
	if !UsesApple(program) {
		return ""
	}
	return "#include \"adamic_apple.h\"\n\n"
}

// nativeCType is the C type a native value crosses objc_msgSend as.
func nativeCType(native ir.NativeType) string {
	switch native.Kind {
	case ir.NativeVoid:
		return "void"
	case ir.NativeDouble:
		return "double"
	case ir.NativeInteger, ir.NativeEnumeration:
		return "long"
	case ir.NativeUnsigned, ir.NativeOptions:
		return "unsigned long"
	case ir.NativeBoolean:
		return "BOOL"
	case ir.NativeRectangle:
		return "adamic_apple_rectangle"
	}
	return "id"
}

// foreignBody emits a foreign function's body: the native arguments, the message, what it gives
// back. Every reference parameter is borrowed: the caller keeps it alive for the call.
func (e *emitter) foreignBody(function ir.Function) {
	foreign := function.Foreign
	e.line("ADAMIC_CHECK_STACK();")
	parameter := func(argument ir.ForeignArgument) string {
		return e.localName(function.Parameters[argument.Parameter])
	}
	natives := []string{}
	types := []string{}
	cleanups := []string{}
	actions := []string{}
	delegates := []string{}
	for index, argument := range foreign.Arguments {
		name := fmt.Sprintf("native_%d", index)
		given := argument.Parameter >= 0
		ctype := nativeCType(argument.Type)
		value := ""
		switch argument.Type.Kind {
		case ir.NativeDouble:
			value = argument.Default
			if given {
				value = parameter(argument)
			}
		case ir.NativeInteger:
			value = "adamic_apple_integer(" + argument.Default + ")"
			if given {
				value = "adamic_apple_integer(" + parameter(argument) + ")"
			}
		case ir.NativeUnsigned:
			value = "adamic_apple_unsigned(" + argument.Default + ")"
			if given {
				value = "adamic_apple_unsigned(" + parameter(argument) + ")"
			}
		case ir.NativeBoolean:
			value = map[string]string{"yes": "YES", "no": "NO"}[argument.Default]
			if given {
				value = "(" + parameter(argument) + " ? YES : NO)"
			}
		case ir.NativeString:
			value = "adamic_apple_string(" + parameter(argument) + ")"
			cleanups = append(cleanups, "adamic_apple_let_go("+name+");")
		case ir.NativeObject:
			value = "nil"
			if given {
				value = "adamic_apple_unbox(" + parameter(argument) + ")"
			}
		case ir.NativeRectangle:
			value = "adamic_apple_rectangle_from(" + parameter(argument) + ")"
		case ir.NativeObjects:
			value = "adamic_apple_objects(" + parameter(argument) + ")"
			cleanups = append(cleanups, "adamic_apple_let_go("+name+");")
		case ir.NativeEnumeration, ir.NativeOptions:
			names, values := []string{}, []string{}
			for member, memberName := range argument.Type.Names {
				names = append(names, cString(memberName))
				values = append(values, fmt.Sprintf("%dL", argument.Type.Values[member]))
			}
			e.line("static const char *const %s_names[] = {%s};", name, strings.Join(names, ", "))
			e.line("static const long %s_values[] = {%s};", name, strings.Join(values, ", "))
			convert := "adamic_apple_enumeration"
			if argument.Type.Kind == ir.NativeOptions {
				convert = "adamic_apple_options"
			}
			value = fmt.Sprintf("%s(%s, %d, %s_names, %s_values)", convert, parameter(argument), len(names), name, name)
		case ir.NativeBlock:
			// A block on this function's stack, holding the closure; Apple copies it if it keeps it.
			block := e.blockFunctions(argument.Type)
			e.line("adamic_apple_block %s;", name)
			e.line("adamic_apple_block_start(&%s, %s, &%s_descriptor, (void (*)(void))%s_invoke);", name, parameter(argument), block, block)
			natives, types = append(natives, "(id)&"+name), append(types, "id")
			cleanups = append(cleanups, "adamic_apple_block_end(&"+name+");")
			continue
		case ir.NativeDelegate:
			// An object of the program's class: an instance of the class made for it, holding it.
			if !given {
				e.line("id %s = nil;", name)
			} else {
				e.line("id %s = adamic_apple_delegate(%s, &%s);", name, parameter(argument), e.delegateClass(argument.Type.Delegate))
				delegates = append(delegates, name)
			}
			natives, types = append(natives, name), append(types, "id")
			continue
		case ir.NativeAction:
			// One closure is two native arguments: the target that calls it, and its action.
			e.line("id %s = adamic_apple_action(%s);", name, parameter(argument))
			natives, types = append(natives, name, "adamic_apple_action_selector()"), append(types, "id", "SEL")
			actions = append(actions, name)
			continue
		}
		e.line("%s %s = %s;", ctype, name, value)
		natives, types = append(natives, name), append(types, ctype)
	}

	if foreign.Kind == ir.CFunction {
		e.foreignFunctionCall(function, natives, types, cleanups)
		return
	}
	selector := "selector"
	e.line("static SEL %s;", selector)
	e.line("if (%s == NULL) {", selector)
	e.line("\t%s = sel_registerName(%s);", selector, cString(foreign.Selector))
	e.line("}")
	receiver := ""
	switch foreign.Kind {
	case ir.InstanceMessage:
		// The receiver is the first native argument.
		receiver, natives, types = natives[0], natives[1:], types[1:]
	case ir.ClassMessage, ir.Construct:
		e.line("static id class;")
		e.line("if (class == nil) {")
		e.line("\tclass = adamic_apple_class(%s);", cString(foreign.Class))
		e.line("}")
		receiver = "class"
		if foreign.Kind == ir.Construct {
			e.line("static SEL allocate;")
			e.line("if (allocate == NULL) {")
			e.line("\tallocate = sel_registerName(\"alloc\");")
			e.line("}")
			e.line("id allocated = ((id (*)(id, SEL))objc_msgSend)(class, allocate);")
			receiver = "allocated"
		}
	}
	returns := nativeCType(foreign.Returns)
	if foreign.Kind == ir.Construct {
		returns = "id"
	}
	call := fmt.Sprintf("((%s (*)(%s))objc_msgSend)(%s)", returns, strings.Join(append([]string{"id", "SEL"}, types...), ", "), strings.Join(append([]string{receiver, selector}, natives...), ", "))
	result := ""
	if returns == "void" {
		e.line("%s;", call)
	} else {
		e.line("%s result = %s;", returns, call)
		result = "result"
	}
	owner := receiver
	if foreign.Kind != ir.InstanceMessage {
		owner = result
	}
	for _, action := range actions {
		// A control holds its target weakly: the object it's given to holds it from here on.
		e.line("adamic_apple_keep(%s, %s, &%s);", owner, action, selector)
		e.line("adamic_apple_let_go(%s);", action)
	}
	for _, delegate := range delegates {
		// Apple holds a delegate weakly too: the object it's given to holds it, one per selector that
		// sets one, so setting another (or undefined) lets the last go.
		e.line("adamic_apple_keep(%s, %s, (const void *)%s);", owner, delegate, selector)
		e.line("if (%s != nil) {", delegate)
		e.line("\tadamic_apple_let_go(%s);", delegate)
		e.line("}")
	}
	for _, cleanup := range cleanups {
		e.line("%s", cleanup)
	}
	e.foreignReturn(function, result)
}

// foreignReturn gives back a foreign call's result, named result, as Adamic holds it.
func (e *emitter) foreignReturn(function ir.Function, result string) {
	foreign := function.Foreign
	// What Apple promises isn't nil, or what new made, is checked: nil there panics, never trusted.
	what := cString(strings.TrimSpace(foreign.Class + " " + foreign.Selector))
	switch {
	case foreign.Kind == ir.Construct && foreign.Returns.Kind == ir.NativeObject:
		e.line("return adamic_apple_box(adamic_apple_constructed(adamic_apple_present(%s, %s)), true);", result, what)
	case function.Returns == 0:
	case foreign.Returns.Kind == ir.NativeObject && !foreign.Returns.Nullable:
		e.line("return adamic_apple_box(adamic_apple_present(%s, %s), %t);", result, what, foreign.Retained)
	case foreign.Returns.Kind == ir.NativeObject:
		e.line("return adamic_apple_box(%s, %t);", result, foreign.Retained)
	case foreign.Returns.Kind == ir.NativeString:
		if foreign.Returns.Nullable {
			e.line("if (%s == nil) {", result)
			e.line("\treturn NULL;")
			e.line("}")
		} else {
			e.line("adamic_apple_present(%s, %s);", result, what)
		}
		e.line("adamic_string *text = adamic_apple_string_from(%s);", result)
		if foreign.Retained {
			e.line("objc_release(%s);", result)
		}
		e.line("return text;")
	case foreign.Returns.Kind == ir.NativeBoolean:
		e.line("return %s != NO;", result)
	default:
		e.line("return (double)%s;", result)
	}
}

// foreignFunctionCall emits a C function's call: declared here with the native types it's called
// with, since the program's C includes none of Apple's headers but libobjc's.
func (e *emitter) foreignFunctionCall(function ir.Function, natives []string, types []string, cleanups []string) {
	foreign := function.Foreign
	returns := nativeCType(foreign.Returns)
	parameters := types
	if len(parameters) == 0 {
		parameters = []string{"void"}
	}
	prototype := fmt.Sprintf("%s %s(%s);", returns, foreign.Selector, strings.Join(parameters, ", "))
	if e.foreignPrototypes == nil {
		e.foreignPrototypes = map[string]string{}
	}
	if declared, seen := e.foreignPrototypes[foreign.Selector]; !seen {
		e.foreignPrototypes[foreign.Selector] = prototype
		e.declarations = append(e.declarations, prototype)
	} else if declared != prototype {
		panic("native: " + foreign.Selector + " is bound twice with different types: " + declared + " and " + prototype)
	}
	call := fmt.Sprintf("%s(%s)", foreign.Selector, strings.Join(natives, ", "))
	if returns == "void" {
		e.line("%s;", call)
	} else {
		e.line("%s result = %s;", returns, call)
	}
	for _, cleanup := range cleanups {
		e.line("%s", cleanup)
	}
	e.foreignReturn(function, "result")
}

// blockFunctions declares, once for each block type, what a block of that type needs at file scope:
// the arguments it's called with, held for the main thread; invoke, which Apple calls on any thread
// and which only retains and hands over; deliver, which runs on the main thread, makes the arguments
// Adamic values and calls the closure; and the block's descriptor, with its signature. It returns the
// name they share as a prefix.
func (e *emitter) blockFunctions(block ir.NativeType) string {
	key := fmt.Sprint(block.Parameters)
	if e.blockTypes == nil {
		e.blockTypes = map[string]string{}
	}
	if name, declared := e.blockTypes[key]; declared {
		return name
	}
	index := len(e.blockTypes)
	name := fmt.Sprintf("adamic_block_%d", index)
	e.blockTypes[key] = name
	fields := []string{"\tid holder;"}
	parameters := []string{"adamic_apple_block *block"}
	held := []string{"\tcall->holder = objc_retain(block->holder);"}
	converted := []string{}
	releases := []string{}
	// The signature: the block itself, then each argument, eight bytes apart.
	signature := fmt.Sprintf("v%d@?0", 8*(len(block.Parameters)+1))
	for position, parameter := range block.Parameters {
		ctype := nativeCType(parameter)
		fields = append(fields, fmt.Sprintf("\t%s p%d;", ctype, position))
		parameters = append(parameters, fmt.Sprintf("%s p%d", ctype, position))
		encoding := map[ir.NativeKind]string{ir.NativeDouble: "d", ir.NativeInteger: "q", ir.NativeUnsigned: "Q", ir.NativeBoolean: "B"}[parameter.Kind]
		if encoding == "" {
			encoding = "@"
		}
		signature += fmt.Sprintf("%s%d", encoding, 8*(position+1))
		value := fmt.Sprintf("call->p%d", position)
		switch parameter.Kind {
		case ir.NativeObject:
			held = append(held, fmt.Sprintf("\tcall->p%d = objc_retain(p%d);", position, position))
			converted = append(converted, fmt.Sprintf("\targuments[%d].reference = adamic_apple_box(%s, true);", position, value))
			releases = append(releases, fmt.Sprintf("\tadamic_release(arguments[%d].reference);", position))
		case ir.NativeString:
			held = append(held, fmt.Sprintf("\tcall->p%d = objc_retain(p%d);", position, position))
			converted = append(converted, fmt.Sprintf("\targuments[%d].reference = adamic_apple_string_from(%s);", position, value), fmt.Sprintf("\tobjc_release(%s);", value))
			releases = append(releases, fmt.Sprintf("\tadamic_release(arguments[%d].reference);", position))
		case ir.NativeBoolean:
			held = append(held, fmt.Sprintf("\tcall->p%d = p%d;", position, position))
			converted = append(converted, fmt.Sprintf("\targuments[%d].boolean = %s != NO;", position, value))
		default:
			held = append(held, fmt.Sprintf("\tcall->p%d = p%d;", position, position))
			converted = append(converted, fmt.Sprintf("\targuments[%d].number = (double)%s;", position, value))
		}
	}
	count := len(block.Parameters)
	if count == 0 {
		count = 1
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "struct %s_call {\n%s\n};\n\n", name, strings.Join(fields, "\n"))
	fmt.Fprintf(&builder, "static void %s_deliver(void *context) {\n\tstruct %s_call *call = context;\n\tadamic_value arguments[%d];\n", name, name, count)
	if len(converted) > 0 {
		builder.WriteString(strings.Join(converted, "\n") + "\n")
	}
	builder.WriteString("\tadamic_apple_block_call(call->holder, arguments);\n")
	if len(releases) > 0 {
		builder.WriteString(strings.Join(releases, "\n") + "\n")
	}
	builder.WriteString("\tobjc_release(call->holder);\n\tadamic_apple_call_free(call);\n}\n\n")
	fmt.Fprintf(&builder, "static void %s_invoke(%s) {\n\tstruct %s_call *call = adamic_apple_call_new(sizeof *call);\n%s\n\tadamic_apple_on_main(call, %s_deliver);\n}\n\n", name, strings.Join(parameters, ", "), name, strings.Join(held, "\n"), name)
	fmt.Fprintf(&builder, "static const adamic_apple_block_descriptor %s_descriptor = {0, sizeof(adamic_apple_block), adamic_apple_block_copy, adamic_apple_block_dispose, %s, NULL};", name, cString(signature))
	e.declarations = append(e.declarations, builder.String())
	return name
}

// delegateClass declares, once for each delegate, its class's description and a method for each
// protocol method the program's class has: each converts what Apple hands it into Adamic values,
// calls the function lowering made for it with the object the delegate holds, and gives back its
// result as Apple takes it. It returns the description's name.
func (e *emitter) delegateClass(delegate *ir.Delegate) string {
	name := "adamic_delegate_" + cIdentifier.ReplaceAllString(delegate.Name, "")
	if e.delegateClasses == nil {
		e.delegateClasses = map[string]bool{}
	}
	if e.delegateClasses[delegate.Name] {
		return name
	}
	e.delegateClasses[delegate.Name] = true
	var builder strings.Builder
	entries := []string{}
	encodings := map[ir.NativeKind]string{ir.NativeVoid: "v", ir.NativeDouble: "d", ir.NativeInteger: "q", ir.NativeUnsigned: "Q", ir.NativeBoolean: "B"}
	encoding := func(native ir.NativeType) string {
		if code, known := encodings[native.Kind]; known {
			return code
		}
		return "@"
	}
	fmt.Fprintf(&builder, "static adamic_apple_delegate_class %s;\n\n", name)
	for index, method := range delegate.Methods {
		implementation := fmt.Sprintf("%s_%d", name, index)
		what := cString(delegate.Name + " " + method.Selector)
		parameters := []string{"id self", "SEL command"}
		types := encoding(method.Returns) + "@:"
		converted, releases, arguments := []string{}, []string{}, []string{"object"}
		for position, native := range method.Parameters {
			parameters = append(parameters, fmt.Sprintf("%s p%d", nativeCType(native), position))
			types += encoding(native)
			value := ""
			switch native.Kind {
			case ir.NativeObject:
				value = fmt.Sprintf("adamic_apple_box(p%d, false)", position)
				if !native.Nullable {
					value = fmt.Sprintf("adamic_apple_box(adamic_apple_present(p%d, %s), false)", position, what)
				}
				releases = append(releases, fmt.Sprintf("\tadamic_release(a%d);", position))
			case ir.NativeString:
				value = fmt.Sprintf("p%d == nil ? NULL : adamic_apple_string_from(p%d)", position, position)
				if !native.Nullable {
					value = fmt.Sprintf("adamic_apple_string_from(adamic_apple_present(p%d, %s))", position, what)
				}
				releases = append(releases, fmt.Sprintf("\tadamic_release(a%d);", position))
			case ir.NativeBoolean:
				value = fmt.Sprintf("p%d != NO", position)
			default:
				value = fmt.Sprintf("(double)p%d", position)
			}
			converted = append(converted, fmt.Sprintf("\t%s a%d = %s;", cType(adamicTypeOf(native)), position, value))
			arguments = append(arguments, fmt.Sprintf("a%d", position))
		}
		fmt.Fprintf(&builder, "static %s %s(%s) {\n\t(void)command;\n", nativeCType(method.Returns), implementation, strings.Join(parameters, ", "))
		fmt.Fprintf(&builder, "\tadamic_object *object = adamic_apple_delegate_object(self, &%s);\n", name)
		for _, line := range converted {
			builder.WriteString(line + "\n")
		}
		call := fmt.Sprintf("%s(%s)", e.functionName(method.Function), strings.Join(arguments, ", "))
		if method.Returns.Kind == ir.NativeVoid {
			fmt.Fprintf(&builder, "\t%s;\n", call)
		} else {
			fmt.Fprintf(&builder, "\t%s result = %s;\n", cType(adamicTypeOf(method.Returns)), call)
		}
		for _, line := range releases {
			builder.WriteString(line + "\n")
		}
		builder.WriteString("\tadamic_apple_delegate_returned();\n")
		switch method.Returns.Kind {
		case ir.NativeVoid:
		case ir.NativeBoolean:
			builder.WriteString("\treturn result ? YES : NO;\n")
		case ir.NativeDouble:
			builder.WriteString("\treturn result;\n")
		case ir.NativeInteger:
			builder.WriteString("\treturn adamic_apple_integer(result);\n")
		case ir.NativeUnsigned:
			builder.WriteString("\treturn adamic_apple_unsigned(result);\n")
		case ir.NativeString:
			// Made for Apple and autoreleased, as a method's result is.
			builder.WriteString("\tid native = result == NULL ? nil : adamic_apple_give_back(adamic_apple_string(result));\n\tadamic_release(result);\n\treturn native;\n")
		default:
			builder.WriteString("\tid native = result == NULL ? nil : objc_autorelease(objc_retain(adamic_apple_unbox(result)));\n\tadamic_release(result);\n\treturn native;\n")
		}
		builder.WriteString("}\n\n")
		entries = append(entries, fmt.Sprintf("{%s, (IMP)%s, %s}", cString(method.Selector), implementation, cString(types)))
	}
	protocols := []string{}
	for _, protocol := range delegate.Protocols {
		protocols = append(protocols, cString(protocol))
	}
	fmt.Fprintf(&builder, "static const char *const %s_protocols[] = {%s};\n", name, strings.Join(append(protocols, "NULL"), ", "))
	if len(entries) == 0 {
		entries = append(entries, "{NULL, NULL, NULL}")
	}
	fmt.Fprintf(&builder, "static const adamic_apple_delegate_method %s_methods[] = {%s};\n", name, strings.Join(entries, ", "))
	fmt.Fprintf(&builder, "static adamic_apple_delegate_class %s = {%s, %d, %s_protocols, %d, %s_methods, Nil, 0};\n", name, cString(delegate.Name), len(delegate.Protocols), name, len(delegate.Methods), name)
	e.declarations = append(e.declarations, builder.String())
	return name
}

// adamicTypeOf is how a native value is held in Adamic, as lowering's adamicType says.
func adamicTypeOf(native ir.NativeType) ir.Type {
	switch native.Kind {
	case ir.NativeDouble, ir.NativeInteger, ir.NativeUnsigned:
		return ir.Number
	case ir.NativeBoolean:
		return ir.Boolean
	case ir.NativeString:
		return ir.String
	}
	return ir.Object
}

// swiftUIShim is swiftui.swift compiled to an object, cached under the user cache directory by the
// source and swiftc's version, so a build after the first takes no Swift compile.
func swiftUIShim() (string, error) {
	source, err := apple.ReadFile("apple/swiftui.swift")
	if err != nil {
		return "", fmt.Errorf("native: %w", err)
	}
	version, err := exec.Command("xcrun", "swiftc", "--version").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("native: swiftc --version: %w\n%s", err, version)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("native: cache directory: %w", err)
	}
	key := sha256.Sum256(append(append([]byte{}, source...), version...))
	directory := filepath.Join(cache, "adamic", "swiftui", hex.EncodeToString(key[:]))
	object := filepath.Join(directory, "swiftui.o")
	if _, err := os.Stat(object); err == nil {
		return object, nil
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", fmt.Errorf("native: %w", err)
	}
	sourcePath := filepath.Join(directory, "swiftui.swift")
	if err := os.WriteFile(sourcePath, source, 0o644); err != nil {
		return "", fmt.Errorf("native: %w", err)
	}
	// Written beside its final name and renamed, so a build running at the same time never links half
	// an object.
	partial := object + fmt.Sprintf(".%d", os.Getpid())
	if combined, err := exec.Command("xcrun", "swiftc", "-parse-as-library", "-emit-object", "-O", "-module-name", "AdamicSwiftUI", "-target", "arm64-apple-macos15.0", sourcePath, "-o", partial).CombinedOutput(); err != nil {
		return "", fmt.Errorf("native: swiftc failed: %w\n%s", err, combined)
	}
	if err := os.Rename(partial, object); err != nil {
		return "", fmt.Errorf("native: %w", err)
	}
	return object, nil
}

// BuildApple compiles a program that calls Apple's frameworks: Build's, with the Objective-C
// conversions compiled beside it and AppKit, Foundation and libobjc linked.
func BuildApple(source string, output string, options Options) error {
	directory, err := os.MkdirTemp("", "adamic-apple-")
	if err != nil {
		return fmt.Errorf("native: %w", err)
	}
	defer os.RemoveAll(directory)
	library, err := RuntimeLibrary("", options)
	if err != nil {
		return err
	}
	units := []string{filepath.Join(directory, "main.c")}
	if err := os.WriteFile(units[0], []byte(source), 0o644); err != nil {
		return fmt.Errorf("native: %w", err)
	}
	entries, err := apple.ReadDir("apple")
	if err != nil {
		return fmt.Errorf("native: %w", err)
	}
	for _, entry := range entries {
		contents, err := apple.ReadFile("apple/" + entry.Name())
		if err != nil {
			return fmt.Errorf("native: %w", err)
		}
		path := filepath.Join(directory, entry.Name())
		if err := os.WriteFile(path, contents, 0o644); err != nil {
			return fmt.Errorf("native: %w", err)
		}
		if strings.HasSuffix(path, ".c") {
			units = append(units, path)
		}
	}
	arguments := append(Flags(options), "-I", filepath.Dir(library), "-I", directory, "-o", output)
	arguments = append(arguments, units...)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-framework", "AppKit", "-framework", "Foundation", "-framework", "CoreFoundation", "-lobjc")
	// SwiftUI is Swift only: a program that names one of swiftui.swift's classes links it, compiled
	// once for each version of it and of swiftc, and Swift's libraries with it.
	if strings.Contains(source, "AdamicSwiftUI") {
		shim, err := swiftUIShim()
		if err != nil {
			return err
		}
		sdk, err := exec.Command("xcrun", "--show-sdk-path").Output()
		if err != nil {
			return fmt.Errorf("native: the SDK's path: %w", err)
		}
		arguments = append(arguments, shim, "-framework", "SwiftUI", "-L", filepath.Join(strings.TrimSpace(string(sdk)), "usr", "lib", "swift"), "-L", "/usr/lib/swift", "-Xlinker", "-rpath", "-Xlinker", "/usr/lib/swift")
	}
	if combined, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
		return fmt.Errorf("native: clang failed: %w\n%s", err, combined)
	}
	return nil
}
