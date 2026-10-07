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
			if err := g.emitDefinition(def); err != nil {
				return Output{}, err
			}
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
	output.Leaves = g.leaves()
	return output, nil
}

// emitDefinition writes a class or protocol once, its superclass first.
func (g *generator) emitDefinition(def *definition) error {
	if def.emitted {
		return nil
	}
	def.emitted = true
	if parent := g.types[def.node.Super.Name]; parent != nil && def.declaration.Kind == naming.Class {
		if err := g.emitDefinition(parent); err != nil {
			return err
		}
	}
	g.declarations = append(g.declarations, def.declaration)
	body, err := g.emitClass(def)
	if err != nil {
		return err
	}
	m := g.getModule(def.output.Module)
	m.declarations = append(m.declarations, body)
	return nil
}

// taken is every name a method of this staticness may not take: properties, its own or inherited,
// and inherited methods none of which it overrides.
func taken(def *definition, static bool, selector string) []string {
	names := []string{}
	for name, members := range def.surface[static] {
		overrides := false
		for _, existing := range members {
			overrides = overrides || !existing.property && existing.selector == selector
		}
		if !overrides {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
func contains(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}
	return false
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
			// A category on the root class is an informal protocol (NSURLClient's
			// url(_:resourceDataDidBecomeAvailable:)): methods a delegate may implement, declared
			// on NSObject for the compiler, not methods NSObject has.
			if n.Name == "NSObject" {
				m.comments = append(m.comments, "// Skipped category "+category.Name+": an informal protocol on NSObject, not its own methods.")
				continue
			}
			_, _, gone, err := g.attributes(category)
			if err != nil {
				return "", err
			}
			if gone != "" {
				m.comments = append(m.comments, "// Skipped category "+category.Name+": "+gone+".")
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
		if child.Kind == "ObjCMethodDecl" && child.Instance && !child.Implicit && !accessors[child.Name] && !child.SetterOnly {
			_, out, _, _, reason, err := g.callable(child, naming.InstanceMethod, n.Name, def.declaration.SwiftName, nil, nil)
			if err == nil && reason == "" && out.Arguments.Constructor {
				initializers[naming.Shape(out.Arguments)] = true
			}
		}
	}
	// A property whose getter's result can't cross but whose setter can (NSView's frame, a
	// rectangle) is offered as its setter, setFrame(frame), Objective-C's own spelling.
	if !protocol {
		for _, p := range properties {
			_, setter := propertySelectors(p)
			if _, err := g.native(p.Type, n.Name, true); err == nil || setter == "" {
				continue
			}
			if _, err := g.native(p.Type, n.Name, false); err != nil {
				continue
			}
			attributes := []*node{}
			for _, child := range p.Children {
				if strings.HasSuffix(child.Kind, "Attr") {
					attributes = append(attributes, child)
				}
			}
			parameter := &node{Kind: "ParmVarDecl", Name: p.Name, Type: p.Type, Begin: p.Begin, End: p.End, Location: p.Location, Framework: n.Framework}
			nodes = append(nodes, &node{Kind: "ObjCMethodDecl", Name: setter, Instance: !p.ClassProperty, Begin: p.Begin, End: p.End, Location: p.Location, Framework: n.Framework, Result: writtenType{Qual: "void"}, Children: append(attributes, parameter), SetterOnly: true})
		}
	}
	members := []string{}
	constructor := false
	// A class sees its ancestors' members, already written: their properties and methods are names
	// its own give way to or reconcile with.
	inherited := map[bool]map[string][]member{false: {}, true: {}}
	def.surface = map[bool]map[string][]member{false: {}, true: {}}
	if parent := g.types[n.Super.Name]; parent != nil && !protocol && parent.emitted {
		for _, static := range []bool{false, true} {
			for name, entries := range parent.surface[static] {
				inherited[static][name] = entries
				def.surface[static][name] = entries
			}
		}
	}
	for _, p := range properties {
		text, entry, name, err := g.emitProperty(def, p, protocol, inherited[p.ClassProperty])
		if err != nil {
			return "", err
		}
		if text != "" {
			members = append(members, text)
			def.surface[p.ClassProperty][name] = []member{entry}
		}
	}
	// Swift's conflict rule: methods whose shortened names collide keep their words, each of them
	// whose name omission changed (addObject: beside add:, both add(_:) once Object is dropped).
	shortened := map[string][]*node{}
	own := map[bool]map[string][]member{false: {}, true: {}}
	for _, child := range nodes {
		if child.Kind != "ObjCMethodDecl" || child.Implicit || accessors[child.Name] && !child.SetterOnly {
			continue
		}
		key, ok := g.methodKey(def, child, taken(def, !child.Instance, child.Name))
		if ok {
			shortened[key] = append(shortened[key], child)
		}
	}
	for _, key := range sortedKeys(shortened) {
		if len(shortened[key]) < 2 {
			continue
		}
		for _, child := range shortened[key] {
			child.KeepNeedlessWords = true
			if kept, _ := g.methodKey(def, child, taken(def, !child.Instance, child.Name)); kept == key {
				child.KeepNeedlessWords = false
			}
		}
	}
	var implemented map[*node]string
	if protocol {
		implemented = g.protocolNames(def, nodes, accessors)
	}
	for _, child := range nodes {
		if child.Kind != "ObjCMethodDecl" || child.Implicit || accessors[child.Name] && !child.SetterOnly {
			continue
		}
		if protocol {
			text, err := g.emitProtocolMethod(def, child, implemented[child])
			if err != nil {
				return "", err
			}
			if text != "" {
				members = append(members, text)
			}
			continue
		}
		text, entry, name, isConstructor, err := g.emitMethod(def, child, initializers, taken(def, !child.Instance, child.Name), inherited[!child.Instance])
		if err != nil {
			return "", err
		}
		if text != "" {
			members = append(members, text)
			constructor = constructor || isConstructor
			if !isConstructor {
				own[!child.Instance][name] = append(own[!child.Instance][name], entry)
			}
		}
	}
	// A method name the class shares with an ancestor's is one overload set: the class re-declares
	// the inherited overloads it doesn't override (NSStackView's remove(view) beside NSView's
	// four), which send the ancestor's selector, as the class inherits them.
	for _, static := range []bool{false, true} {
		for _, name := range sortedKeys(own[static]) {
			entries := own[static][name]
			for _, existing := range inherited[static][name] {
				overridden := existing.property
				for _, entry := range entries {
					overridden = overridden || entry.selector == existing.selector
				}
				if !overridden {
					members = append(members, existing.text)
					for _, reference := range existing.references {
						g.addImport(m, def.output.Module, reference)
					}
					entries = append(entries, existing)
				}
			}
			def.surface[static][name] = entries
		}
	}
	sortMembers(members)
	var text strings.Builder
	if protocol {
		text.WriteString(doc(n.Name, []string{"protocol " + n.Name}, "\t"))
		text.WriteString("\texport interface " + def.output.Name + " {\n")
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

// sortMembers orders a class's members by name, and a name's overloads by most fields first:
// tsc types a closure's parameters from the first overload it tries, so the one that takes the
// closure (dataTask({ with, completionHandler })) comes before the one that doesn't.
func sortMembers(members []string) {
	key := func(member string) (string, int) {
		lines := strings.Split(strings.TrimRight(member, "\n"), "\n")
		declaration := strings.TrimSpace(lines[len(lines)-1])
		for _, prefix := range []string{"static ", "readonly "} {
			declaration = strings.TrimPrefix(declaration, prefix)
		}
		name := declaration
		if end := strings.IndexAny(declaration, "(:"); end >= 0 {
			name = declaration[:end]
		}
		return name, strings.Count(declaration, ": ")
	}
	sort.SliceStable(members, func(i, j int) bool {
		left, leftFields := key(members[i])
		right, rightFields := key(members[j])
		if left != right {
			return left < right
		}
		if leftFields != rightFields {
			return leftFields > rightFields
		}
		return members[i] < members[j]
	})
}

// methodKey is a method's slot and shape as it would be emitted, static and instance apart.
func (g *generator) methodKey(owner *definition, n *node, values []string) (string, bool) {
	kind := naming.InstanceMethod
	if !n.Instance {
		kind = naming.ClassMethod
	}
	_, out, _, _, reason, err := g.callable(n, kind, owner.node.Name, owner.declaration.SwiftName, g.propertyNames(owner, map[string]bool{}), values)
	if err != nil || reason != "" {
		return "", false
	}
	return fmt.Sprintf("%t %s%s", n.Instance, out.Name, naming.Shape(out.Arguments)), true
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
func (g *generator) emitProperty(owner *definition, n *node, protocol bool, inherited map[string][]member) (string, member, string, error) {
	d, gone, err := g.description(n, naming.Property, owner.node.Name)
	d.ParentSwiftName = owner.declaration.SwiftName
	d.ClassProperty = n.ClassProperty
	if err != nil {
		return "", member{}, "", err
	}
	m := g.getModule(owner.output.Module)
	skip := func(reason string) {
		m.comments = append(m.comments, "// Skipped "+owner.node.Name+"."+n.Name+": "+reason+".")
	}
	if gone != "" {
		skip(gone)
		return "", member{}, "", nil
	}
	if protocol {
		optional, err := g.protocolOptional(owner.node, n)
		if err != nil {
			return "", member{}, "", err
		}
		if optional {
			skip("optional protocol properties are not proven present")
			return "", member{}, "", nil
		}
	}
	native, err := g.native(n.Type, owner.node.Name, true)
	if err != nil {
		skip(err.Error())
		return "", member{}, "", nil
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
		return "", member{}, "", nil
	}
	// A property an ancestor already has stands as the ancestor declared it when this one disagrees:
	// NSMatrix's selectedCell, a method in NSControl, or NSSavePanel's title, nullable where
	// NSWindow's isn't. The ancestor's declaration sends the same getter.
	for _, existing := range inherited[out.Name] {
		if !existing.property {
			skip("an ancestor declares " + out.Name + " as a method, which stands")
			return "", member{}, "", nil
		}
		if existing.adamic != native.adamic {
			skip("an ancestor declares " + out.Name + " as " + existing.adamic + ", which stands")
			return "", member{}, "", nil
		}
	}
	g.declarations = append(g.declarations, d)
	for _, reference := range native.references() {
		g.addImport(m, owner.output.Module, reference)
	}
	tags := []string{"get " + getter + " -> " + native.tag}
	prefix := ""
	if n.ClassProperty {
		prefix = "static "
		if protocol {
			skip("class protocol properties have no receiver class")
			return "", member{}, "", nil
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
	return doc(original, tags, "\t\t") + "\t\t" + prefix + readonly + out.Name + ": " + native.adamic + ";\n", member{selector: getter, property: true, adamic: native.adamic}, out.Name, nil
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
	d, gone, err := g.description(n, kind, parent)
	d.ParentSwiftName = parentSwiftName
	if err != nil {
		return d, naming.Output{}, nil, nativeType{}, "", err
	}
	if gone != "" {
		return d, naming.Output{}, nil, nativeType{}, gone, nil
	}
	if kind != naming.CFunction && (n.Name == "retain" || n.Name == "release" || n.Name == "autorelease" || n.Name == "dealloc") {
		return d, naming.Output{}, nil, nativeType{}, "manual reference-count messages bypass Adamic ownership", nil
	}
	if kind == naming.ClassMethod && (n.Name == "initialize" || n.Name == "load") {
		return d, naming.Output{}, nil, nativeType{}, "the runtime sends this message itself, once", nil
	}
	if n.Variadic {
		return d, naming.Output{}, nil, nativeType{}, "variadic declarations are not carried by the bridge", nil
	}
	var natives []nativeType
	parameters := children(n, "ParmVarDecl")
	for i := 0; i < len(parameters); i++ {
		p := parameters[i]
		// A target and its action (buttonWithTitle:target:action:) are one closure, which the
		// bridge hands Apple as both: an object that calls it, and the selector it answers.
		if cleanType(p.Type.Qual) == "id" && i+1 < len(parameters) && cleanType(parameters[i+1].Type.Qual) == "SEL" {
			natives = append(natives, nativeType{hidden: true, c: "id", header: "id"}, nativeType{tag: "action", adamic: "() => void", c: "id", header: "SEL"})
			d.Parameters = append(d.Parameters,
				naming.Parameter{Name: p.Name, Type: naming.Type{Spelling: "id", Nullability: naming.Nonnull, Framework: n.Framework}},
				naming.Parameter{Name: parameters[i+1].Name, Type: naming.Type{Spelling: "SEL", Nullability: naming.Nonnull, Framework: n.Framework}})
			i++
			continue
		}
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
		parameterType := naming.Type{Spelling: namingSpelling(p.Type, native), Object: native.c == "id", Nullability: naming.Nonnull, Framework: referenceFramework(native, n.Framework)}
		if strings.HasPrefix(native.tag, "block(") {
			native = blockNames(native, source)
			// The naming layer reads a block as a function of what it's handed.
			function := &naming.FunctionType{Result: naming.Type{Spelling: "void"}}
			for _, member := range native.members {
				function.Parameters = append(function.Parameters, naming.Type{Spelling: member.header, Object: member.c == "id", Nullability: naming.Nonnull, Framework: referenceFramework(member, n.Framework)})
			}
			parameterType = naming.Type{Spelling: "", Function: function, Nullability: naming.Nonnull, Framework: n.Framework}
		}
		natives = append(natives, native)
		d.Parameters = append(d.Parameters, naming.Parameter{Name: p.Name, Type: parameterType})
	}
	result, e := g.native(n.Result, parent, true)
	if e != nil {
		return d, naming.Output{}, nil, nativeType{}, e.Error(), nil
	}
	d.Result = naming.Type{Spelling: namingSpelling(n.Result, result), Object: result.c == "id", Framework: referenceFramework(result, n.Framework), Nullability: naming.Nonnull}
	d.PropertyNames = properties
	d.TakenNames = values
	d.KeepNeedlessWords = n.KeepNeedlessWords
	out, e := naming.Name(d)
	if e != nil {
		return d, out, nil, result, "no Adamic name yet: " + e.Error(), nil
	}
	return d, out, natives, result, "", nil
}
func signature(out naming.Output, natives []nativeType) (string, string) {
	parameters := []string{}
	sources := make([]string, len(natives))
	// A target an action carries is no argument of its own; its action's source stands for both.
	for _, arg := range out.Arguments.Positional {
		if natives[arg.Index].hidden {
			continue
		}
		sources[arg.Index] = fmt.Sprintf("%d:%s", len(parameters), natives[arg.Index].tag)
		parameters = append(parameters, arg.Name+": "+natives[arg.Index].adamic)
	}
	visible := 0
	for _, arg := range out.Arguments.Options {
		if !natives[arg.Index].hidden {
			visible++
		}
	}
	if visible > 0 {
		fields := []string{}
		position := len(parameters)
		for _, arg := range out.Arguments.Options {
			if natives[arg.Index].hidden {
				continue
			}
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
	written := []string{}
	for _, source := range sources {
		if source != "" {
			written = append(written, source)
		}
	}
	suffix := ""
	if len(written) > 0 {
		suffix = " " + strings.Join(written, " ")
	}
	return strings.Join(parameters, ", "), suffix
}
func (g *generator) emitMethod(owner *definition, n *node, initializers map[string]bool, values []string, inherited map[string][]member) (string, member, string, bool, error) {
	kind := naming.InstanceMethod
	if !n.Instance {
		kind = naming.ClassMethod
	}
	d, out, natives, result, reason, err := g.callable(n, kind, owner.node.Name, owner.declaration.SwiftName, g.propertyNames(owner, map[string]bool{}), values)
	if err != nil {
		return "", member{}, "", false, err
	}
	m := g.getModule(owner.output.Module)
	if reason != "" {
		m.comments = append(m.comments, "// Skipped "+methodOriginal(owner.node.Name, n.Name, n.Instance)+": "+reason+".")
		return "", member{}, "", false, nil
	}
	constructor := out.Arguments.Constructor
	if constructor && !n.Instance && initializers[naming.Shape(out.Arguments)] {
		m.comments = append(m.comments, "// Skipped "+methodOriginal(owner.node.Name, n.Name, false)+": an initializer of the same shape is the constructor.")
		return "", member{}, "", false, nil
	}
	if has(n, "NSConsumesSelfAttr") && !constructor {
		m.comments = append(m.comments, "// Skipped "+n.Name+": a consumed receiver cannot be passed borrowed.")
		return "", member{}, "", false, nil
	}
	for _, existing := range inherited[out.Name] {
		// +[NSCalendarDate distantFuture] beside NSDate's class property: the ancestor's stands.
		if existing.property && !constructor {
			m.comments = append(m.comments, "// Skipped "+methodOriginal(owner.node.Name, n.Name, n.Instance)+": an ancestor declares "+out.Name+" as a property, which stands.")
			return "", member{}, "", false, nil
		}
	}
	g.declarations = append(g.declarations, d)
	for _, native := range append(append([]nativeType{}, natives...), result) {
		for _, reference := range native.references() {
			g.addImport(m, owner.output.Module, reference)
		}
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
	g.bound[methodOriginal(owner.node.Name, n.Name, n.Instance)] = true
	text := doc(methodOriginal(owner.node.Name, n.Name, n.Instance), []string{tag}, "\t\t")
	if constructor {
		text += "\t\tconstructor(" + parameters + ");\n"
	} else {
		text += "\t\t" + prefix + out.Name + "(" + parameters + "): " + result.adamic + ";\n"
	}
	references := []*definition{}
	for _, native := range append(append([]nativeType{}, natives...), result) {
		references = append(references, native.references()...)
	}
	return text, member{selector: n.Name, text: text, references: references}, out.Name, constructor, nil
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
		for _, reference := range native.references() {
			g.addImport(m, out.Module, reference)
		}
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
