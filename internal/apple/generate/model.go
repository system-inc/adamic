package generate

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/system-inc/adamic/internal/apple/naming"
)

type definition struct {
	node        *node
	declaration naming.Declaration
	output      naming.Output
	members     []string
	values      []string
	options     bool
	// emitted marks a class or protocol written out; a class is written after its superclass, so
	// it can see what its ancestors' members are called.
	emitted bool
	// surface is every member the class has by Adamic name, its own and its ancestors', static
	// (true) and instance apart, so a subclass can reconcile its members with what it inherits.
	surface map[bool]map[string][]member
}

// member is one declaration under a name: a property, or one overload of a method.
type member struct {
	selector   string
	property   bool
	adamic     string        // a property's type
	text       string        // a method's declaration, re-declared by a subclass that overloads its name
	references []*definition // the types that text names, for the subclass's imports
}
type module struct {
	imports      map[string]string
	declarations []string
	comments     []string
}
type generator struct {
	configuration Configuration
	nodes         []*node
	types         map[string]*definition
	sources       map[string][]byte
	modules       map[string]*module
	declarations  []naming.Declaration
	checks        []checkCall
	implements    []implementCheck
	// bound is every method a binding sends, -[Class selector], what leaves.go counts a program can
	// hand an instance.
	bound map[string]bool
	// holdsTable is what leaves derives every non-leaf class holds (leaves.go).
	holdsTable  []byte
	importError error
}
type nativeType struct {
	tag, adamic, c, header string
	reference              *definition
	nullable               bool
	// A block's parameters, what Apple hands the closure; their types are imports too.
	members []nativeType
	// hidden is a target whose action argument carries it: the two are one closure.
	hidden bool
}

// references are the bound types a native type names, its own and a block's members'.
func (t nativeType) references() []*definition {
	references := []*definition{}
	if t.reference != nil {
		references = append(references, t.reference)
	}
	for _, member := range t.members {
		references = append(references, member.references()...)
	}
	return references
}

func cleanType(s string) string {
	for _, qualifier := range []string{"_Nonnull", "_Nullable", "_Null_unspecified", "const ", "volatile ", "__kindof "} {
		s = strings.ReplaceAll(s, qualifier, "")
	}
	return strings.TrimSpace(s)
}
func (g *generator) getModule(path string) *module {
	m := g.modules[path]
	if m == nil {
		m = &module{imports: map[string]string{}}
		g.modules[path] = m
	}
	return m
}
func (g *generator) addImport(m *module, own string, d *definition) {
	if d != nil && d.output.Module != own {
		if other, exists := m.imports[d.output.Name]; exists && other != d.output.Module {
			g.importError = fmt.Errorf("module %s imports %s from both %s and %s", own, d.output.Name, other, d.output.Module)
			return
		}
		for _, local := range g.types {
			if local.output.Module == own && local.output.Name == d.output.Name {
				g.importError = fmt.Errorf("module %s imports %s from %s but already exports that name", own, d.output.Name, d.output.Module)
				return
			}
		}
		m.imports[d.output.Name] = d.output.Module
	}
}
func children(n *node, kind string) []*node {
	var result []*node
	for _, child := range n.Children {
		if child.Kind == kind {
			result = append(result, child)
		}
	}
	return result
}
func has(n *node, kind string) bool { return len(children(n, kind)) > 0 }
func (g *generator) text(l location) (string, error) {
	if !l.Valid {
		return "", fmt.Errorf("attribute has no source range")
	}
	name, err := filepath.Abs(l.File)
	if err != nil {
		return "", err
	}
	source, ok := g.sources[name]
	if !ok {
		source, err = g.configuration.ReadSource(name)
		if err != nil {
			return "", err
		}
		g.sources[name] = source
	}
	if l.Offset < 0 || l.Offset >= len(source) {
		return "", fmt.Errorf("attribute offset outside %s", name)
	}
	end := l.Offset
	// A macro's name takes digits after its first character
	// (DEPRECATED_IN_MAC_OS_X_VERSION_10_0_AND_LATER).
	for end < len(source) && (source[end] >= 'a' && source[end] <= 'z' || source[end] >= 'A' && source[end] <= 'Z' || source[end] == '_' || end > l.Offset && source[end] >= '0' && source[end] <= '9') {
		end++
	}
	for end < len(source) && (source[end] == ' ' || source[end] == '\t') {
		end++
	}
	if end < len(source) && source[end] == '(' {
		depth := 0
		quote := byte(0)
		for ; end < len(source); end++ {
			c := source[end]
			if quote != 0 {
				if c == '\\' {
					end++
					continue
				}
				if c == quote {
					quote = 0
				}
				continue
			}
			if c == '"' || c == '\'' {
				quote = c
				continue
			}
			if c == '(' {
				depth++
			}
			if c == ')' {
				depth--
				if depth == 0 {
					end++
					break
				}
			}
		}
		if depth != 0 {
			return "", fmt.Errorf("unterminated attribute in %s", name)
		}
	}
	return string(source[l.Offset:end]), nil
}

// regionMacro is an availability macro written without arguments, its platforms in its name.
// The word may open the name (DEPRECATED_IN_MAC_OS_X_VERSION_10_1_AND_LATER).
var regionMacro = regexp.MustCompile(`^[A-Z0-9_]*?(AVAILABLE|DEPRECATED)[A-Z0-9_]*`)

// unavailableByName reports whether a macro like APPKIT_API_UNAVAILABLE_BEGIN_MACCATALYST makes its
// declarations unavailable on the platform being generated: the words after UNAVAILABLE name the
// platforms it removes. AVAILABLE and DEPRECATED regions keep them.
func (g *generator) unavailableByName(macro string) bool {
	at := strings.Index(macro, "UNAVAILABLE")
	if at < 0 {
		return false
	}
	platform := g.configuration.Platform
	for _, word := range strings.Split(strings.ToLower(macro[at+len("UNAVAILABLE"):]), "_") {
		if word == platform || platform == "macos" && word == "macosx" || platform == "visionos" && word == "xros" {
			return true
		}
	}
	return false
}

// attributes reads a declaration's Swift name and availability. gone says why the declaration
// isn't bound on the platform: "unavailable on macos", or "deprecated on macos", since Apple's
// deprecation is its "don't use" and the bindings leave those out (docs/apple.md), or "" when it is
// bound.
func (g *generator) attributes(n *node) (swift string, refined bool, gone string, err error) {
	unavailable, deprecated := false, false
	defer func() {
		switch {
		case unavailable:
			gone = "unavailable on " + g.configuration.Platform
		case deprecated:
			gone = "deprecated on " + g.configuration.Platform
		}
	}()
	for _, child := range n.Children {
		switch child.Kind {
		case "SwiftPrivateAttr":
			refined = true
		case "UnavailableAttr":
			unavailable = true
		case "DeprecatedAttr":
			// __attribute__((deprecated)), DEPRECATED_ATTRIBUTE: deprecated everywhere.
			deprecated = true
		case "SwiftNameAttr", "AvailabilityAttr":
			var text string
			text, err = g.text(child.Begin)
			if err != nil {
				return
			}
			left := strings.IndexByte(text, '(')
			right := strings.LastIndexByte(text, ')')
			if (left < 0 || right <= left) && child.Kind == "AvailabilityAttr" && regionMacro.MatchString(text) {
				// A region's macro names its platforms instead of taking them:
				// APPKIT_API_UNAVAILABLE_BEGIN_MACCATALYST covers the declarations up to its END.
				if g.unavailableByName(regionMacro.FindString(text)) {
					unavailable = true
				}
				deprecated = deprecated || g.deprecatedByName(regionMacro.FindString(text))
				continue
			}
			if child.Kind == "AvailabilityAttr" {
				deprecated = deprecated || g.deprecatedIn(text, child.Begin.File)
			}
			if left < 0 || right <= left {
				err = fmt.Errorf("cannot read attribute %q", text)
				return
			}
			arguments := text[left+1 : right]
			if child.Kind == "SwiftNameAttr" {
				if strings.HasPrefix(arguments, "\"") {
					swift, err = strconv.Unquote(arguments)
				} else {
					swift = strings.TrimSpace(arguments)
				}
			} else {
				upper := strings.ToUpper(text)
				if strings.Contains(upper, "UNAVAILABLE") {
					platform := g.configuration.Platform
					for _, part := range strings.FieldsFunc(strings.ToLower(arguments), func(c rune) bool { return c == ',' || c == ' ' || c == '\t' || c == '(' || c == ')' }) {
						if part == platform || platform == "macos" && part == "macosx" || platform == "visionos" && part == "xros" {
							unavailable = true
						}
					}
				} else if strings.HasPrefix(text, "availability(") {
					unavailable = unavailable || g.unavailableIn(text)
				} else if !strings.Contains(upper, "AVAILABLE") && !strings.Contains(upper, "DEPRECATED") {
					// A framework's own macro over availability (CoreGraphics'
					// SCREEN_CAPTURE_OBSOLETE(10.5,14.0,15.0)): expand it from its #define.
					expanded, found := g.expandMacro(text, child.Begin.File)
					if !found || !strings.Contains(expanded, "availability(") {
						err = fmt.Errorf("unrecognized availability attribute %q", text)
						return
					}
					unavailable = unavailable || g.unavailableIn(expanded)
				}
			}
		}
	}
	return
}

// unavailableIn reads availability(platform, ...) clauses: a declaration is unavailable on this
// platform when its clause says unavailable or obsoleted, since every obsoleted version precedes
// the SDK being generated.
func (g *generator) unavailableIn(text string) bool {
	platform := g.configuration.Platform
	for rest := text; ; {
		at := strings.Index(rest, "availability(")
		if at < 0 {
			return false
		}
		rest = rest[at+len("availability("):]
		end := strings.IndexByte(rest, ')')
		if end < 0 {
			return false
		}
		clauses := strings.Split(rest[:end], ",")
		name := strings.TrimSpace(clauses[0])
		if name == platform || platform == "macos" && name == "macosx" || platform == "visionos" && name == "xros" {
			for _, clause := range clauses[1:] {
				clause = strings.Join(strings.Fields(clause), "")
				if clause == "unavailable" || strings.HasPrefix(clause, "obsoleted=") {
					return true
				}
			}
		}
		rest = rest[end:]
	}
}

// expandMacro expands one use of a function-like macro, NAME(arguments), from its #define in the
// use's own header or any header of the frameworks being generated. It substitutes whole-word
// parameters and nothing else: enough for a framework's availability macros, which is all it's for.
func (g *generator) expandMacro(use, file string) (string, bool) {
	left := strings.IndexByte(use, '(')
	if left <= 0 || !strings.HasSuffix(use, ")") {
		return "", false
	}
	name := use[:left]
	arguments := strings.Split(use[left+1:len(use)-1], ",")
	definition := regexp.MustCompile(`(?m)^[ \t]*#[ \t]*define[ \t]+` + regexp.QuoteMeta(name) + `\(([^)]*)\)((?:[^\n]*\\\n)*[^\n]*)`)
	files := []string{file}
	for _, framework := range g.configuration.Frameworks {
		entries, _ := os.ReadDir(framework.Headers)
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".h") {
				files = append(files, filepath.Join(framework.Headers, entry.Name()))
			}
		}
	}
	for _, candidate := range files {
		source, err := g.configuration.ReadSource(candidate)
		if err != nil {
			continue
		}
		match := definition.FindSubmatch(source)
		if match == nil {
			continue
		}
		parameters := strings.Split(string(match[1]), ",")
		if len(parameters) != len(arguments) {
			return "", false
		}
		body := strings.ReplaceAll(string(match[2]), "\\\n", " ")
		for i, parameter := range parameters {
			word := regexp.MustCompile(`\b` + regexp.QuoteMeta(strings.TrimSpace(parameter)) + `\b`)
			body = word.ReplaceAllLiteralString(body, strings.TrimSpace(arguments[i]))
		}
		return body, true
	}
	return "", false
}

func (g *generator) description(n *node, kind naming.Kind, parent string) (naming.Declaration, string, error) {
	swift, refined, gone, err := g.attributes(n)
	return naming.Declaration{Kind: kind, Name: n.Name, Parent: parent, Framework: n.Framework, SwiftName: swift, RefinedForSwift: refined}, gone, err
}
func (g *generator) skipped(n *node, reason string) {
	path := "apple/" + strings.ToLower(n.Framework) + "/unsupported"
	g.getModule(path).comments = append(g.getModule(path).comments, "// Skipped "+n.Name+": "+strings.ReplaceAll(reason, "\n", " ")+".")
}
func enumValue(n *node) (string, bool) {
	if n.Kind == "ConstantExpr" && n.Value != "" {
		return n.Value, true
	}
	for _, child := range n.Children {
		if v, ok := enumValue(child); ok {
			return v, true
		}
	}
	if n.Kind == "IntegerLiteral" && n.Value != "" {
		return n.Value, true
	}
	return "", false
}
func (g *generator) inventory() error {
	// Definitions are registered before references are mapped, regardless of
	// header order. Forward declarations never replace a complete declaration.
	for _, n := range g.nodes {
		var kind naming.Kind
		switch n.Kind {
		case "ObjCInterfaceDecl", "ObjCProtocolDecl":
			if len(n.Children) == 0 {
				source, err := g.sourceRange(n)
				if err != nil {
					return err
				}
				if !strings.Contains(source, "@end") {
					continue
				}
			}
			kind = naming.Class
			if n.Kind == "ObjCProtocolDecl" {
				kind = naming.Protocol
			}
		case "EnumDecl":
			if n.Name == "" {
				g.skipped(n, "anonymous enum constants have no native load tag")
				continue
			}
			if len(children(n, "EnumConstantDecl")) == 0 {
				continue
			}
			kind = naming.Enum
			if has(n, "FlagEnumAttr") {
				kind = naming.OptionSet
			}
		case "RecordDecl":
			if n.Name == "" {
				g.skipped(n, "anonymous structs have no bridge representation")
				continue
			}
			if !n.Complete {
				continue
			}
			if n.Tag == "union" {
				g.skipped(n, "unions are not carried by the bridge")
				continue
			}
			kind = naming.Struct
		case "TypedefDecl":
			continue
		default:
			continue
		}
		d, gone, err := g.description(n, kind, "")
		if err != nil {
			return err
		}
		if gone != "" {
			g.skipped(n, gone)
			continue
		}
		key := n.Name
		if old := g.types[key]; old != nil && old.declaration.Kind != d.Kind {
			// Swift's rule: a protocol that shares a class's name imports as
			// ...Protocol (NSObjectProtocol, NSAccessibilityElementProtocol), and both are kept.
			switch {
			case d.Kind == naming.Protocol && old.declaration.Kind == naming.Class:
				key = protocolKey(n.Name)
			case d.Kind == naming.Class && old.declaration.Kind == naming.Protocol:
				delete(g.types, key)
				protocolSwiftName(&old.declaration)
				renamed, err := naming.Name(old.declaration)
				if err != nil {
					return err
				}
				old.output = renamed
				g.types[protocolKey(n.Name)] = old
			default:
				return fmt.Errorf("ambiguous type kind for %s; refusing to discard either declaration", n.Name)
			}
		}
		if key != n.Name {
			protocolSwiftName(&d)
		}
		out, err := naming.Name(d)
		if err != nil {
			return err
		}
		if old := g.types[key]; old != nil {
			if len(n.Children) <= len(old.node.Children) {
				continue
			}
		}
		g.types[key] = &definition{node: n, declaration: d, output: out, options: kind == naming.OptionSet}
	}
	// Read every enumerator's evaluated value. Implicit values use the previous
	// evaluated value plus one, not a search for an arbitrary integer leaf.
	for _, name := range sortedKeys(g.types) {
		def := g.types[name]
		if def.declaration.Kind != naming.Enum && def.declaration.Kind != naming.OptionSet {
			continue
		}
		siblings := []naming.Enumerator{}
		for _, child := range children(def.node, "EnumConstantDecl") {
			swift, _, gone, err := g.attributes(child)
			if err != nil {
				return err
			}
			if strings.HasPrefix(gone, "deprecated") {
				// Listed, as every declaration left out for its deprecation is.
				child.Framework = def.node.Framework
				g.skipped(child, gone)
			}
			siblings = append(siblings, naming.Enumerator{Name: child.Name, SwiftName: swift, Unavailable: gone != "", Deprecated: has(child, "DeprecatedAttr")})
		}
		next := big.NewInt(0)
		for i, child := range children(def.node, "EnumConstantDecl") {
			value, found := enumValue(child)
			if found {
				if _, ok := next.SetString(value, 0); !ok {
					return fmt.Errorf("invalid enum value %q", value)
				}
			}
			current := next.String()
			next.Add(next, big.NewInt(1))
			if siblings[i].Unavailable {
				continue
			}
			// An option set's bits are unsigned and may reach bit 63 (NSAlignRectFlipped); an
			// enumeration's values are signed.
			bound := ""
			if def.options {
				parsed, err := strconv.ParseUint(current, 10, 64)
				if err != nil {
					return fmt.Errorf("%s: bridge option table requires unsigned 64-bit values: %w", child.Name, err)
				}
				bound = strconv.FormatUint(parsed, 10)
			} else {
				parsed, err := strconv.ParseInt(current, 10, 64)
				if err != nil {
					return fmt.Errorf("%s: bridge enum table requires signed 64-bit values: %w", child.Name, err)
				}
				bound = strconv.FormatInt(parsed, 10)
			}
			kind := naming.EnumConstant
			if def.options {
				kind = naming.OptionConstant
			}
			d := naming.Declaration{Kind: kind, Name: child.Name, Parent: def.node.Name, ParentSwiftName: def.output.SwiftName, Framework: def.node.Framework, SwiftName: siblings[i].SwiftName, Enumerators: siblings}
			out, err := naming.Name(d)
			if err != nil {
				return err
			}
			def.members = append(def.members, out.Name)
			def.values = append(def.values, bound)
			g.declarations = append(g.declarations, d)
		}
	}
	return nil
}

// protocolKey keys a protocol in the type table when a class holds its Objective-C name.
// It can't collide with a declaration's own name, which is always an identifier.
func protocolKey(name string) string { return "@protocol " + name }

// protocol finds the protocol written id<name>, under its own name or beside a class of that name.
func (g *generator) protocol(name string) *definition {
	if def := g.types[protocolKey(name)]; def != nil {
		return def
	}
	if def := g.types[name]; def != nil && def.declaration.Kind == naming.Protocol {
		return def
	}
	return nil
}

// protocolSwiftName gives a protocol beside a same-named class Swift's imported name, unless
// the header already names it.
func protocolSwiftName(d *naming.Declaration) {
	if d.SwiftName == "" {
		d.SwiftName = d.Name + "Protocol"
	}
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func (g *generator) native(t writtenType, parent string, result bool) (nativeType, error) {
	q := cleanType(t.Qual)
	desugared := cleanType(t.Desugared)
	if desugared == "" {
		desugared = q
	}
	if desugared == q {
		switch q {
		case "NSInteger", "long long", "int64_t":
			desugared = "long"
		case "NSUInteger":
			desugared = "unsigned long"
		case "CGFloat":
			desugared = "double"
		}
	}
	mapped := nativeType{header: q, nullable: !strings.Contains(t.Qual, "_Nonnull")}
	if strings.Contains(q, "^") {
		if result {
			return mapped, fmt.Errorf("block results are not carried by the bridge")
		}
		return g.block(t.Qual, parent)
	}
	if strings.ContainsAny(q, "[]") {
		return mapped, fmt.Errorf("C arrays are not carried by the bridge")
	}
	if strings.Contains(q, "(") || strings.Contains(desugared, "(") {
		return mapped, fmt.Errorf("function pointers are not carried by the bridge")
	}
	if strings.HasPrefix(q, "union ") || strings.HasPrefix(desugared, "union ") {
		return mapped, fmt.Errorf("unions are not carried by the bridge")
	}
	if q == "instancetype" {
		q = parent + " *"
		mapped.header = q
	}
	if q == "void" {
		mapped.tag = "void"
		mapped.adamic = "void"
		mapped.c = "void"
		return mapped, nil
	}
	if q == "BOOL" {
		mapped.tag = "boolean"
		mapped.adamic = "boolean"
		mapped.c = "BOOL"
		return mapped, nil
	}
	switch desugared {
	case "double":
		mapped.tag = "double"
		mapped.adamic = "number"
		mapped.c = "double"
		return mapped, nil
	case "long", "long long":
		// long long is long's 64 bits on every Apple target the bridge builds for.
		mapped.tag = "integer"
		mapped.adamic = "number"
		mapped.c = "long"
		return mapped, nil
	case "unsigned long", "unsigned long long":
		mapped.tag = "unsigned"
		mapped.adamic = "number"
		mapped.c = "unsigned long"
		return mapped, nil
	}
	base := strings.TrimPrefix(q, "enum ")
	if def := g.types[base]; def != nil && (def.declaration.Kind == naming.Enum || def.options) {
		underlying := cleanType(def.node.Underlying.Desugared)
		if underlying == "" {
			underlying = cleanType(def.node.Underlying.Qual)
		}
		if underlying == "NSInteger" {
			underlying = "long"
		}
		if underlying == "NSUInteger" {
			underlying = "unsigned long"
		}
		want := "long"
		tag := "enum"
		mapped.adamic = def.output.Name
		if def.options {
			want = "unsigned long"
			tag = "options"
			mapped.adamic = "readonly " + def.output.Name + "[]"
		}
		// An enumeration over NSUInteger crosses in the bridge's long, and an option set over
		// unsigned long long (NSEventMask) in its unsigned long: the same register and bits, since
		// every case's value is held by the check file.
		if underlying != want && !(tag == "enum" && underlying == "unsigned long") && !(tag == "options" && underlying == "unsigned long long") {
			return mapped, fmt.Errorf("enum width/sign %s cannot cross as %s", underlying, want)
		}
		if result {
			return mapped, fmt.Errorf("enumeration and option results are not carried by the bridge")
		}
		if len(def.members) == 0 {
			return mapped, fmt.Errorf("enum has no available cases")
		}
		pairs := []string{}
		for i, name := range def.members {
			pairs = append(pairs, strings.Trim(name, "'")+"="+def.values[i])
		}
		mapped.tag = tag + "(" + strings.Join(pairs, ",") + ")"
		mapped.c = want
		mapped.reference = def
		return mapped, nil
	}
	if q == "CGRect" || q == "NSRect" || desugared == "struct CGRect" {
		if result {
			return mapped, fmt.Errorf("rectangle results are not carried by the bridge")
		}
		def := g.types["CGRect"]
		if def == nil {
			return mapped, fmt.Errorf("CGRect definition not among generated headers")
		}
		mapped.tag = "rectangle"
		mapped.adamic = def.output.Name
		mapped.c = "adamic_apple_rectangle"
		mapped.reference = def
		return mapped, nil
	}
	if q == "id" {
		// Any object: every object the bridge holds is an NSObject (FoundationObject).
		q = "NSObject *"
	}
	if q == "NSString *" {
		mapped.tag = "string"
		mapped.adamic = "string"
		mapped.c = "id"
	} else {
		if strings.Count(q, "*") > 1 {
			return mapped, fmt.Errorf("pointer-to-pointer types are not carried by the bridge")
		}
		base = strings.TrimSpace(strings.TrimSuffix(q, "*"))
		def := g.types[base]
		if strings.HasPrefix(base, "id<") {
			def = g.protocol(strings.TrimSuffix(strings.TrimPrefix(base, "id<"), ">"))
		}
		if def == nil || def.declaration.Kind != naming.Class && def.declaration.Kind != naming.Protocol {
			return mapped, fmt.Errorf("unsupported or unbound native type %q", t.Qual)
		}
		if !strings.HasSuffix(q, "*") && !strings.HasPrefix(q, "id<") {
			return mapped, fmt.Errorf("object type needs a pointer")
		}
		mapped.tag = "object"
		mapped.adamic = def.output.Name
		mapped.c = "id"
		mapped.reference = def
	}
	if mapped.nullable {
		mapped.tag += "?"
		mapped.adamic += " | undefined"
	}
	return mapped, nil
}

// block maps a block parameter, a closure Apple calls with the values its parameters list:
// void (^)(NSData * _Nullable, NSURLResponse * _Nullable, NSError * _Nullable) is
// block(object?,object?,object?). Its parameters are named argument1 onward until blockNames
// reads the header's names.
func (g *generator) block(written, parent string) (nativeType, error) {
	mapped := nativeType{header: cleanType(written), c: "id"}
	caret := strings.Index(written, "(^")
	close := strings.Index(written[caret:], ")")
	if caret < 0 || close < 0 {
		return mapped, fmt.Errorf("unreadable block type %q", written)
	}
	if returned := withoutAttributes(cleanType(written[:caret])); returned != "void" {
		return mapped, fmt.Errorf("blocks returning %s are not carried by the bridge", returned)
	}
	list := strings.TrimSpace(written[caret+close+1:])
	if !strings.HasPrefix(list, "(") || !strings.HasSuffix(list, ")") {
		return mapped, fmt.Errorf("unreadable block parameters %q", written)
	}
	tags, types := []string{}, []string{}
	for i, parameter := range splitTopLevel(list[1 : len(list)-1]) {
		if cleanType(parameter) == "void" && i == 0 {
			break
		}
		member, err := g.native(writtenType{Qual: parameter}, parent, false)
		if err != nil {
			return mapped, fmt.Errorf("block parameter: %w", err)
		}
		switch {
		case member.tag == "void", member.tag == "action", member.tag == "rectangle", strings.HasPrefix(member.tag, "block("), strings.HasPrefix(member.tag, "enum("), strings.HasPrefix(member.tag, "options("):
			return mapped, fmt.Errorf("a block can't take a %s yet", member.tag)
		}
		mapped.members = append(mapped.members, member)
		tags = append(tags, member.tag)
		types = append(types, fmt.Sprintf("argument%d: %s", i+1, member.adamic))
	}
	mapped.tag = "block(" + strings.Join(tags, ",") + ")"
	mapped.adamic = "(" + strings.Join(types, ", ") + ") => void"
	return mapped, nil
}

// withoutAttributes drops what only annotates a type: __attribute__((...)) groups, and the macros
// that spell them (NS_SWIFT_SENDABLE void is a void).
func withoutAttributes(text string) string {
	text = attributeGroup.ReplaceAllString(text, " ")
	words := []string{}
	for _, word := range strings.Fields(text) {
		if !macroWord.MatchString(word) {
			words = append(words, word)
		}
	}
	return strings.Join(words, " ")
}

var attributeGroup = regexp.MustCompile(`__attribute__\(\([^)]*(?:\([^)]*\)[^)]*)*\)\)`)
var macroWord = regexp.MustCompile(`^[A-Z][A-Z0-9_]*_[A-Z0-9_]*$`)

// splitTopLevel splits a parameter list at the commas outside any parentheses or angle brackets.
func splitTopLevel(list string) []string {
	parts, depth, start := []string{}, 0, 0
	for i, c := range list {
		switch c {
		case '(', '<':
			depth++
		case ')', '>':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(list[start:i]))
				start = i + 1
			}
		}
	}
	return append(parts, strings.TrimSpace(list[start:]))
}

// blockNames gives a block's closure type the parameter names its header wrote
// (completionHandler:(void (^)(NSData *data, NSURLResponse *response, NSError *error))), where
// it wrote them.
func blockNames(native nativeType, source string) nativeType {
	// The caret may follow an attribute: (void (NS_SWIFT_SENDABLE ^)(NSData *data, ...)).
	caret := strings.Index(source, "^")
	if caret < 0 || len(native.members) == 0 {
		return native
	}
	close := strings.Index(source[caret:], ")")
	rest := strings.TrimSpace(source[caret+close+1:])
	if !strings.HasPrefix(rest, "(") {
		return native
	}
	// The list ends at its own closing parenthesis, not the parameter's after it
	// (length))completionHandler).
	end, depth := -1, 0
	for i, c := range rest {
		if c == '(' {
			depth++
		} else if c == ')' {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}
	if end < 0 {
		return native
	}
	written := splitTopLevel(rest[1:end])
	if len(written) != len(native.members) {
		return native
	}
	types := []string{}
	for i, member := range native.members {
		name := fmt.Sprintf("argument%d", i+1)
		words := strings.FieldsFunc(written[i], func(c rune) bool { return c == ' ' || c == '*' || c == '\t' })
		if len(words) > 1 {
			last := words[len(words)-1]
			if identifierPattern.MatchString(last) && !strings.HasPrefix(last, "_") && last != "const" {
				name = last
			}
		}
		types = append(types, name+": "+member.adamic)
	}
	native.adamic = "(" + strings.Join(types, ", ") + ") => void"
	return native
}

var identifierPattern = regexp.MustCompile(`^[a-z][A-Za-z0-9]*$`)

func (g *generator) sourceRange(n *node) (string, error) {
	if !n.Begin.Valid || !n.End.Valid || n.Begin.File != n.End.File {
		return "", nil
	}
	name, err := filepath.Abs(n.Begin.File)
	if err != nil {
		return "", err
	}
	source, ok := g.sources[name]
	if !ok {
		source, err = g.configuration.ReadSource(name)
		if err != nil {
			return "", err
		}
		g.sources[name] = source
	}
	end := n.End.Offset + n.End.Length
	if n.Begin.Offset < 0 || end < n.Begin.Offset || end > len(source) {
		return "", fmt.Errorf("source range outside %s", name)
	}
	return string(source[n.Begin.Offset:end]), nil
}

// JSON omits protocol @optional state. Read markers before the member, outside
// comments and strings, rather than promising an optional method is present.
func (g *generator) protocolOptional(owner, member *node) (bool, error) {
	source, err := g.sourceRange(owner)
	if err != nil {
		return false, err
	}
	offset := member.Begin.Offset - owner.Begin.Offset
	if !member.Begin.Valid || member.Begin.File != owner.Begin.File || offset < 0 || offset > len(source) {
		return false, fmt.Errorf("cannot recover protocol optionality for %s", member.Name)
	}
	text := []byte(source[:offset])
	for i := 0; i < len(text); {
		if i+1 < len(text) && text[i] == '/' && text[i+1] == '/' {
			for i < len(text) && text[i] != '\n' {
				text[i] = ' '
				i++
			}
			continue
		}
		if i+1 < len(text) && text[i] == '/' && text[i+1] == '*' {
			text[i], text[i+1] = ' ', ' '
			i += 2
			for i < len(text) {
				if i+1 < len(text) && text[i] == '*' && text[i+1] == '/' {
					text[i], text[i+1] = ' ', ' '
					i += 2
					break
				}
				text[i] = ' '
				i++
			}
			continue
		}
		if text[i] == '"' || text[i] == '\'' {
			quote := text[i]
			text[i] = ' '
			i++
			for i < len(text) {
				c := text[i]
				text[i] = ' '
				i++
				if c == '\\' && i < len(text) {
					text[i] = ' '
					i++
					continue
				}
				if c == quote {
					break
				}
			}
			continue
		}
		i++
	}
	optional := strings.LastIndex(string(text), "@optional")
	required := strings.LastIndex(string(text), "@required")
	return optional > required, nil
}
