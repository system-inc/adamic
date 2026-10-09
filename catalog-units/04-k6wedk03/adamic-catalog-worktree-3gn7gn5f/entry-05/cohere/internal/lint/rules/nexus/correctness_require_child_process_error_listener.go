package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessRequireChildProcessErrorListenerId = "childProcessWithoutErrorListener"

// correctnessRequireChildProcessErrorListenerText is the rule's message, whose wording lives in
// `policy/messages/correctness-require-child-process-error-listener.json`.
var correctnessRequireChildProcessErrorListenerText = policy.MessageOf("nexus/correctness-require-child-process-error-listener", correctnessRequireChildProcessErrorListenerId)

func correctnessRequireChildProcessErrorListenerMessage() rule.Message {
	return rule.Message{
		Id:          correctnessRequireChildProcessErrorListenerId,
		Description: correctnessRequireChildProcessErrorListenerText.Render(nil),
	}
}

// CorrectnessRequireChildProcessErrorListener reports a ChildProcess from `node:child_process`'s
// `spawn` or `fork` that never gets an `'error'` listener and never leaves the file's hands.
//
//	invalid: const child = spawn('gh', ['pr', 'checks', pr, '--watch']); child.stdout.on('data', read);
//	invalid: spawn('open', [path], { detached: true, stdio: 'ignore' }).unref();
//	invalid: const player = spawn('afplay', [soundPath], { stdio: 'ignore' });
//	invalid: spawn('ollama', ['serve']).on('close', done);
//	valid:   const child = spawn('gh', arguments); child.on('error', reject);
//	valid:   spawn('open', [path], { detached: true, stdio: 'ignore' }).on('error', ignore).unref();
//	valid:   return spawn('claude', arguments);                          (the caller may attach it)
//	valid:   const child = spawn('tar', arguments); await once(child, 'exit');      (handed away)
//	valid:   execFile('gh', arguments);                (execFile attaches its own `'error'` listener)
//
// # Where it came from
//
// `modules/os/sensation/AhraOsMonitors.ts:140` in ahra: the `prCheckStream` monitor source spawns
// `gh pr checks --watch` and listens for stdout, stderr and `'exit'`, never `'error'`. With `gh` off
// the daemon's PATH, or an EAGAIN when the process table is full, node emits `'error'` on the next
// tick, nothing listens, and the throw takes down the whole sensation daemon. Found by the
// cross-language pass of the new-rules sweep (`#tevhg3f`, after clippy's `zombie_processes`), built
// as task `#8gcxzsw`.
//
// # Which calls produce a child, and which do not
//
// The call's resolved signature must be `spawn` or `fork` declared directly in `declare module
// "node:child_process"` (or `"child_process"`), read through the same resolution
// `nexus/security-no-interpolated-shell-command` uses, so a namespace import, a renamed import and
// `process.getBuiltinModule('node:child_process')` typed as the module are all seen, and a local
// function named `spawn` never is.
//
// `exec` and `execFile` are declared in the same module and return a ChildProcess too, and they are
// deliberately not producers. node's own `execFile` (which `exec` calls) runs
// `child.addListener('error', errorhandler)` on every child it spawns, with or without a callback,
// read from the `lib/child_process.js` embedded in node 24.14.1. A missing binary there reaches the
// callback, or nowhere when there is none, and never throws. Reporting them would be false.
// `spawnSync`, `execSync` and `execFileSync` return no emitter at all.
//
// # What the child's uses say
//
// Each use of the child, the call itself and every reference to a variable it is stored in, is read
// outward through parentheses, `!`, `as` and `satisfies`, and along any chain of the emitter methods
// that return the child again:
//
//   - **Handled**: `.on`, `.once`, `.addListener`, `.prependListener` or `.prependOnceListener`
//     called with `'error'`. An event argument whose type is not made only of string literals (a
//     `string` variable, a spread) counts as handled too, since it may be `'error'`.
//   - **Kept**: the chain goes on through `.off`, `.removeListener`, `.removeAllListeners`,
//     `.setMaxListeners` and the other listener methods with another event, which all return the
//     child.
//   - **Escaped**: anything that hands the child somewhere this rule cannot follow. Returned,
//     passed to a function (`once(child, 'exit')`, a helper that wires it up), stored in an object,
//     an array or a property, awaited, copied into another variable, exported, destructured, or an
//     emitter method read without being called. The receiver may attach the listener, so an
//     escape silences the child, and a receiver that does not is a missed finding.
//   - **Neither**: everything else, which reads the child without letting it go. Another member
//     (`.stdout`, `.pid`, `.kill()`, `.unref()`), a statement on its own, a truth test, `!`,
//     `typeof`, an equality comparison, `instanceof`, and an assignment into the variable.
//
// A child is reported when its call is neither handled nor escaped and is not stored, or when it is
// stored in a variable none of whose references, anywhere in the file, is handled or escaped. A
// variable with no references at all is reported.
//
// # Where the line is drawn, and why
//
//   - **Any listener anywhere counts.** node emits a spawn failure on the next tick, so a listener
//     attached later (inside an `'exit'` handler, after an `await`) is too late, and one attached on
//     one branch leaves the other. The rule still counts every `'error'` registration in the file
//     as handling the child: placing it in time is control flow the rule does not claim, and
//     counting more as handled can only silence a site, never report one.
//   - **Only a variable the file owns is followed.** The child is followed into a `const`, `let` or
//     `var` declared from the call, or assigned to a local variable or parameter. A variable
//     declared in another file, exported, or global in a script file is an escape. A
//     conditional (`flag ? spawn(a) : spawn(b)`) or logical expression around the call is an
//     escape too.
//   - **A process-wide `uncaughtException` handler does not count.** It is not a listener on the
//     child, the failure still reaches it as an uncaught exception, and whether one is installed
//     is not visible from the file.
//
// # No fix
//
// What a failed spawn should do is the author's call: reject the promise around it, log and carry
// on, or fall back to another binary. A listener written in by a fixer would swallow the failure.
var CorrectnessRequireChildProcessErrorListener = rule.Rule{
	Name:             "nexus/correctness-require-child-process-error-listener",
	NeedsTypeChecker: true,

	// Left on the default, Contents. The producer test resolves an import through to its declaration,
	// and the dispatch scan cannot prove that read stops short of a body, so a shape-keyed replay is not
	// claimed (TestRulesClaimShapesOnlyWhereTheScanAllowsIt).

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			// Telling node's `spawn` from any other function of that name is a question for the
			// checker. With none, the rule declines rather than guess by name.
			return nil
		}
		verdicts := map[*ast.Symbol]bool{}
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				switch securityNoInterpolatedShellCommandFunction(ctx, node) {
				case "spawn", "fork":
				default:
					return
				}
				fate, bound := correctnessRequireChildProcessErrorListenerFollow(ctx, node)
				switch fate {
				case correctnessRequireChildProcessErrorListenerHandled, correctnessRequireChildProcessErrorListenerEscaped:
					return
				case correctnessRequireChildProcessErrorListenerBound:
					symbol := ctx.TypeChecker.GetSymbolAtLocation(bound)
					if symbol == nil {
						return
					}
					handled, seen := verdicts[symbol]
					if !seen {
						handled = correctnessRequireChildProcessErrorListenerVariableIsHandled(ctx, symbol)
						verdicts[symbol] = handled
					}
					if handled {
						return
					}
				}
				ctx.ReportNode(node, correctnessRequireChildProcessErrorListenerMessage())
			},
		}
	},
}

// correctnessRequireChildProcessErrorListenerFate is what one use does with a child.
type correctnessRequireChildProcessErrorListenerFate uint8

const (
	// correctnessRequireChildProcessErrorListenerNeither reads the child and lets it go nowhere.
	correctnessRequireChildProcessErrorListenerNeither correctnessRequireChildProcessErrorListenerFate = iota
	correctnessRequireChildProcessErrorListenerHandled
	correctnessRequireChildProcessErrorListenerEscaped
	// correctnessRequireChildProcessErrorListenerBound stores the child in a plain identifier.
	correctnessRequireChildProcessErrorListenerBound
)

// correctnessRequireChildProcessErrorListenerRegistrations are the EventEmitter methods that attach
// a listener. Each returns the emitter.
var correctnessRequireChildProcessErrorListenerRegistrations = map[string]bool{
	"on": true, "once": true, "addListener": true, "prependListener": true, "prependOnceListener": true,
}

// correctnessRequireChildProcessErrorListenerReturnsEmitter are the other EventEmitter methods that
// return the emitter, so a chain through them still holds the child.
var correctnessRequireChildProcessErrorListenerReturnsEmitter = map[string]bool{
	"off": true, "removeListener": true, "removeAllListeners": true, "setMaxListeners": true,
}

// correctnessRequireChildProcessErrorListenerFollow reads what the surroundings of one expression
// holding a child do with it, following the chain of emitter methods that hand the child back. For
// Bound, it also returns the identifier the child is stored in.
func correctnessRequireChildProcessErrorListenerFollow(ctx rule.Context, expression *ast.Node) (correctnessRequireChildProcessErrorListenerFate, *ast.Node) {
	current := expression
	for {
		current = correctnessRequireChildProcessErrorListenerOutermost(current)
		parent := current.Parent
		if parent == nil {
			return correctnessRequireChildProcessErrorListenerEscaped, nil
		}
		switch parent.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			if (parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression != current) ||
				(parent.Kind == ast.KindElementAccessExpression && parent.AsElementAccessExpression().Expression != current) {
				// The child is the subscript, `table[child]`, a key into something else.
				return correctnessRequireChildProcessErrorListenerEscaped, nil
			}
			member, named := property.AccessedName(parent, property.Textual)
			if !named {
				return correctnessRequireChildProcessErrorListenerEscaped, nil
			}
			registers := correctnessRequireChildProcessErrorListenerRegistrations[member]
			if !registers && !correctnessRequireChildProcessErrorListenerReturnsEmitter[member] {
				// `.stdout`, `.pid`, `.kill()`, `.unref()`: none of them hands the child on.
				return correctnessRequireChildProcessErrorListenerNeither, nil
			}
			call := parent.Parent
			if call == nil || call.Kind != ast.KindCallExpression || call.AsCallExpression().Expression != parent {
				// The method itself handed somewhere, bound or called later: it may register anything.
				return correctnessRequireChildProcessErrorListenerEscaped, nil
			}
			if registers && correctnessRequireChildProcessErrorListenerNamesError(ctx, call) {
				return correctnessRequireChildProcessErrorListenerHandled, nil
			}
			current = call
			continue
		case ast.KindExpressionStatement, ast.KindVoidExpression, ast.KindTypeOfExpression:
			return correctnessRequireChildProcessErrorListenerNeither, nil
		case ast.KindPrefixUnaryExpression:
			if parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
				return correctnessRequireChildProcessErrorListenerNeither, nil
			}
		case ast.KindIfStatement:
			if parent.AsIfStatement().Expression == current {
				return correctnessRequireChildProcessErrorListenerNeither, nil
			}
		case ast.KindWhileStatement:
			if parent.AsWhileStatement().Expression == current {
				return correctnessRequireChildProcessErrorListenerNeither, nil
			}
		case ast.KindDoStatement:
			if parent.AsDoStatement().Expression == current {
				return correctnessRequireChildProcessErrorListenerNeither, nil
			}
		case ast.KindConditionalExpression:
			if parent.AsConditionalExpression().Condition == current {
				return correctnessRequireChildProcessErrorListenerNeither, nil
			}
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken,
				ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
				return correctnessRequireChildProcessErrorListenerNeither, nil
			case ast.KindInstanceOfKeyword, ast.KindAmpersandAmpersandToken:
				// `child && child.kill()`: a ChildProcess is always truthy, so the left side of `&&`
				// never becomes its value.
				if binary.Left == current {
					return correctnessRequireChildProcessErrorListenerNeither, nil
				}
			case ast.KindEqualsToken:
				if target := ast.SkipParentheses(binary.Left); binary.Right == current && target.Kind == ast.KindIdentifier {
					return correctnessRequireChildProcessErrorListenerBound, target
				}
			}
		case ast.KindVariableDeclaration:
			declaration := parent.AsVariableDeclaration()
			if declaration.Initializer == current && parent.Name() != nil && parent.Name().Kind == ast.KindIdentifier {
				return correctnessRequireChildProcessErrorListenerBound, parent.Name()
			}
		}
		return correctnessRequireChildProcessErrorListenerEscaped, nil
	}
}

// correctnessRequireChildProcessErrorListenerOutermost climbs from an expression through the
// wrappers that pass its value through unchanged.
func correctnessRequireChildProcessErrorListenerOutermost(expression *ast.Node) *ast.Node {
	for expression.Parent != nil {
		switch expression.Parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression,
			ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
			expression = expression.Parent
			continue
		}
		return expression
	}
	return expression
}

// correctnessRequireChildProcessErrorListenerNamesError says whether a listener registration may be
// for `'error'`: its event is the literal `'error'`, or its type is not made only of string literals
// that exclude it.
func correctnessRequireChildProcessErrorListenerNamesError(ctx rule.Context, call *ast.Node) bool {
	arguments := call.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return true
	}
	event := ast.SkipParentheses(arguments.Nodes[0])
	switch event.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return event.Text() == "error"
	}
	eventType := ctx.TypeChecker.GetTypeAtLocation(event)
	if eventType == nil {
		return true
	}
	for part := range type_checking.UnionTypePartsSeq(eventType) {
		if !type_checking.IsTypeFlagSet(part, checker.TypeFlagsStringLiteral) {
			return true
		}
		if value, _ := part.AsLiteralType().Value().(string); value == "error" {
			return true
		}
	}
	return false
}

// correctnessRequireChildProcessErrorListenerVariableIsHandled says whether a variable a child is
// stored in is handled or escapes through any reference in the file. A variable the file does not
// wholly own (declared elsewhere, exported, a script-file global, not a plain variable or
// parameter) counts as escaped.
func correctnessRequireChildProcessErrorListenerVariableIsHandled(ctx rule.Context, symbol *ast.Symbol) bool {
	sourceFile := ctx.SourceFile
	declarations := rule.DeclarationsIn(sourceFile, symbol)
	if len(declarations) == 0 || len(declarations) != len(symbol.Declarations) {
		return true
	}
	declarationNames := map[*ast.Node]bool{}
	for _, declaration := range declarations {
		switch declaration.Kind {
		case ast.KindParameter:
		case ast.KindVariableDeclaration:
			list := declaration.Parent
			if list == nil || list.Kind != ast.KindVariableDeclarationList {
				return true
			}
			if statement := list.Parent; statement != nil && statement.Kind == ast.KindVariableStatement {
				if module.IsExported(statement) {
					return true
				}
				if statement.Parent != nil && statement.Parent.Kind == ast.KindSourceFile && !ast.IsExternalModule(sourceFile) {
					return true
				}
			}
		default:
			return true
		}
		if declaration.Name() == nil || declaration.Name().Kind != ast.KindIdentifier {
			return true
		}
		declarationNames[declaration.Name()] = true
	}
	name := symbol.Name

	handled := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if handled {
			return true
		}
		if node.Kind == ast.KindIdentifier && node.Text() == name && !declarationNames[node] {
			if fate := correctnessRequireChildProcessErrorListenerReference(ctx, node, symbol); fate == correctnessRequireChildProcessErrorListenerHandled ||
				fate == correctnessRequireChildProcessErrorListenerEscaped {
				handled = true
				return true
			}
		}
		node.ForEachChild(visit)
		return handled
	}
	sourceFile.AsNode().ForEachChild(visit)
	return handled
}

// correctnessRequireChildProcessErrorListenerReference reads one identifier spelled like the
// variable: Neither when it is another binding, a type position or a write into the variable, and
// otherwise what its surroundings do with the child.
func correctnessRequireChildProcessErrorListenerReference(ctx rule.Context, identifier *ast.Node, symbol *ast.Symbol) correctnessRequireChildProcessErrorListenerFate {
	parent := identifier.Parent
	if parent != nil {
		switch parent.Kind {
		case ast.KindShorthandPropertyAssignment:
			// `{ child }` stores it in an object.
			if ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent) == symbol {
				return correctnessRequireChildProcessErrorListenerEscaped
			}
			return correctnessRequireChildProcessErrorListenerNeither
		case ast.KindExportSpecifier:
			if ctx.TypeChecker.GetExportSpecifierLocalTargetSymbol(parent) == symbol {
				return correctnessRequireChildProcessErrorListenerEscaped
			}
			return correctnessRequireChildProcessErrorListenerNeither
		}
	}
	if ctx.TypeChecker.GetSymbolAtLocation(identifier) != symbol {
		return correctnessRequireChildProcessErrorListenerNeither
	}
	// `typeof child` and `typeof child.stdout` name a type and run nothing.
	typePosition := identifier.Parent
	for typePosition != nil && typePosition.Kind == ast.KindQualifiedName {
		typePosition = typePosition.Parent
	}
	if typePosition != nil && typePosition.Kind == ast.KindTypeQuery {
		return correctnessRequireChildProcessErrorListenerNeither
	}
	outer := correctnessRequireChildProcessErrorListenerOutermost(identifier)
	if assignment := outer.Parent; assignment != nil && assignment.Kind == ast.KindBinaryExpression &&
		assignment.AsBinaryExpression().Left == outer && ast.IsAssignmentOperator(assignment.AsBinaryExpression().OperatorToken.Kind) {
		// A write into the variable: the child it held, if any, is not read here.
		return correctnessRequireChildProcessErrorListenerNeither
	}
	fate, _ := correctnessRequireChildProcessErrorListenerFollow(ctx, identifier)
	if fate == correctnessRequireChildProcessErrorListenerBound {
		// Copied into another variable, which this rule does not follow.
		return correctnessRequireChildProcessErrorListenerEscaped
	}
	return fate
}
