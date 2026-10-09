package core

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageMissingCause = rule.Message{
	Id: "missingCause",
	Description: "This throws a new error out of a catch block without attaching the error that " +
		"was caught. The original stack, the original message, and whatever the underlying " +
		"library said about what actually went wrong are all discarded at this line, and what " +
		"reaches the log is a sentence somebody wrote by hand months ago. The failure that " +
		"matters is the one you can no longer see. Pass the caught error as the `cause` option " +
		"so both halves survive.",
}

var messageIncorrectCause = rule.Message{
	Id: "incorrectCause",
	Description: "This error is thrown out of a catch block with a `cause`, but the cause is not " +
		"the error that was caught. A property of it, a copy made earlier, or an unrelated value " +
		"each loses something the original carried, usually its stack, and the reader of the log " +
		"has no way to tell that the chain was cut here. Attach the caught error itself.",
}

var messageMissingCatchErrorParam = rule.Message{
	Id: "missingCatchErrorParam",
	Description: "This throws a new error out of a catch clause that declares no parameter, so the " +
		"error it caught cannot be reached at all. Nothing in the block can log it, inspect it, or " +
		"attach it as the `cause` of what is thrown here, and the information is gone before the " +
		"first statement runs. Name the parameter so the caught error stays available.",
}

var messagePartiallyLostError = rule.Message{
	Id: "partiallyLostError",
	Description: "This catch clause destructures the error it caught, so only the parts it names " +
		"survive. The stack, and whatever else the pattern does not pick out, is gone before the " +
		"block runs, and an error thrown from here cannot attach the original as its `cause`. Bind " +
		"the whole error to a name and destructure it inside the block where that is still useful.",
}

var messageCaughtErrorShadowed = rule.Message{
	Id: "caughtErrorShadowed",
	Description: "This attaches a `cause` spelled like the caught error, but a closer declaration " +
		"of the same name shadows the catch parameter, so what is attached is that other value. " +
		"The line reads as preserving the original and does not. Rename the inner declaration so " +
		"the caught error is the one attached.",
}

var messageIncludeCause = rule.Message{
	Id: "includeCause",
	Description: "Attach the caught error as the `cause` option. This changes what the thrown " +
		"error carries, so it is offered rather than applied.",
}

// PreserveCaughtErrorOptions mirrors ESLint's two options. Both default off, which is the Go zero
// value, so a rule configured as bare `"error"` and an empty object read the same.
type PreserveCaughtErrorOptions struct {
	// RequireCatchParameter makes a throw of a new error inside a catch clause with no parameter a
	// finding of its own. It reports at each such throw, not at the clause, so a parameterless catch
	// that throws nothing is still silent.
	RequireCatchParameter bool `json:"requireCatchParameter"`

	// ErrorClassNames adds constructors beyond the built-in error types, each with the position its
	// options argument takes. A bare name takes position 2, the built-in Error's own shape.
	ErrorClassNames []PreserveCaughtErrorClassName `json:"errorClassNames"`
}

// PreserveCaughtErrorClassName is one entry of `errorClassNames`, in either of its two spellings.
type PreserveCaughtErrorClassName struct {
	Name string

	// ArgumentPosition is one-based, as ESLint's schema writes it: 2 means the second argument.
	ArgumentPosition int
}

// UnmarshalJSON accepts a bare string or a `{ name, argumentPosition }` object, which is ESLint's
// `oneOf`.
//
// The object is read strictly, which is ESLint's `additionalProperties: false`, and both keys are
// required. A position below 1 is refused rather than clamped: it would index the arguments at -1,
// and ESLint's schema sets `minimum: 1` for the same reason.
func (entry *PreserveCaughtErrorClassName) UnmarshalJSON(raw []byte) error {
	var name string
	if err := rule.UnmarshalOptions(raw, &name); err == nil {
		entry.Name = name
		entry.ArgumentPosition = 2
		return nil
	}
	var object struct {
		Name             *string `json:"name"`
		ArgumentPosition *int    `json:"argumentPosition"`
	}
	if err := rule.UnmarshalOptions(raw, &object); err != nil {
		return err
	}
	if object.Name == nil || object.ArgumentPosition == nil {
		return fmt.Errorf("an errorClassNames object needs both name and argumentPosition")
	}
	if *object.ArgumentPosition < 1 {
		return fmt.Errorf("errorClassNames argumentPosition is %d, and the first argument is 1", *object.ArgumentPosition)
	}
	entry.Name = *object.Name
	entry.ArgumentPosition = *object.ArgumentPosition
	return nil
}

// builtInErrorTypes are the global constructors that take a `cause` option, which is ESLint's list
// and TypeScript's es2022.error.d.ts. AggregateError takes it third, after the errors iterable and
// the message; every other one takes it second.
var builtInErrorTypes = map[string]bool{
	"Error":          true,
	"EvalError":      true,
	"RangeError":     true,
	"ReferenceError": true,
	"SyntaxError":    true,
	"TypeError":      true,
	"URIError":       true,
	"AggregateError": true,
}

// PreserveCaughtError flags a new error thrown out of a catch block that does not carry the caught
// error as its `cause`, following ESLint's rule of the same name.
//
//	valid:   try { a() } catch (err) { throw new Error("m", { cause: err }) }
//	valid:   try { a() } catch (err) { throw new Error("m", { "cause": err }) }   // any static key
//	valid:   try { a() } catch (err) { throw new Error("m", { ...opts }) }        // spread may carry it
//	valid:   try { a() } catch (err) { throw new Error("m", o) }                  // options it can't see
//	valid:   try { a() } catch (err) { const f = () => { throw new Error() } }   // a nested function
//	invalid: try { a() } catch (err) { throw new RangeError("m") }               // missingCause
//	invalid: try { a() } catch (err) { throw new Error("m", { cause: err.message }) }  // incorrectCause
//	invalid: try { a() } catch ({ message }) { throw new Error(message) }        // partiallyLostError
//
// # Which throws are examined
//
// A throw is examined when its nearest enclosing catch clause is reached without first crossing a
// function or a class static block. A throw inside a function defined in the catch block runs
// whenever that function is called, which is not necessarily while the caught error means anything,
// and that holds for an arrow as much as for a `function`. The port this replaced followed oxc,
// which stops at a `function` and walks through an arrow; ESLint stops at both, and so does this.
//
// A throw inside a nested catch is examined against that nested catch, since it is the nearest. A
// throw in a nested try block or finally block inside an outer catch is examined against the outer
// one, because no catch clause sits between them.
//
// # Which constructors count
//
// The eight built-in error types, when the name is the global rather than a local binding that
// shares its spelling: `import { Error } from "./mine"` puts a constructor in scope whose signature
// nobody here knows. `errorClassNames` adds others by name, as a bare identifier or the property of
// a dotted member (`new errors.AppError()`), whether or not the name is global, since a project
// naming its own class means its own class. A built-in global takes its own position even when
// `errorClassNames` also names it.
//
// Parentheses are transparent, as they are to ESLint, whose tree has no node for them:
// `throw (new (Error)("m"))` is a new Error. An optional call (`Error?.("m")`) is not a construction
// ESLint recognizes, and neither is a type assertion around the thrown value.
//
// # Reading the options argument
//
// The `cause` is the LAST property whose static key is `cause`, because the last one is what the
// object holds at runtime. A quoted key, a template key and a bracketed literal (`{ ["cause"]: err }`)
// all name it; a bracketed variable does not, since it names whatever the variable holds. The port
// this replaced read only a plain identifier key and the first match, as oxc does, so
// `{ "cause": err }` reported and `{ cause: err, cause: other }` was clean while attaching `other`.
//
// Three shapes are declined rather than guessed at: a spread argument at or before the options
// position (the positions can't be counted), a spread property in the options (it may carry the
// cause), and an options argument that is not an object literal (`new Error("m", o)`).
//
// # What each finding names
//
//	missingCause            no `cause` at all, at the throw
//	incorrectCause          a `cause` naming something else, at that value
//	caughtErrorShadowed     a `cause` spelled like the parameter but bound by a closer declaration, at the throw
//	missingCatchErrorParam  a catch with no parameter, under requireCatchParameter, at the throw
//	partiallyLostError      a destructured catch parameter, at the catch clause, once per such throw
//
// The value of a method or accessor named `cause` is its function, which ESLint's tree spans from the
// type parameters or the opening parenthesis to the end of the body, so that is the span reported.
// Two `cause` keys report the last one's value with no suggestion, as ESLint does, since rewriting one
// of two keys reads as fixing a line that still holds a duplicate.
//
// # Shadowing
//
// A `cause` spelled like the caught error is compared by symbol, not by name. Upstream's corpus
// carries the case that settles it:
//
//	catch (error) { if (whatever) { const error = anotherError; throw new Error("m", { cause: error }); } }
//
// A shorthand `{ cause }` names the variable `cause`, so it is the caught error exactly when the
// parameter is spelled `cause`; its symbol is the value's, read through the shorthand.
//
// # The repair is a suggestion
//
// Attaching a cause, or replacing a wrong one, changes what the thrown error carries, so the repair
// needs a person to agree, which is ESLint's reading too. Its shapes are ESLint's: missing arguments
// before the options slot are written as `""` (and `[]` for AggregateError's errors), an existing
// options object gains `cause` after its last property (`{}` becomes `{cause: err}`, ESLint's own
// spacing), and a method, accessor or shorthand `cause` is replaced whole. A custom class missing an
// argument before its options slot gets no suggestion, since its signature is unknown.
//
// One place differs from ESLint on purpose. With type arguments and no arguments,
// `new Error<() => void>()`, ESLint inserts at the first parenthesis after the callee, which is the
// one inside the type arguments, and writes `new Error<("", { cause: err }) => void>()`. This inserts
// at the argument list the parser found, which is the code ESLint meant to write.
var PreserveCaughtError = rule.Rule{
	Name: "preserve-caught-error",

	// Whether a constructor is the global and whether a `cause` names the caught binding are both
	// symbol questions.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as bare `"error"` is handed nil options, and the zero struct is the
		// documented default for both options.
		parsed, _ := rule.OptionsAs[PreserveCaughtErrorOptions](options)

		// Later entries win, as ESLint's Map.set does: naming one class twice with two positions
		// takes the second.
		errorClassPositions := map[string]int{}
		for _, entry := range parsed.ErrorClassNames {
			errorClassPositions[entry.Name] = entry.ArgumentPosition
		}

		return rule.Listeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				// Without a checker neither the global test nor the cause comparison can answer, and
				// a file whose program failed to build should read as a rule that could not run, not
				// as one complaining only about the findings that need no types. Silence throughout.
				if ctx.TypeChecker == nil {
					return
				}
				checkThrowInCatch(ctx, node, parsed.RequireCatchParameter, errorClassPositions)
			},
		}
	},
}

// constructedError is a thrown `new X(...)` or `X(...)`, with the pieces the rule reads off it.
type constructedError struct {
	node          *ast.Node
	callee        *ast.Node
	typeArguments *ast.NodeList
	argumentList  *ast.NodeList
	arguments     []*ast.Node
}

// thrownConstruction reads a thrown expression as a construction or call, or answers false.
//
// An optional call is declined: ESLint's tree wraps it in a chain expression, which is neither a new
// expression nor a call, so `throw Error?.("m")` is silent there.
func thrownConstruction(thrown *ast.Node) (constructedError, bool) {
	thrown = ast.SkipParentheses(thrown)
	if thrown == nil {
		return constructedError{}, false
	}
	construction := constructedError{node: thrown}
	switch thrown.Kind {
	case ast.KindNewExpression:
		expression := thrown.AsNewExpression()
		construction.callee = expression.Expression
		construction.typeArguments = expression.TypeArguments
		construction.argumentList = expression.Arguments
	case ast.KindCallExpression:
		if ast.IsOptionalChain(thrown) {
			return constructedError{}, false
		}
		expression := thrown.AsCallExpression()
		construction.callee = expression.Expression
		construction.typeArguments = expression.TypeArguments
		construction.argumentList = expression.Arguments
	default:
		return constructedError{}, false
	}
	if construction.argumentList != nil {
		construction.arguments = construction.argumentList.Nodes
	}
	return construction, true
}

// enclosingCatchClause returns the catch clause a throw belongs to, or nil when it is in none or a
// function or static block comes first.
func enclosingCatchClause(throwStatement *ast.Node) *ast.Node {
	for current := throwStatement.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindCatchClause:
			return current
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindConstructor, ast.KindClassStaticBlockDeclaration:
			return nil
		}
	}
	return nil
}

// errorClassName returns the name a constructor is recognized by: the identifier itself, or the
// property of a dotted member. A bracketed member and a private name answer empty.
func errorClassName(callee *ast.Node) string {
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text()
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		if name.Kind == ast.KindIdentifier {
			return name.Text()
		}
	}
	return ""
}

// checkThrowInCatch decides whether one throw statement reports, and with what.
func checkThrowInCatch(ctx rule.Context, throwStatement *ast.Node, requireCatchParameter bool, errorClassPositions map[string]int) {
	construction, isConstruction := thrownConstruction(throwStatement.AsThrowStatement().Expression)
	if !isConstruction {
		return
	}
	catchClause := enclosingCatchClause(throwStatement)
	if catchClause == nil {
		return
	}

	// The structural tests come first and the checker last: most throws in a catch block construct
	// nothing this rule names.
	callee := ast.SkipParentheses(construction.callee)
	className := errorClassName(callee)
	if className == "" {
		return
	}
	builtIn := callee.Kind == ast.KindIdentifier && builtInErrorTypes[className] && resolvesToAGlobal(ctx, callee)
	customPosition, isCustom := errorClassPositions[className]
	if !builtIn && !isCustom {
		return
	}

	declaration := catchClause.AsCatchClause().VariableDeclaration
	if declaration == nil {
		if requireCatchParameter {
			ctx.ReportNode(throwStatement, messageMissingCatchErrorParam)
		}
		return
	}
	binding := declaration.Name()
	if binding.Kind != ast.KindIdentifier {
		ctx.ReportNode(catchClause, messagePartiallyLostError)
		return
	}

	optionsIndex := customPosition - 1
	if builtIn {
		optionsIndex = 1
		if className == "AggregateError" {
			optionsIndex = 2
		}
	}

	causeProperty, known, multipleDefinitions := errorCauseProperty(construction.arguments, optionsIndex)
	if !known {
		return
	}

	if causeProperty == nil {
		fix, hasFix := missingCauseFix(ctx, construction, builtIn, className == "AggregateError", optionsIndex, binding.Text())
		if !hasFix {
			ctx.ReportNode(throwStatement, messageMissingCause)
			return
		}
		ctx.ReportNodeWithSuggestions(throwStatement, messageMissingCause,
			rule.Suggestion{Message: messageIncludeCause, Fixes: []rule.Fix{fix}})
		return
	}

	value := causePropertyValue(causeProperty)
	if value == nil || value.Kind != ast.KindIdentifier || value.Text() != binding.Text() {
		reportIncorrectCause(ctx, causeProperty, value, multipleDefinitions, binding.Text())
		return
	}

	var valueSymbol *ast.Symbol
	if causeProperty.Kind == ast.KindShorthandPropertyAssignment {
		valueSymbol = ctx.TypeChecker.GetShorthandAssignmentValueSymbol(causeProperty)
	} else {
		valueSymbol = ctx.TypeChecker.GetSymbolAtLocation(value)
	}
	if valueSymbol != nil && valueSymbol != ctx.TypeChecker.GetSymbolAtLocation(binding) {
		ctx.ReportNode(throwStatement, messageCaughtErrorShadowed)
	}
}

// errorCauseProperty finds the property supplying `cause` in the options argument.
//
// It answers known false where ESLint answers UNKNOWN_CAUSE and stays silent: a spread argument at or
// before the options position, a spread property in the options, or options that are not an object
// literal. Known with a nil property is "no cause at all".
func errorCauseProperty(arguments []*ast.Node, optionsIndex int) (causeProperty *ast.Node, known bool, multipleDefinitions bool) {
	for index, argument := range arguments {
		if argument.Kind == ast.KindSpreadElement && index <= optionsIndex {
			return nil, false, false
		}
	}
	if optionsIndex >= len(arguments) {
		return nil, true, false
	}

	options := ast.SkipParentheses(arguments[optionsIndex])
	if options.Kind != ast.KindObjectLiteralExpression {
		return nil, false, false
	}

	count := 0
	for _, member := range options.AsObjectLiteralExpression().Properties.Nodes {
		if member.Kind == ast.KindSpreadAssignment {
			return nil, false, false
		}
		if name, isStatic := property.Name(member.Name(), property.Named|property.Quoted|property.Templated|property.Computed); isStatic && name == "cause" {
			causeProperty = member
			count++
		}
	}
	return causeProperty, true, count > 1
}

// causePropertyValue returns the expression a `cause` property holds, with parentheses read through:
// the initializer, or the name of a shorthand. A method or accessor answers nil, since its value is
// a function and never the caught error.
func causePropertyValue(causeProperty *ast.Node) *ast.Node {
	switch causeProperty.Kind {
	case ast.KindPropertyAssignment:
		return ast.SkipParentheses(causeProperty.AsPropertyAssignment().Initializer)
	case ast.KindShorthandPropertyAssignment:
		return causeProperty.Name()
	}
	return nil
}

// reportIncorrectCause reports a `cause` that names something other than the caught error, at its
// value, with the suggestion to attach the caught one unless there are two `cause` keys.
func reportIncorrectCause(ctx rule.Context, causeProperty *ast.Node, value *ast.Node, multipleDefinitions bool, caughtName string) {
	var reported core.TextRange
	var fix rule.Fix
	switch {
	case value == nil:
		// A method or accessor. ESLint's tree spans the function from its type parameters, or its
		// opening parenthesis, to the end of its body, and the repair replaces the whole member.
		start := causeProperty.ParameterList().Pos() - 1
		if typeParameters := causeProperty.TypeParameterList(); typeParameters != nil {
			start = typeParameters.Pos() - 1
		}
		reported = core.NewTextRange(start, causeProperty.End())
		fix = ctx.ReplaceNode(causeProperty, "cause: "+caughtName)
	case causeProperty.Kind == ast.KindShorthandPropertyAssignment:
		reported = rule.TokenRange(ctx.SourceFile, value)
		fix = ctx.ReplaceNode(causeProperty, "cause: "+caughtName)
	default:
		reported = rule.TokenRange(ctx.SourceFile, value)
		fix = ctx.ReplaceNode(value, caughtName)
	}

	if multipleDefinitions {
		ctx.ReportRange(reported, messageIncorrectCause)
		return
	}
	ctx.ReportRangeWithSuggestions(reported, messageIncorrectCause,
		rule.Suggestion{Message: messageIncludeCause, Fixes: []rule.Fix{fix}})
}

// missingCauseFix builds the edit that attaches the caught error to a construction carrying none, or
// answers false where ESLint offers nothing.
func missingCauseFix(ctx rule.Context, construction constructedError, builtIn bool, aggregate bool, optionsIndex int, caughtName string) (rule.Fix, bool) {
	causeObject := "{ cause: " + caughtName + " }"
	arguments := construction.arguments

	// The positional arguments a built-in needs before its options, written as ESLint writes them.
	if builtIn {
		leading := []string{`""`}
		if aggregate {
			leading = []string{"[]", `""`}
		}
		if len(arguments) < len(leading) {
			missing := append(leading[len(arguments):], causeObject)
			if len(arguments) == 0 {
				return insertIntoEmptyCall(construction, strings.Join(missing, ", ")), true
			}
			return appendAfterLastArgument(arguments, ", "+strings.Join(missing, ", ")), true
		}
	} else if len(arguments) < optionsIndex {
		// A custom signature is unknown, so placeholder arguments would be a guess.
		return rule.Fix{}, false
	}

	if optionsIndex >= len(arguments) {
		if len(arguments) == 0 {
			return insertIntoEmptyCall(construction, causeObject), true
		}
		return appendAfterLastArgument(arguments, ", "+causeObject), true
	}

	options := ast.SkipParentheses(arguments[optionsIndex])
	if options.Kind != ast.KindObjectLiteralExpression {
		return rule.Fix{}, false
	}
	properties := options.AsObjectLiteralExpression().Properties
	if len(properties.Nodes) == 0 {
		// Directly after the brace, as ESLint inserts: `{}` becomes `{cause: err}`. The list's
		// position is the point just past the opening brace.
		return rule.ReplaceRange(core.NewTextRange(properties.Pos(), properties.Pos()), "cause: "+caughtName), true
	}
	last := properties.Nodes[len(properties.Nodes)-1]
	return rule.ReplaceRange(core.NewTextRange(last.End(), last.End()), ", cause: "+caughtName), true
}

// insertIntoEmptyCall writes arguments into a construction that has none.
//
// The argument list's position is the point just past its opening parenthesis, which the parser
// found, so a parenthesis inside a comment (`new Error/* ( */()`) or inside type arguments
// (`new Error<() => void>()`) is never mistaken for it. With no argument list at all,
// `new Error`, the parentheses are written too, after the callee and any type arguments.
func insertIntoEmptyCall(construction constructedError, text string) rule.Fix {
	if construction.argumentList == nil {
		end := construction.node.End()
		return rule.ReplaceRange(core.NewTextRange(end, end), "("+text+")")
	}
	position := construction.argumentList.Pos()
	return rule.ReplaceRange(core.NewTextRange(position, position), text)
}

// appendAfterLastArgument inserts after the last argument, past any parentheses wrapping it and
// before a trailing comma or comment.
func appendAfterLastArgument(arguments []*ast.Node, text string) rule.Fix {
	end := arguments[len(arguments)-1].End()
	return rule.ReplaceRange(core.NewTextRange(end, end), text)
}
