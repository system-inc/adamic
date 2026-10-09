package typescript

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// InitDeclarationsMode is which way the rule points.
type InitDeclarationsMode string

const (
	// InitDeclarationsUnconfigured is what a rule named without options does, which is nothing.
	//
	// It exists because upstream's `defaultOptions: ['always']` never reaches the rule, for the
	// reason recorded on DefaultInitDeclarationsSettings below. Modelling that as its own mode rather
	// than as an absent value keeps the inertness deliberate and visible at every call site.
	InitDeclarationsUnconfigured InitDeclarationsMode = ""
	// InitDeclarationsAlways requires every declaration to carry an initializer.
	InitDeclarationsAlways InitDeclarationsMode = "always"
	// InitDeclarationsNever forbids one.
	InitDeclarationsNever InitDeclarationsMode = "never"
)

// InitDeclarationsOptions is the rule's option surface.
type InitDeclarationsOptions struct {
	// Mode defaults to always.
	Mode InitDeclarationsMode
	// IgnoreForLoopInit exempts a for loop's own initializer, and is read only under never.
	IgnoreForLoopInit bool
}

// DefaultInitDeclarationsSettings is what an UNCONFIGURED rule does, which is nothing.
//
// This is not upstream's stated default and the difference is measured rather than chosen. Upstream
// writes `defaultOptions: ['always']`, but it writes it as a createRule property rather than as
// `meta.defaultOptions`, and ESLint 10 applies only the latter. So a rule configured as a bare
// severity string is handed no options at all: the wrapper's `mode === 'always'` guard is skipped and
// the core rule's own mode matches neither branch, and the rule reports nothing.
//
// Measured on the installed 8.x build across three spellings of the same input, `var foo; var bar = 1;`:
//
//	"error"                 no findings
//	["error", "always"]     reports foo
//	["error", "never"]      reports bar
//
// So an unconfigured mode is a third state rather than a synonym for always, and it is reproduced as
// one. Defaulting to always instead would have been a defensible reading of the source and would have
// put 512 findings on this tree that the gate being replaced does not report. That number is how the
// difference was found: the audit predicted zero, this port produced 512, and the audit was right.
func DefaultInitDeclarationsSettings() InitDeclarationsOptions {
	return InitDeclarationsOptions{Mode: InitDeclarationsUnconfigured}
}

// DecodeInitDeclarationsOptions reads upstream's option list: `["always"]`, `["never"]`,
// `["never", {"ignoreForLoopInit": true}]`, or nothing.
//
// # The list arrives whole, and the first port of this decoder is why that matters
//
// Upstream's `meta.schema` is an `anyOf` of two positional lists: the mode is `options[0]` and an
// object holding `ignoreForLoopInit` is `options[1]`, legal only beside "never". The rule registers
// with `DecodeOptionList`, so it is handed exactly that.
//
// The config layer used to store `tuple[1]` and nothing after it, so a decoder written to upstream's
// list shape failed at run time on the bare string `"never"` while every fixture passed, and the
// repair was a cohere-only `{"mode": ..., "ignoreForLoopInit": ...}` object to carry the second
// element through one slot. Both the bare string and the object are refused now: neither is
// upstream's, and the config layer no longer delivers either.
//
// A mode outside the enum is refused rather than resolved to the inert unconfigured mode. Upstream's
// schema refuses it before the rule runs, and resolving it to "never ran" was the same silence one
// layer further in. So is a second element beside "always", a key the second element's schema does
// not declare, and a third element.
func DecodeInitDeclarationsOptions(list []byte) (any, error) {
	options := DefaultInitDeclarationsSettings()
	elements, err := rule.OptionElements(list, 2)
	if err != nil || len(elements) == 0 {
		return options, err
	}

	var mode string
	if err := json.Unmarshal(elements[0], &mode); err != nil {
		return options, fmt.Errorf("init-declarations element 1 takes a mode string, got %s", elements[0])
	}
	options.Mode = initDeclarationsModeOf(mode)
	if options.Mode == InitDeclarationsUnconfigured {
		return DefaultInitDeclarationsSettings(), fmt.Errorf(
			"init-declarations: unknown mode %q, wanted always or never", mode)
	}
	if len(elements) < 2 {
		return options, nil
	}
	if options.Mode != InitDeclarationsNever {
		return DefaultInitDeclarationsSettings(), fmt.Errorf(
			"init-declarations: %q takes no second element, so %s would never be read; only "+
				"\"never\" reads ignoreForLoopInit", mode, elements[1])
	}

	var second struct {
		IgnoreForLoopInit *bool `json:"ignoreForLoopInit"`
	}
	if err := rule.UnmarshalOptions(elements[1], &second); err != nil {
		return DefaultInitDeclarationsSettings(), fmt.Errorf("init-declarations element 2: %w", err)
	}
	if second.IgnoreForLoopInit != nil {
		options.IgnoreForLoopInit = *second.IgnoreForLoopInit
	}
	return options, nil
}

// initDeclarationsModeOf reads a mode name, and answers the unconfigured mode for anything else, which
// the decoder refuses: upstream's schema is an enum, so a value outside it is a configuration error.
func initDeclarationsModeOf(name string) InitDeclarationsMode {
	switch InitDeclarationsMode(name) {
	case InitDeclarationsAlways:
		return InitDeclarationsAlways
	case InitDeclarationsNever:
		return InitDeclarationsNever
	}
	return InitDeclarationsUnconfigured
}

// InitDeclarations requires, or forbids, an initializer on a variable declaration.
//
//	always mode:
//	  valid:   var foo = 1;
//	  valid:   for (var foo in []) {}          the loop supplies the value
//	  valid:   declare var foo;                an ambient declaration has no value to give
//	  invalid: var foo;
//	  invalid: let arr: string;
//
//	never mode:
//	  valid:   var foo;
//	  valid:   const foo = 1;                  a constant must be initialized, so it is exempt
//	  invalid: var foo = 1;
//
// The two modes are opposite house styles rather than a correctness judgment, which is why the same
// source appears in both halves of upstream's corpus. Under `always` a declaration without a value is
// a variable whose type the reader has to infer from a later assignment; under `never` an initializer
// on a `var` is a value that will be reassigned before it is read.
//
// # An unconfigured rule does NOTHING, and that is upstream's behavior rather than a decision here
//
// Upstream declares `defaultOptions: ['always']` and it never takes effect, because it is written as a
// createRule property rather than as `meta.defaultOptions` and ESLint 10 reads only the latter. A rule
// named as a bare severity string is therefore handed no mode, and reports nothing at all. Measured on
// the installed build rather than inferred, and the full reasoning is on
// DefaultInitDeclarationsSettings.
//
// This matters more than a usual divergence because the natural port is the wrong one. Reading
// `defaultOptions: ['always']` and defaulting to always is the obvious choice, it passes every one of
// upstream's seventy-seven cases since each names its mode, and it puts 512 findings on this tree that
// the gate being replaced does not report. The audit predicted zero, the first version of this port
// produced 512, and the audit was right.
//
// # The span differs by MODE, and reproducing that is most of this port
//
// This is an extension rule: upstream wraps ESLint's core rule and its only live contribution is
// narrowing where the finding points. Under `always` the report covers the identifier alone, so
// `let arr: string;` underlines three characters and leaves the annotation out; under `never` the
// core rule's own span is kept and covers the whole declarator, annotation and initializer included.
//
//	always:  let arr: string;              reports `arr`
//	never:   let arr: string = ['a'];      reports `arr: string = ['a']`
//
// Both are in the imported corpus and no message id assertion can tell them apart, so both are
// asserted as text. The narrowing exists because an uninitialized declaration's whole point is that
// there is nothing after the name worth underlining.
//
// # Where our tree answers directly what upstream has to walk for
//
// Upstream's wrapper walks every ancestor looking for a `declare`d namespace, and the core rule it
// wraps keeps a boolean it sets and clears on entering and leaving one. Both exist because an ESTree
// carries `declare` only on the node that was written with the keyword.
//
// Our parser propagates it. A declaration inside `declare namespace n {}` carries NodeFlagsAmbient
// itself, and so does one nested three namespaces deep, one inside `declare module`, one inside
// `declare global`, and one written `declare var` inside a plain namespace. Measured over six shapes
// rather than assumed. So the ancestor walk and the enter/exit bookkeeping are both replaced by a
// flag test on the declaration list, which is the same DECISION reached without reproducing a
// workaround for a constraint we do not have.
//
// # `never` exempts a constant binding, and there are three of them
//
// The core rule keeps a set of `const`, `using` and `await using`, because a binding in any of those
// forms must be initialized and cannot be reported for being so. Our parser puts all three on the
// same flag field as `let`, so the test is a flag test rather than a string comparison, and
// `await using` is the using flag with another bit set.
//
// # A for loop supplies the value, so the loop head is not a declaration without one
//
// `for (var foo in [])` and `for (var foo of [])` bind `foo` on each pass, and `for (var i = 0;;)`
// carries its initializer in the head. Upstream expresses this as "if the enclosing block is a loop,
// ask whether this declaration is the loop's own binding position rather than whether it has an
// initializer". Reproduced exactly, including that a declaration in a loop's BODY is not in that
// position: `for (var a in []) var foo;` reports, because `var foo` is the loop's statement rather
// than its binding.
//
// # Cost
//
// One listener on the declaration list, no checker, no program, no per-file state.
var InitDeclarations = rule.Rule{
	Name: "@typescript-eslint/init-declarations",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[InitDeclarationsOptions](options)
		if !isSettings {
			settings = DefaultInitDeclarationsSettings()
		}

		// A rule named without a mode does nothing, matching the installed build. Declining here
		// rather than inside the listener means an unconfigured rule registers no listener at all,
		// which is also what makes the inertness visible in `cohere --timing` rather than silent.
		if settings.Mode != InitDeclarationsAlways && settings.Mode != InitDeclarationsNever {
			return nil
		}

		return rule.Listeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				list := node.AsVariableDeclarationList()

				// An ambient declaration has no value to give, and the flag is already propagated
				// into every nested namespace, module and global block, so this one test does the
				// work of upstream's ancestor walk and its enter-and-exit bookkeeping both.
				if list.Flags&ast.NodeFlagsAmbient != 0 {
					return
				}
				if list.Declarations == nil {
					return
				}

				enclosingLoop := enclosingLoopOfDeclarationList(node)

				for _, declarator := range list.Declarations.Nodes {
					declaration := declarator.AsVariableDeclaration()

					// Upstream reports only on an identifier binding, so a destructuring pattern is
					// left alone in both modes. Its comment says the base rule guards on this and
					// the span narrowing relies on it, which is true here too: the narrowed span is
					// the identifier's own text and there is none to take from a pattern.
					name := declaration.Name()
					if name == nil || name.Kind != ast.KindIdentifier {
						continue
					}

					initialized := isDeclarationInitialized(declaration, node, enclosingLoop)

					if settings.Mode == InitDeclarationsAlways {
						if initialized {
							continue
						}
						// The narrowed span: the identifier alone, excluding any type annotation.
						ctx.ReportNode(name, buildInitDeclarationsMessage(
							"initialized", name.AsIdentifier().Text))
						continue
					}

					if !initialized {
						continue
					}
					// A constant binding must be initialized, so reporting it for being so would be
					// asking for source that does not compile. Three forms, matching upstream's set.
					//
					// Naming NodeFlagsAwaitUsing is redundant: it is NodeFlagsUsing with a second bit
					// set, measured as 6 against 4, so masking on the using flag alone already
					// catches it and a mutation dropping it from this expression cannot change any
					// answer. It is written because upstream's set has three members and reading two
					// here would invite the next reader to conclude the third was overlooked.
					if list.Flags&(ast.NodeFlagsConst|ast.NodeFlagsUsing|ast.NodeFlagsAwaitUsing) != 0 {
						continue
					}
					if settings.IgnoreForLoopInit && enclosingLoop != nil {
						continue
					}
					// The core rule's own span, kept: the whole declarator.
					ctx.ReportNode(declarator, buildInitDeclarationsMessage(
						"notInitialized", name.AsIdentifier().Text))
				}
			},
		}
	},
}

// buildInitDeclarationsMessage renders upstream's two messages, which interpolate the variable name.
func buildInitDeclarationsMessage(messageId string, variableName string) rule.Message {
	if messageId == "initialized" {
		return rule.Message{
			Id:          "initialized",
			Description: "Variable '" + variableName + "' should be initialized on declaration.",
		}
	}
	return rule.Message{
		Id:          "notInitialized",
		Description: "Variable '" + variableName + "' should not be initialized on declaration.",
	}
}

// enclosingLoopOfDeclarationList answers the loop a declaration list is the HEAD of, if any.
//
// Upstream reads `declaration.parent` and asks whether it is a loop. Ours has a variable statement
// between a declaration list and its enclosing statement when the list is a statement of its own, and
// no such node when the list is a loop's own head, so the parent is the loop exactly when the list is
// its binding position. That is what keeps `for (var a in []) var foo;` reporting on `foo`: the inner
// list's parent is a variable statement, not the loop.
func enclosingLoopOfDeclarationList(list *ast.Node) *ast.Node {
	parent := list.Parent
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		return parent
	}
	return nil
}

// isDeclarationInitialized is upstream's `isInitialized`.
//
// Inside a loop head the question is positional rather than syntactic: a `for-in` or `for-of` binding
// is given a value on every pass, and a `for` statement's initializer is the loop's own. Everywhere
// else it is simply whether an initializer was written.
func isDeclarationInitialized(
	declaration *ast.VariableDeclaration,
	list *ast.Node,
	enclosingLoop *ast.Node,
) bool {
	if enclosingLoop != nil {
		if enclosingLoop.Kind == ast.KindForStatement {
			return enclosingLoop.AsForStatement().Initializer == list
		}
		// A for-in or for-of binds on each pass, so its binding position counts as initialized.
		// Upstream compares against `block.left`; ours is the loop's own initializer field.
		//
		// The comparison is SUBSUMED and kept anyway. A mutation replacing it with a bare true
		// survives every fixture, so eight parses were run to ask why: a for-in or for-of statement
		// has exactly one declaration-list child and it is always the initializer, including in four
		// malformed shapes the parser has to recover from, a missing binding name, a missing right
		// side, two declarators, and a written initializer. So the caller having found this loop as
		// the list's parent already answers the question.
		//
		// Written out because it is upstream's comparison and because the for-statement arm above it
		// genuinely needs one, where a list can be the initializer or the whole loop's own statement.
		// Reading the two arms as one shape is what makes this the natural line to delete.
		if enclosingLoop.Kind == ast.KindForInStatement || enclosingLoop.Kind == ast.KindForOfStatement {
			return enclosingLoop.AsForInOrOfStatement().Initializer == list
		}
	}
	return declaration.Initializer != nil
}
