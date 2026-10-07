package native

import (
	"embed"
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

//go:embed apple/*.c apple/*.h
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
	case ir.CFunction:
		panic("native: C functions from bindings aren't emitted yet")
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
	for _, cleanup := range cleanups {
		e.line("%s", cleanup)
	}
	switch {
	case foreign.Kind == ir.Construct:
		e.line("return adamic_apple_box(adamic_apple_constructed(result), true);")
	case function.Returns == 0:
	case foreign.Returns.Kind == ir.NativeObject:
		e.line("return adamic_apple_box(result, %t);", foreign.Retained)
	case foreign.Returns.Kind == ir.NativeString:
		e.line("adamic_string *text = adamic_apple_string_from(result);")
		if foreign.Retained {
			e.line("objc_release(result);")
		}
		e.line("return text;")
	case foreign.Returns.Kind == ir.NativeBoolean:
		e.line("return result != NO;")
	default:
		e.line("return (double)result;")
	}
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
	if combined, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
		return fmt.Errorf("native: clang failed: %w\n%s", err, combined)
	}
	return nil
}
