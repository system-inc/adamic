package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageMaxClassesPerFileId is the id, kept separate because the message itself is built per
// finding: the rendered text carries the count and the limit, so there is no single constant to
// report and nothing a reader could compare a diagnostic against by identity.
const messageMaxClassesPerFileId = "maximumExceeded"

// maxClassesPerFileMessage renders the finding.
//
// Upstream's template is `File has too many classes ({{ classCount }}). Maximum allowed is {{ max
// }}.` and both values move, which is why this is a function and why the tests assert the rendered
// string rather than only the id. A message-id fixture cannot see the two arguments being passed in
// the wrong order, and that swap renders as a grammatical sentence.
func maxClassesPerFileMessage(classCount int, maximum int) rule.Message {
	return rule.Message{
		Id: messageMaxClassesPerFileId,
		Description: fmt.Sprintf("File has too many classes (%d). Maximum allowed is %d. "+
			"A file holding several classes makes the reader hold several vocabularies at once, "+
			"and the import that names the file no longer says which one it wanted. Split them "+
			"into a file each, named for the class.", classCount, maximum),
	}
}

// MaxClassesPerFileOptions carries upstream's one positional option, which has two shapes.
//
// Upstream's schema is a `oneOf` over an integer and an object, so `[2]` and `[{max: 2}]` are both
// legal and mean the same thing. Both are kept as pointers rather than plain values, because the
// zero value of `Maximum` is 0 and upstream's schema forbids 0 outright: a config that reached the
// rule with a zero maximum would report every file holding a single class, which is the loudest
// possible way for a decode to fail silently.
type MaxClassesPerFileOptions struct {
	// Maximum is how many classes a file may hold. Upstream defaults it to 1.
	Maximum *int

	// IgnoreExpressions drops class EXPRESSIONS from the count, leaving only declarations.
	IgnoreExpressions *bool
}

// DefaultMaxClassesPerFileSettings is upstream's `defaultOptions: [1]`.
func DefaultMaxClassesPerFileSettings() MaxClassesPerFileOptions {
	maximum := 1
	ignoreExpressions := false
	return MaxClassesPerFileOptions{Maximum: &maximum, IgnoreExpressions: &ignoreExpressions}
}

// maxClassesPerFileObjectShape is the object arm of upstream's `oneOf`, named so the decoder can try
// it after the integer arm fails.
type maxClassesPerFileObjectShape struct {
	Max               *int  `json:"max"`
	IgnoreExpressions *bool `json:"ignoreExpressions"`
}

// DecodeMaxClassesPerFileOptions turns the configured value into options.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` because the wire value is polymorphic and the
// default is not the zero value. Three things have to be right and none of them falls out of a
// struct tag:
//
// The value may be a bare INTEGER. cohere's config layer strips the severity tuple, so a config
// writing `["error", 2]` hands this the JSON `2`, which no struct can unmarshal.
//
// An ABSENT value must mean 1 rather than 0. The generic helper errors on empty input and the config
// layer turns that into nil, which a plain int field would read as 0, a maximum of zero, reporting
// every file with a single class in it. That is the live config's own shape for a rule configured as
// a bare severity, so it is the ordinary path rather than a corner.
//
// An object with no `max` must also mean 1. Upstream writes `option.max || 1`, so `{}` and
// `{ignoreExpressions: true}` both leave the limit at one, measured against the installed build at
// 10.8.1 rather than read off the source.
//
// A maximum below 1 is refused, which is upstream's `minimum: 1` in the schema. Measured: eslint
// rejects both `0` and `{max: 0}` at configuration load rather than at lint time, so refusing here
// is the same decision arriving through the only surface this port has.
func DecodeMaxClassesPerFileOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return DefaultMaxClassesPerFileSettings(), nil
	}

	settings := DefaultMaxClassesPerFileSettings()

	// The integer arm first, because it is the narrower shape: a JSON number cannot also decode as
	// an object, so trying it first cannot swallow a case belonging to the other arm.
	var configuredMaximum int
	if err := json.Unmarshal(raw, &configuredMaximum); err == nil {
		if configuredMaximum < 1 {
			return DefaultMaxClassesPerFileSettings(), fmt.Errorf(
				"max-classes-per-file takes a maximum of at least 1, got %d", configuredMaximum)
		}
		settings.Maximum = &configuredMaximum
		return settings, nil
	}

	var object maxClassesPerFileObjectShape
	if err := rule.UnmarshalOptions(raw, &object); err != nil {
		return DefaultMaxClassesPerFileSettings(), err
	}
	// A nil Max is upstream's `option.max || 1` and stays at the default rather than becoming zero.
	if object.Max != nil {
		if *object.Max < 1 {
			return DefaultMaxClassesPerFileSettings(), fmt.Errorf(
				"max-classes-per-file takes a maximum of at least 1, got %d", *object.Max)
		}
		settings.Maximum = object.Max
	}
	if object.IgnoreExpressions != nil {
		settings.IgnoreExpressions = object.IgnoreExpressions
	}
	return settings, nil
}

// MaxClassesPerFile flags a file declaring more classes than the configured limit.
//
//	valid:   class Foo {}
//	valid:   var x = class {};
//	valid:   class Foo {}\nclass Bar {}                       under a maximum of 2
//	valid:   class Foo {}\nconst e = class {}                 under ignoreExpressions
//	invalid: class Foo {}\nclass Bar {}
//	invalid: var x = class {};\nvar y = class {};
//
// # A class declared inside a function does not count, which is where cohere leaves upstream
//
// Upstream hooks the two class node kinds and counts every one in the file. Measured against the
// installed build at 10.8.1, a class nested inside another class's method body and a class inside a
// function both report. Kirk ruled 2026-10-02 (#mbbg6js) that they should not: "one class per file"
// names the design unit a reader greps for and whose history they read, and a class declared
// inside a function, the shape a test's `describe` or `it` fixture takes, is not one. Four nexus
// type tests carried nothing but those. So a declaration counts only outside every function body; a
// class in a namespace or a static block still counts, since neither is a function. ESLint draws the
// same line through Nexus's `nexus/max-classes-per-file` twin, which shares this rule's fixture rows.
//
// Upstream's own `ignoreExpressions` is the other half of the ruling, and it is configuration rather
// than code here: ahra sets it, and so does the ESLint side, so expressions need no divergence.
//
// A class EXPRESSION counts too, and that is the half `ignoreExpressions` exists to turn off. The
// option drops expressions from the count while leaving declarations, rather than lowering the
// limit.
//
// # Where the finding points, which is not the node it is reported on
//
// Upstream reports on the program while overriding the location to span the program's BODY: the
// start of the first statement to the end of the last. The difference is trivia. A file opening with
// a comment has a program starting at position zero and a first statement starting after the
// comment, and upstream's own last corpus case pins that distinction by asserting line 2 column 1
// for a file whose first line is `/* comment */`.
//
// `rule.TokenRange` on the first statement is the same span, and it was checked against all three
// shapes upstream asserts rather than assumed: a plain two-class file, a comment-wrapped one, and an
// indented one inside a template. All three reproduce upstream's line and column exactly.
//
// An empty file cannot report, since a file with no statements has no classes and therefore never
// exceeds a limit of at least one. That is measured rather than argued, and the measurement is
// recorded at the indexing site below, which is why the body span carries no length guard.
//
// # No fixer, matching upstream
//
// Upstream declares none, and the repair is a human decision: which class moves, to which file, and
// what the new file is called. The audit measured 18 sites in this tree for exactly that reason.
var MaxClassesPerFile = rule.Rule{
	Name: "max-classes-per-file",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as a bare severity is handed nil, which the assertion turns into a zero
		// struct whose pointers are nil. Falling back matters more here than usual: a nil Maximum
		// read as 0 would report every file holding one class.
		settings, _ := rule.OptionsAs[MaxClassesPerFileOptions](options)
		defaults := DefaultMaxClassesPerFileSettings()
		if settings.Maximum == nil {
			settings.Maximum = defaults.Maximum
		}
		if settings.IgnoreExpressions == nil {
			settings.IgnoreExpressions = defaults.IgnoreExpressions
		}
		maximum := *settings.Maximum
		ignoreExpressions := *settings.IgnoreExpressions

		return rule.Listeners{
			// The verdict is about the whole file, so it cannot be decided at any one class. The
			// walk is pre-order and a source-file listener fires before its children, so counting
			// and judging happen in the same place, which is what an exit hook would otherwise buy.
			ast.KindSourceFile: func(node *ast.Node) {
				classCount := 0

				// insideFunction is whether the walk is inside any function body, a method, a
				// constructor or an arrow included. See the rule's doc comment for why a declaration
				// there does not count.
				insideFunction := false
				var visit func(*ast.Node) bool
				visit = func(current *ast.Node) bool {
					switch current.Kind {
					case ast.KindClassDeclaration:
						if !insideFunction {
							classCount++
						}
					case ast.KindClassExpression:
						if !ignoreExpressions {
							classCount++
						}
					}
					if ast.IsFunctionLike(current) && !insideFunction {
						insideFunction = true
						current.ForEachChild(visit)
						insideFunction = false
						return false
					}
					current.ForEachChild(visit)
					return false
				}
				node.ForEachChild(visit)

				if classCount <= maximum {
					return
				}

				// Indexed without a length guard, which is deliberate and measured rather than
				// assumed. Reaching here needs `classCount > maximum` and the maximum is at least
				// 1, so at least two classes exist, and a class node cannot exist in a file with
				// no statements. Probed over 13 shapes including six the parser had to recover
				// from: `class`, `class {}`, `class A`, `@dec class`, `export class` and a stray
				// `}` all either produce a statement alongside the class or produce neither, and
				// every zero-statement input produced zero classes. An empty guard here survived
				// the whole fixture set while inverting it failed 25 lines, which is the signature
				// of an unreachable branch rather than a blind spot.
				//
				// The verdict names its callers: this listener is the only reader of this slice,
				// and it is reached only through the count test above. A second caller, or a change
				// letting the count rise without a statement, voids the argument and the guard has
				// to come back.
				statements := ctx.SourceFile.Statements.Nodes
				first := rule.TokenRange(ctx.SourceFile, statements[0])
				last := statements[len(statements)-1]

				ctx.ReportRange(core.NewTextRange(first.Pos(), last.End()),
					maxClassesPerFileMessage(classCount, maximum))
			},
		}
	},
}
