package type_checking

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

/*
 * StableTypeText prints a type the same way whichever checker holds it and whatever that checker did before, so
 * a message that names a type reads the same in every run (#dq881as).
 *
 * TypeToString can't promise that. It prints a union's members and an object's members in the checker's own
 * order, and that order moves with the checker's history:
 *   - a union's members are ordered by type id, which follows creation order;
 *   - an anonymous object's members are ordered by declaration, and the `name?: undefined` members TypeScript
 *     adds when it normalizes a union of object literals borrow the declaration of whichever sibling property the
 *     checker first made one from (getUndefinedProperty caches it by name, per checker).
 * With the walk and the type check sharing checkers concurrently, the same literal printed its members in
 * another order from run to run on www-phi-health at GOMAXPROCS=16, while single-threaded runs agreed.
 *
 * The order here uses no type id and no checker history, all the way down:
 *   - an anonymous object's own members (declared directly in the literal, namespace or file it came from) by
 *     their position, then every other member (the synthesized ones) by name;
 *   - a union's and an intersection's members by their own printed text, null and undefined last, which is a
 *     total order: two members that print alike are interchangeable in the text;
 *   - a named type by its name, with its type arguments printed by these same rules.
 *
 * This is not TypeScript's single-threaded order, which no type can recover (the cache's first entry is a fact
 * about the whole program's check order). So it is for messages that answer to nobody else's text: the adamic
 * rules. A ported rule's message is held to ESLint's, which prints the checker's own order.
 */
func StableTypeText(typeChecker *checker.Checker, t *checker.Type) string {
	var text string
	for _, limit := range stableTypeLimits {
		printer := stableTypePrinter{typeChecker: typeChecker, maximumMembers: limit.members, maximumDepth: limit.depth}
		if text = printer.print(t, 0); len(text) <= stableTypeTextBudget {
			break
		}
	}
	return text
}

/*
 * A type is printed within stableTypeTextBudget characters, as TypeToString keeps one short: first with the widest
 * limits, then with each narrower pair in turn, until it fits or the narrowest is used. Members past the member
 * limit become `... N more ...`, and nesting past the depth limit becomes the type's name, or `...`. The steps are
 * fixed, so the text is the same in every run. Without a budget, one union of column definitions printed 6,247
 * characters where TypeToString had printed at most 1,484.
 */
const stableTypeTextBudget = 400

var stableTypeLimits = []struct{ members, depth int }{{8, 5}, {4, 4}, {3, 3}, {2, 2}, {1, 1}}

type stableTypePrinter struct {
	typeChecker *checker.Checker

	// maximumMembers and maximumDepth are this printing's limits. See stableTypeLimits.
	maximumMembers int
	maximumDepth   int
	// stack is the types being printed, so a type that contains itself prints by name the second time.
	stack []*checker.Type
}

func (p *stableTypePrinter) print(t *checker.Type, depth int) string {
	if t == nil {
		return "unknown"
	}
	if alias := t.Alias(); alias != nil && alias.Symbol() != nil {
		return p.named(alias.Symbol().Name, alias.TypeArguments(), depth)
	}
	switch {
	case t.Flags()&checker.TypeFlagsUnion != 0:
		return p.union(t, depth)
	case t.Flags()&checker.TypeFlagsIntersection != 0:
		return p.joined(t.Types(), " & ", depth)
	case t.Flags()&checker.TypeFlagsObject == 0:
		// Primitives, literals, enums, type parameters and the rest print a name of their own, which no other
		// type's order reaches.
		return p.typeChecker.TypeToString(t)
	}
	if slices.Contains(p.stack, t) || depth > p.maximumDepth {
		if symbol := t.Symbol(); symbol != nil && t.ObjectFlags()&checker.ObjectFlagsAnonymous == 0 {
			return symbol.Name
		}
		return "..."
	}
	p.stack = append(p.stack, t)
	defer func() { p.stack = p.stack[:len(p.stack)-1] }()

	if checker.Checker_isArrayType(p.typeChecker, t) {
		arguments := p.typeChecker.GetTypeArguments(t)
		element := "unknown"
		if len(arguments) > 0 {
			element = p.operand(arguments[0], depth+1)
		}
		if t.Target() != nil && t.Target().Symbol() != nil && t.Target().Symbol().Name == "ReadonlyArray" {
			return "readonly " + element + "[]"
		}
		return element + "[]"
	}
	if checker.IsTupleType(t) {
		return p.tuple(t, depth)
	}
	if t.ObjectFlags()&checker.ObjectFlagsReference != 0 && t.Target() != nil && t.Target().Symbol() != nil {
		arguments := p.typeChecker.GetTypeArguments(t)
		if target := t.Target().TargetInterfaceType(); target != nil && len(target.TypeParameters()) <= len(arguments) {
			arguments = arguments[:len(target.TypeParameters())]
		}
		return p.named(t.Target().Symbol().Name, arguments, depth)
	}
	if t.ObjectFlags()&checker.ObjectFlagsAnonymous == 0 && t.Symbol() != nil {
		return p.typeChecker.TypeToString(t)
	}
	return p.anonymous(t, depth)
}

// named prints a name with its type arguments, each by these rules.
func (p *stableTypePrinter) named(name string, arguments []*checker.Type, depth int) string {
	if len(arguments) == 0 {
		return name
	}
	printed := make([]string, len(arguments))
	for index, argument := range arguments {
		printed[index] = p.print(argument, depth+1)
	}
	return name + "<" + strings.Join(printed, ", ") + ">"
}

// operand is a type printed where an operator follows it, `T[]`: a union, an intersection or a function is
// parenthesized, as TypeScript prints it.
func (p *stableTypePrinter) operand(t *checker.Type, depth int) string {
	text := p.print(t, depth)
	if t.Alias() == nil && (t.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 || strings.Contains(text, "=>")) {
		return "(" + text + ")"
	}
	return text
}

// union prints a union's members in their stable order. `true | false` is `boolean`, as TypeScript prints it.
func (p *stableTypePrinter) union(t *checker.Type, depth int) string {
	var printed []string
	var booleans []*checker.Type
	for _, member := range t.Types() {
		if member.Flags()&checker.TypeFlagsBooleanLiteral != 0 {
			booleans = append(booleans, member)
			continue
		}
		printed = append(printed, p.print(member, depth+1))
	}
	switch len(booleans) {
	case 2:
		printed = append(printed, "boolean")
	case 1:
		printed = append(printed, p.typeChecker.TypeToString(booleans[0]))
	}
	return p.sortedJoin(printed, " | ")
}

// joined prints members each by these rules and joins them in their stable order.
func (p *stableTypePrinter) joined(members []*checker.Type, separator string, depth int) string {
	printed := make([]string, len(members))
	for index, member := range members {
		if separator == " & " {
			printed[index] = p.operand(member, depth+1)
		} else {
			printed[index] = p.print(member, depth+1)
		}
	}
	return p.sortedJoin(printed, separator)
}

// sortedJoin orders printed members by their text, with null and undefined last, and keeps the first
// maximumMembers of them.
func (p *stableTypePrinter) sortedJoin(printed []string, separator string) string {
	rank := func(text string) int {
		switch text {
		case "null":
			return 1
		case "undefined":
			return 2
		}
		return 0
	}
	slices.SortStableFunc(printed, func(a, b string) int {
		if rank(a) != rank(b) {
			return rank(a) - rank(b)
		}
		return strings.Compare(a, b)
	})
	if len(printed) > p.maximumMembers {
		more := len(printed) - p.maximumMembers + 1
		printed = append(printed[:p.maximumMembers-1], "... "+strconv.Itoa(more)+" more ...")
	}
	return strings.Join(printed, separator)
}

// tuple prints a tuple's elements in their positions, which no checker reorders.
func (p *stableTypePrinter) tuple(t *checker.Type, depth int) string {
	arguments := p.typeChecker.GetTypeArguments(t)
	tuple := t.TargetTupleType()
	flags := tuple.ElementFlags()
	printed := make([]string, 0, len(flags))
	for index := range flags {
		if index >= len(arguments) {
			break
		}
		element := p.print(arguments[index], depth+1)
		switch {
		case flags[index]&checker.ElementFlagsOptional != 0:
			element = p.operand(arguments[index], depth+1) + "?"
		case flags[index]&checker.ElementFlagsVariable != 0:
			element = "..." + element
		}
		printed = append(printed, element)
	}
	text := "[" + strings.Join(printed, ", ") + "]"
	if tuple.IsReadonly() {
		return "readonly " + text
	}
	return text
}

// anonymous prints an object type that has no name: a function as `(p: T) => R`, and anything else as its
// signatures, index signatures and members in braces.
func (p *stableTypePrinter) anonymous(t *checker.Type, depth int) string {
	calls := p.typeChecker.GetSignaturesOfType(t, checker.SignatureKindCall)
	constructs := p.typeChecker.GetSignaturesOfType(t, checker.SignatureKindConstruct)
	indexes := p.typeChecker.GetIndexInfosOfType(t)
	properties := p.members(t)
	if len(calls) == 1 && len(constructs) == 0 && len(indexes) == 0 && len(properties) == 0 {
		return p.signature(calls[0], " => ", depth)
	}
	if len(constructs) == 1 && len(calls) == 0 && len(indexes) == 0 && len(properties) == 0 {
		return "new " + p.signature(constructs[0], " => ", depth)
	}
	var parts []string
	for _, call := range calls {
		parts = append(parts, p.signature(call, ": ", depth))
	}
	for _, construct := range constructs {
		parts = append(parts, "new "+p.signature(construct, ": ", depth))
	}
	for _, index := range indexes {
		part := "[key: " + p.print(index.KeyType(), depth+1) + "]: " + p.print(index.ValueType(), depth+1)
		if index.IsReadonly() {
			part = "readonly " + part
		}
		parts = append(parts, part)
	}
	for _, property := range properties {
		parts = append(parts, p.member(property, depth))
	}
	if len(parts) == 0 {
		return "{}"
	}
	if len(parts) > p.maximumMembers {
		more := len(parts) - p.maximumMembers + 1
		parts = append(parts[:p.maximumMembers-1], "... "+strconv.Itoa(more)+" more ...")
	}
	return "{ " + strings.Join(parts, "; ") + "; }"
}

// members is an anonymous object's properties in their stable order: its own, declared directly in the declaration
// it came from, by position, then every other one by name. The others are the `name?: undefined` members that
// normalization adds, whose declaration is a sibling's, and which sibling is the checker's history.
//
// Own is where the declaration sits, not its place in the source (#ncz8caa). A borrowed sibling can sit in a literal
// nested inside this one, inside its range but not among its members, and then whether it read as own was the
// checker's history again, as a spread of a nested literal's member was on www-phi-health. memberHome is exact for
// every container measured there: object literals, type literals, a class's static side, namespaces and files.
func (p *stableTypePrinter) members(t *checker.Type) []*ast.Symbol {
	properties := slices.Clone(p.typeChecker.GetPropertiesOfType(t))
	var container *ast.Node
	if symbol := t.Symbol(); symbol != nil && len(symbol.Declarations) > 0 {
		container = symbol.Declarations[0]
	}
	own := func(property *ast.Symbol) (int, bool) {
		if container == nil || len(property.Declarations) == 0 || memberHome(property.Declarations[0]) != container {
			return 0, false
		}
		return property.Declarations[0].Pos(), true
	}
	slices.SortStableFunc(properties, func(a, b *ast.Symbol) int {
		positionA, ownA := own(a)
		positionB, ownB := own(b)
		switch {
		case ownA && ownB:
			return positionA - positionB
		case ownA:
			return -1
		case ownB:
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})
	return properties
}

// memberHome is the declaration a member is declared directly in: its parent, or for a namespace's and a file's
// statements the namespace or file around their block and variable statement.
func memberHome(declaration *ast.Node) *ast.Node {
	home := declaration.Parent
	for home != nil && (home.Kind == ast.KindModuleBlock || home.Kind == ast.KindVariableStatement || home.Kind == ast.KindVariableDeclarationList) {
		home = home.Parent
	}
	return home
}

// member prints one property: `readonly name?: T`, or a method as `name(p: T): R`.
//
// An optional member's type is the one TypeScript's own printer starts from, without the missing type the `?` already
// says (#sp4xwtj). Under exactOptionalPropertyTypes that is only the marker, so `y?: number` prints as written and a
// written `y?: number | undefined` keeps its undefined, and a synthesized `y?:` member prints `never`, as TypeToString
// prints it. Without the option the marker is undefined itself, `y?: number` and `y?: number | undefined` are one
// type, and it prints `| undefined`: TypeToString drops it only where it reuses a written annotation.
func (p *stableTypePrinter) member(property *ast.Symbol, depth int) string {
	name := StablePropertyName(property.Name)
	if property.Flags&ast.SymbolFlagsOptional != 0 {
		name += "?"
	}
	propertyType := checker.Checker_getNonMissingTypeOfSymbol(p.typeChecker, property)
	if property.Flags&ast.SymbolFlagsMethod != 0 {
		if signatures := p.typeChecker.GetSignaturesOfType(propertyType, checker.SignatureKindCall); len(signatures) == 1 {
			return name + p.signature(signatures[0], ": ", depth)
		}
	}
	text := name + ": " + p.print(propertyType, depth+1)
	if checker.Checker_isReadonlySymbol(p.typeChecker, property) {
		text = "readonly " + text
	}
	return text
}

// signature prints `<T>(a: A, b?: B, ...rest: C[])` followed by the arrow or colon and the return type.
func (p *stableTypePrinter) signature(signature *checker.Signature, arrow string, depth int) string {
	var text strings.Builder
	if typeParameters := signature.TypeParameters(); len(typeParameters) > 0 {
		names := make([]string, len(typeParameters))
		for index, typeParameter := range typeParameters {
			names[index] = p.typeChecker.TypeToString(typeParameter)
		}
		text.WriteString("<" + strings.Join(names, ", ") + ">")
	}
	parameters := signature.Parameters()
	printed := make([]string, len(parameters))
	for index, parameter := range parameters {
		prefix, suffix := "", ""
		switch {
		case signature.HasRestParameter() && index == len(parameters)-1:
			prefix = "..."
		case index >= signature.MinArgumentCount():
			suffix = "?"
		}
		printed[index] = prefix + parameter.Name + suffix + ": " + p.print(p.typeChecker.GetTypeOfSymbol(parameter), depth+1)
	}
	text.WriteString("(" + strings.Join(printed, ", ") + ")")
	text.WriteString(arrow + p.print(p.typeChecker.GetReturnTypeOfSignature(signature), depth+1))
	return text.String()
}

// StablePropertyName prints a property's name as a member is written: bare when it's an identifier, quoted when
// it isn't, and `[name]` for a member keyed by a symbol.
//
// typescript-go names a symbol-keyed member with the raw byte 0xFE, which is not UTF-8 so no identifier can hold
// it: `\xfe@name@id` for a unique symbol (getESSymbolLikeTypeForNode), whose id is the symbol's creation order and
// moves from run to run, and `\xfe@name` for a known symbol the library doesn't declare. It is matched as bytes
// (#z9jcxp1): a regexp reads the invalid byte as U+FFFD, and the pattern this replaced matched U+00FE, so no
// name ever matched and the id was printed.
func StablePropertyName(name string) string {
	if symbolName, ok := strings.CutPrefix(name, ast.InternalSymbolNamePrefix+"@"); ok {
		if at := strings.LastIndexByte(symbolName, '@'); at >= 0 && isDecimal(symbolName[at+1:]) {
			symbolName = symbolName[:at]
		}
		return "[" + symbolName + "]"
	}
	if !isIdentifierName(name) {
		return strconv.Quote(name)
	}
	return name
}

// identifierName is a property name TypeScript prints bare.
var identifierName = regexp.MustCompile(`^[\p{L}_$][\p{L}\p{N}_$]*$`)

func isIdentifierName(name string) bool {
	return identifierName.MatchString(name)
}

// isDecimal is a symbol id: one or more ASCII digits.
func isDecimal(text string) bool {
	return text != "" && strings.Trim(text, "0123456789") == ""
}
