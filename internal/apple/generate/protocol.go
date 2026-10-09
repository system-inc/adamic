package generate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/system-inc/adamic/internal/apple/naming"
)

// A protocol's methods are what a program's class implements, as a delegate (docs/apple.md,
// "Delegates"): Apple calls them, so each takes its arguments in the selector's order, positional,
// and its @objc implement tag says how each crosses back into Adamic. A required one keeps its
// method tag too, for an object of Apple's that conforms. An optional one is an optional member,
// present when it can cross and left out with its reason when it can't.

// implementCheck is one implement tag, held to its header by a class in the check file that
// conforms to the protocol and implements the method in the bridge's C types.
type implementCheck struct {
	owner      *definition
	selector   string
	parameters []nativeType
	result     nativeType
}

// protocolNames names a protocol's methods. Swift tells them apart by their labels
// (tableView(_:objectValueFor:row:) beside tableView(_:viewFor:row:)), and a class has one method
// per name, so where two share a name each folds its labels into it: tableViewObjectValueForRow.
func (g *generator) protocolNames(def *definition, nodes []*node, accessors map[string]bool) map[*node]string {
	names := map[*node]string{}
	folded := map[*node]string{}
	selectors := map[string]map[string]bool{}
	for _, child := range nodes {
		if child.Kind != "ObjCMethodDecl" || child.Implicit || !child.Instance || accessors[child.Name] {
			continue
		}
		_, out, _, _, reason, err := g.callable(child, naming.InstanceMethod, def.node.Name, def.declaration.SwiftName, g.propertyNames(def, map[string]bool{}), nil)
		if err != nil || reason != "" {
			continue
		}
		labels := out.Name
		for _, argument := range out.Arguments.Options {
			labels += strings.ToUpper(argument.Name[:1]) + argument.Name[1:]
		}
		names[child], folded[child] = out.Name, labels
		if selectors[out.Name] == nil {
			selectors[out.Name] = map[string]bool{}
		}
		selectors[out.Name][child.Name] = true
	}
	for child, name := range names {
		if len(selectors[name]) > 1 {
			names[child] = folded[child]
		}
	}
	// A folded name can land on another's (application(_:didUpdate:) folds to applicationDidUpdate,
	// beside applicationDidUpdate(_:)): the folded one is then the selector's own words,
	// applicationDidUpdateUserActivity. A name still shared by two is no one's.
	owners := map[string][]*node{}
	for child, name := range names {
		owners[name] = append(owners[name], child)
	}
	for _, children := range owners {
		for _, child := range children {
			if len(children) > 1 && names[child] == folded[child] && folded[child] != "" {
				names[child] = selectorWords(child.Name)
			}
		}
	}
	owners = map[string][]*node{}
	for child, name := range names {
		owners[name] = append(owners[name], child)
	}
	for _, children := range owners {
		if len(children) > 1 {
			for _, child := range children {
				names[child] = ""
			}
		}
	}
	return names
}

// selectorWords is a selector's pieces as one name: application:didUpdateUserActivity: is
// applicationDidUpdateUserActivity.
func selectorWords(selector string) string {
	name := ""
	for _, piece := range strings.Split(selector, ":") {
		if piece == "" {
			continue
		}
		if name != "" {
			piece = strings.ToUpper(piece[:1]) + piece[1:]
		}
		name += piece
	}
	return name
}

// implementable is why a program's class can't implement a method with these types yet, or "".
func implementable(selector string, natives []nativeType, result nativeType) string {
	for _, native := range natives {
		switch strings.TrimSuffix(native.tag, "?") {
		case "object", "string", "boolean", "double", "integer", "unsigned":
		default:
			if native.hidden {
				return "a target and its action can't be handed to a program's class yet"
			}
			return "a program's class can't be handed a " + native.tag + " yet"
		}
	}
	switch strings.TrimSuffix(result.tag, "?") {
	case "void", "object", "string", "boolean", "double", "integer", "unsigned":
	default:
		return "a program's class can't give back a " + result.tag + " yet"
	}
	if result.c == "id" && retainedFamily(selector) {
		return "a result its caller owns can't be given back by a program's class yet"
	}
	return ""
}

// parameterName is an Objective-C parameter's name as an Adamic parameter: the same word, unless
// it's one JavaScript reserves.
func parameterName(name string, used map[string]bool) string {
	switch name {
	case "", "break", "case", "catch", "class", "const", "continue", "debugger", "default", "delete", "do", "else", "enum", "export", "extends", "false", "finally", "for", "function", "if", "import", "in", "instanceof", "new", "null", "return", "super", "switch", "this", "throw", "true", "try", "typeof", "var", "void", "while", "with", "yield", "let", "static", "implements", "interface", "package", "private", "protected", "public", "await":
		name += "Value"
	}
	for base, index := name, 2; used[name]; index++ {
		name = fmt.Sprintf("%s%d", base, index)
	}
	used[name] = true
	return name
}

// emitProtocolMethod writes one of a protocol's methods, named by protocolNames.
func (g *generator) emitProtocolMethod(owner *definition, n *node, name string) (string, error) {
	m := g.getModule(owner.output.Module)
	original := methodOriginal(owner.node.Name, n.Name, n.Instance)
	skip := func(reason string) (string, error) {
		m.comments = append(m.comments, "// Skipped "+original+": "+reason+".")
		return "", nil
	}
	optional, err := g.protocolOptional(owner.node, n)
	if err != nil {
		return "", err
	}
	if !n.Instance {
		return skip("protocol class messages need a concrete class")
	}
	d, out, natives, result, reason, err := g.callable(n, naming.InstanceMethod, owner.node.Name, owner.declaration.SwiftName, g.propertyNames(owner, map[string]bool{}), nil)
	if err != nil {
		return "", err
	}
	if reason != "" {
		return skip(reason)
	}
	if out.Arguments.Constructor {
		return skip("protocol construction needs a concrete class")
	}
	if name == "" {
		return skip("another of the protocol's methods has its name")
	}
	cannot := implementable(n.Name, natives, result)
	if optional && cannot != "" {
		return skip(cannot)
	}
	for _, native := range natives {
		if native.hidden {
			return skip("a target and its action can't be handed to a protocol's method yet")
		}
	}
	g.declarations = append(g.declarations, d)
	for _, native := range append(append([]nativeType{}, natives...), result) {
		for _, reference := range native.references() {
			g.addImport(m, owner.output.Module, reference)
		}
	}
	parameters := children(n, "ParmVarDecl")
	written, sources := []string{}, []string{}
	used := map[string]bool{}
	for index, native := range natives {
		written = append(written, parameterName(parameters[index].Name, used)+": "+native.adamic)
		sources = append(sources, fmt.Sprintf(" %d:%s", index, native.tag))
	}
	tags := []string{}
	if !optional {
		returns := " -> " + result.tag
		if result.c == "id" && (has(n, "NSReturnsRetainedAttr") || retainedFamily(n.Name) && !has(n, "NSReturnsNotRetainedAttr")) {
			returns = " -> new " + result.tag
		}
		tags = append(tags, "method "+n.Name+strings.Join(sources, "")+returns)
		g.checks = append(g.checks, checkCall{owner: owner, selector: n.Name, instance: true, parameters: natives, result: result})
	}
	if cannot == "" {
		tags = append(tags, "implement "+n.Name+strings.Join(sources, "")+" -> "+result.tag)
		g.implements = append(g.implements, implementCheck{owner: owner, selector: n.Name, parameters: natives, result: result})
	}
	mark := ""
	if optional {
		mark = "?"
	}
	return doc(original, tags, "\t\t") + "\t\t" + name + mark + "(" + strings.Join(written, ", ") + "): " + result.adamic + ";\n", nil
}

// emitImplementChecks writes, for each protocol with an implement tag, a class conforming to it
// that implements each such method in the bridge's C types: clang refuses a type that conflicts
// with the header's (-Wmismatched-return-types, -Wmismatched-parameter-types), so a delegate's
// method is called with what its tag says or the bindings aren't served. Methods a check class
// leaves out are the protocol's business, not the bridge's (-Wprotocol is off around them, and
// -Wobjc-protocol-property-synthesis and -Wobjc-property-implementation for the properties it
// requires).
func (g *generator) emitImplementChecks(text *strings.Builder) {
	byProtocol := map[string][]implementCheck{}
	for _, check := range g.implements {
		byProtocol[check.owner.node.Name] = append(byProtocol[check.owner.node.Name], check)
	}
	protocols := sortedKeys(byProtocol)
	if len(protocols) == 0 {
		return
	}
	text.WriteString("\n#pragma clang diagnostic push\n#pragma clang diagnostic ignored \"-Wprotocol\"\n#pragma clang diagnostic ignored \"-Wobjc-protocol-property-synthesis\"\n#pragma clang diagnostic ignored \"-Wobjc-property-implementation\"\n")
	for index, protocol := range protocols {
		checks := byProtocol[protocol]
		sort.Slice(checks, func(i, j int) bool { return checks[i].selector < checks[j].selector })
		methods := strings.Builder{}
		// A root class of its own: what it inherits is nobody's business here.
		fmt.Fprintf(&methods, "__attribute__((objc_root_class))\n@interface AdamicImplementCheck%d <%s>\n@end\n@implementation AdamicImplementCheck%d\n", index, protocol, index)
		for _, check := range checks {
			pieces := strings.Split(check.selector, ":")
			declaration := check.selector
			if len(check.parameters) > 0 {
				parts := []string{}
				for position, native := range check.parameters {
					parts = append(parts, fmt.Sprintf("%s:(%s)argument%d", pieces[position], implementType(text, check.selector, native), position))
				}
				declaration = strings.Join(parts, " ")
			}
			result := implementType(text, check.selector, check.result)
			body := " return 0; "
			if result == "void" {
				body = " "
			}
			fmt.Fprintf(&methods, "- (%s)%s {%s}\n", result, declaration, body)
		}
		text.WriteString(methods.String())
		text.WriteString("@end\n")
	}
	text.WriteString("#pragma clang diagnostic pop\n")
}

// implementType is the C type a check class's method writes for one of an implement tag's types:
// the bridge's, except that a 64-bit integer is spelled as the header spells it (long long, int64_t)
// once an assertion written before the class holds it to the bridge's long, the same 64 bits, which
// is all a method's conflicting-type check can't be told.
func implementType(text *strings.Builder, selector string, native nativeType) string {
	c := tagCType(strings.TrimSuffix(native.tag, "?"))
	if (c == "long" || c == "unsigned long") && native.header != c {
		fmt.Fprintf(text, "_Static_assert(%s, \"%s implemented ABI\");\n", compatible(c, native.header), selector)
		return native.header
	}
	return c
}
