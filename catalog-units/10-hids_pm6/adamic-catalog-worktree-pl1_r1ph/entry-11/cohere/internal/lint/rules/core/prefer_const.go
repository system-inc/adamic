package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messagePreferConst = rule.Message{
	Id: "preferConst",
	Description: "This binding is declared `let` but nothing ever reassigns it, so the `let` " +
		"claims a mutability the code does not use. A reader has to hold open the possibility " +
		"that the value changes somewhere, and has to scan the rest of the scope to find out that " +
		"it does not. `const` answers that question at the declaration and lets the compiler " +
		"reject a future write that was not intended. Declare it `const`.",
}

// PreferConstDestructuring selects how a destructuring declaration is judged.
type PreferConstDestructuring int

const (
	// PreferConstDestructuringAny reports each binding in a pattern on its own merits, which is
	// upstream's default. `let {a, b} = obj; b = 0;` reports `a`.
	PreferConstDestructuringAny PreferConstDestructuring = iota
	// PreferConstDestructuringAll reports a pattern only when every binding in it could be const,
	// so a pattern that mixes reassigned and never-reassigned bindings is left alone. `const`
	// applies to the whole declarator rather than to one name, so under `any` the finding names
	// something the reader cannot act on without splitting the declaration.
	PreferConstDestructuringAll
)

// UnmarshalJSON decodes the option's two spellings from a configuration file.
//
// The option arrives as the JSON string "any" or "all" and the decoder is rule.UnmarshalOptions
// into this struct, so without this the string lands on an int field and the whole option object
// fails to decode. That failure is quiet in the direction that matters: the rule then runs on its
// zero value, which is "any", and a project that asked for "all" silently gets the other answer.
//
// An unrecognized value is an error rather than a fallback to the default, because ESLint's schema
// is an enum with additionalProperties false: a typo is a configuration mistake and reporting it is
// more useful than guessing which of the two the author meant.
func (destructuring *PreferConstDestructuring) UnmarshalJSON(raw []byte) error {
	var spelling string
	if err := json.Unmarshal(raw, &spelling); err != nil {
		return err
	}
	switch spelling {
	case "any":
		*destructuring = PreferConstDestructuringAny
	case "all":
		*destructuring = PreferConstDestructuringAll
	default:
		return fmt.Errorf("destructuring is %q, which is neither \"any\" nor \"all\"", spelling)
	}
	return nil
}

// PreferConstOptions is the decoded option object.
//
// Both fields are read from ESLint's `meta.schema` rather than from the inventory column, per the
// brief. The schema is one object with exactly two properties and `additionalProperties: false`:
//
//	destructuring: { enum: ["any", "all"] }
//	ignoreReadBeforeAssign: { type: "boolean" }
//
// oxc's `PreferConstConfig` carries the same two under camelCase serde renaming, so the two sources
// agree here and there is nothing to reconcile.
type PreferConstOptions struct {
	Destructuring PreferConstDestructuring `json:"destructuring"`

	// IgnoreReadBeforeAssign suppresses a finding when the binding is read at a source position
	// before its first assignment: before the declaration when it has an initializer, and before its
	// one later write when it has none. Upstream's stated purpose is avoiding a conflict with
	// `no-use-before-define`: turning such a `let` into a `const` moves the temporal dead zone in a
	// way the author may not want, so the option exists to leave those alone.
	IgnoreReadBeforeAssign bool `json:"ignoreReadBeforeAssign"`
}

// PreferConst flags a `let` binding that is never reassigned after it is initialized.
//
//	valid:   const x = 0;
//	valid:   let x = 0; x = 1;
//	valid:   let x; x = 0; x += 1;
//	valid:   let x; { x = 0; } foo(x);
//	valid:   for (let x of [1,2,3]) { x = 0; }
//	valid:   let predicate; [typeNode.returnType, ...predicate] = foo();
//	valid:   let x; foo() || (x = 0);
//	invalid: let x = 1; foo(x);
//	invalid: let x; x = 0;
//	invalid: let a; ({a} = obj);
//	invalid: for (let x of [1,2,3]) { foo(x); }
//	invalid: let { foo, bar } = baz;
//
// # This rule inverts its siblings, and the inversion decides the whole design
//
// `no-const-assign` and `no-class-assign` report a write that should not have happened, so a write
// their detection misses is a missed report on code that was already wrong. This rule reports the
// *absence* of any write, so a write its detection misses turns correct code into a finding, and
// the fix it offers rewrites that correct code into a `const` that throws at runtime. A missed
// write category here is a false positive with a broken repair attached, which is the strictly
// worse direction.
//
// Every judgment below therefore fails toward silence. Where the rule cannot see the whole picture,
// it says nothing rather than guessing.
//
// # Why this reads the checker
//
// Same reason as the siblings: the discrimination is name resolution. Upstream's clean cases
// include `let x = 0; { let x = 1; foo(x); } x = 0;`, where the inner `x` genuinely is const-able
// and the outer one is not, and the two are spelled identically. Node identity rather than
// declaration kind, because both anchors here are `KindVariableDeclaration` and kind cannot
// separate them at all. The distinguishing input is `let a = 1; { let a = 1; a = 2; }`, where a
// kind comparison sees the inner write as a write to the outer binding and silences a finding that
// identity reports; it has its own fixture and its own mutant.
//
// # The structural half does not come off the shelf here, and that is measured
//
// `no-const-assign` uses `ast.IsWriteAccess` as its write detector and documents it as probed. It
// was probed again for this rule, on this rule's own corpus, because the brief is right that a
// shelf function's doc comment is a claim. It has a gap:
//
//	[...w] = []        IsWriteAccess reports false
//	({...w} = {})      IsWriteAccess reports false
//	[a, ...w] = []     IsWriteAccess reports false
//
// A rest element inside a destructuring *assignment target* is a `KindSpreadElement` or
// `KindSpreadAssignment` whose identifier the accessor declines. For the sibling rules that gap is
// a missed report. For this rule it is exactly the false positive described above: upstream carries
// `let predicate; [typeNode.returnType, ...predicate] = foo();` as a *clean* case, and a rule
// trusting the shelf sees zero writes to `predicate` and reports it. So the write detection here is
// `reference.WritesToBinding`, which is `ast.IsWriteAccess` with the spread arms added, and the gap
// has its own fixture in both corpora.
//
// That wrapper was written three times, once per rule, before a census read the whole corpus at
// once and lifted the union onto the shelf. The lifted version also closes a gap this rule's own
// wrapper had: a parenthesis sitting directly around the identifier made the climb bail on its
// first step, so `let w = 1; [...(w)] = xs;` read as never written and was reported as convertible
// to `const`. Applying that suggestion produces code that throws `TypeError: Assignment to constant
// variable` at runtime. Measured with a control on the unparenthesized form, which was always
// correct.
//
// # What counts as never reassigned
//
// Two shapes qualify, and they are upstream's:
//
//	an initializer and no writes at all       `let x = 1; foo(x);`
//	no initializer and exactly one write      `let x; x = 0;`
//
// The second needs more than a count, because `const` has to be initialized where it is declared.
// A single write only converts when it is an assignment that is a statement of its own, in the same
// statement list as the declaration, so `let x; { x = 0; } foo(x);` and `let a; while (a = foo());`
// and `let a; if (true) a = 0; foo(a);` and `let a; foo() || (a = 0);` are all clean upstream and
// all clean here. `initializingAssignment` is that test, and it is ESLint's canBecomeVariableDeclaration:
// the assignment could be rewritten into the declaration where it stands, and nowhere else could.
//
// A destructuring assignment converts too. `let a; ({a} = obj);` becomes `const {a} = obj;`, which is
// why ESLint reports it, as long as the pattern could become a declaration's pattern: every target in it
// a plain name, and every name either undeclared or a `let` beside this one in the same statement list.
// A property target, a parameter or a binding from an enclosing scope cannot be declared there, so
// `let predicate; [typeNode.returnType, predicate] = foo();` stays clean. ESLint asks the outer-scope
// question of the pattern's top level only; this asks it at every depth, which can only fall silent.
//
// Where the finding points follows ESLint. With an initializer it names the declaration. Without one it
// names the write, since that is the line that becomes the declaration, unless the binding is read
// before that write, in which case it names the declaration, where the read-before-write is visible.
//
// A for-in or for-of head is its own shape. `for (let x of [1,2,3]) { foo(x); }` binds a fresh `x`
// per iteration with no initializer to speak of, so zero writes is enough and upstream reports it.
//
// # What it does not catch, on purpose
//
// `var` is untouched. Upstream leaves `var x = 0;` clean, and converting a `var` changes hoisting
// and scope rather than only mutability, which is a different rule's judgment.
//
// A binding written from a nested function is never reported, whatever the count. `let a; function
// foo() { a = bar(); }` is clean upstream, and the reason is not stylistic: the write may run any
// number of times including zero, and `const` cannot express it. The write detector sees these, and
// `initializingAssignment` refuses them, since the write's statement is not in the declaration's list.
//
// A pattern whose bindings disagree is where the `destructuring` option lives, and the two answers
// are both defensible. See PreferConstDestructuring.
//
// # The fix, and when there is not one
//
// The repair is `let` to `const` over the declaration list's keyword, and it is a fix rather than a
// suggestion: it preserves meaning wherever the rule fires, since the rule only fires when nothing
// reassigns the binding. But it is *conditional*, and the conditions are upstream's:
//
//	every declarator in the list must be const-able     `let x = 1, y = 2; y = 3;` rewrites `y`
//	                                                    to const and breaks the file
//	every declarator must have an initializer, unless   `const y;` does not parse
//	the list is a for-in/of head
//
// Where either fails the finding is still reported and simply carries no fix, which is what
// upstream does and is the honest answer: the diagnosis is right and the repair is a judgment call
// about splitting the declaration.
var PreferConst = rule.Rule{
	Name: "prefer-const",

	// See the doc above: upstream's shadow cases are textually identical to reporting ones and
	// differ only in what the name resolves to.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := PreferConstOptions{}
		if configured, ok := rule.OptionsAs[PreferConstOptions](options); ok {
			settings = configured
		}

		return rule.Listeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				// The engine hands every rule a nil checker when the program could not be built,
				// and this rule can answer nothing without one. Silence is the correct failure
				// here rather than a guess, for the reason in the doc above.
				if ctx.TypeChecker == nil {
					return
				}

				// `let` only. The flags live on the list rather than on the declarator, which is
				// why this anchors on the list. `const` is already right, `var` is a different
				// judgment, and `using` bindings are already constant.
				if node.Flags&ast.NodeFlagsLet == 0 {
					return
				}

				// There is deliberately no `declare` guard here, and the absence is measured rather
				// than an oversight. An ambient `let` looks like it needs one, but neither
				// reporting shape can be reached from an ambient context: TypeScript forbids an
				// initializer on an ambient declaration, which rules out "has an initializer and no
				// writes", and an ambient context holds no statements, which rules out "no
				// initializer and exactly one unconditional write". A guard was written first and
				// then mutated out, and every ambient shape tried still reported nothing, so it was
				// dead code standing in front of a decision already made below. The clean fixture
				// stays, because it pins the OUTCOME rather than the guard.

				if ctx.SourceFile == nil {
					return
				}

				isLoopHead := node.Parent != nil &&
					(node.Parent.Kind == ast.KindForInStatement ||
						node.Parent.Kind == ast.KindForOfStatement)

				declarations := node.AsVariableDeclarationList().Declarations.Nodes

				// A multi-declarator classic `for` head is all-or-nothing. `for (let i = 0, end =
				// 10; i < end; ++i) {}` has an `end` nothing writes and an `i` the update clause
				// writes every iteration, and upstream reports neither: the keyword is shared, so
				// naming `end` alone offers advice the reader cannot take without splitting a
				// declaration that the loop's own grammar holds together. Measured against
				// upstream's corpus, which carries this exact input and its in-a-function twin as
				// clean, and which the first execution of these fixtures caught.
				if node.Parent != nil && node.Parent.Kind == ast.KindForStatement &&
					len(declarations) > 1 {
					for _, declaration := range declarations {
						anyDeclined := false
						forEachBoundName(declaration.AsVariableDeclaration().Name(),
							func(boundName *ast.Node) {
								if judgeConstCandidate(ctx, settings, boundName,
									declaration, isLoopHead) == nil {
									anyDeclined = true
								}
							})
						if anyDeclined {
							return
						}
					}
				}

				// Fix eligibility is a property of the whole list rather than of one declarator,
				// because the keyword being rewritten is shared. Computed once, before any
				// reporting, so that a finding on the first declarator knows about the third.
				listIsFixable := true
				for _, declaration := range declarations {
					hasInitializer := declaration.AsVariableDeclaration().Initializer != nil
					// `const y;` does not parse. A for-in/of head is the exception: it has no
					// initializer and does not need one.
					if !hasInitializer && !isLoopHead {
						listIsFixable = false
					}
					forEachBoundName(declaration.AsVariableDeclaration().Name(),
						func(boundName *ast.Node) {
							if judgeConstCandidate(ctx, settings, boundName,
								declaration, isLoopHead) == nil {
								listIsFixable = false
							}
						})
				}

				// The repair rewrites one keyword that the whole list shares, so it is attached to
				// exactly one of the list's findings however many bindings report. Attaching it to
				// each one produced several identical rewrites of the same three bytes, which the
				// fixture harness correctly refuses to apply rather than guessing which copy wins,
				// and which in the engine would be a pile of overlapping edits for no reason. The
				// first reported binding carries it; the rest carry the diagnosis alone.
				fixAlreadyOffered := false
				for _, declaration := range declarations {
					reportConstCandidates(ctx, settings, node, declaration, isLoopHead,
						listIsFixable, &fixAlreadyOffered)
				}
			},
		}
	},
}

// reportConstCandidates reports the bindings of one declarator that nothing ever reassigns.
//
// Split out from the listener because the `destructuring` option changes the *unit* of the
// judgment, not the judgment itself: under "any" each bound name answers for itself, and under
// "all" the pattern answers as a whole. Both arms call the same predicate.
func reportConstCandidates(ctx rule.Context, settings PreferConstOptions, list *ast.Node,
	declaration *ast.Node, isLoopHead bool, listIsFixable bool, fixAlreadyOffered *bool) {
	name := declaration.AsVariableDeclaration().Name()
	if name == nil {
		return
	}

	var boundNames []*ast.Node
	forEachBoundName(name, func(boundName *ast.Node) {
		boundNames = append(boundNames, boundName)
	})
	if len(boundNames) == 0 {
		return
	}

	isDestructuring := name.Kind == ast.KindObjectBindingPattern ||
		name.Kind == ast.KindArrayBindingPattern

	candidates := make([]*constCandidate, len(boundNames))
	allConstable := true
	for index, boundName := range boundNames {
		candidates[index] = judgeConstCandidate(ctx, settings, boundName, declaration,
			isLoopHead)
		if candidates[index] == nil {
			allConstable = false
		}
	}

	// Under "all", a pattern that mixes the two answers reports nothing. `const` applies to the
	// declarator, so naming one binding inside a pattern the reader cannot convert without
	// restructuring is advice they cannot take.
	if settings.Destructuring == PreferConstDestructuringAll && isDestructuring && !allConstable {
		return
	}

	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		// The same judgment for a destructuring assignment: under "all", a binding whose only write is
		// a pattern reports only when every binding that pattern declares here could be const too,
		// since the pattern moves into one declaration whole. `let a, b; ({a, b} = obj); b = 0;` names
		// neither. ESLint groups by the assignment the same way.
		if settings.Destructuring == PreferConstDestructuringAll && candidate.assignment != nil &&
			!assignmentConvertsWhole(ctx, settings, candidate.assignment) {
			continue
		}
		if listIsFixable && !*fixAlreadyOffered {
			*fixAlreadyOffered = true
			ctx.ReportNodeWithFixes(candidate.reportAt, messagePreferConst, letKeywordToConst(ctx, list))
			continue
		}
		ctx.ReportNode(candidate.reportAt, messagePreferConst)
	}
}

// constCandidate is a binding that could have been declared `const`, and where to say so.
type constCandidate struct {
	// reportAt is the identifier the finding names: the declaration's name, or the single write that
	// would become the declaration. See "Where the finding points" on PreferConst.
	reportAt *ast.Node

	// assignment is the destructuring assignment that carries the binding's only write, when that is
	// how it is written. Under "all" the bindings that assignment declares answer together.
	assignment *ast.Node
}

// judgeConstCandidate decides whether one bound name could have been declared `const`, and where a
// finding about it points. Nil means it could not.
//
// The whole rule's judgment is here, and every branch that returns nil is a place the rule declines
// to speak. That asymmetry is deliberate: see the inversion note on PreferConst.
func judgeConstCandidate(ctx rule.Context, settings PreferConstOptions, boundName *ast.Node,
	declaration *ast.Node, isLoopHead bool) *constCandidate {
	if boundName == nil || boundName.Kind != ast.KindIdentifier {
		return nil
	}

	// Resolved through the checker rather than taken as the declarator node directly, so that both
	// sides of the later comparison answer the same question. Asking the AST for one side and the
	// checker for the other compares two things that happen to agree today.
	anchor := declarationAnchoredAt(ctx, boundName)
	if anchor == nil {
		return nil
	}

	writes := writesResolvingTo(ctx, boundName, anchor)

	if settings.IgnoreReadBeforeAssign {
		// Upstream splits this option into two tests by whether the declarator has an initializer,
		// and the split is load-bearing rather than tidy. With an initializer the write is the
		// declaration itself, so "read before assign" can only mean a read at a source position
		// before the declaration. With no initializer the assignment is a separate statement, so
		// the question is whether any read sits before THAT write, and a read after the declaration
		// still counts: `let x; function foo() { bar(x); } x = 0;` is upstream's own case, clean
		// under the option and reported without it, and a single position test against the
		// declaration answers it wrong whichever position it picks.
		boundary := declaration.Pos()
		if declaration.AsVariableDeclaration().Initializer == nil && len(writes) == 1 {
			boundary = writes[0].Pos()
		}
		if readsBeforePosition(ctx, boundName, anchor, boundary) {
			return nil
		}
	}

	// A for-in/of head rebinds on every iteration and needs no initializer, so no writes at all is
	// the whole test. `for (let x of [1,2,3]) { x = 0; }` has one and is correctly declined. The same
	// holds with an initializer, and both name the declaration.
	if isLoopHead || declaration.AsVariableDeclaration().Initializer != nil {
		if len(writes) != 0 {
			return nil
		}
		return &constCandidate{reportAt: boundName}
	}

	// No initializer. A `const` must be initialized at its declaration, so exactly one write can
	// stand in for the initializer and only where that write could be rewritten into the declaration.
	if len(writes) != 1 {
		return nil
	}

	// A read-write such as `x += 1` or `x++` cannot be the initializing write: it reads the binding
	// before the declaration ever gave it a value. `let x; x += 1;` is clean upstream.
	if !isWriteOnly(writes[0]) {
		return nil
	}

	assignment := initializingAssignment(ctx, writes[0], declaration)
	if assignment == nil {
		return nil
	}
	candidate := &constCandidate{reportAt: writes[0]}
	if pattern := ast.SkipParentheses(assignment.AsBinaryExpression().Left); pattern.Kind == ast.KindArrayLiteralExpression ||
		pattern.Kind == ast.KindObjectLiteralExpression {
		candidate.assignment = assignment
	}

	// A read before the write moves the finding to the declaration, as ESLint does: the reader acting
	// on it has to see that read, and it is at the declaration's end of the code, not the write's.
	if readsBeforePosition(ctx, boundName, anchor, writes[0].Pos()) {
		candidate.reportAt = boundName
	}
	return candidate
}

// writesResolvingTo collects every occurrence in the file that assigns to one declaration.
//
// The whole file rather than any bounded subtree, for the sibling rules' reason: a write can sit
// before the declaration, after it, or several scopes down inside a callback, and all three are the
// same binding. That matters more here than there, because a write this misses is a false positive
// rather than a missed report. The file's identifiers come from reference.IdentifiersNamed, which
// walks it once for every binding rather than once per binding, in the same order (#hekjpw3).
func writesResolvingTo(ctx rule.Context, boundName *ast.Node, anchor *ast.Node) []*ast.Node {
	var found []*ast.Node
	for _, current := range reference.IdentifiersNamed(ctx, boundName.Text()) {
		// The declarator's own name is not a write: its parent is the declarator, which
		// `reference.WritesToBinding` declines. No explicit exclusion is needed and adding one would be
		// untested code.
		if reference.WritesToBinding(current) && resolvesToDeclaration(ctx, current, anchor) {
			found = append(found, current)
		}
	}
	return found
}

// readsBeforePosition reports whether the binding is read at a source position before `position`.
//
// Only consulted under `ignoreReadBeforeAssign`. Upstream frames this as avoiding a conflict with
// `no-use-before-define`, and the position comparison is upstream's too: it compares source offsets
// rather than reasoning about execution order, so `function foo() { bar(x); } let x = 0;` counts as
// a read before the declaration even though the call happens later. That is a coarse test and it is
// reproduced rather than improved on, because the option exists to suppress a class of findings and
// a narrower test would suppress fewer of them than the option's users expect.
func readsBeforePosition(ctx rule.Context, boundName *ast.Node, anchor *ast.Node, position int) bool {
	for _, current := range reference.IdentifiersNamed(ctx, boundName.Text()) {
		// The declaration's own name resolves to the declaration too, and is not a read. Counting it
		// made every uninitialized binding read before its write, so under the option `let x; x = 0;`
		// went silent, and ESLint reports it.
		if current != boundName &&
			current.Pos() < position &&
			!reference.WritesToBinding(current) &&
			resolvesToDeclaration(ctx, current, anchor) {
			return true
		}
	}
	return false
}

// isWriteOnly reports whether an occurrence writes without first reading.
//
// `x = 0` qualifies; `x += 1` and `x++` do not, because they read the binding to compute the value
// they store. That distinction is what keeps `let x; x = 0; x += 1;` and `let a; for (const x of
// [1,2,3]) { foo(++a); }` clean, and it is upstream's `is_write() && !is_read()`.
//
// A destructuring target is write-only: `({a} = obj)` stores into `a` without reading it. The
// default in `({a = 0} = obj)` is a read of the *source*, not of `a`, so it stays write-only, and
// upstream reports that case.
func isWriteOnly(identifier *ast.Node) bool {
	for parent := identifier.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			return binary.OperatorToken != nil &&
				binary.OperatorToken.Kind == ast.KindEqualsToken

		case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
			// `++x` and `x--` read before they write.
			return false

		case ast.KindForInStatement, ast.KindForOfStatement:
			return true

		case ast.KindArrayLiteralExpression,
			ast.KindObjectLiteralExpression,
			ast.KindParenthesizedExpression,
			ast.KindSpreadElement,
			ast.KindSpreadAssignment,
			ast.KindShorthandPropertyAssignment,
			ast.KindPropertyAssignment:
			// Destructuring wrappers. The question is decided at the assignment above them.

		default:
			return false
		}
	}
	return false
}

// initializingAssignment returns the assignment a single write makes, when that assignment could be
// rewritten into the declaration where it stands, and nil when it could not.
//
// That is ESLint's canBecomeVariableDeclaration, and it is a whitelist on purpose: the assignment must
// be a plain `=`, the whole of its own expression statement, and that statement must sit directly in
// the declaration's statement list. Anything else reads as conditional or repeated, which is the silent
// direction. `let x; { x = 0; }` fails the list, `let a; while (a = foo());` and `let a; if (c) a = 0;`
// fail the statement, and `let a; foo() || (a = 0);` and `let a; b = a = 0;` fail being the whole of
// it. The last two used to report: the climb this replaces passed through any binary expression on the
// way up, so a write a `||` guards read as unconditional.
//
// A destructuring assignment qualifies when its pattern could become a declaration's pattern. See
// patternDeclaresOnlyNeighbors.
func initializingAssignment(ctx rule.Context, write *ast.Node, declaration *ast.Node) *ast.Node {
	declarationScope := enclosingStatementList(declaration)
	if declarationScope == nil {
		return nil
	}

	// No operator check: isWriteOnly has already refused every compound and logical assignment, and
	// a mutant dropping one here survived the suite as the dead code it was.
	assignment := assignmentWriting(write)
	if assignment == nil {
		return nil
	}

	statement := assignment.Parent
	for statement != nil && statement.Kind == ast.KindParenthesizedExpression {
		statement = statement.Parent
	}
	if statement == nil || statement.Kind != ast.KindExpressionStatement ||
		statement.Parent != declarationScope {
		return nil
	}

	pattern := ast.SkipParentheses(assignment.AsBinaryExpression().Left)
	if (pattern.Kind == ast.KindArrayLiteralExpression || pattern.Kind == ast.KindObjectLiteralExpression) &&
		!patternDeclaresOnlyNeighbors(ctx, pattern, declarationScope) {
		return nil
	}
	return assignment
}

// assignmentWriting returns the assignment that writes an identifier, climbing out of any destructuring
// pattern the identifier is a target in, or nil when the write is not an assignment's.
//
// It is only asked about writes, so a property key or a shorthand's default, which are never written,
// cannot arrive here and the wrappers need no test of which side the identifier is on. A for-in/of
// target is a write that no assignment makes, and the climb ends at its head with nil.
//
// TypeScript's GetAssignmentTarget climbs the same wrappers but stops at a default: in `[a = 0] = xs`
// it answers the `a = 0`, which is the element's default value and not an assignment anyone wrote. An
// `=` standing where a pattern element goes is read here as the default it is and climbed past. If the
// literal around it turns out not to be a pattern, as in `foo([a = 0])`, the climb ends at something
// that is not an assignment and the answer is nil, which is right: that write is no statement a
// declaration could replace.
func assignmentWriting(write *ast.Node) *ast.Node {
	child := write
	for parent := write.Parent; parent != nil; child, parent = parent, parent.Parent {
		switch parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindArrayLiteralExpression,
			ast.KindObjectLiteralExpression,
			ast.KindSpreadElement,
			ast.KindSpreadAssignment,
			ast.KindShorthandPropertyAssignment,
			ast.KindPropertyAssignment:
			// Wrappers a target sits inside. The question is decided further up.

		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken == nil || !ast.IsAssignmentOperator(binary.OperatorToken.Kind) ||
				binary.Left != child {
				return nil
			}
			if isPatternElementDefault(parent) {
				continue
			}
			return parent

		default:
			return nil
		}
	}
	return nil
}

// isPatternElementDefault reports whether an `=` stands where an array element or a property's value
// goes, which in a pattern makes it that element's default: `[a = 0]`, `{k: a = 0}`.
func isPatternElementDefault(binary *ast.Node) bool {
	if binary.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken || binary.Parent == nil {
		return false
	}
	switch binary.Parent.Kind {
	case ast.KindArrayLiteralExpression:
		return true
	case ast.KindPropertyAssignment:
		return binary.Parent.Name() != binary
	}
	return false
}

// patternDeclaresOnlyNeighbors reports whether a destructuring assignment's pattern could move into a
// declaration in this statement list unchanged.
//
// Every target must be a plain name, at any depth, since a declaration's pattern binds names and
// nothing else: `[typeNode.returnType, predicate] = foo()` cannot become one. That is ESLint's
// hasMemberExpressionAssignment, which also reads every depth. And every name must be one the new
// declaration may declare: undeclared, or a `let` declared directly in the same statement list, which
// the rewrite would fold in. A parameter, a `var`, an import, a function, or a binding from an
// enclosing scope cannot be redeclared there. ESLint checks this only at the pattern's top level, and
// for a `var` or a same-scope function it lets the name through; both are reported here as not
// convertible, the silent direction, since the rewrite they invite does not compile.
func patternDeclaresOnlyNeighbors(ctx rule.Context, pattern *ast.Node, declarationScope *ast.Node) bool {
	return forEachAssignmentTarget(ctx, pattern, func(symbol *ast.Symbol) bool {
		if !hasDeclaration(symbol) {
			return true
		}
		return neighboringLet(ctx, symbol, declarationScope) != nil
	})
}

// assignmentConvertsWhole reports whether every binding a destructuring assignment declares could be
// const, which is what "all" asks of the pattern before naming any of them.
func assignmentConvertsWhole(ctx rule.Context, settings PreferConstOptions, assignment *ast.Node) bool {
	pattern := ast.SkipParentheses(assignment.AsBinaryExpression().Left)
	return forEachAssignmentTarget(ctx, pattern, func(symbol *ast.Symbol) bool {
		if !hasDeclaration(symbol) {
			return true
		}
		// The pattern already passed patternDeclaresOnlyNeighbors, so every declared name in it is a
		// `let` this file holds.
		own := rule.DeclarationsIn(ctx.SourceFile, symbol)
		if len(own) == 0 {
			return false
		}
		declaration := own[0]
		for declaration != nil && declaration.Kind != ast.KindVariableDeclaration {
			declaration = declaration.Parent
		}
		if declaration == nil {
			return false
		}
		return judgeConstCandidate(ctx, settings, own[0].Name(), declaration, false) != nil
	})
}

// forEachAssignmentTarget calls back with the symbol each name in a destructuring pattern writes, and
// answers false as soon as a target is not a plain name or the callback answers false.
func forEachAssignmentTarget(ctx rule.Context, target *ast.Node, callback func(*ast.Symbol) bool) bool {
	switch target.Kind {
	case ast.KindParenthesizedExpression:
		return forEachAssignmentTarget(ctx, target.AsParenthesizedExpression().Expression, callback)

	case ast.KindArrayLiteralExpression:
		for _, element := range target.AsArrayLiteralExpression().Elements.Nodes {
			if element.Kind == ast.KindOmittedExpression {
				continue
			}
			if !forEachAssignmentTarget(ctx, element, callback) {
				return false
			}
		}
		return true

	case ast.KindObjectLiteralExpression:
		for _, property := range target.AsObjectLiteralExpression().Properties.Nodes {
			switch property.Kind {
			case ast.KindShorthandPropertyAssignment:
				// The name resolves to the property, so the value it writes is asked for separately.
				if !callback(ctx.TypeChecker.GetShorthandAssignmentValueSymbol(property)) {
					return false
				}
			case ast.KindPropertyAssignment:
				if !forEachAssignmentTarget(ctx, property.AsPropertyAssignment().Initializer, callback) {
					return false
				}
			case ast.KindSpreadAssignment:
				if !forEachAssignmentTarget(ctx, property.AsSpreadAssignment().Expression, callback) {
					return false
				}
			default:
				return false
			}
		}
		return true

	case ast.KindSpreadElement:
		return forEachAssignmentTarget(ctx, target.AsSpreadElement().Expression, callback)

	case ast.KindBinaryExpression:
		// A default: `[a = 0]`. The target is its left side.
		binary := target.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindEqualsToken {
			return false
		}
		return forEachAssignmentTarget(ctx, binary.Left, callback)

	case ast.KindIdentifier:
		return callback(ctx.TypeChecker.GetSymbolAtLocation(target))
	}
	return false
}

// hasDeclaration reports whether a name resolves to something declared, here or in a lib. A name
// nothing declares resolves to no symbol, or to one without declarations, and ESLint treats it as free
// to declare. Asked through the shelf's file questions, which hand back no node of another file.
func hasDeclaration(symbol *ast.Symbol) bool {
	return rule.IsDeclaredInASourceFile(symbol) || rule.IsDeclaredOnlyInDeclarationFiles(symbol)
}

// neighboringLet returns a symbol's declaration when it is a `let` statement's binding in this file,
// declared directly in the given statement list, and nil otherwise.
//
// A `let` cannot merge with another declaration, so its one declaration in this file is the whole of
// it, and a name declared only elsewhere has none here.
func neighboringLet(ctx rule.Context, symbol *ast.Symbol, declarationScope *ast.Node) *ast.Node {
	own := rule.DeclarationsIn(ctx.SourceFile, symbol)
	if len(own) == 0 {
		return nil
	}
	declaration := own[0]
	list := declaration.Parent
	for list != nil && list.Kind != ast.KindVariableDeclarationList {
		list = list.Parent
	}
	if list == nil || list.Flags&ast.NodeFlagsLet == 0 || list.Parent == nil ||
		list.Parent.Kind != ast.KindVariableStatement {
		return nil
	}
	if enclosingStatementList(declaration) != declarationScope {
		return nil
	}
	return declaration
}

// enclosingStatementList returns the construct whose statement list directly holds a node.
//
// Deliberately the *directly* enclosing one rather than the enclosing function: a write inside a
// nested block of the same function is still conditional in the sense that matters, because the
// block may be a loop body or an `if` consequent. Upstream reaches the same answers through a scope
// comparison plus a control-flow list; this reaches them with one test, and where the two disagree
// this one is quieter.
func enclosingStatementList(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindSourceFile,
			ast.KindBlock,
			ast.KindModuleBlock,
			ast.KindCaseClause,
			ast.KindDefaultClause,
			ast.KindClassStaticBlockDeclaration:
			return current
		}
	}
	return nil
}

// letKeywordToConst builds the repair that rewrites the list's `let` keyword.
//
// The span is the keyword alone rather than the whole declaration. oxc replaces the entire
// declaration span deliberately, so that an overlapping fix from another rule conflicts and only
// one lands; our engine resolves overlaps itself, so the narrow span is both correct and easier to
// read in a diff.
//
// The keyword is located by scanning forward from the list's start for the three characters, rather
// than assuming offset zero. A list's `Pos()` sits before leading trivia, so `/* c */ let x = 1`
// would otherwise rewrite the comment.
func letKeywordToConst(ctx rule.Context, list *ast.Node) rule.Fix {
	keyword := rule.TokenRange(ctx.SourceFile, list)
	return rule.ReplaceRange(core.NewTextRange(keyword.Pos(), keyword.Pos()+len("let")), "const")
}
