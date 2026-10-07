package generate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/system-inc/adamic/internal/apple/naming"
)

func (g *generator) build() (Output, error) {
	if err := g.inventory(); err != nil {
		return Output{}, err
	}
	for _, name := range sortedKeys(g.types) {
		def := g.types[name]
		m := g.getModule(def.output.Module)
		switch def.declaration.Kind {
		case naming.Enum, naming.OptionSet:
			if len(def.members) > 0 {
				g.declarations = append(g.declarations, def.declaration)
				m.declarations = append(m.declarations, fmt.Sprintf("\t/** %s */\n\texport type %s = %s;\n", name, def.output.Name, strings.Join(def.members, " | ")))
			}
		case naming.Struct:
			if name == "CGRect" {
				g.declarations = append(g.declarations, def.declaration)
				m.declarations = append(m.declarations, "\t/** CGRect */\n\texport interface "+def.output.Name+" {\n\t\treadonly x: number;\n\t\treadonly y: number;\n\t\treadonly width: number;\n\t\treadonly height: number;\n\t}\n")
			} else {
				m.comments = append(m.comments, "// Skipped "+name+": only CGRect structs cross the bridge.")
			}
		case naming.Class, naming.Protocol:
			g.declarations = append(g.declarations, def.declaration)
			body, err := g.emitClass(def)
			if err != nil {
				return Output{}, err
			}
			m.declarations = append(m.declarations, body)
		}
	}
	for _, n := range g.nodes {
		switch n.Kind {
		case "ObjCCategoryDecl":
			// Category methods are emitted while visiting their owning class.
			if g.types[n.Interface.Name] == nil {
				g.skipped(n, "category owner is not bound")
			}
		case "FunctionDecl":
			if err := g.emitFunction(n); err != nil {
				return Output{}, err
			}
		case "VarDecl":
			g.skipped(n, "global constants have no native load tag yet")
		case "TypedefDecl":
			q := cleanType(n.Type.Qual)
			if g.types[n.Name] == nil && q != "double" && q != "long" && q != "unsigned long" && q != "signed char" {
				g.skipped(n, "typedef alias has no independent bridge representation")
			}
		}
	}
	if g.importError != nil {
		return Output{}, g.importError
	}
	if _, err := naming.Build(g.declarations); err != nil {
		return Output{}, fmt.Errorf("binding surface: %w", err)
	}
	output := Output{}
	for _, path := range sortedKeys(g.modules) {
		m := g.modules[path]
		var content strings.Builder
		content.WriteString("// Generated from Objective-C declaration facts. Do not edit.\n")
		sort.Strings(m.comments)
		for _, comment := range m.comments {
			content.WriteString(comment + "\n")
		}
		content.WriteString("\ndeclare module '" + path + "' {\n")
		for _, name := range sortedKeys(m.imports) {
			content.WriteString("\timport type { " + name + " } from '" + m.imports[name] + "';\n")
		}
		if len(m.imports) > 0 {
			content.WriteByte('\n')
		}
		sort.Strings(m.declarations)
		for i, declaration := range m.declarations {
			if i > 0 {
				content.WriteByte('\n')
			}
			content.WriteString(declaration)
		}
		content.WriteString("}\n")
		output.Files = append(output.Files, File{Path: strings.TrimPrefix(path, "apple/") + ".d.ts", Content: []byte(content.String())})
	}
	check, err := g.emitCheck()
	if err != nil {
		return Output{}, err
	}
	output.Check = []byte(check)
	return output, nil
}
func doc(original string, tags []string, indent string) string {
	text := indent + "/**\n" + indent + " * " + original + "\n"
	for _, tag := range tags {
		text += indent + " * @objc " + tag + "\n"
	}
	return text + indent + " */\n"
}
func (g *generator) propertyNames(def *definition, seen map[string]bool) []string {
	if seen[def.node.Name] {
		return nil
	}
	seen[def.node.Name] = true
	var names []string
	for _, child := range children(def.node, "ObjCPropertyDecl") {
		names = append(names, child.Name)
	}
	if parent := g.types[def.node.Super.Name]; parent != nil {
		names = append(names, g.propertyNames(parent, seen)...)
	}
	return names
}
func (g *generator) emitClass(def *definition) (string, error) {
	n := def.node
	m := g.getModule(def.output.Module)
	protocol := def.declaration.Kind == naming.Protocol
	nodes := append([]*node{}, n.Children...)
	for _, category := range g.nodes {
		if !protocol && category.Kind == "ObjCCategoryDecl" && category.Interface.Name == n.Name {
			_, _, unavailable, err := g.attributes(category)
			if err != nil {
				return "", err
			}
			if unavailable {
				m.comments = append(m.comments, "// Skipped category "+category.Name+": unavailable on "+g.configuration.Platform+".")
				continue
			}
			for _, member := range category.Children {
				member.Framework = n.Framework
				nodes = append(nodes, member)
			}
		}
	}
	for _, child := range nodes {
		child.Framework = n.Framework
	}
	// Clang synthesizes implicit property accessors. Explicit getter/setter pairs
	// without @property are also combined, with both selectors in the tag.
	properties := children(n, "ObjCPropertyDecl")
	for _, child := range nodes {
		if child.Kind != "ObjCPropertyDecl" || containsNode(properties, child) {
			continue
		}
		// A category may declare a property the class already has (NSSlider's vertical, readwrite in
		// the class, readonly in NSSliderVerticalGetter): the class's declaration stands unless the
		// category's widens it to readwrite.
		redeclared := false
		for i, existing := range properties {
			if existing.Name == child.Name && existing.ClassProperty == child.ClassProperty {
				redeclared = true
				if existing.Readonly && !child.Readonly {
					properties[i] = child
				}
			}
		}
		if !redeclared {
			properties = append(properties, child)
		}
	}
	accessors := map[string]bool{}
	for _, p := range properties {
		getter, setter := propertySelectors(p)
		accessors[getter] = true
		if setter != "" {
			accessors[setter] = true
		}
	}
	for _, getter := range nodes {
		if getter.Kind != "ObjCMethodDecl" || getter.Implicit || len(children(getter, "ParmVarDecl")) != 0 || getter.Result.Qual == "void" || accessors[getter.Name] || getter.Name == "" {
			continue
		}
		setterName := "set" + strings.ToUpper(getter.Name[:1]) + getter.Name[1:] + ":"
		for _, setter := range nodes {
			parameters := children(setter, "ParmVarDecl")
			if setter.Kind == "ObjCMethodDecl" && !setter.Implicit && setter.Name == setterName && setter.Instance == getter.Instance && setter.Result.Qual == "void" && len(parameters) == 1 && cleanType(parameters[0].Type.Qual) == cleanType(getter.Result.Qual) {
				if protocol {
					getterOptional, err := g.protocolOptional(n, getter)
					if err != nil {
						return "", err
					}
					setterOptional, err := g.protocolOptional(n, setter)
					if err != nil {
						return "", err
					}
					if getterOptional || setterOptional {
						continue
					}
				}
				property := &node{Begin: getter.Begin, End: getter.End, Location: getter.Location, Kind: "ObjCPropertyDecl", Name: getter.Name, Framework: n.Framework, Type: getter.Result, Getter: reference{Name: getter.Name}, Setter: reference{Name: setter.Name}, ClassProperty: !getter.Instance, Children: append(append([]*node{}, getter.Children...), setter.Children...)}
				properties = append(properties, property)
				accessors[getter.Name] = true
				accessors[setter.Name] = true
				break
			}
		}
	}
	// Swift's rule: a factory method imported as an initializer gives way to a real initializer
	// of the same shape (+[NSAffineTransform transform] beside -init).
	initializers := map[string]bool{}
	for _, child := range nodes {
		if child.Kind == "ObjCMethodDecl" && child.Instance && !child.Implicit && !accessors[child.Name] {
			_, out, _, _, reason, err := g.callable(child, naming.InstanceMethod, n.Name, def.declaration.SwiftName, nil, nil)
			if err == nil && reason == "" && out.Arguments.Constructor {
				initializers[naming.Shape(out.Arguments)] = true
			}
		}
	}
	members := []string{}
	constructor := false
	// values holds the Adamic names of the properties, static and instance apart, so a method that
	// would share one's name can give way.
	values := map[bool][]string{}
	for _, p := range properties {
		member, name, err := g.emitProperty(def, p, protocol)
		if err != nil {
			return "", err
		}
		if member != "" {
			members = append(members, member)
			values[p.ClassProperty] = append(values[p.ClassProperty], name)
		}
	}
	for _, child := range nodes {
		if child.Kind != "ObjCMethodDecl" || child.Implicit || accessors[child.Name] {
			continue
		}
		member, isConstructor, err := g.emitMethod(def, child, protocol, initializers, values[!child.Instance])
		if err != nil {
			return "", err
		}
		if member != "" {
			members = append(members, member)
			constructor = constructor || isConstructor
		}
	}
	sort.Strings(members)
	var text strings.Builder
	if protocol {
		text.WriteString("\t/** " + n.Name + " */\n\texport interface " + def.output.Name + " {\n")
	} else {
		text.WriteString(doc(n.Name, []string{"class " + n.Name}, "\t"))
		text.WriteString("\texport class " + def.output.Name)
		if parent := g.types[n.Super.Name]; parent != nil {
			g.addImport(m, def.output.Module, parent)
			text.WriteString(" extends " + parent.output.Name)
		}
		text.WriteString(" {\n")
		if !constructor {
			visibility := "private"
			for _, candidate := range g.types {
				if candidate.node.Super.Name == n.Name {
					visibility = "protected"
				}
			}
			text.WriteString("\t\t" + visibility + " constructor();\n")
		}
	}
	for _, member := range members {
		text.WriteByte('\n')
		text.WriteString(member)
	}
	text.WriteString("\t}\n")
	return text.String(), nil
}
func containsNode(nodes []*node, n *node) bool {
	for _, other := range nodes {
		if n == other {
			return true
		}
	}
	return false
}
func propertySelectors(n *node) (string, string) {
	getter := n.Getter.Name
	if getter == "" {
		getter = n.Name
	}
	setter := n.Setter.Name
	if setter == "" && !n.Readonly {
		setter = "set" + strings.ToUpper(n.Name[:1]) + n.Name[1:] + ":"
	}
	return getter, setter
}
func (g *generator) emitProperty(owner *definition, n *node, protocol bool) (string, string, error) {
	d, unavailable, err := g.description(n, naming.Property, owner.node.Name)
	d.ParentSwiftName = owner.declaration.SwiftName
	d.ClassProperty = n.ClassProperty
	if err != nil {
		return "", "", err
	}
	m := g.getModule(owner.output.Module)
	skip := func(reason string) {
		m.comments = append(m.comments, "// Skipped "+owner.node.Name+"."+n.Name+": "+reason+".")
	}
	if unavailable {
		skip("unavailable on " + g.configuration.Platform)
		return "", "", nil
	}
	if protocol {
		optional, err := g.protocolOptional(owner.node, n)
		if err != nil {
			return "", "", err
		}
		if optional {
			skip("optional protocol properties are not proven present")
			return "", "", nil
		}
	}
	native, err := g.native(n.Type, owner.node.Name, true)
	if err != nil {
		skip(err.Error())
		return "", "", nil
	}
	getter, setter := propertySelectors(n)
	d.Getter = getter
	d.Setter = setter
	d.Result = naming.Type{Spelling: namingSpelling(n.Type, native), Object: native.c == "id", Framework: referenceFramework(native, n.Framework), Nullability: naming.Nonnull}
	out, err := naming.Name(d)
	if err != nil {
		// One member the naming layer can't name yet is left out with its reason; a collision
		// between names that were given stays fatal in naming.Build.
		skip("no Adamic name yet: " + err.Error())
		return "", "", nil
	}
	g.declarations = append(g.declarations, d)
	g.addImport(m, owner.output.Module, native.reference)
	tags := []string{"get " + getter + " -> " + native.tag}
	prefix := ""
	if n.ClassProperty {
		prefix = "static "
		if protocol {
			skip("class protocol properties have no receiver class")
			return "", "", nil
		}
	}
	readonly := ""
	if setter == "" {
		readonly = "readonly "
	} else {
		tags = append(tags, "set "+setter+" "+native.tag)
	}
	original := methodOriginal(owner.node.Name, getter, !n.ClassProperty)
	g.checks = append(g.checks, checkCall{owner: owner, selector: getter, instance: !n.ClassProperty, result: native})
	if setter != "" {
		original += ", " + methodOriginal(owner.node.Name, setter, !n.ClassProperty)
		g.checks = append(g.checks, checkCall{owner: owner, selector: setter, instance: !n.ClassProperty, parameters: []nativeType{native}, result: nativeType{tag: "void", c: "void", header: "void"}})
	}
	return doc(original, tags, "\t\t") + "\t\t" + prefix + readonly + out.Name + ": " + native.adamic + ";\n", out.Name, nil
}
func methodOriginal(parent, selector string, instance bool) string {
	prefix := "+"
	if instance {
		prefix = "-"
	}
	return prefix + "[" + parent + " " + selector + "]"
}

// parentSwiftName is the owner's explicit imported name, which tells a protocol's members from
// those of a class sharing its Objective-C name.
func (g *generator) callable(n *node, kind naming.Kind, parent, parentSwiftName string, properties, values []string) (naming.Declaration, naming.Output, []nativeType, nativeType, string, error) {
	d, unavailable, err := g.description(n, kind, parent)
	d.ParentSwiftName = parentSwiftName
	if err != nil {
		return d, naming.Output{}, nil, nativeType{}, "", err
	}
	if unavailable {
		return d, naming.Output{}, nil, nativeType{}, "unavailable on " + g.configuration.Platform, nil
	}
	if kind != naming.CFunction && (n.Name == "retain" || n.Name == "release" || n.Name == "autorelease" || n.Name == "dealloc") {
		return d, naming.Output{}, nil, nativeType{}, "manual reference-count messages bypass Adamic ownership", nil
	}
	if n.Variadic {
		return d, naming.Output{}, nil, nativeType{}, "variadic declarations are not carried by the bridge", nil
	}
	var natives []nativeType
	for _, p := range children(n, "ParmVarDecl") {
		if has(p, "NSConsumedAttr") || has(p, "CFConsumedAttr") {
			return d, naming.Output{}, nil, nativeType{}, "consumed parameters cannot use the bridge's borrowed arguments", nil
		}
		source, sourceError := g.sourceRange(p)
		if sourceError != nil {
			return d, naming.Output{}, nil, nativeType{}, "", sourceError
		}
		if strings.Contains(source, "[") {
			return d, naming.Output{}, nil, nativeType{}, "C arrays are not carried by the bridge", nil
		}
		native, e := g.native(p.Type, parent, false)
		if e != nil {
			return d, naming.Output{}, nil, nativeType{}, e.Error(), nil
		}
		natives = append(natives, native)
		d.Parameters = append(d.Parameters, naming.Parameter{Name: p.Name, Type: naming.Type{Spelling: namingSpelling(p.Type, native), Object: native.c == "id", Nullability: naming.Nonnull, Framework: referenceFramework(native, n.Framework)}})
	}
	result, e := g.native(n.Result, parent, true)
	if e != nil {
		return d, naming.Output{}, nil, nativeType{}, e.Error(), nil
	}
	d.Result = naming.Type{Spelling: namingSpelling(n.Result, result), Object: result.c == "id", Framework: referenceFramework(result, n.Framework), Nullability: naming.Nonnull}
	d.PropertyNames = properties
	d.ValueNames = values
	out, e := naming.Name(d)
	if e != nil {
		return d, out, nil, result, "no Adamic name yet: " + e.Error(), nil
	}
	return d, out, natives, result, "", nil
}
func signature(out naming.Output, natives []nativeType) (string, string) {
	parameters := []string{}
	sources := make([]string, len(natives))
	for i, arg := range out.Arguments.Positional {
		parameters = append(parameters, arg.Name+": "+natives[arg.Index].adamic)
		sources[arg.Index] = fmt.Sprintf("%d:%s", i, natives[arg.Index].tag)
	}
	if len(out.Arguments.Options) > 0 {
		fields := []string{}
		position := len(parameters)
		for _, arg := range out.Arguments.Options {
			fields = append(fields, "readonly "+arg.Name+": "+natives[arg.Index].adamic)
			sources[arg.Index] = fmt.Sprintf("%d.%s:%s", position, arg.Name, natives[arg.Index].tag)
		}
		optionsName := "options"
		for _, arg := range out.Arguments.Positional {
			if arg.Name == optionsName {
				optionsName = "argumentOptions"
			}
		}
		parameters = append(parameters, optionsName+": { "+strings.Join(fields, "; ")+" }")
	}
	suffix := ""
	if len(sources) > 0 {
		suffix = " " + strings.Join(sources, " ")
	}
	return strings.Join(parameters, ", "), suffix
}
func (g *generator) emitMethod(owner *definition, n *node, protocol bool, initializers map[string]bool, values []string) (string, bool, error) {
	if protocol {
		optional, err := g.protocolOptional(owner.node, n)
		if err != nil {
			return "", false, err
		}
		if optional {
			m := g.getModule(owner.output.Module)
			m.comments = append(m.comments, "// Skipped "+n.Name+": optional protocol methods are not proven present.")
			return "", false, nil
		}
	}
	kind := naming.InstanceMethod
	if !n.Instance {
		kind = naming.ClassMethod
	}
	d, out, natives, result, reason, err := g.callable(n, kind, owner.node.Name, owner.declaration.SwiftName, g.propertyNames(owner, map[string]bool{}), values)
	if err != nil {
		return "", false, err
	}
	m := g.getModule(owner.output.Module)
	if reason != "" {
		m.comments = append(m.comments, "// Skipped "+methodOriginal(owner.node.Name, n.Name, n.Instance)+": "+reason+".")
		return "", false, nil
	}
	constructor := out.Arguments.Constructor
	if constructor && !n.Instance && initializers[naming.Shape(out.Arguments)] {
		m.comments = append(m.comments, "// Skipped "+methodOriginal(owner.node.Name, n.Name, false)+": an initializer of the same shape is the constructor.")
		return "", false, nil
	}
	if has(n, "NSConsumesSelfAttr") && !constructor {
		m.comments = append(m.comments, "// Skipped "+n.Name+": a consumed receiver cannot be passed borrowed.")
		return "", false, nil
	}
	if protocol && (!n.Instance || constructor) {
		m.comments = append(m.comments, "// Skipped "+n.Name+": protocol construction and class messages need a concrete class.")
		return "", false, nil
	}
	g.declarations = append(g.declarations, d)
	for _, native := range append(append([]nativeType{}, natives...), result) {
		g.addImport(m, owner.output.Module, native.reference)
	}
	parameters, sources := signature(out, natives)
	tagKind := "method"
	prefix := ""
	if !n.Instance {
		tagKind = "static"
		prefix = "static "
	}
	returns := " -> " + result.tag
	if constructor && n.Instance {
		tagKind = "init"
		returns = ""
	}
	if !constructor && result.c == "id" && (has(n, "NSReturnsRetainedAttr") || retainedFamily(n.Name) && !has(n, "NSReturnsNotRetainedAttr")) {
		returns = " -> new " + result.tag
	}
	tag := tagKind + " " + n.Name + sources + returns
	g.checks = append(g.checks, checkCall{owner: owner, selector: n.Name, instance: n.Instance, parameters: natives, result: result})
	text := doc(methodOriginal(owner.node.Name, n.Name, n.Instance), []string{tag}, "\t\t")
	if constructor {
		text += "\t\tconstructor(" + parameters + ");\n"
	} else if protocol {
		text += "\t\treadonly " + out.Name + ": (" + parameters + ") => " + result.adamic + ";\n"
	} else {
		text += "\t\t" + prefix + out.Name + "(" + parameters + "): " + result.adamic + ";\n"
	}
	return text, constructor, nil
}
func retainedFamily(selector string) bool {
	for _, family := range []string{"alloc", "new", "copy", "mutableCopy"} {
		if strings.HasPrefix(selector, family) && (len(selector) == len(family) || selector[len(family)] < 'a' || selector[len(family)] > 'z') {
			return true
		}
	}
	return false
}
func (g *generator) emitFunction(n *node) error {
	q := n.Type.Qual
	left := strings.IndexByte(q, '(')
	if left < 0 {
		g.skipped(n, "cannot recover C function result")
		return nil
	}
	n.Result = writtenType{Qual: strings.TrimSpace(q[:left])}
	d, out, natives, result, reason, err := g.callable(n, naming.CFunction, "", "", nil, nil)
	if err != nil {
		return err
	}
	if reason != "" {
		g.skipped(n, reason)
		return nil
	}
	g.declarations = append(g.declarations, d)
	m := g.getModule(out.Module)
	for _, native := range append(append([]nativeType{}, natives...), result) {
		g.addImport(m, out.Module, native.reference)
	}
	parameters, sources := signature(out, natives)
	returns := " -> " + result.tag
	if result.c == "id" && has(n, "NSReturnsRetainedAttr") {
		returns = " -> new " + result.tag
	}
	m.declarations = append(m.declarations, doc(n.Name, []string{"function " + n.Name + sources + returns}, "\t")+"\texport function "+out.Name+"("+parameters+"): "+result.adamic+";\n")
	g.checks = append(g.checks, checkCall{selector: n.Name, parameters: natives, result: result, function: true})
	return nil
}

// namingSpelling is the written type the naming layer reads for omission, except that references
// to protocols and rectangle aliases use their established concrete spelling.
func namingSpelling(written writtenType, native nativeType) string {
	spelling := cleanType(written.Qual)
	if strings.HasPrefix(spelling, "id<") && native.reference != nil {
		spelling = native.reference.node.Name + " *"
	}
	if native.tag == "rectangle" {
		spelling = "CGRect"
	}
	if native.reference != nil && (native.reference.declaration.Kind == naming.Enum || native.reference.options) {
		spelling = native.reference.node.Name
	}
	return spelling
}

func referenceFramework(native nativeType, fallback string) string {
	if native.reference != nil && native.reference.node.Framework != "" {
		return native.reference.node.Framework
	}
	return fallback
}
