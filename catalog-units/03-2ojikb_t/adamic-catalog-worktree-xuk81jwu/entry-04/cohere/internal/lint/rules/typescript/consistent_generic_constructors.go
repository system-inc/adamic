package typescript

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ConsistentGenericConstructorsMode selects which side of the assignment carries the type arguments.
type ConsistentGenericConstructorsMode string

const (
	// ConsistentGenericConstructorsConstructor wants `const a = new Map<string, number>()`.
	ConsistentGenericConstructorsConstructor ConsistentGenericConstructorsMode = "Constructor"

	// ConsistentGenericConstructorsTypeAnnotation wants `const a: Map<string, number> = new Map()`.
	ConsistentGenericConstructorsTypeAnnotation ConsistentGenericConstructorsMode = "TypeAnnotation"
)

// ConsistentGenericConstructorsOptions configures which spelling the rule enforces.
type ConsistentGenericConstructorsOptions struct {
	// Mode is the spelling to prefer. Upstream's default is the constructor side.
	Mode ConsistentGenericConstructorsMode
}

// DefaultConsistentGenericConstructorsSettings matches upstream's `defaultOptions: ['constructor']`.
func DefaultConsistentGenericConstructorsSettings() ConsistentGenericConstructorsOptions {
	return ConsistentGenericConstructorsOptions{Mode: ConsistentGenericConstructorsConstructor}
}

// DecodeConsistentGenericConstructorsOptions reads the rule's configuration.
//
// Upstream's `meta.schema` is a one-element positional tuple holding a bare string, so the ESLint
// spelling is `["error", "type-annotation"]`. cohere's config layer unwraps the
// `[severity, options]` pair before dispatch, so a decoder here is handed the bare string rather than
// an array. A one-element array is accepted too, so a setting copied out of an ESLint config is read
// rather than refused.
//
// The nil case matters more than it looks: a rule configured as a bare `"error"` is handed empty
// input, and a decoder erroring there yields a zero-value struct whose Mode is the empty string,
// matching neither arm and making the rule silently inert. Empty input and `null` are upstream's
// default.
//
// Anything else is refused, naming the value. This read an unknown mode, an object or a number as
// the default, on the reasoning that a typo should not turn the rule off. It did not turn it off; it
// turned it to a mode nobody wrote, and loaded clean while doing it. Upstream's schema is a string
// enum and refuses all three at load, and the config layer now refuses a decoder's error by rule
// name, so refusing is what tells the author (#p9s1131).
func DecodeConsistentGenericConstructorsOptions(raw []byte) (any, error) {
	options := DefaultConsistentGenericConstructorsSettings()
	// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return options, nil
	}

	// The spelling cohere actually delivers: the mode as a bare string.
	var mode string
	if err := json.Unmarshal(raw, &mode); err == nil {
		return consistentGenericConstructorsOptionsFor(mode)
	}

	// The spelling an ESLint config is written in, accepted so a copied setting still works.
	var modes []string
	if err := json.Unmarshal(raw, &modes); err == nil && len(modes) == 1 {
		return consistentGenericConstructorsOptionsFor(modes[0])
	}

	return options, fmt.Errorf(`expected a mode string, "constructor" or "type-annotation", got %s`, trimmed)
}

// consistentGenericConstructorsOptionsFor maps upstream's wire spelling to the internal one, refusing
// a spelling upstream's schema would refuse.
func consistentGenericConstructorsOptionsFor(mode string) (any, error) {
	switch mode {
	case "constructor":
		return ConsistentGenericConstructorsOptions{Mode: ConsistentGenericConstructorsConstructor}, nil
	case "type-annotation":
		return ConsistentGenericConstructorsOptions{Mode: ConsistentGenericConstructorsTypeAnnotation}, nil
	}
	return DefaultConsistentGenericConstructorsSettings(),
		fmt.Errorf(`mode %q is not "constructor" or "type-annotation"`, mode)
}

// ConsistentGenericConstructors enforces that a generic constructor call writes its type arguments on
// one side of the assignment rather than the other, whichever side the project picked.
//
//	valid (constructor mode):      const a = new Map<string, number>();
//	invalid (constructor mode):    const a: Map<string, number> = new Map();
//	valid (type-annotation mode):  const a: Map<string, number> = new Map();
//	invalid (type-annotation):     const a = new Map<string, number>();
//
// Writing them on both sides is legal and redundant, writing them on neither leaves the checker to
// infer, and the rule is about picking one so a codebase reads the same way throughout. It has no
// opinion about which, which is why the mode is an option and the default is upstream's.
//
// # Four anchors, and one of them is not a declaration
//
// Upstream's selector matches a variable declarator, a property definition, an accessor property,
// and a DEFAULT PARAMETER of a function declaration or function expression. That last one is easy to
// miss and the corpus tests it in both modes: `function foo(a: Foo<string> = new Foo()) {}` reports.
// A default parameter of an ARROW function is deliberately excluded by that selector, and the
// exclusion is reproduced rather than tidied away.
//
// # The two sides must name the same thing, which is the whole guard against reporting nonsense
//
// The annotation has to be a plain type reference whose name matches the constructor's callee name,
// so `const a: Foo<string> = new Bar()` is silent: moving type arguments between two different names
// would produce code that does not compile. That test also excludes a qualified name, an array type,
// a union, and anything else that is not a bare identifier reference.
//
// # The typed-array exemption, which exists for a real conflict
//
// Nine built-in array types are exempt when the annotation names them: `Uint8Array` and its family.
// Upstream added this because those types became generic over their buffer, so `const a:
// Uint8Array = new Uint8Array()` and `new Uint8Array<ArrayBuffer>()` are not interchangeable the way
// `Map` is. The exemption applies only when the name resolves to the GLOBAL, so a locally declared
// `class Uint8Array` is still reported. Reproduced here through the checker rather than through
// ESLint's scope analysis, which is the same question asked of a different instrument.
//
// # isolatedDeclarations, read from the compiler rather than from the lint config
//
// Upstream reads `parserOptions.isolatedDeclarations` and, when it is on, declines to move type
// arguments ONTO the constructor, because that mode needs the annotation to stay explicit. We do not
// have that lint-config field, and we do not need it: the real compiler option is on the program, so
// this reads `ctx.Program.Options().IsolatedDeclarations`. That is the brief's standing instruction
// that fidelity is to what the rule DECIDES rather than to how it obtains what it needs.
//
// It is also the one branch no fixture can reach. `rule_testing` writes its tsconfig AFTER the setup
// hook runs and always with the default settings, so a fixture cannot turn the option on; measured,
// with a setup hook writing its own tsconfig and the rule still seeing the option unset. Recorded in
// its own test with controls rather than left as an untested claim.
//
// # The repair, and the comments it goes out of its way to keep
//
// Both directions are FIXES rather than suggestions, matching upstream, because moving type arguments
// between two spellings of the same thing does not change what the code means.
//
// Moving them onto the constructor deletes the whole annotation, and an annotation can contain
// comments that are not inside the type arguments. Upstream collects those and re-emits them after
// the callee so they are not silently deleted, and this port does the same: a fixer that eats a
// comment is a fixer that loses something a person wrote, and it is applied unattended.
//
// The parenthesis case is the other detail: `new Foo` is a legal construction with no argument list,
// so after inserting type arguments the fixer has to add `()` as well, or it writes `new Foo<string>`
// which parses as something else entirely.
var ConsistentGenericConstructors = rule.Rule{
	Name: "@typescript-eslint/consistent-generic-constructors",

	// The typed-array exemption asks whether a name resolves to the global rather than to a local
	// declaration, which is a resolution question. Everything else is syntax.
	NeedsTypeChecker: true,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[ConsistentGenericConstructorsOptions](options)
		if !isSettings {
			// A rule configured as a bare severity reaches here with nil options, so the assertion
			// fails and the fallback is what `"error"` means.
			//
			// EQUIVALENT today, and kept anyway. Measured: a mutant replacing this with the zero
			// value changes no verdict, because an empty mode is not the type-annotation mode and
			// therefore falls into the constructor branch, which is already the default. So the
			// line is load-bearing only against a future change to which mode is default, and that
			// is exactly the change most likely to be made without reading this function. Without
			// it, flipping the default would leave a bare `"error"` silently enforcing the old one.
			settings = DefaultConsistentGenericConstructorsSettings()
		}

		check := func(node *ast.Node, annotationHost *ast.Node, typeAnnotation *ast.Node,
			initializer *ast.Node) {
			if ctx.TypeChecker == nil || initializer == nil {
				return
			}

			// The right-hand side has to be `new Identifier(...)`. A namespaced constructor, a
			// call expression, or anything else is out of scope for both directions.
			if initializer.Kind != ast.KindNewExpression {
				return
			}
			newExpression := initializer.AsNewExpression()
			callee := newExpression.Expression
			if callee == nil || !ast.IsIdentifier(callee) {
				return
			}

			// When there IS an annotation it has to name the same thing as the callee, otherwise
			// moving type arguments across would produce code that does not compile.
			//
			// Parentheses are unwrapped first, because estree has no parenthesized-type node and
			// ours does. Measured: `const a: (Foo<string>) = new Foo();` REPORTS on the installed
			// build and was silent here until this unwrap existed. The corpus writes no
			// parenthesized annotation anywhere, so nothing imported could see it.
			//
			// This is the mirror of the same hazard in expression position: unwrap before a
			// question you put to a node KIND, which this is, and never before one you put to the
			// checker.
			//
			// The OUTER node is kept for the deletion range and the unwrapped one is used for every
			// question about the type, because the parentheses are part of what gets deleted:
			// deleting only the inner reference leaves `const a) = new Foo<string>()`.
			unwrappedAnnotation := skipParenthesesInTypeSafely(typeAnnotation)
			if unwrappedAnnotation != nil {
				if unwrappedAnnotation.Kind != ast.KindTypeReference {
					return
				}
				annotationName := unwrappedAnnotation.AsTypeReferenceNode().TypeName
				if annotationName == nil || !ast.IsIdentifier(annotationName) {
					return
				}
				if annotationName.Text() != callee.Text() {
					return
				}
				if isExemptBuiltInArray(ctx, annotationName) {
					return
				}
			}

			if settings.Mode == ConsistentGenericConstructorsTypeAnnotation {
				// Only the shape with no annotation and type arguments on the constructor.
				if typeAnnotation != nil || newExpression.TypeArguments == nil {
					return
				}
				reportPreferTypeAnnotation(ctx, node, annotationHost, callee,
					newExpression.TypeArguments)
				return
			}

			// Constructor mode. `isolatedDeclarations` needs the annotation to stay explicit, so
			// the repair that removes it is declined outright.
			if ctx.Program != nil && ctx.Program.Options() != nil &&
				ctx.Program.Options().IsolatedDeclarations.IsTrue() {
				return
			}

			annotationArguments := typeReferenceTypeArguments(unwrappedAnnotation)
			if annotationArguments == nil || newExpression.TypeArguments != nil {
				return
			}

			reportPreferConstructor(ctx, node, typeAnnotation, callee, annotationArguments,
				newExpression)
		}

		checkVariableLike := func(node *ast.Node) {
			// A variable declarator and a default parameter have the same shape here: a name that
			// may carry an annotation, and an initializer.
			name := node.Name()
			if name == nil {
				return
			}
			check(node, name, node.Type(), node.Initializer())
		}

		checkPropertyLike := func(node *ast.Node) {
			// A class property attaches its annotation to the property itself rather than to a
			// separate binding name, so the host for a new annotation is the KEY.
			check(node, node.Name(), node.Type(), node.Initializer())
		}

		return rule.Listeners{
			ast.KindVariableDeclaration: checkVariableLike,
			ast.KindPropertyDeclaration: checkPropertyLike,
			ast.KindParameter: func(node *ast.Node) {
				// Upstream's selector is
				// `:matches(FunctionDeclaration,FunctionExpression) > AssignmentPattern`, and the
				// set of parents that names is WIDER than it reads, which cost this port two
				// fixtures. In estree a class method and a constructor are a MethodDefinition whose
				// `value` is a FunctionExpression, so a default parameter of either matches the
				// second branch. Measured against the installed build:
				//
				//	class A { constructor(a: Foo<string> = new Foo()) {} }   reports
				//	class A { foo(a: Foo<string> = new Foo()) {} }           reports
				//	const foo = (a: Foo<string> = new Foo()) => {};          SILENT
				//
				// The arrow is the only exclusion, and it is a real one rather than an oversight to
				// tidy away: an arrow has no MethodDefinition wrapper and is not a FunctionExpression
				// either, so upstream's selector never reaches it.
				parent := node.Parent
				if parent == nil {
					return
				}
				switch parent.Kind {
				case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
					ast.KindMethodDeclaration, ast.KindConstructor,
					ast.KindGetAccessor, ast.KindSetAccessor:
				default:
					return
				}
				checkVariableLike(node)
			},
		}
	},
}

// reportPreferTypeAnnotation moves the type arguments from the constructor onto a new annotation.
//
// Two edits: delete the type arguments, and write `: Callee<Args>` after the name. The annotation
// text is copied from the source rather than rebuilt through the checker, so whatever spacing the
// author wrote inside the type arguments survives the move.
func reportPreferTypeAnnotation(ctx rule.Context, node *ast.Node, annotationHost *ast.Node,
	callee *ast.Node, typeArguments *ast.NodeList) {
	sourceText := ctx.SourceFile.Text()
	calleeRange := rule.TokenRange(ctx.SourceFile, callee)

	// The type argument list's own text including its angle brackets, which are not part of any
	// node in the list.
	argumentsRange := typeArgumentListRange(ctx, typeArguments)
	annotationText := sourceText[calleeRange.Pos():calleeRange.End()] +
		sourceText[argumentsRange.Pos():argumentsRange.End()]

	hostRange := rule.TokenRange(ctx.SourceFile, annotationHost)

	ctx.ReportNodeWithFixes(node, buildPreferTypeAnnotationMessage(),
		rule.RemoveRange(argumentsRange),
		rule.ReplaceRange(core.NewTextRange(hostRange.End(), hostRange.End()),
			": "+annotationText))
}

// reportPreferConstructor moves the type arguments from the annotation onto the constructor.
//
// The annotation is deleted whole, which is why any comment inside it that is NOT inside the type
// arguments has to be carried over: deleting it would silently drop something a person wrote, and
// this repair is a fix rather than a suggestion, so nobody is watching when it lands.
func reportPreferConstructor(ctx rule.Context, node *ast.Node, typeAnnotation *ast.Node,
	callee *ast.Node, annotationArguments *ast.NodeList, newExpression *ast.NewExpression) {
	sourceText := ctx.SourceFile.Text()

	// The annotation including its leading colon, which is what upstream deletes: its `lhs.parent`
	// is the type annotation node, and estree's type annotation starts at the colon.
	annotationRange := typeAnnotationRangeIncludingColon(ctx, typeAnnotation)
	argumentsRange := typeArgumentListRange(ctx, annotationArguments)
	calleeRange := rule.TokenRange(ctx.SourceFile, callee)

	// Everything written after the callee lands at ONE offset, so it is assembled as a single
	// string rather than as several insertions. Two separate defects came from not doing that:
	// appending the parentheses as their own fix wrote `new Foo()<number>`, and appending the
	// comments as their own fixes put them after the type arguments and in reverse.
	//
	// The order upstream writes, and therefore the order here: the carried comments in source
	// order, then the type arguments, then the parentheses if there were none.
	insertion := ""

	// Comments inside the annotation but outside the type arguments. The annotation is deleted
	// whole, so without this they vanish, and this repair is a FIX rather than a suggestion, so
	// nobody is watching when it lands.
	//
	// `comments.ForFile` rather than `type_checking.GetCommentsInRange`, and that is measured
	// rather than preferred. The latter reads only the leading and trailing trivia AT its range's
	// start position, so over `: /* comment */ Foo/* another */ <string>` it yields nothing and
	// both of upstream's comment cases lost their comments silently while every message-id
	// assertion stayed green. `ForFile` is a cached whole-file scan, so the cost is shared with
	// every other rule that wants comments, and it returns them already sorted by position.
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() < annotationRange.Pos() || comment.Range.End() > annotationRange.End() {
			continue
		}
		if comment.Range.Pos() >= argumentsRange.Pos() && comment.Range.End() <= argumentsRange.End() {
			continue
		}
		insertion += comment.Text
	}

	insertion += sourceText[argumentsRange.Pos():argumentsRange.End()]

	// `new Foo` is a legal construction with no argument list, and `new Foo<string>` without one is
	// not the same expression, so the parentheses are not cosmetic.
	if newExpression.Arguments == nil {
		insertion += "()"
	}

	fixes := []rule.Fix{
		rule.RemoveRange(annotationRange),
		rule.ReplaceRange(core.NewTextRange(calleeRange.End(), calleeRange.End()), insertion),
	}

	ctx.ReportNodeWithFixes(node, buildPreferConstructorMessage(), fixes...)
}

// typeReferenceTypeArguments reads the type arguments off a type reference, or nil.
func typeReferenceTypeArguments(typeAnnotation *ast.Node) *ast.NodeList {
	if typeAnnotation == nil || typeAnnotation.Kind != ast.KindTypeReference {
		return nil
	}
	return typeAnnotation.AsTypeReferenceNode().TypeArguments
}

// typeArgumentListRange spans a type argument list INCLUDING its angle brackets.
//
// The list's own positions cover the arguments and not the brackets, and the brackets are exactly
// what makes the copied text a usable type argument list at its destination. Found by scanning
// outward from the first and last argument rather than by arithmetic, because the author may have
// written whitespace inside the brackets and the corpus tests that in both directions.
func typeArgumentListRange(ctx rule.Context, typeArguments *ast.NodeList) core.TextRange {
	sourceText := ctx.SourceFile.Text()

	start := typeArguments.Pos()
	for start > 0 && sourceText[start-1] != '<' {
		start--
	}
	if start > 0 {
		start--
	}

	end := typeArguments.End()
	for end < len(sourceText) && sourceText[end] != '>' {
		end++
	}
	if end < len(sourceText) {
		end++
	}

	return core.NewTextRange(start, end)
}

// typeAnnotationRangeIncludingColon spans an annotation together with the colon that introduces it.
//
// Upstream deletes `lhs.parent`, which in estree is the TSTypeAnnotation node and starts AT the
// colon. Our parser has no such node: the annotation is the type itself, hanging off the declaration,
// so the colon has to be found by scanning back. Without it the repair leaves a dangling `:`.
func typeAnnotationRangeIncludingColon(ctx rule.Context, typeAnnotation *ast.Node) core.TextRange {
	sourceText := ctx.SourceFile.Text()
	annotationRange := rule.TokenRange(ctx.SourceFile, typeAnnotation)

	start := annotationRange.Pos()
	for start > 0 && sourceText[start-1] != ':' {
		start--
	}
	if start > 0 {
		start--
	}

	return core.NewTextRange(start, annotationRange.End())
}

// consistentGenericConstructorsBuiltInArrays is upstream's exempt set, verbatim.
//
// These became generic over their backing buffer, so unlike `Map` the two spellings are not
// interchangeable and moving the type arguments would change what the code means.
var consistentGenericConstructorsBuiltInArrays = map[string]struct{}{
	"Float32Array":      {},
	"Float64Array":      {},
	"Int16Array":        {},
	"Int32Array":        {},
	"Int8Array":         {},
	"Uint16Array":       {},
	"Uint32Array":       {},
	"Uint8Array":        {},
	"Uint8ClampedArray": {},
}

// isExemptBuiltInArray answers upstream's `isBuiltInArray`: the name is one of the nine AND it
// resolves to the global rather than to something declared in the file.
//
// The resolution half is load-bearing rather than defensive. Upstream asks ESLint's scope analysis
// whether the reference reaches a global; we ask the checker whether the symbol has a declaration in
// source. A locally declared `class Uint8Array` is reported, which the corpus pins, and a port
// testing only the name would go silent on it.
func isExemptBuiltInArray(ctx rule.Context, annotationName *ast.Node) bool {
	if _, isBuiltIn := consistentGenericConstructorsBuiltInArrays[annotationName.Text()]; !isBuiltIn {
		return false
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(annotationName)
	if symbol == nil {
		// No symbol at all means nothing in this file declares the name, so it is the global.
		return true
	}

	// A declaration in a real source file means the name was shadowed locally. A declaration only in
	// a default library file is the global itself.
	for _, declaration := range symbol.Declarations {
		declarationFile := ast.GetSourceFileOfNode(declaration)
		if declarationFile == nil {
			continue
		}
		if !type_checking.IsSourceFileDefaultLibrary(ctx.Program, declarationFile) {
			return false
		}
	}
	return true
}

func buildPreferConstructorMessage() rule.Message {
	return rule.Message{
		Id: "preferConstructor",
		Description: "The generic type arguments should be specified as part of the constructor " +
			"type arguments.",
	}
}

func buildPreferTypeAnnotationMessage() rule.Message {
	return rule.Message{
		Id: "preferTypeAnnotation",
		Description: "The generic type arguments should be specified as part of the type " +
			"annotation.",
	}
}
