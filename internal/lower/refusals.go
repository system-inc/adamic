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
	ast.KindAnyKeyword:        {"any", "name the proven type, or use unknown and narrow it"},
	ast.KindAwaitExpression:   {"await", "0.1 has no async; it arrives with the concurrency model (adamic/no-async)"},
	ast.KindYieldExpression:   {"yield (generators)", "build an array, or call a function per item (adamic/no-generators)"},
	ast.KindDecorator:         {"a decorator", "write the behavior where it applies; 0.1 doesn't rewrite classes at runtime (adamic/no-decorators)"},
	ast.KindWithStatement:     {"with", "name the object you mean (adamic/no-with)"},
	ast.KindDebuggerStatement: {"debugger", "remove it (adamic/no-debugger)"},
	ast.KindExportAssignment:  {"export default", "export by name: one name for one thing (adamic/named-declaration-export)"},
}

// refusedOperators are binary operators 0.1 refuses.
var refusedOperators = map[ast.Kind]refusal{
	ast.KindEqualsEqualsToken:      {"==", "use ===, which doesn't coerce (adamic/strict-equality)"},
	ast.KindExclamationEqualsToken: {"!=", "use !==, which doesn't coerce (adamic/strict-equality)"},
}

// refuse walks a module for permanent refusals first, then record operations not
// implemented yet, then temporary record declaration refusals. Each names its fix.
func (l *lowering) refuse(module *ast.SourceFile) error {
	if err := l.refuseDiscriminantWrites(module); err != nil {
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
		return &Refused{Where: fmt.Sprintf("%s:%d:%d", l.program.FileName(module), line+1, column+1), What: name + " suppression directive", Fix: "remove it and fix the type error (ban-ts-comment)"}
	}
	// File-level checking pragmas are separate from line-suppression directives. Use every
	// parsed pragma, including one overridden by a later pragma, rather than just CheckJsDirective.
	for _, pragma := range module.Pragmas {
		if pragma.Name == "ts-nocheck" || pragma.Name == "ts-check" {
			line, column := scanner.GetLineAndCharacterOfPosition(module, pragma.Pos())
			return &Refused{Where: fmt.Sprintf("%s:%d:%d", l.program.FileName(module), line+1, column+1), What: "@" + pragma.Name + " checking pragma", Fix: "remove it and fix any type errors"}
		}
	}
	var found error
	var recordDeclaration error
	var recordOperation error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if branch := l.literalCallableBranch(node); branch != nil {
			return visit(branch)
		}
		if err := l.recordLiteralRead(node); err != nil {
			found = err
			return true
		}
		if node.Kind == ast.KindIndexSignature {
			if _, readonly := l.recordInfo(l.checker.GetTypeAtLocation(node.Parent)); !readonly && l.recordElement(l.checker.GetTypeAtLocation(node.Parent)) == nil {
				found = l.notYet(node, "an index signature beside named members, or a non-mutable unrestricted string signature (dictionary storage cannot preserve named-property contracts)")
				return true
			}
		}
		if err := l.nodeLibraryRefusal(node); err != nil {
			found = err
			return true
		}
		if err := l.refuseClassMerge(node); err != nil {
			found = err
			return true
		}
		if err := l.refuseUnsafeDeclaration(node); err != nil {
			found = err
			return true
		}
		if err := l.refuseRecordDeclaration(node); err != nil && recordDeclaration == nil {
			recordDeclaration = err
		}
		if err := l.recordOperationNotYet(node); err != nil && recordOperation == nil {
			recordOperation = err
		}
		if refused, isRefused := refusals[node.Kind]; isRefused {
			found = &Refused{Where: l.program.Where(node), What: refused.what, Fix: refused.fix}
			return true
		}
		if node.Kind == ast.KindIndexSignature {
			if _, supported := l.recordInfo(l.checker.GetTypeAtLocation(node.Parent)); !supported && l.recordElement(l.checker.GetTypeAtLocation(node.Parent)) == nil {
				found = l.notYet(node, recordLimit)
				return true
			}
		}
		if err := l.namespaceRefusal(node); err != nil {
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
		if err := l.recordStorageView(node); err != nil {
			found = err
			return true
		}
		if node.Kind == ast.KindBinaryExpression {
			if refused, isRefused := refusedOperators[node.AsBinaryExpression().OperatorToken.Kind]; isRefused && !l.nullishComparison(node) && !(node.AsBinaryExpression().OperatorToken.Kind == ast.KindInKeyword && l.recordElement(l.checker.GetTypeAtLocation(node.AsBinaryExpression().Right)) != nil) {
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
			found = &Refused{Where: l.program.Where(node), What: "a generator function", Fix: "use an explicit iterator object; suspended frames need ownership and cancellation rules before generators can be compiled without a collector (docs/user-iterators.md) (adamic/no-generators)"}
			return true
		}
		if ast.IsFunctionLike(node) && ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
			found = &Refused{Where: l.program.Where(node), What: "an async function", Fix: "0.1 has no async; it arrives with the concurrency model (adamic/no-async)"}
			return true
		}
		if node.Kind == ast.KindIdentifier && node.Text() == "arguments" {
			// JavaScript's arguments object, not a variable the program named arguments.
			if symbol := l.checker.GetSymbolAtLocation(node); symbol != nil && len(symbol.Declarations) == 0 {
				found = &Refused{Where: l.program.Where(node), What: "arguments", Fix: "name the parameters, or take a rest parameter (adamic/no-arguments)"}
				return true
			}
		}
		if (node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression) && !called(node) {
			if err := l.nodeBufferUnsupportedUse(node); err != nil {
				found = err
				return true
			}
		}
		if node.Kind == ast.KindPropertyAccessExpression && !called(node) && !l.libraryNumberBoundMethod(node) && !l.stringMethodObservation(node) && !l.libraryArrayObservedMethod(node) {
			// A method read as a value loses its object: this is undefined when it's called.
			access := node.AsPropertyAccessExpression()
			if access.Name().Text() == "isPrototypeOf" && l.libraryMember(node) {
				found = l.prototypeRead(node, "isPrototypeOf", nil)
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
	if found != nil {
		return found
	}
	// Permanent refusals win. Otherwise name the unimplemented record operation
	// before falling back to the temporary refusal of its annotation.
	if recordOperation != nil {
		return recordOperation
	}
	return recordDeclaration
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

// Class members need storage or an implementation. A merged type cannot supply either.
// Use the bound symbol, so declaration order and imported augmentations do not hide a merge.
func (l *lowering) refuseClassMerge(node *ast.Node) error {
	if (node.Kind != ast.KindClassDeclaration && node.Kind != ast.KindInterfaceDeclaration && node.Kind != ast.KindModuleDeclaration) || node.Name() == nil {
		return nil
	}
	symbol := l.checker.GetSymbolAtLocation(node.Name())
	if symbol == nil {
		return nil
	}
	var class *ast.Node
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindClassDeclaration {
			class = declaration
			break
		}
	}
	if class == nil {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		kind := ""
		switch declaration.Kind {
		case ast.KindModuleDeclaration:
			kind = "namespace"
		case ast.KindInterfaceDeclaration:
			proven := l.checker.GetTypeAtLocation(declaration.Name())
			if len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0 || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 {
				kind = "interface"
			}
			// A repeated description of a real member is harmless. Inherited interface
			// members count too; looking only at the written member list would miss them.
			for _, member := range l.checker.GetPropertiesOfType(proven) {
				implemented := false
				for _, site := range member.Declarations {
					if site.Parent != nil && (site.Parent.Kind == ast.KindClassDeclaration || site.Parent.Kind == ast.KindClassExpression) {
						implemented = true
					}
				}
				if !implemented {
					kind = "interface"
					break
				}
			}
		}
		if kind != "" {
			name := class.Name().Text()
			return &Refused{Where: l.program.Where(declaration), What: "declaration merging of class " + name + " with " + kind + " " + name, Fix: "declare and initialize members in class " + name + "; use a separate interface for an object's checked shape, or a module for namespace exports"}
		}
	}
	return nil
}

// These checks belong to the refusal pass even when the declaration is unused or
// appears after syntax lowering has not learned. Never turn a permanent refusal into NotYet.
func (l *lowering) refuseUnsafeDeclaration(node *ast.Node) error {
	if node.Kind == ast.KindTypeReference {
		reference := node.AsTypeReferenceNode()
		if l.isLibraryGlobal(reference.TypeName, "Function") {
			return &Refused{Where: l.program.Where(node), What: "the Function type", Fix: "write a function type with its parameters and result, like (value: number) => number"}
		}

	}
	if node.Kind == ast.KindIdentifier && l.isLibraryGlobal(node, "eval") {
		return &Refused{Where: l.program.Where(node), What: "eval", Fix: "write the code as a function; native programs have no compiler at runtime"}
	}
	if node.Kind == ast.KindIdentifier && l.isLibraryGlobal(node, "Function") && (node.Parent == nil || node.Parent.Kind != ast.KindTypeReference) {
		return &Refused{Where: l.program.Where(node), What: "the Function constructor (new Function)", Fix: "write a function with explicit parameters and a checked body; native programs have no compiler at runtime"}
	}
	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		symbol := l.checker.GetSymbolAtLocation(node)
		if node.Kind == ast.KindElementAccessExpression {
			access := node.AsElementAccessExpression()
			if access.ArgumentExpression.Kind == ast.KindStringLiteral {
				symbol = l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(access.Expression), access.ArgumentExpression.Text())
			}
		}
		if symbol != nil {
			for _, declaration := range symbol.Declarations {
				if ast.IsExpandoPropertyDeclaration(declaration) {
					return &Refused{Where: l.program.Where(node), What: "properties added after creation (expando)", Fix: "declare the properties when the object is created, or use a Map for dynamic keys"}
				}
			}
		}
	}
	return nil
}

// Unused string-indexed declarations and named missing reads stay temporarily
// refused until #p9v82wa lands. Finite literal-key Records are checked object shapes.
func (l *lowering) refuseRecordDeclaration(node *ast.Node) error {
	if node.Kind == ast.KindTypeReference {
		// Record<string, T> stays refused until #p9v82wa lowers own keys only,
		// dynamic reads as T | undefined, and loudly stops prototype-name reads and in.
		// A finite Record literal-key union is an ordinary checked object shape.
		if _, readonly := l.recordInfo(l.checker.GetTypeAtLocation(node)); !readonly && l.checker.GetStringIndexType(l.checker.GetTypeAtLocation(node)) != nil && l.recordElement(l.checker.GetTypeAtLocation(node)) == nil {
			return &Refused{Where: l.program.Where(node), What: "Record<string, T> (an index signature without own-key record lowering)", Fix: "use Map<string, T> until records have own keys, T | undefined dynamic reads, and loud prototype-name checks (#p9v82wa)"}
		}
	}
	return nil
}

// These operations have not been implemented, rather than forbidden. Diagnose
// them before a record can become ordinary fixed-shape IR and fail at runtime.
func (l *lowering) recordOperationNotYet(node *ast.Node) error {
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		if _, readonly := l.recordInfo(l.checker.GetTypeAtLocation(access.Expression)); !readonly && l.checker.GetStringIndexType(l.checker.GetTypeAtLocation(access.Expression)) != nil && l.recordElement(l.checker.GetTypeAtLocation(access.Expression)) == nil {
			parent := node.Parent
			if parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == node && ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind) {
				return l.notYet(node, "dynamic record writes (own-key record storage is not implemented; use Map.set)")
			}
			return l.notYet(node, "dynamic record reads for a value of type "+l.checker.TypeToString(l.checker.GetTypeAtLocation(access.Expression))+" (own-key lookup returning T | undefined is not implemented for this representation; use Map.get)")
		}
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			access := callee.AsPropertyAccessExpression()
			name := access.Name().Text()
			if l.isLibraryGlobal(access.Expression, "Object") && (name == "values" || name == "entries") && l.checker.GetStringIndexType(l.checker.GetTypeAtLocation(call.Arguments.Nodes[0])) != nil && l.recordElement(l.checker.GetTypeAtLocation(call.Arguments.Nodes[0])) == nil && !l.readonlyRecordSupported(l.checker.GetTypeAtLocation(call.Arguments.Nodes[0])) {
				return l.notYet(node, "Object."+name+" on a record (own-key record enumeration is not implemented; use Map."+name+")")
			}
		}
	}
	return nil
}
