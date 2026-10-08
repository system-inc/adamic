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
	ast.KindYieldExpression:   {"yield (generators)", "build an array, or call a function per item"},
	ast.KindDecorator:         {"a decorator", "write the behavior where it applies; 0.1 doesn't rewrite classes at runtime"},
	ast.KindWithStatement:     {"with", "name the object you mean"},
	ast.KindDeleteExpression:  {"delete", "an object's shape is fixed; use a Map for keys that come and go"},
	ast.KindDebuggerStatement: {"debugger", "remove it"},
	ast.KindModuleDeclaration: {"a namespace", "use a module: a file of its own, with named exports"},
	ast.KindVoidExpression:    {"the void operator", "evaluate the expression as a statement"},
	ast.KindIndexSignature:    {"an index signature", "use a Map, which keeps keys in the order they were added"},
	ast.KindExportAssignment:  {"export default", "export by name: one name for one thing"},
	ast.KindNonNullExpression: {"the non-null assertion !", "write ?? panic('why it can't be missing'), or narrow and handle the missing case"},
}

// refusedOperators are binary operators 0.1 refuses.
var refusedOperators = map[ast.Kind]refusal{
	ast.KindEqualsEqualsToken:             {"==", "use ===, which doesn't coerce"},
	ast.KindExclamationEqualsToken:        {"!=", "use !==, which doesn't coerce"},
	ast.KindInKeyword:                     {"in", "an object's shape is known; use a discriminant, or a Map"},
	ast.KindCommaToken:                    {"the comma operator", "write each expression as its own statement"},
	ast.KindAmpersandAmpersandEqualsToken: {"&&=", "write the if"},
	ast.KindBarBarEqualsToken:             {"||=", "write the if"},
}

// refuse walks a module for what 0.1 refuses and returns the first, with where it is and the fix.
func (l *lowering) refuse(module *ast.SourceFile) error {
	if err := l.parallelPreflight(module); err != nil {
		return err
	}
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
	if err := l.parallelPreflight(module); err != nil {
		return err
	}
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if refused, isRefused := refusals[node.Kind]; isRefused {
			found = &Refused{Where: l.program.Where(node), What: refused.what, Fix: refused.fix}
			return true
		}
		if l.isJSONSchemaCall(node) {
			_, found = l.jsonSchemaCallType(node)
			if found != nil {
				return true
			}
		}
		if err := l.typedArrayUnsupported(node); err != nil {
			found = err
			return true
		}
		if node.Kind == ast.KindTypePredicate {
			if err := l.provePredicate(node); err != nil {
				found = err
				return true
			}
		}
		if err := l.enumRefusal(node); err != nil {
			found = err
			return true
		}
		checkedCast := false
		if node.Kind == ast.KindAsExpression {
			proof, err := l.castProof(node)
			if err != nil {
				found = err
				return true
			}
			checkedCast = len(proof.allowed) > 0 || len(proof.classes) > 0
		}
		var assertion *ast.Node
		if node.Kind == ast.KindPropertyDeclaration {
			if token := node.PostfixToken(); token != nil && token.Kind == ast.KindExclamationToken {
				assertion = token
			}
		}
		if node.Kind == ast.KindVariableDeclaration {
			assertion = node.AsVariableDeclaration().ExclamationToken
		}
		if assertion != nil {
			found = &Refused{Where: l.program.Where(assertion), What: "a definite assignment assertion !", Fix: "remove ! and initialize it where it is declared or in the constructor, or type it T | undefined"}
			return true
		}
		if node.Kind == ast.KindBinaryExpression {
			if refused, isRefused := refusedOperators[node.AsBinaryExpression().OperatorToken.Kind]; isRefused {
				found = &Refused{Where: l.program.Where(node.AsBinaryExpression().OperatorToken), What: refused.what, Fix: refused.fix}
				return true
			}
		}
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
			found = &Refused{Where: l.program.Where(node), What: "a generator function", Fix: "use an explicit iterator object; suspended frames need ownership and cancellation rules before generators can be compiled without a collector (docs/user-iterators.md)"}
			return true
		}
		// Covered async syntax is lowered by async.go. Unsupported lifecycle and Promise
		// operations get specific NotYet there; permanent refusals above still apply.

		if node.Kind == ast.KindIdentifier && node.Text() == "arguments" {
			// JavaScript's arguments object, not a variable the program named arguments.
			if symbol := l.checker.GetSymbolAtLocation(node); symbol != nil && len(symbol.Declarations) == 0 {
				found = &Refused{Where: l.program.Where(node), What: "arguments", Fix: "name the parameters, or take a rest parameter"}
				return true
			}
		}
		if node.Kind == ast.KindPropertyAccessExpression && !called(node) && !l.libraryNumberBoundMethod(node) && !l.stringMethodObservation(node) && !l.libraryArrayObservedMethod(node) {
			// A method read as a value loses its object: this is undefined when it's called.
			access := node.AsPropertyAccessExpression()
			if access.Name().Text() == "isPrototypeOf" && l.libraryMember(node) {
				found = l.prototypeRead(node, "isPrototypeOf")
				return true
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
				found = &Refused{Where: l.program.Where(node), What: "a method read as a value (" + symbol.Name + " would lose its object, and this with it)", Fix: "call it in an arrow that keeps the object: (" + parameters + ") => " + object + "." + symbol.Name + "(" + parameters + ") (unbound-method)"}
				return true
			}
		}
		if node.Kind == ast.KindClassDeclaration {
			if err := l.checkOverrides(node, l.checker.GetTypeAtLocation(node.Name())); err != nil {
				found = err
				return true
			}
		}
		// Prefer the writable-slot explanation when both a mutable view and nominal ancestry fail.
		// A checked cast relates only members selected by its tag, not excluded source members.
		if !checkedCast {
			if err := l.refuseStringWidening(node); err != nil {
				found = err
				return true
			}
		}
		if err := l.refuseOptionalWidening(node); err != nil {
			found = err
			return true
		}
		if err := l.classViewRefusal(node); err != nil {
			found = err
			return true
		}
		node.ForEachChild(visit)
		return false
	}
	module.AsNode().ForEachChild(visit)
	return found
}

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
