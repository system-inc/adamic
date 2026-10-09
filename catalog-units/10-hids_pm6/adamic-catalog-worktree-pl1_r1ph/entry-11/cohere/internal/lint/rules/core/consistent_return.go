package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/consistentreturn"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ConsistentReturnSettings is the decoded option surface.
//
// One boolean, defaulting to false, which is the zero value. Unlike no-useless-computed-key the
// generic decoder would answer correctly here, and it is still hand rolled so that an explicit
// `false` and an absent key stay distinguishable in the wire type rather than by accident.
type ConsistentReturnSettings struct {
	// TreatUndefinedAsUnspecified makes `return undefined` and `return void 0` count as returning
	// nothing, so a function mixing them with a bare `return` is consistent.
	TreatUndefinedAsUnspecified bool
}

// DefaultConsistentReturnSettings is upstream's `defaultOptions: [{treatUndefinedAsUnspecified: false}]`.
func DefaultConsistentReturnSettings() ConsistentReturnSettings {
	return ConsistentReturnSettings{TreatUndefinedAsUnspecified: false}
}

type consistentReturnWire struct {
	TreatUndefinedAsUnspecified *bool `json:"treatUndefinedAsUnspecified"`
}

// DecodeConsistentReturnOptions reads the option object off the config.
func DecodeConsistentReturnOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return DefaultConsistentReturnSettings(), nil
	}
	var wire consistentReturnWire
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return DefaultConsistentReturnSettings(), err
	}
	settings := DefaultConsistentReturnSettings()
	if wire.TreatUndefinedAsUnspecified != nil {
		settings.TreatUndefinedAsUnspecified = *wire.TreatUndefinedAsUnspecified
	}
	return settings, nil
}

var messageConsistentReturnMissingReturn = rule.Message{
	Id: "missingReturn",
	Description: "Some paths through this function return a value and some run off the end, which " +
		"returns undefined. A caller reading one branch cannot tell which kind of function this " +
		"is, so the undefined arrives somewhere far from here.",
}

var messageConsistentReturnMissingReturnValue = rule.Message{
	Id: "missingReturnValue",
	Description: "This returns nothing while another return in the same function returns a value. " +
		"The caller receives undefined from one path and a value from another, and nothing in the " +
		"signature says which.",
}

var messageConsistentReturnUnexpectedReturnValue = rule.Message{
	Id: "unexpectedReturnValue",
	Description: "This returns a value while another return in the same function returns nothing. " +
		"A reader who saw the bare return will not expect a value to come back from here.",
}

// ConsistentReturn requires every `return` in a function to agree about whether it carries a value.
//
//	valid:   function foo() { if (true) return; else return; }
//	valid:   function foo() { if (true) return true; else return false; }
//	invalid: function foo() { if (true) return true; else return; }
//	invalid: function foo() { if (a) return true; }
//
// # Two judgments, and the second is why this needs a control-flow graph
//
// The first is syntactic: within one function, the FIRST `return` sets the expectation and every
// later one that disagrees is reported, at that return statement. The second asks whether a
// function that returns a value anywhere can also run off its end, which is not a syntactic
// question at all. `function foo() { if (a) return true; }` reports and
// `function foo() { if (a) return true; else throw 1; }` does not, and the difference is whether
// any path reaches the closing brace.
//
// Upstream asks `isAnySegmentReachable(funcInfo.currentSegments)` at the function's exit.
// `control_flow_graph.Graph.EndReachable` is documented as answering that exact question and is
// already used for it by `array_callback_return`, so this rule is a caller rather than a
// reimplementation. Measured agreement on six shapes including a trailing `throw`, an
// `if`/`else throw`, and `while (true) { return 1; }`, which upstream and the graph both call
// unreachable.
//
// # The message text is three messages and a rendered name, and the casing differs between them
//
// `missingReturn` renders the name in lower case -- "at the end of function 'foo'" -- while the two
// per-return messages render it capitalised -- "Function 'foo' expected a return value". Upstream
// gets this from `upperCaseFirst` applied at one of the two sites and not the other. The corpus
// asserts both spellings, and the `Program` case asserts "Program" capitalised against "program"
// lower case in the same file, so the two cannot be collapsed.
//
// # Constructors are exempt, and only some of them
//
// A class constructor is skipped because `return` in one is a language feature rather than a value
// channel. An object literal's method NAMED `constructor` is not a constructor and is not exempt:
// measured against the installed rule, `({ constructor() { if (a) return true; } })` reports as
// "method 'constructor'" while `class A { constructor() { if (a) return true; } }` is silent. The
// corpus asserts both.
//
// Upstream also exempts an ES5 constructor, which is `astUtils.isES5Constructor`: a function
// declaration or expression whose name begins with a capital letter. The corpus's
// `function Foo() { if (!(this instanceof Foo)) return new Foo(); }` is clean for that reason and
// for no other, and `class A { CapitalizedFunction() {...} }` REPORTS, which is what shows the
// exemption is about the function's own name rather than about capitalisation anywhere.
//
// # What this does NOT ask
//
// Upstream never consults a type annotation, so `function foo(): number | undefined { if (a)
// return true; }` reports even though the annotation makes the mixed return correct TypeScript.
// Measured against the installed rule. That is the extension rule's job:
// `@typescript-eslint/consistent-return` wraps this one and filters on the declared return type.
// Reproduced as-is rather than improved on, because narrowing here would silently change what the
// extension rule is a wrapper around.
var ConsistentReturn = rule.Rule{
	Name: "consistent-return",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[ConsistentReturnSettings](options)
		if !ok {
			settings = DefaultConsistentReturnSettings()
		}

		return rule.Listeners{
			// One listener over the whole file rather than one per function kind, because the
			// judgment is about a function as a whole and the walk here is pre-order: a listener
			// on the function cannot see the returns that follow it.
			ast.KindSourceFile: func(node *ast.Node) {
				// A zero Hooks is the syntactic judgment with nothing added, which is what this
				// rule is. The extension supplies the two type-driven ones.
				consistentreturn.Judge(node,
					consistentreturn.Settings{
						TreatUndefinedAsUnspecified: settings.TreatUndefinedAsUnspecified,
					},
					consistentreturn.Hooks{},
					ConsistentReturnReporter(ctx),
					ConsistentReturnDescribe)
			},
		}
	},
}

// ConsistentReturnReporter wires the shared judgment's findings back into a rule context.
//
// Exported so the typescript extension builds an identical one rather than writing its own: the
// three spans this rule computes are the part a second implementation would most easily get wrong,
// and they are already measured against the installed build here.
func ConsistentReturnReporter(ctx rule.Context) consistentreturn.Reporter {
	return consistentreturn.Reporter{
		ReportNode: func(node *ast.Node, messageId string, description string) {
			ctx.ReportNode(node, rule.Message{Id: messageId, Description: description})
		},
		ReportRange: func(textRange core.TextRange, messageId string, description string) {
			ctx.ReportRange(textRange, rule.Message{Id: messageId, Description: description})
		},
		TokenRange: func(node *ast.Node) core.TextRange {
			return rule.TokenRange(ctx.SourceFile, node)
		},
	}
}

// ConsistentReturnDescribe renders the sentence for one finding.
//
// The casing difference between the two shapes is upstream's and the corpus asserts both: the
// per-return messages take a capitalised name, `missingReturn` takes a lower-case one, and the
// Program case asserts "Program" against "program" in the same file so they cannot be collapsed.
func ConsistentReturnDescribe(messageId string, name string) string {
	if messageId == consistentreturn.MessageIdMissingReturn {
		return fmt.Sprintf("Expected to return a value at the end of %s. %s",
			name, messageConsistentReturnMissingReturn.Description)
	}
	description := messageConsistentReturnMissingReturnValue.Description
	if messageId == consistentreturn.MessageIdUnexpectedReturnValue {
		description = messageConsistentReturnUnexpectedReturnValue.Description
	}
	return fmt.Sprintf("%s %s %s", name, consistentreturn.Verb(messageId), description)
}

// consistentReturnIsGenerator says whether a function carries the `*`.
func consistentReturnIsGenerator(node *ast.Node) bool {
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
