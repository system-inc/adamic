package lower

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A delegate is an object of one of the program's classes handed to Apple where it takes an object
// (docs/apple.md, "Delegates"):
//
//	class Keeper implements WindowDelegate {
//		windowShouldClose(sender: Window): boolean { ... }
//	}
//	window.delegate = new Keeper();
//
// The class names the Apple protocols it implements, and each protocol's members carry an @objc
// implement tag. Apple is handed an instance of an Objective-C class made at runtime for the class
// (ir.Delegate), which holds the object, counted, and whose method for each protocol member the
// class has calls a function made here: ordinary Adamic, its parameters the object and what Apple
// hands it, and its body a call of the class's method, virtual, as any call through the class is.
// So every analysis sees the class's method called as code. Apple calls that function, and nothing
// in the program does: the object has escaped to Apple, as any object handed to a foreign call has.

// delegateHold is a delegate handed to Apple: the delegate's type, and the type of the Apple object
// that holds it from then on, the receiver of the message or what the message makes.
type delegateHold struct {
	delegate, holder *checker.Type
	node             *ast.Node
}

// foreignDelegate is the delegate an argument makes, when its value is an object of one of the
// program's classes; isDelegate is false for anything else, which crosses as Apple's own object.
func (l *lowering) foreignDelegate(argument *ast.Node) (*ir.Delegate, bool, error) {
	if argument == nil {
		return nil, false, nil
	}
	proven := l.concrete(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(argument)))
	if !isClassInstance(proven) || proven.Symbol() == nil {
		return nil, false, nil
	}
	declaration, isClass := l.classes[proven.Symbol()]
	if !isClass {
		return nil, false, nil
	}
	lowered, err := l.instantiate(declaration, proven, argument)
	if err != nil {
		return nil, true, err
	}
	if l.foreignDelegates == nil {
		l.foreignDelegates = map[*instance]*ir.Delegate{}
	}
	if delegate, made := l.foreignDelegates[lowered]; made {
		return delegate, true, nil
	}
	protocols := l.appleProtocols(declaration)
	if len(protocols) == 0 {
		return nil, true, l.notYet(argument, "an object of the program's own class handed to Apple without an Apple protocol it implements (write class "+declaration.Name().Text()+" implements WindowDelegate, or the protocol Apple takes)")
	}
	className := declaration.Name().Text()
	delegate := &ir.Delegate{Name: "AdamicDelegate_" + className + "_" + strconv.Itoa(len(l.foreignDelegates))}
	selectors := map[string]bool{}
	for _, protocol := range protocols {
		tag, err := l.foreignTag(argument, protocol, "protocol")
		if err != nil {
			return nil, true, err
		}
		delegate.Protocols = append(delegate.Protocols, tag.selector)
		for _, member := range protocol.AsInterfaceDeclaration().Members.Nodes {
			if err := l.delegateMethod(argument, delegate, selectors, lowered, proven, member); err != nil {
				return nil, true, err
			}
		}
	}
	l.foreignDelegates[lowered] = delegate
	return delegate, true, nil
}

// appleProtocols are the binding files' protocols a class says it implements, its own and its
// bases'.
func (l *lowering) appleProtocols(declaration *ast.Node) []*ast.Node {
	protocols := []*ast.Node{}
	for declaration != nil {
		var base *ast.Node
		for _, clause := range nodesOf(declaration.AsClassDeclaration().HeritageClauses) {
			for _, written := range clause.AsHeritageClause().Types.Nodes {
				// An extends clause is an expression; an implements clause names a type.
				var named *ast.Node
				switch written.Kind {
				case ast.KindExpressionWithTypeArguments:
					named = ast.SkipParentheses(written.AsExpressionWithTypeArguments().Expression)
				case ast.KindTypeReference:
					named = written.AsTypeReferenceNode().TypeName
				default:
					continue
				}
				symbol := l.symbol(named)
				if symbol == nil || len(symbol.Declarations) == 0 {
					continue
				}
				if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
					base = l.classes[symbol]
					continue
				}
				if found := symbol.Declarations[0]; isForeign(symbol) && found.Kind == ast.KindInterfaceDeclaration {
					protocols = append(protocols, found)
				}
			}
		}
		declaration = base
	}
	return protocols
}

// delegateMethod adds to a delegate the method one protocol member makes, when the class has it.
func (l *lowering) delegateMethod(argument *ast.Node, delegate *ir.Delegate, selectors map[string]bool, lowered *instance, proven *checker.Type, member *ast.Node) error {
	name := member.Name()
	if name == nil || !ast.IsIdentifier(name) {
		return nil
	}
	required := member.PostfixToken() == nil || member.PostfixToken().Kind != ast.KindQuestionToken
	var implement *foreignTag
	if member.Kind == ast.KindMethodSignature {
		tags, err := l.foreignTags(member)
		if err != nil {
			return err
		}
		for index := range tags {
			if tags[index].kind == "implement" {
				implement = &tags[index]
			}
		}
	}
	property := l.checker.GetPropertyOfType(proven, name.Text())
	if property == nil {
		if required {
			return l.notYet(argument, "a delegate missing "+name.Text()+", which its protocol requires")
		}
		return nil
	}
	if implement == nil {
		return l.notYet(argument, "a delegate implementing "+name.Text()+", which a program's class can't implement for Apple yet")
	}
	if len(property.Declarations) == 0 || property.Declarations[0].Kind != ast.KindMethodDeclaration || ast.HasSyntacticModifier(property.Declarations[0], ast.ModifierFlagsStatic) {
		return l.notYet(argument, "a delegate whose "+name.Text()+" isn't one of its class's methods")
	}
	if selectors[implement.selector] {
		return nil
	}
	selectors[implement.selector] = true
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(property), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return l.notYet(argument, "a delegate whose "+name.Text()+" is overloaded")
	}
	parameters := signatures[0].Parameters()
	natives := []ir.NativeType{}
	for _, source := range implement.arguments {
		natives = append(natives, source.native)
	}
	if len(parameters) > len(natives) {
		return l.notYet(argument, fmt.Sprintf("a delegate's %s taking %d parameters where Apple calls it with %d", name.Text(), len(parameters), len(natives)))
	}
	for index, parameter := range parameters {
		held, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		if want := adamicType(natives[index]); !known || held != want {
			return l.notYet(argument, fmt.Sprintf("a delegate's %s whose parameter %s isn't held as Apple hands it", name.Text(), parameter.Name))
		}
	}
	returns := adamicType(implement.returns)
	if returns != 0 {
		held, known := l.representation(l.checker.GetReturnTypeOfSignature(signatures[0]))
		if !known || held != returns {
			return l.notYet(argument, fmt.Sprintf("a delegate's %s whose result isn't held as Apple takes it", name.Text()))
		}
	}

	// The function Apple's call reaches: the object, then what Apple hands it.
	method := l.fieldName(property.Declarations[0].Name())
	function := len(l.result.Functions)
	declared := ir.Function{Name: "apple_" + delegate.Name + "_" + strings.ReplaceAll(implement.selector, ":", "_"), Returns: returns}
	self := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "self", Type: ir.Object, Function: function, Borrowed: true})
	declared.Parameters = append(declared.Parameters, self)
	call := ir.Call{Function: lowered.methods[method], Arguments: []ir.Expression{ir.Read{Local: self, Of: ir.Object}}, Returns: l.result.Functions[lowered.methods[method]].Returns, Virtual: lowered.slots[method] + 1}
	for index, native := range natives {
		local := len(l.result.Locals)
		held := adamicType(native)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: fmt.Sprintf("argument%d", index), Type: held, Function: function, Borrowed: held.IsReference()})
		declared.Parameters = append(declared.Parameters, local)
		if index < len(parameters) {
			call.Arguments = append(call.Arguments, ir.Read{Local: local, Of: held})
		}
	}
	if returns == 0 {
		declared.Body = []ir.Statement{ir.Evaluate{Value: call}}
	} else {
		declared.Body = []ir.Statement{ir.Return{Value: call}}
	}
	l.result.Functions = append(l.result.Functions, declared)
	delegate.Methods = append(delegate.Methods, ir.DelegateMethod{Selector: implement.selector, Function: function, Parameters: natives, Returns: implement.returns})
	return nil
}

// foreignClassName is the Objective-C class a type is declared as in a binding file, or "".
func (l *lowering) foreignClassName(proven *checker.Type) string {
	symbol := proven.Symbol()
	if symbol == nil || !isForeign(symbol) || len(symbol.Declarations) == 0 || symbol.Declarations[0].Kind != ast.KindClassDeclaration {
		return ""
	}
	tags, err := l.foreignTags(symbol.Declarations[0])
	if err != nil {
		return ""
	}
	for _, tag := range tags {
		if tag.kind == "class" {
			return tag.selector
		}
	}
	return ""
}

// foreignLeaf reports whether a type is an Apple class the leaf table names (load.AppleLeaves): one
// whose instances hold strong references only to other leaves, which the cycle finder doesn't walk
// through.
func (l *lowering) foreignLeaf(proven *checker.Type) bool {
	return l.appleLeaf(l.foreignClassName(proven))
}

// appleLeaf reports whether the leaf table names an Objective-C class.
func (l *lowering) appleLeaf(class string) bool {
	if class == "" {
		return false
	}
	leaves, err := load.AppleLeaves()
	return err == nil && leaves[class]
}

// appleType is the type of the class or protocol a binding file the program loaded declares, by its
// Objective-C name ("class NSURL", "protocol NSWindowDelegate"), or nil when no loaded file declares
// it: then nothing of the program's is one, though Apple's objects may be.
func (l *lowering) appleType(name string) *checker.Type {
	if l.appleTypes == nil {
		l.appleTypes = map[string]*checker.Type{}
		for _, sourceFile := range l.program.AppleFiles() {
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if (node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindInterfaceDeclaration) && node.Name() != nil {
					if tags, err := l.foreignTags(node); err == nil {
						for _, tag := range tags {
							if tag.kind == "class" || tag.kind == "protocol" {
								l.appleTypes[tag.kind+" "+tag.selector] = l.checker.GetTypeAtLocation(node.Name())
							}
						}
					}
				}
				return node.ForEachChild(visit)
			}
			sourceFile.AsNode().ForEachChild(visit)
		}
	}
	return l.appleTypes[name]
}

// extendsApple refuses a class whose base is one of Apple's: Apple's classes are declared, not
// lowered, and a program's class implements Apple's protocols but doesn't extend Apple's classes
// yet. Until it can, no class of the program's can be seen as one of Apple's leaves, which is what
// lets the cycle finder take a leaf's own fields as all it holds.
func (l *lowering) extendsApple(where *ast.Node, base *ast.Symbol) error {
	if !isForeign(base) {
		return nil
	}
	return l.notYet(where, "a class extending Apple's "+base.Name+" (a program's class can implement an Apple protocol)")
}
