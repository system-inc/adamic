package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// MaxDepthSettings is the decoded option surface: how deeply blocks may nest.
type MaxDepthSettings struct {
	// Maximum is the deepest nesting allowed before a block reports. Upstream's default is 4.
	Maximum int

	// Disabled reproduces the one option shape under which upstream reports NOTHING at all.
	//
	// `{"maximum": 0}` makes upstream's `option.maximum || option.max` evaluate to `undefined`,
	// and every `depth > undefined` comparison is false, so the rule goes silent rather than
	// reporting every block. Modelled as its own flag because no integer value of Maximum can
	// express it: 0 means "report every block" and there is no number meaning "report none".
	Disabled bool
}

// DefaultMaxDepthSettings is upstream's `defaultOptions: [4]`, spelled out because the zero value
// of the struct is not it: a Maximum of 0 means "no block may nest at all", which is a legal
// configuration and the opposite of the default.
func DefaultMaxDepthSettings() MaxDepthSettings {
	return MaxDepthSettings{Maximum: 4}
}

// DecodeMaxDepthOptions reads the depth limit off the config.
//
// Hand rolled because upstream's schema is a `oneOf`: the option is either a bare integer or an
// object carrying `maximum` or `max`, and `rule.DecodeOptionsInto` expresses neither alternative.
// The corpus configures all three spellings, so all three are accepted:
//
//	["error", 2]                the bare integer
//	["error", { "maximum": 2 }]
//	["error", { "max": 2 }]     an older alias upstream still honours
//
// # The object form has a JavaScript quirk in it, and the two zero spellings do OPPOSITE things
//
// Upstream's handling is two statements:
//
//	if (typeof option === "object" && (hasOwn(option,"maximum") || hasOwn(option,"max")))
//	    maxDepth = option.maximum || option.max;
//
// The `hasOwn` guard decides whether the branch is entered; the `||` inside decides the value, and
// it is a TRUTHINESS pick rather than a presence one. Those two tests disagree exactly when a zero
// is written, and the result is not what either reading alone predicts. Measured against the
// installed rule, reading the limit out of its own message rather than inferring it from a count:
//
//	2                       limit 2      the ordinary object-free spelling
//	0                       limit 0      a bare zero reports every block
//	{"max": 0}              limit 0      `undefined || 0` is 0, so zero WINS here
//	{"maximum": 0}          SILENT       `0 || undefined` is undefined, and every
//	                                     `depth > undefined` is false, so the rule reports nothing
//	{"max": 0, "maximum": 3}  limit 3    the `||` consults maximum first
//	{}                      limit 4      neither key present, so the branch is not entered
//
// Both zero spellings are in upstream's corpus. `{"maximum": 0}` silencing the rule entirely is a
// defect rather than a design, but it is a VERDICT and this port reproduces it: a rule that
// disagreed with the tool it replaces on a configuration somebody wrote would be the worse outcome,
// and the divergence is recorded here rather than silently improved. An earlier draft of this
// decoder read the `||` as a plain truthiness fallback and got BOTH zero rows wrong, which the
// corpus caught.
//
// Empty input decodes to the default rather than failing, which is what a rule configured as a bare
// "error" is handed and what upstream's `defaultOptions` says it means.
func DecodeMaxDepthOptions(raw []byte) (any, error) {
	settings := DefaultMaxDepthSettings()
	if len(raw) == 0 {
		return settings, nil
	}

	var maximum int
	if err := json.Unmarshal(raw, &maximum); err == nil {
		settings.Maximum = maximum
		return settings, nil
	}

	var object struct {
		Maximum *int `json:"maximum"`
		Max     *int `json:"max"`
	}
	if err := rule.UnmarshalOptions(raw, &object); err != nil {
		return settings, err
	}

	// Upstream's `hasOwn` guard: neither key present means the branch is never entered and the
	// initialized default of 4 stands.
	if object.Maximum == nil && object.Max == nil {
		return settings, nil
	}

	// Upstream's `option.maximum || option.max`, evaluated the way JavaScript evaluates it.
	switch {
	case object.Maximum != nil && *object.Maximum != 0:
		settings.Maximum = *object.Maximum
	case object.Max != nil && *object.Max != 0:
		settings.Maximum = *object.Max
	case object.Max != nil:
		// `undefined || 0` and `0 || 0` both yield 0, so a `max` of zero wins whenever `maximum`
		// is absent or is itself zero.
		settings.Maximum = 0
	default:
		// Only `maximum` was written and it is zero, so the expression is `0 || undefined` and the
		// limit becomes undefined. Every comparison against it is false.
		settings.Disabled = true
	}
	return settings, nil
}

// maxDepthSettingsFrom recovers the settings from whatever the config layer handed over.
func maxDepthSettingsFrom(options any) MaxDepthSettings {
	if settings, ok := rule.OptionsAs[MaxDepthSettings](options); ok {
		return settings
	}
	return DefaultMaxDepthSettings()
}

var messageMaxDepthTooDeeply = rule.Message{
	Id: "tooDeeply",
	Description: "Each level of nesting is another condition the reader has to hold in their head " +
		"to know whether this line runs at all. Lift a branch into a named function, return early " +
		"to flatten the happy path, or invert a condition so the exceptional case exits first.",
}

// MaxDepth reports a block nested more deeply than the configured limit.
//
//	valid:   function f() { if (a) { if (b) {} } }         (limit 2)
//	valid:   function f() { if (a) {} else if (b) {} }     (an else-if is one chain, not two levels)
//	valid:   function f() { if (a) {} } function g() { if (b) {} }  (each function counts alone)
//	invalid: function f() { if (a) { if (b) { if (c) {} } } }       (limit 2)
//
// # The depth is per FUNCTION, not per file, and that is the whole shape of the rule
//
// Upstream keeps a stack of counters, pushes one when it enters any function-like node, and counts
// blocks against the top of that stack. So two sibling functions each nesting three deep are two
// separate measurements rather than one of six, and a nested function RESETS the count rather than
// inheriting it. The corpus pins that: `function f() { if (a) { function g() { if (b) {} } } }`
// is clean at limit 1, because `g` starts over.
//
// # Enter and exit, with no exit hook, so the walk is explicit
//
// Upstream is a listener pair -- `IfStatement` and `IfStatement:exit` -- and this tree has neither
// an exit hook nor anything but a pre-order walk. So the whole rule runs inside one
// `KindSourceFile` listener that recurses itself, incrementing on the way down and decrementing on
// the way back up. That is the shape the standard names for a rule that has to know what it is
// INSIDE, and it is why there are no per-kind listeners here.
//
// # `else if` is one construct wearing two nodes, and it is the rule's only real discrimination
//
// `if (a) {} else if (b) {}` parses as an IfStatement whose else-branch is another IfStatement, so
// a naive count reads a two-level nest where a reader sees one flat chain. Upstream's `isElseIf`
// tests `node.parent.type === "IfStatement" && node.parent.alternate === node`, and that maps here
// exactly: measured over the corpus shapes, an `else if` chain gives each inner IfStatement a
// parent IfStatement whose `ElseStatement` is that node, while `else { if (b) {} }` -- braces --
// gives the inner one a Block parent and correctly counts as a new level.
//
// # The span is the KEYWORD, not the statement
//
// Upstream reports `loc: sourceCode.getFirstToken(node).loc`, so the finding covers `if`, `for`,
// `switch` and nothing else. Measured against the installed rule: columns 43-45 for an `if`,
// 27-33 for a `switch`. `rule.TokenRange` is the wrong helper here because it extends to
// `node.End()` and would underline the entire statement including its body;
// `scanner.GetRangeOfTokenAtPosition` alone is the token.
//
// # No fix
//
// Upstream offers none and none is possible: flattening a nest means extracting a function or
// inverting a condition, and only the author knows which.
var MaxDepth = rule.Rule{
	Name: "max-depth",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := maxDepthSettingsFrom(options)
		if settings.Disabled {
			// `{"maximum": 0}` leaves upstream comparing against `undefined`, so it reports
			// nothing at all. Declining the file is the same verdict paid once rather than once
			// per block. See the decoder's option table for the measurement.
			return rule.Listeners{}
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// The walk is done here rather than through per-kind listeners because the rule
				// needs a counter that rises entering a block and falls leaving it, and the
				// listener walk is pre-order with no exit hook to decrement on.
				walker := &maxDepthWalker{ctx: ctx, settings: settings}
				// Upstream's `Program: startFunction`: the file itself is a function frame, so
				// top-level blocks are measured rather than ignored.
				walker.depths = append(walker.depths, 0)
				walker.walk(node)
			},
		}
	},
}

// maxDepthWalker carries the per-function depth stack across the recursive walk.
type maxDepthWalker struct {
	ctx      rule.Context
	settings MaxDepthSettings

	// depths is upstream's `functionStack`. Its top is the nesting depth in force inside the
	// innermost enclosing function, and entering a function pushes a fresh zero so a nested
	// function starts over rather than inheriting its parent's depth.
	depths []int
}

func (w *maxDepthWalker) walk(node *ast.Node) {
	if node == nil {
		return
	}

	pushedFunction := false
	pushedBlock := false

	switch node.Kind {
	// Upstream's startFunction set. `Program` is handled by the caller, since the walk starts
	// there.
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindClassStaticBlockDeclaration:
		w.depths = append(w.depths, 0)
		pushedFunction = true

	// Upstream's pushBlock set, with the else-if exemption folded into the IfStatement arm.
	case ast.KindIfStatement:
		if !maxDepthIsElseIf(node) {
			w.pushBlock(node)
			pushedBlock = true
		}

	case ast.KindSwitchStatement, ast.KindTryStatement, ast.KindDoStatement,
		ast.KindWhileStatement, ast.KindWithStatement, ast.KindForStatement,
		ast.KindForInStatement, ast.KindForOfStatement:
		w.pushBlock(node)
		pushedBlock = true

	// A class member's body is its own frame, and this arm is what makes that true here.
	//
	// Upstream needs no such arm because ESTree wraps every method body in a FunctionExpression,
	// which its startFunction set already covers, so the frame is pushed by the body rather than by
	// the member. Our parser hangs the body directly off the member -- probed, there is no wrapping
	// node at all -- so the frame has to be pushed here instead.
	//
	// The distinguishing shape is a class nested INSIDE an already-deep block, and finding it took
	// a correction worth recording. The obvious argument -- "without a frame, two sibling methods
	// would accumulate depth" -- is WRONG, because the depth decrements on the way out, so siblings
	// never accumulate whether or not a frame is pushed. A mutation deleting this arm survived both
	// the corpus and a first set of sibling-method fixtures written from that argument.
	//
	// What a frame actually does is RESET the count, so it only shows where there is a count to
	// reset. Measured against the installed rule, which reports zero on every row:
	//
	//	function f() { if (a) { class C { m() { if (b) {} } } } }             limit 1, clean
	//	function f() { if (a) { if (b) { class C { m() { if (c) {} } } } } }  limit 2, clean
	//	function f() { if (a) { class C { get g() { if (b) {} } } } }         limit 1, clean
	//	function f() { if (a) { class C { constructor() { if (b) {} } } } }   limit 1, clean
	//
	// Without this arm the port reports all four, because the inner `if` inherits the enclosing
	// block's depth instead of starting over. Verified by deleting the arm and watching exactly
	// those four rows fail.
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		w.depths = append(w.depths, 0)
		pushedFunction = true
	}

	node.ForEachChild(func(child *ast.Node) bool {
		w.walk(child)
		return false
	})

	if pushedBlock {
		w.depths[len(w.depths)-1]--
	}
	if pushedFunction {
		w.depths = w.depths[:len(w.depths)-1]
	}
}

// pushBlock is upstream's `pushBlock`: increment the current function's depth and report if the new
// depth exceeds the limit.
//
// Reporting on the way IN rather than on the way out is upstream's order and it matters for a file
// with several findings, since it decides the sequence they arrive in.
func (w *maxDepthWalker) pushBlock(node *ast.Node) {
	if len(w.depths) == 0 {
		return
	}
	w.depths[len(w.depths)-1]++
	depth := w.depths[len(w.depths)-1]
	if depth <= w.settings.Maximum {
		return
	}

	// The keyword alone, which is upstream's `getFirstToken(node).loc`. See the span section on the
	// rule for why `rule.TokenRange` is the wrong helper.
	tokenRange := scanner.GetRangeOfTokenAtPosition(w.ctx.SourceFile, node.Pos())
	w.ctx.Report(rule.Diagnostic{
		Range:      tokenRange,
		SourceFile: w.ctx.SourceFile,
		Message: rule.Message{
			Id: messageMaxDepthTooDeeply.Id,
			Description: fmt.Sprintf("Blocks are nested too deeply (%d). Maximum allowed is %d. %s",
				depth, w.settings.Maximum, messageMaxDepthTooDeeply.Description),
		},
	})
}

// maxDepthIsElseIf is upstream's `isElseIf`: this IfStatement is the else-branch of another one, so
// it continues a chain rather than opening a new level.
func maxDepthIsElseIf(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindIfStatement {
		return false
	}
	statement := parent.AsIfStatement()
	return statement != nil && statement.ElseStatement == node
}
