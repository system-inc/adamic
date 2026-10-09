package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedLabel = rule.Message{
	Id: "unexpectedLabel",
	Description: "This labels a statement. A label is a second naming system that only `break` and " +
		"`continue` can read, so control can leave from a line that carries no visible marker and " +
		"land somewhere the reader has to search the file to find. Reading the code top to bottom " +
		"no longer tells you where it goes. Extract the labeled region into a function and return " +
		"from it, or restructure the loops so the exit is local.",
}

var messageUnexpectedLabelInBreak = rule.Message{
	Id: "unexpectedLabelInBreak",
	Description: "This `break` names a label, so it exits an enclosing block or loop rather than " +
		"the nearest one. The jump target is not on this line and not adjacent to it, which is " +
		"what makes the reading cost real: the reader has to find the label before they know what " +
		"was left. Return from an extracted function, or use a flag the loops already test.",
}

var messageUnexpectedLabelInContinue = rule.Message{
	Id: "unexpectedLabelInContinue",
	Description: "This `continue` names a label, so it resumes an outer loop rather than the one " +
		"it sits in. Which loop advances is decided by a name written elsewhere, so the iteration " +
		"a reader traces is not the iteration that runs. Restructure so the inner work is a " +
		"function that returns, or move the condition out to the loop it belongs to.",
}

// NoLabelsOptions is the decoded option object.
//
// Both fields default to false, which is upstream's default and also Go's zero value, so a rule
// configured as a bare `"error"` and handed nil options lands on the strict setting the same way
// upstream does. That agreement is a coincidence worth naming rather than relying on silently: see
// the nil-options fallback in `Run` and the fixture that bypasses the decoder.
type NoLabelsOptions struct {
	AllowLoop   bool `json:"allowLoop"`
	AllowSwitch bool `json:"allowSwitch"`
}

// NoLabels flags a labeled statement, and a `break` or `continue` that names a label.
//
//	valid:   var f = { label: foo ()}
//	valid:   while (true) { break; }
//	valid:   A: while (a) { break A; }                 (allowLoop)
//	valid:   A: switch (a) { case 0: break A; }        (allowSwitch)
//	invalid: label: while(true) {}
//	invalid: label: while (true) { break label; }      two findings
//	invalid: A: var foo = 0;
//	invalid: A: switch (a) { case 0: B: { break A; } default: break; };   three findings
//
// A label is a control-flow name that only `break` and `continue` can read, so a jump can leave
// from a line carrying no visible marker and resume somewhere the reader has to go find. The rule
// exists because reading top to bottom stops telling you where control goes.
//
// # One input, up to three findings
//
// The label and the jumps that name it are separate findings, which the corpus states outright:
// `label: while (true) { break label; }` names two ids and the nested-label case names three. A
// port reporting once per input passes nothing here, and a port reporting once per LABEL still
// misses the jump.
//
// # The permission is decided by what the label WRAPS, not by what the jump sits in
//
// `allowLoop` exempts a label whose body is a loop; `allowSwitch` exempts one whose body is a
// switch. The corpus pins that the test is on the labeled statement's own body and nothing else:
//
//	A: for (var a in obj) { for (;;) { switch (a) { case 0: continue A; } } }   allowLoop: clean
//	A: for (var a in obj) { for (;;) { switch (a) { case 0: break A; } } }      allowSwitch: reports
//
// Same shape, same nesting, opposite verdicts, and the only difference is which option is set. The
// `continue A` in the first sits lexically inside a switch and is still permitted, because `A`
// labels a `for`. So a port asking "is this jump inside a loop" reports the first and is wrong; the
// question is what `A` names.
//
// That is why the jumps resolve through a label STACK rather than through their own ancestors. A
// jump names a label; the label's kind decides.
//
// # Reporting the label on exit, which is where the finding ORDER comes from
//
// Upstream reports a labeled statement from `LabeledStatement:exit` and a labeled jump on entry.
// The corpus's three-finding case is what pins it:
//
//	A: switch (a) { case 0: B: { break A; } default: break; };
//	unexpectedLabel, unexpectedLabel, unexpectedLabelInBreak
//
// Both labels report before the break, and `B` is nested inside `A`. Entry order would give A, B,
// break. Exit order gives B, A, break -- still two labels first. The break comes last either way
// because it is deeper than both, but the label pair only precedes it under exit ordering, since on
// entry `A` fires before the break is reached and `B` fires after it. Reproduced here by collecting
// during the walk and emitting labels before jumps, which is the same observable order without an
// exit hook.
//
// There is no `rule.OnExit` and the walk is pre-order, so the whole judgment happens inside one
// `KindSourceFile` listener: it fires before its children, so the walk, the label stack, and the
// emission all live there.
//
// # The trap the first clean case sets
//
// `var f = { label: foo ()}` is an object property named `label`. A rule keyed on the identifier,
// or on a colon, reports it. Anchoring on `KindLabeledStatement` declines it by kind.
//
// No fix. Removing a label means restructuring the control flow it names, which is a rewrite rather
// than a spelling change.
var NoLabels = rule.Rule{
	Name: "no-labels",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as a bare `"error"` is handed NIL rather than a decoded struct, because
		// `DecodeOptionsInto` errors on empty input and the config turns that into nil. Without this
		// the type assertion yields the zero value anyway, which happens to be correct here, but
		// relying on that means the rule is right by accident. Naming it makes the default a
		// decision. `TestNoLabelsDefaultsWithoutTheDecoder` bypasses the decoder to pin it.
		settings := NoLabelsOptions{}
		if decoded, ok := rule.OptionsAs[NoLabelsOptions](options); ok {
			settings = decoded
		}

		// isAllowed answers upstream's `isAllowed(kind)`: a label wrapping a loop is permitted by
		// `allowLoop`, one wrapping a switch by `allowSwitch`, and everything else never.
		isAllowed := func(body *ast.Node) bool {
			// No nil guard, measured rather than assumed. A mutant flipping one to `return true`
			// survived the whole fixture set, so the parse was probed instead of a fixture written:
			// `A:`, `A:}`, `A: else` all recover to a synthesized `KindExpressionStatement` in the
			// body slot, `A:;` gives `KindEmptyStatement`, and nothing produces nil. This matches
			// what `guard-for-in` measured for a for-in body and an if consequent, so it looks like
			// a property of typescript-go's statement recovery rather than of this node type.
			//
			// The verdict names its callers: the two calls below, both on a
			// `KindLabeledStatement`'s own `Statement`.
			// `false` for lookInLabeledStatements, matching upstream, which tests the body node's
			// own type and does not see through a nested label. `A: B: while (a) {}` therefore has
			// `A` wrapping a LabeledStatement rather than a loop, so `allowLoop` does not exempt A.
			// Measured against the installed build; see the test of the same name.
			if ast.IsIterationStatement(body, false) {
				return settings.AllowLoop
			}
			if body.Kind == ast.KindSwitchStatement {
				return settings.AllowSwitch
			}
			return false
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// Two buckets rather than one, so labels emit before jumps and the corpus's stated
				// finding order comes out without an exit hook. See the doc above.
				var labelFindings []*ast.Node
				var jumpFindings []struct {
					node    *ast.Node
					message rule.Message
				}

				// The label stack, innermost last. A jump reads it by name, which is what makes the
				// permission depend on what the label WRAPS rather than on where the jump sits.
				type labelEntry struct {
					name string
					body *ast.Node
				}
				var stack []labelEntry

				// kindOf resolves a jump's label name against the stack, innermost first, and
				// answers whether that label is permitted. A name with no entry answers false,
				// matching upstream's `/* c8 ignore */ return "other"` -- unreachable in valid
				// source, since a jump naming an undeclared label does not parse as one, but the
				// parser recovers from invalid source and this must not index into nothing.
				kindOf := func(name string) bool {
					for index := len(stack) - 1; index >= 0; index-- {
						if stack[index].name == name {
							return isAllowed(stack[index].body)
						}
					}
					return false
				}

				var visit func(*ast.Node) bool
				visit = func(current *ast.Node) bool {
					switch current.Kind {
					case ast.KindLabeledStatement:
						labeled := current.AsLabeledStatement()
						if labeled == nil || labeled.Label == nil {
							break
						}
						if !isAllowed(labeled.Statement) {
							labelFindings = append(labelFindings, current)
						}
						stack = append(stack, labelEntry{
							name: labeled.Label.Text(),
							body: labeled.Statement,
						})
						current.ForEachChild(visit)
						// The pop is inert ON ITS OWN and load-bearing in PAIR with the search
						// direction above, which is a coupling worth stating because a reader
						// deleting either half alone will measure no change.
						//
						// Measured: deleting this line leaves all 40 fixtures green and produces
						// byte-identical findings on five hand-built sibling-label inputs,
						// including `A: switch(a) {} A: while(b) { break A; }`, which was written
						// specifically to catch it and does not. The reason is that `kindOf` scans
						// innermost-first, so it reaches the most recently pushed entry before any
						// stale one and never consults what is beneath. A stale entry is
						// unreachable rather than harmless.
						//
						// Deleting this line AND reversing the search to outermost-first is caught
						// by five lines, which is what makes the equivalence a measurement rather
						// than an argument. The pop stays because the invariant it maintains is
						// what the search is entitled to assume, and a later reader changing the
						// search direction should not have to rediscover this.
						stack = stack[:len(stack)-1]
						return false

					case ast.KindBreakStatement:
						if label := current.AsBreakStatement().Label; label != nil &&
							!kindOf(label.Text()) {
							jumpFindings = append(jumpFindings, struct {
								node    *ast.Node
								message rule.Message
							}{current, messageUnexpectedLabelInBreak})
						}

					case ast.KindContinueStatement:
						if label := current.AsContinueStatement().Label; label != nil &&
							!kindOf(label.Text()) {
							jumpFindings = append(jumpFindings, struct {
								node    *ast.Node
								message rule.Message
							}{current, messageUnexpectedLabelInContinue})
						}
					}

					current.ForEachChild(visit)
					return false
				}
				node.ForEachChild(visit)

				for _, labeled := range labelFindings {
					ctx.ReportNode(labeled, messageUnexpectedLabel)
				}
				for _, jump := range jumpFindings {
					ctx.ReportNode(jump.node, jump.message)
				}
			},
		}
	},
}
