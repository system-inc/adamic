package lower

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// refusal is one construct Adamic 0.1 doesn't allow (docs/0.1.md, "What's refused in 0.1"), and the
// fix to offer for it.
type refusal struct {
	what string
	fix  string
}

// refusals by syntax kind. Each is checked before lowering, so a program learns it has written
// something 0.1 refuses for good, never that stage 0 hasn't got to it yet.
var refusals = map[ast.Kind]refusal{
	ast.KindNonNullExpression: {"the non-null assertion !", "write ?? panic('why it can't be missing'), or narrow and handle the missing case"},
	ast.KindYieldExpression:   {"yield (generators)", "build an array, or call a function per item"},
	ast.KindDecorator:         {"a decorator", "write the behavior where it applies; 0.1 doesn't rewrite classes at runtime"},
	ast.KindWithStatement:     {"with", "name the object you mean"},
	ast.KindExportAssignment:  {"export default", "export by name: one name for one thing"},
}

// refusedOperators are binary operators 0.1 refuses.
var refusedOperators = map[ast.Kind]refusal{
	ast.KindEqualsEqualsToken:      {"==", "use ===, which doesn't coerce"},
	ast.KindExclamationEqualsToken: {"!=", "use !==, which doesn't coerce"},
}

// refuse walks a module for what 0.1 refuses and returns the first, with where it is and the fix.
func (l *lowering) refuse(module *ast.SourceFile) error {
	// Use the parser's directives, which also recognize the block forms honored by the checker.
	// Text in a string or a prose comment never enters this list.
	if len(module.CommentDirectives) > 0 {
		directive := module.CommentDirectives[0]
		name := "@ts-ignore"
		if directive.Kind == ast.CommentDirectiveKindExpectError {
			name = "@ts-expect-error"
		}
		line, column := scanner.GetLineAndCharacterOfPosition(module, directive.Loc.Pos())
		return &Refused{Where: fmt.Sprintf("%s:%d:%d", l.program.FileName(module), line+1, column+1), What: name + " suppression directive", Fix: "remove it and fix the type error"}
	}
	// File-level checking pragmas are separate from line-suppression directives. Use every
	// parsed pragma, including one overridden by a later pragma, rather than just CheckJsDirective.
	for _, pragma := range module.Pragmas {
		if pragma.Name == "ts-nocheck" || pragma.Name == "ts-check" {
			line, column := scanner.GetLineAndCharacterOfPosition(module, pragma.Pos())
			return &Refused{Where: fmt.Sprintf("%s:%d:%d", l.program.FileName(module), line+1, column+1), What: "@" + pragma.Name + " checking pragma", Fix: "remove it and fix any type errors"}
		}
	}
	// After the directives and pragmas, which name a line the checker already honors.
	if err := l.referenceRefusal(module); err != nil {
		return err
	}
	if err := l.parallelPreflight(module); err != nil {
		return err
	}
	if err := l.mergedFieldRulings(module); err != nil {
		return err
	}
	if err := l.functionAnnotations(module); err != nil {
		return err
	}
	// Before any call's own contract: a read narrowed on the false side of a call through a
	// predicate value is stopped at the read, naming the call (#vdt9f1q).
	if err := l.predicateFalseReads(module); err != nil {
		return err
	}
	// Validate arguments before visiting their annotations, so a failed contract
	// names the actual argument and parameter even for an inline arrow.
	var contractError error
	var contracts ast.Visitor
	contracts = func(node *ast.Node) bool {
		if contractError != nil {
			return true
		}
		if node.Kind == ast.KindCallExpression {
			contractError = l.predicateArguments(node)
		}
		if contractError == nil {
			node.ForEachChild(contracts)
		}
		return contractError != nil
	}
	module.AsNode().ForEachChild(contracts)
	if contractError != nil {
		return contractError
	}
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if branch := l.literalCallableBranch(node); branch != nil {
			return visit(branch)
		}
		// Each rule registers from its own file (refusal_rules.go); the first that refuses a node wins.
		for _, rule := range refusalRules.rules {
			prune, err := rule.check(l, node)
			if err != nil {
				found = err
				return true
			}
			if prune {
				return false
			}
		}
		node.ForEachChild(visit)
		return false
	}
	module.AsNode().ForEachChild(visit)
	return found
}

// syntaxRefusal refuses the constructs refusals names by kind.
func (l *lowering) syntaxRefusal(node *ast.Node) error {
	// The index signature clause is dead, since refusals names no index signature; it's kept for order.
	if refused, isRefused := refusals[node.Kind]; isRefused && !(node.Kind == ast.KindIndexSignature && l.enumerationIndexSignature(node)) && (node.Kind != ast.KindNonNullExpression || !l.checkedAssertionSource(node)) {
		return &Refused{Where: l.program.Where(node), What: refused.what, Fix: refused.fix}
	}
	return nil
}

func init() { refusalRules.register(1400, "syntaxRefusal", (*lowering).syntaxRefusal) }

// operatorRefusal refuses the binary operators refusedOperators names, at the operator.
func (l *lowering) operatorRefusal(node *ast.Node) error {
	if node.Kind == ast.KindBinaryExpression {
		// The in clause is dead, since refusedOperators doesn't name in; it's kept for order.
		if refused, isRefused := refusedOperators[node.AsBinaryExpression().OperatorToken.Kind]; isRefused && !(node.AsBinaryExpression().OperatorToken.Kind == ast.KindInKeyword && l.recordElement(l.checker.GetTypeAtLocation(node.AsBinaryExpression().Right)) != nil) {
			return &Refused{Where: l.program.Where(node.AsBinaryExpression().OperatorToken), What: refused.what, Fix: refused.fix}
		}
	}
	return nil
}

func init() { refusalRules.register(2300, "operatorRefusal", (*lowering).operatorRefusal) }

// Async generators have a separate ruling from synchronous generators.
// Keep this refusal when synchronous generator lowering removes its check below.
func (l *lowering) asyncGeneratorRefusal(node *ast.Node) error {
	if asyncGenerator(node) {
		return &Refused{Where: l.program.Where(node), What: "an async generator function", Fix: "use synchronous iteration; async generators need a separate ownership and cancellation ruling"}
	}
	return nil
}

func init() { refusalRules.register(2400, "asyncGeneratorRefusal", (*lowering).asyncGeneratorRefusal) }

// generatorRefusal refuses a synchronous generator function. Covered async syntax is lowered by
// async.go. Unsupported lifecycle and Promise operations get specific NotYet there; permanent
// refusals here still apply.
func (l *lowering) generatorRefusal(node *ast.Node) error {
	generator := false
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		generator = node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		generator = node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		generator = node.AsMethodDeclaration().AsteriskToken != nil
	}
	if generator {
		return &Refused{Where: l.program.Where(node), What: "a generator function", Fix: "use an explicit iterator object; suspended frames need ownership and cancellation rules before generators can be compiled without a collector (docs/user-iterators.md)"}
	}
	return nil
}

func init() { refusalRules.register(2500, "generatorRefusal", (*lowering).generatorRefusal) }

// shorthandLibraryMethodGap stops a shorthand property whose variable holds a library method value.
func (l *lowering) shorthandLibraryMethodGap(node *ast.Node) error {
	if node.Kind == ast.KindShorthandPropertyAssignment {
		if symbol := l.checker.GetShorthandAssignmentValueSymbol(node); symbol != nil && len(symbol.Declarations) == 1 && symbol.Declarations[0].Kind == ast.KindVariableDeclaration {
			if initializer := symbol.Declarations[0].AsVariableDeclaration().Initializer; initializer != nil {
				if _, known := l.libraryMethod(initializer, map[*ast.Symbol]bool{}); known {
					return l.notYet(node, "a library method value outside a const alias (an object field erases its receiver and callable ABI); use an arrow")
				}
			}
		}
	}
	return nil
}

func init() {
	refusalRules.register(2800, "shorthandLibraryMethodGap", (*lowering).shorthandLibraryMethodGap)
}

// methodValueRefusal refuses a method read as a value, which loses its object.
func (l *lowering) methodValueRefusal(node *ast.Node) error {
	if node.Kind == ast.KindPropertyAccessExpression && !called(node) && !l.libraryNumberBoundMethod(node) && !l.stringMethodObservation(node) && !l.libraryArrayObservedMethod(node) && !l.libraryMethodReadAllowed(node) && !l.detachedOwnMethod(node) {
		// A method read as a value loses its object: this is undefined when it's called.
		access := node.AsPropertyAccessExpression()
		if access.Name().Text() == "isPrototypeOf" && l.libraryMember(node) {
			return l.prototypeRead(node, "isPrototypeOf")
		}
		// Math.random has its own refusal, called or not.
		isRandom := l.isLibraryGlobal(access.Expression, "Math") && access.Name().Text() == "random"
		if symbol := l.checker.GetSymbolAtLocation(node); symbol != nil && symbol.Flags&ast.SymbolFlagsMethod != 0 && !isRandom {
			object := "its object"
			if ast.IsIdentifier(access.Expression) || access.Expression.Kind == ast.KindThisKeyword {
				object = scannedText(access.Expression)
			}
			parameters := "value"
			if signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(symbol), checker.SignatureKindCall); len(signatures) > 0 {
				names := []string{}
				for _, parameter := range signatures[0].Parameters() {
					names = append(names, parameter.Name)
				}
				parameters = strings.Join(names, ", ")
			}
			return &Refused{Where: l.program.Where(node), What: "a method read as a value (" + symbol.Name + " would lose its object, and this with it)", Fix: "call it in an arrow that keeps the object: (" + parameters + ") => " + object + "." + symbol.Name + "(" + parameters + ") (unbound-method)"}
		}
	}
	return nil
}

func init() { refusalRules.register(2900, "methodValueRefusal", (*lowering).methodValueRefusal) }

// called reports whether a property access is what a call calls, through any parentheses around it:
// a method called on its object, which keeps it.
func called(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node.Parent != nil && node.Parent.Kind == ast.KindCallExpression && node.Parent.AsCallExpression().Expression == node
}

// scannedText is an identifier's name, or this.
func scannedText(node *ast.Node) string {
	if node.Kind == ast.KindThisKeyword {
		return "this"
	}
	return node.Text()
}

// asyncGenerator recognizes each function form independently of the synchronous generator gate.
func asyncGenerator(node *ast.Node) bool {
	if !ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

// dynamicImportGap runs before declaration lowering: the checker can already
// expose types from a dynamically imported module, which has no static route.
func (l *lowering) dynamicImportGap(module *ast.SourceFile) error {
	if module.AsNode().Flags&ast.NodeFlagsPossiblyContainsDynamicImport == 0 {
		return nil
	}
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression && node.AsCallExpression().Expression.Kind == ast.KindImportKeyword {
			found = l.notYet(node, "a dynamic import()")
			return true
		}
		return node.ForEachChild(visit)
	}
	module.AsNode().ForEachChild(visit)
	return found
}
