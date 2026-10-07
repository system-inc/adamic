package lower

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Apple's frameworks are declared in binding files (internal/load/apple), and each declaration
// carries the Objective-C it calls in an @objc tag in its doc comment (docs/apple.md):
//
//	@objc init <selector> <arguments>                 a constructor: alloc, then the init
//	@objc static <selector> <arguments> -> <result>   a message to the class
//	@objc method <selector> <arguments> -> <result>   a message to the object
//	@objc get <selector> -> <result>                  a property's getter
//	@objc set <selector> <type>                       a property's setter
//
// Each argument is <source>:<type>, in the selector's order. The source is the Adamic argument's
// position (0), a field of an options object written at the call (1.styleMask, or 1.defer?=no
// when it may be left out), or a constant (nil, yes, no). The types are ir.NativeKind's: double,
// integer, unsigned, boolean, string, object (object? where nil is a value), rectangle,
// enum(Name=1,...), options(Name=1,...) and action, and a result may be void. A result marked new
// comes back retained, as alloc, new and copy give it.
//
// A call lowers to an ordinary ir.Call of a function whose body is foreign (ir.Foreign), made once
// for each shape of call: its parameters are the receiver, if any, then the values the call
// evaluates, in the order the source evaluates them.

// foreignTag is one parsed @objc tag.
type foreignTag struct {
	kind      string // init, static, method, get or set
	selector  string
	arguments []foreignSource
	returns   ir.NativeType
	retained  bool
}

// foreignSource is one native argument and where its value comes from.
type foreignSource struct {
	position int    // the Adamic argument, or -1 for a constant
	field    string // the options object's field, or ""
	optional bool
	value    string // a constant's or a missing optional field's native value
	native   ir.NativeType
}

// isForeign reports whether a symbol is declared in one of Apple's binding files.
func isForeign(symbol *ast.Symbol) bool {
	return symbol != nil && len(symbol.Declarations) > 0 && load.IsApple(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// foreignTags reads a declaration's @objc tags.
func (l *lowering) foreignTags(declaration *ast.Node) ([]foreignTag, error) {
	sourceFile := ast.GetSourceFileOfNode(declaration)
	tags := []foreignTag{}
	for _, documentation := range declaration.JSDoc(sourceFile) {
		text := sourceFile.Text()[documentation.Pos():documentation.End()]
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "/*"))
			if !strings.HasPrefix(line, "@objc ") {
				continue
			}
			tag, err := parseForeignTag(strings.TrimPrefix(line, "@objc "))
			if err != nil {
				return nil, fmt.Errorf("lower: %s: the binding's tag %q: %w", l.program.Where(declaration), line, err)
			}
			tags = append(tags, tag)
		}
	}
	return tags, nil
}

func parseForeignTag(text string) (foreignTag, error) {
	tag := foreignTag{}
	words := strings.Fields(text)
	if len(words) < 2 {
		return tag, errors.New("it names no selector")
	}
	tag.kind, tag.selector = words[0], words[1]
	words = words[2:]
	switch tag.kind {
	case "class":
		return tag, nil
	case "init":
		tag.retained = true
		tag.returns = ir.NativeType{Kind: ir.NativeObject}
	case "set":
		if len(words) != 1 {
			return tag, errors.New("a setter takes one type")
		}
		native, err := parseNativeType(words[0])
		if err != nil {
			return tag, err
		}
		tag.arguments = []foreignSource{{position: 0, native: native}}
		tag.returns = ir.NativeType{Kind: ir.NativeVoid}
		return tag, nil
	case "static", "method", "get":
	default:
		return tag, fmt.Errorf("%s is not a tag kind", tag.kind)
	}
	for index := 0; index < len(words); index++ {
		if words[index] != "->" {
			source, err := parseForeignSource(words[index])
			if err != nil {
				return tag, err
			}
			tag.arguments = append(tag.arguments, source)
			continue
		}
		rest := words[index+1:]
		if len(rest) == 2 && rest[0] == "new" {
			tag.retained, rest = true, rest[1:]
		}
		if len(rest) != 1 {
			return tag, errors.New("-> takes one result type")
		}
		native, err := parseNativeType(rest[0])
		if err != nil {
			return tag, err
		}
		tag.returns = native
		return tag, nil
	}
	if tag.kind != "init" {
		return tag, errors.New("it has no -> result")
	}
	return tag, nil
}

func parseForeignSource(text string) (foreignSource, error) {
	source := foreignSource{position: -1}
	place, typeText, found := strings.Cut(text, ":")
	if !found {
		return source, fmt.Errorf("argument %s has no type", text)
	}
	// A default follows the type, and an enumeration's members have = of their own inside it.
	if close := strings.LastIndex(typeText, ")"); close >= 0 {
		typeText, source.value = typeText[:close+1], strings.TrimPrefix(typeText[close+1:], "=")
	} else {
		typeText, source.value, _ = strings.Cut(typeText, "=")
	}
	native, err := parseNativeType(typeText)
	if err != nil {
		return source, err
	}
	source.native = native
	switch place {
	case "nil", "yes", "no":
		source.value = place
		return source, nil
	}
	position, field, hasField := strings.Cut(place, ".")
	if source.position, err = strconv.Atoi(position); err != nil || source.position < 0 {
		return source, fmt.Errorf("argument %s names no position", text)
	}
	if hasField {
		source.field, source.optional = strings.CutSuffix(field, "?")
	}
	if source.optional != (source.value != "") {
		return source, fmt.Errorf("argument %s: an optional field needs a default, and only it", text)
	}
	return source, nil
}

func parseNativeType(text string) (ir.NativeType, error) {
	native := ir.NativeType{}
	if name, members, isMembers := strings.Cut(strings.TrimSuffix(text, ")"), "("); isMembers {
		switch name {
		case "enum":
			native.Kind = ir.NativeEnumeration
		case "options":
			native.Kind = ir.NativeOptions
		default:
			return native, fmt.Errorf("%s is not a native type", text)
		}
		for _, member := range strings.Split(members, ",") {
			memberName, valueText, found := strings.Cut(member, "=")
			value, err := strconv.ParseInt(valueText, 0, 64)
			if !found || err != nil {
				return native, fmt.Errorf("%s: member %s has no integer value", text, member)
			}
			native.Names = append(native.Names, memberName)
			native.Values = append(native.Values, value)
		}
		return native, nil
	}
	kinds := map[string]ir.NativeKind{
		"void": ir.NativeVoid, "double": ir.NativeDouble, "integer": ir.NativeInteger, "unsigned": ir.NativeUnsigned,
		"boolean": ir.NativeBoolean, "string": ir.NativeString, "object": ir.NativeObject, "object?": ir.NativeObject,
		"rectangle": ir.NativeRectangle, "action": ir.NativeAction,
	}
	kind, isKind := kinds[text]
	if !isKind {
		return native, fmt.Errorf("%s is not a native type", text)
	}
	native.Kind, native.Nullable = kind, text == "object?"
	return native, nil
}

// adamicType is how a native value is held in Adamic.
func adamicType(native ir.NativeType) ir.Type {
	switch native.Kind {
	case ir.NativeDouble, ir.NativeInteger, ir.NativeUnsigned:
		return ir.Number
	case ir.NativeBoolean:
		return ir.Boolean
	case ir.NativeString, ir.NativeEnumeration:
		return ir.String
	case ir.NativeOptions:
		return ir.Array
	case ir.NativeAction:
		return ir.Closure
	case ir.NativeVoid:
		return 0
	}
	return ir.Object
}

// foreignTag finds a declaration's tag of a kind.
func (l *lowering) foreignTag(node *ast.Node, declaration *ast.Node, kinds ...string) (foreignTag, error) {
	tags, err := l.foreignTags(declaration)
	if err != nil {
		return foreignTag{}, err
	}
	for _, tag := range tags {
		for _, kind := range kinds {
			if tag.kind == kind {
				return tag, nil
			}
		}
	}
	return foreignTag{}, l.notYet(node, "an Apple declaration whose binding has no @objc "+strings.Join(kinds, " or ")+" tag")
}

// foreignClass is the Objective-C class a class declared in a binding file stands for.
func (l *lowering) foreignClass(node *ast.Node, symbol *ast.Symbol) (string, error) {
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindClassDeclaration {
			continue
		}
		tag, err := l.foreignTag(node, declaration, "class")
		if err != nil {
			return "", err
		}
		return tag.selector, nil
	}
	return "", l.notYet(node, "an Apple declaration that isn't a class's member")
}

// foreignCall lowers a call of a method declared in a binding file. handled is false for any other
// call.
func (l *lowering) foreignCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	method := l.symbol(callee.Name())
	if !isForeign(method) {
		return nil, false, nil
	}
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil, true, l.notYet(node, "an optional call of an Apple method")
	}
	tag, err := l.foreignTag(node, method.Declarations[0], "method", "static")
	if err != nil {
		return nil, true, err
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	value, err := l.foreignSend(node, tag, receiver, node.AsCallExpression().Arguments.Nodes)
	return value, true, err
}

// foreignProperty lowers a read of a property declared in a binding file, through its getter.
func (l *lowering) foreignProperty(node *ast.Node) (ir.Expression, bool, error) {
	property := l.symbol(node.Name())
	if !isForeign(property) || property.Flags&ast.SymbolFlagsMethod != 0 {
		return nil, false, nil
	}
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil, true, l.notYet(node, "an optional read of an Apple property")
	}
	tag, err := l.foreignTag(node, property.Declarations[0], "get")
	if err != nil {
		return nil, true, err
	}
	value, err := l.foreignSend(node, tag, node.AsPropertyAccessExpression().Expression, nil)
	return value, true, err
}

// foreignSetProperty lowers object.name = value for a property declared in a binding file, through
// its setter.
func (l *lowering) foreignSetProperty(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, bool, error) {
	property := l.symbol(target.Name())
	if !isForeign(property) {
		return nil, false, nil
	}
	tag, err := l.foreignTag(target, property.Declarations[0], "set")
	if err != nil {
		return nil, true, err
	}
	value, err := l.foreignSend(target, tag, target.AsPropertyAccessExpression().Expression, []*ast.Node{valueNode})
	if err != nil {
		return nil, true, err
	}
	return []ir.Statement{ir.Evaluate{Value: value}}, true, nil
}

// foreignUpdateProperty refuses object.name += value on an Apple property, for now: it reads and
// writes through two messages, and stage 0 doesn't compose them yet.
func (l *lowering) foreignUpdateProperty(target *ast.Node) error {
	if !isForeign(l.symbol(target.Name())) {
		return nil
	}
	return l.notYet(target, "a compound assignment to an Apple property (write object.name = object.name + value)")
}

// foreignNew lowers new Class(...) for a class declared in a binding file.
func (l *lowering) foreignNew(node *ast.Node) (ir.Expression, bool, error) {
	created := node.AsNewExpression()
	class := l.symbol(ast.SkipParentheses(created.Expression))
	if !isForeign(class) {
		return nil, false, nil
	}
	signature := l.checker.GetResolvedSignature(node)
	if signature == nil || signature.Declaration() == nil || signature.Declaration().Kind != ast.KindConstructor {
		return nil, true, l.notYet(node, "an Apple class constructed without a declared constructor")
	}
	tag, err := l.foreignTag(node, signature.Declaration(), "init", "static")
	if err != nil {
		return nil, true, err
	}
	arguments := []*ast.Node{}
	if created.Arguments != nil {
		arguments = created.Arguments.Nodes
	}
	value, err := l.foreignSend(node, tag, created.Expression, arguments)
	return value, true, err
}

// foreignSend lowers one message: receiver is the object or the class it's sent to, and arguments
// the call's, evaluated in the order they're written, an options object's fields in the order the
// literal writes them.
func (l *lowering) foreignSend(node *ast.Node, tag foreignTag, receiver *ast.Node, arguments []*ast.Node) (ir.Expression, error) {
	foreign := &ir.Foreign{Selector: tag.selector, Returns: tag.returns}
	values := []ir.Expression{}
	types := []ir.Type{}
	shape := []string{}
	receiverSymbol := l.symbol(ast.SkipParentheses(receiver))
	switch {
	case tag.kind == "init":
		foreign.Kind = ir.Construct
	case receiverSymbol != nil && receiverSymbol.Flags&ast.SymbolFlagsClass != 0 && isForeign(receiverSymbol):
		foreign.Kind = ir.ClassMessage
	default:
		foreign.Kind = ir.InstanceMessage
		value, err := l.expression(receiver)
		if err != nil {
			return nil, err
		}
		if value.Type() != ir.Object {
			return nil, l.notYet(receiver, "an Apple message to a value held as something other than an object")
		}
		values, types, shape = append(values, value), append(types, ir.Object), append(shape, "receiver")
	}
	if foreign.Kind != ir.InstanceMessage {
		class, err := l.foreignClass(node, l.foreignReceiverClass(receiverSymbol, receiver))
		if err != nil {
			return nil, err
		}
		foreign.Class = class
	}

	// Where each native argument's value is: a parameter, by the source's position and field.
	parameters := map[string]int{}
	used := map[int]bool{}
	for _, source := range tag.arguments {
		if source.position >= 0 {
			used[source.position] = true
		}
	}
	for position, argument := range arguments {
		if !used[position] {
			return nil, l.notYet(argument, "an argument the Apple binding doesn't place")
		}
		fielded := false
		for _, source := range tag.arguments {
			fielded = fielded || (source.position == position && source.field != "")
		}
		if !fielded {
			value, err := l.expression(argument)
			if err != nil {
				return nil, err
			}
			parameters[strconv.Itoa(position)] = len(values)
			values, types, shape = append(values, value), append(types, value.Type()), append(shape, strconv.Itoa(position))
			continue
		}
		literal := ast.SkipParentheses(argument)
		if literal.Kind != ast.KindObjectLiteralExpression {
			return nil, l.notYet(argument, "an Apple options object that isn't written at the call")
		}
		for _, property := range literal.AsObjectLiteralExpression().Properties.Nodes {
			var initializer *ast.Node
			switch property.Kind {
			case ast.KindPropertyAssignment:
				initializer = property.AsPropertyAssignment().Initializer
			case ast.KindShorthandPropertyAssignment:
				initializer = property.Name()
			default:
				return nil, l.notYet(property, "an Apple options object with "+describe(property))
			}
			if !ast.IsIdentifier(property.Name()) {
				return nil, l.notYet(property, "an Apple option named by anything but an identifier")
			}
			value, err := l.expression(initializer)
			if err != nil {
				return nil, err
			}
			key := strconv.Itoa(position) + "." + property.Name().Text()
			parameters[key] = len(values)
			values, types, shape = append(values, value), append(types, value.Type()), append(shape, key)
		}
	}
	for _, source := range tag.arguments {
		argument := ir.ForeignArgument{Type: source.native, Parameter: -1, Default: source.value}
		if source.position >= 0 {
			key := strconv.Itoa(source.position)
			if source.field != "" {
				key += "." + source.field
			}
			parameter, given := parameters[key]
			switch {
			case given:
				argument.Parameter = parameter
				if want := adamicType(source.native); types[parameter] != want {
					return nil, l.notYet(node, fmt.Sprintf("an Apple argument %s held as %v where the binding takes %v", key, types[parameter], want))
				}
			case !source.optional:
				return nil, l.notYet(node, "an Apple call that leaves out "+key)
			}
		}
		foreign.Arguments = append(foreign.Arguments, argument)
	}
	if foreign.Kind == ir.InstanceMessage {
		// The receiver is the first parameter, and every argument comes after it.
		foreign.Arguments = append([]ir.ForeignArgument{{Type: ir.NativeType{Kind: ir.NativeObject}, Parameter: 0}}, foreign.Arguments...)
	}
	switch tag.returns.Kind {
	case ir.NativeRectangle, ir.NativeEnumeration, ir.NativeOptions, ir.NativeAction:
		return nil, l.notYet(node, "an Apple result held as a "+tag.selector+" gives it (rectangles, enumerations, options and actions come back later)")
	}
	returns := adamicType(tag.returns)
	if tag.kind == "init" || (tag.kind == "static" && node.Kind == ast.KindNewExpression) {
		returns = ir.Object
		if tag.returns.Nullable {
			return nil, l.notYet(node, "an Apple constructor that may give nil")
		}
	}
	foreign.Retained = tag.retained
	function := l.foreignFunction(foreign, types, returns, strings.Join(shape, ","))
	return ir.Call{Function: function, Arguments: values, Returns: returns}, nil
}

// foreignReceiverClass is the class symbol a class message goes to: the receiver's, or for new, the
// class constructed.
func (l *lowering) foreignReceiverClass(symbol *ast.Symbol, receiver *ast.Node) *ast.Symbol {
	if symbol != nil && symbol.Flags&ast.SymbolFlagsClass != 0 {
		return symbol
	}
	return l.symbol(ast.SkipParentheses(receiver))
}

// foreignFunction is the function a foreign call calls, made once for each class, selector and
// shape of call.
func (l *lowering) foreignFunction(foreign *ir.Foreign, types []ir.Type, returns ir.Type, shape string) int {
	key := fmt.Sprintf("%d %s %s %s", foreign.Kind, foreign.Class, foreign.Selector, shape)
	if l.foreignFunctions == nil {
		l.foreignFunctions = map[string]int{}
	}
	if function, made := l.foreignFunctions[key]; made {
		return function
	}
	function := len(l.result.Functions)
	name := "apple_" + strings.ReplaceAll(foreign.Class+"_"+foreign.Selector, ":", "_")
	declared := ir.Function{Name: name, Returns: returns, Foreign: foreign}
	for index, of := range types {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: fmt.Sprintf("argument%d", index), Type: of, Function: function, Borrowed: of.IsReference()})
		declared.Parameters = append(declared.Parameters, local)
	}
	// The body a backend runs when it can't make the call: JavaScript can't reach AppKit.
	message := l.constant("an Apple framework call needs the native backend on an Apple platform: " + foreign.Class + " " + foreign.Selector)
	declared.Body = []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: message}}}
	switch returns {
	case ir.Number:
		declared.Body = append(declared.Body, ir.Return{Value: ir.NumberConstant{}})
	case ir.Boolean:
		declared.Body = append(declared.Body, ir.Return{Value: ir.BooleanConstant{}})
	case ir.String:
		declared.Body = append(declared.Body, ir.Return{Value: ir.StringConstant{Index: l.constant("")}})
	case ir.Object:
		declared.Body = append(declared.Body, ir.Return{Value: ir.Undefined{Of: ir.Object}})
	}
	l.result.Functions = append(l.result.Functions, declared)
	l.foreignFunctions[key] = function
	return function
}
