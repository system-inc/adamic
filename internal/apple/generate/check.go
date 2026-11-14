package generate

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type checkCall struct {
	owner              *definition
	selector           string
	instance, function bool
	parameters         []nativeType
	result             nativeType
}

func (g *generator) emitCheck() (string, error) {
	var text strings.Builder
	text.WriteString("// Generated header witness. Compile with clang -fsyntax-only -Werror.\n")
	if g.configuration.Umbrella != "" {
		if strings.ContainsAny(g.configuration.Umbrella, "\n\r\"") {
			return "", fmt.Errorf("invalid umbrella include")
		}
		text.WriteString("#include " + strconv.Quote(g.configuration.Umbrella) + "\n")
	} else {
		names := []string{}
		for _, f := range g.configuration.Frameworks {
			names = append(names, f.Name)
		}
		sort.Strings(names)
		for _, name := range names {
			text.WriteString("#import <" + name + "/" + name + ".h>\n")
		}
	}
	text.WriteString("#include <stddef.h>\n\ntypedef struct adamic_apple_rectangle { double x, y, width, height; } adamic_apple_rectangle;\n")
	text.WriteString("_Static_assert(sizeof(long) == 8 && sizeof(void *) == 8, \"the bridge requires a 64-bit Apple ABI\");\n")
	for _, name := range sortedKeys(g.types) {
		d := g.types[name]
		if !d.options && d.declaration.Kind != "enum" {
			continue
		}
		constants := children(d.node, "EnumConstantDecl")
		available := []*node{}
		for _, constant := range constants {
			_, _, gone, err := g.attributes(constant)
			if err != nil {
				return "", err
			}
			if gone == "" {
				available = append(available, constant)
			}
		}
		// Option values are written unsigned, so bit 63 needs its suffix to stay a literal C reads.
		suffix := ""
		if d.options {
			suffix = "UL"
		}
		for i, constant := range available {
			if i < len(d.values) {
				fmt.Fprintf(&text, "_Static_assert(%s == %s%s, \"%s enum value\");\n", constant.Name, d.values[i], suffix, constant.Name)
			}
		}
	}
	sort.Slice(g.checks, func(i, j int) bool { return checkKey(g.checks[i]) < checkKey(g.checks[j]) })
	for i, call := range g.checks {
		parameters := []string{}
		receiver := ""
		if !call.function {
			if call.instance {
				receiver = "receiver"
				typ := call.owner.node.Name + " *"
				if call.owner.declaration.Kind == "protocol" {
					typ = "id<" + call.owner.node.Name + ">"
				}
				parameters = append(parameters, typ+" _Nonnull receiver")
			} else {
				receiver = call.owner.node.Name
			}
		}
		for index, native := range call.parameters {
			if members, isBlock := strings.CutPrefix(native.tag, "block("); isBlock {
				// A block in the bridge's C types: clang refuses it where the header's block differs.
				types := []string{}
				for _, member := range splitTopLevel(strings.TrimSuffix(members, ")")) {
					if member != "" {
						types = append(types, tagCType(member))
					}
				}
				if len(types) == 0 {
					types = append(types, "void")
				}
				parameters = append(parameters, fmt.Sprintf("void (^argument%d)(%s)", index, strings.Join(types, ", ")))
				continue
			}
			parameters = append(parameters, fmt.Sprintf("%s argument%d", tagCType(native.tag), index))
		}
		if len(parameters) == 0 {
			parameters = append(parameters, "void")
		}
		fmt.Fprintf(&text, "\nvoid adamic_binding_check_%d(%s) {\n", i, strings.Join(parameters, ", "))
		arguments := []string{}
		for index, native := range call.parameters {
			argument := fmt.Sprintf("argument%d", index)
			if native.tag == "rectangle" {
				// C's nominal struct types differ although their ABI agrees. Compare
				// each coordinate's size, offset and alignment before copying the bits.
				header := native.header
				fmt.Fprintf(&text, "\t_Static_assert(sizeof(%s) == sizeof(adamic_apple_rectangle) && _Alignof(%s) == _Alignof(adamic_apple_rectangle), \"rectangle ABI\");\n", header, header)
				for _, field := range []struct{ name, path string }{{"x", "origin.x"}, {"y", "origin.y"}, {"width", "size.width"}, {"height", "size.height"}} {
					fmt.Fprintf(&text, "\t_Static_assert(offsetof(%s, %s) == offsetof(adamic_apple_rectangle, %s) && __builtin_types_compatible_p(__typeof__(((%s *)0)->%s), double), \"rectangle %s\");\n", header, field.path, field.name, header, field.path, field.name)
				}
				fmt.Fprintf(&text, "\t%s header%d;\n\t__builtin_memcpy(&header%d, &%s, sizeof(header%d));\n", header, index, index, argument, index)
				argument = fmt.Sprintf("header%d", index)
			} else if strings.HasPrefix(native.tag, "enum(") {
				// An enumeration crosses as a long whichever its signedness: 64 bits either way.
				fmt.Fprintf(&text, "\t_Static_assert(__builtin_types_compatible_p(long, %s) || __builtin_types_compatible_p(unsigned long, %s), \"%s parameter %d ABI\");\n", native.header, native.header, call.selector, index)
			} else if strings.HasPrefix(native.tag, "options(") {
				// Option bits cross as an unsigned long; unsigned long long is the same 64 bits.
				fmt.Fprintf(&text, "\t_Static_assert(__builtin_types_compatible_p(unsigned long, %s) || __builtin_types_compatible_p(unsigned long long, %s), \"%s parameter %d ABI\");\n", native.header, native.header, call.selector, index)
			} else if native.c != "id" {
				fmt.Fprintf(&text, "\t_Static_assert(%s, \"%s parameter %d ABI\");\n", compatible(tagCType(native.tag), native.header), call.selector, index)
			}
			arguments = append(arguments, argument)
		}
		expression := ""
		if call.function {
			expression = call.selector + "(" + strings.Join(arguments, ", ") + ")"
		} else {
			pieces := strings.Split(call.selector, ":")
			message := call.selector
			if len(arguments) > 0 {
				if len(pieces)-1 != len(arguments) {
					return "", fmt.Errorf("selector %s has wrong parameter count", call.selector)
				}
				parts := []string{}
				for j, argument := range arguments {
					parts = append(parts, pieces[j]+":"+argument)
				}
				message = strings.Join(parts, " ")
			}
			expression = "[" + receiver + " " + message + "]"
		}
		if call.result.c == "void" {
			text.WriteString("\t" + expression + ";\n")
		} else {
			if call.result.c != "id" {
				fmt.Fprintf(&text, "\t_Static_assert(%s, \"%s result ABI\");\n", compatible(tagCType(call.result.tag), "__typeof__("+expression+")"), call.selector)
			}
			fmt.Fprintf(&text, "\t%s result = %s;\n\t(void)result;\n", tagCType(call.result.tag), expression)
		}
		text.WriteString("}\n")
	}
	g.emitImplementChecks(&text)
	return text.String(), nil
}
func checkKey(call checkCall) string {
	owner := ""
	if call.owner != nil {
		// A protocol and a class may share an Objective-C name; the kind keeps their order fixed.
		owner = call.owner.node.Name + ":" + string(call.owner.declaration.Kind)
	}
	return fmt.Sprintf("%s:%t:%s", owner, call.instance, call.selector)
}

// compatible asserts that header is the C type a tag crosses as; a 64-bit integer may be spelled
// long or long long (int64_t, the progress counts), the same 64 bits.
func compatible(c, header string) string {
	text := fmt.Sprintf("__builtin_types_compatible_p(%s, %s)", c, header)
	if c == "long" || c == "unsigned long" {
		text += fmt.Sprintf(" || __builtin_types_compatible_p(%s long, %s)", c, header)
	}
	return text
}

// Derive the witness's native types from the tags themselves, independently
// of the mapper's metadata. This mirrors the documented bridge ABI.
func tagCType(tag string) string {
	switch tag {
	case "void":
		return "void"
	case "double":
		return "double"
	case "integer":
		return "long"
	case "unsigned":
		return "unsigned long"
	case "boolean":
		return "BOOL"
	case "rectangle":
		return "adamic_apple_rectangle"
	case "action":
		// An action is two arguments: the target before it, an object, and this, its selector.
		return "SEL"
	}
	if strings.HasPrefix(tag, "enum(") {
		return "long"
	}
	if strings.HasPrefix(tag, "options(") {
		return "unsigned long"
	}
	return "id"
}
