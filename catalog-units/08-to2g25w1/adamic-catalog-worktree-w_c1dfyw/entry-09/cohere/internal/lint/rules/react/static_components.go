package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/static_single_assignment"
)

// StaticComponents flags a component value created during render and then used as a JSX tag.
//
//	valid:   function Inner(p) { return <div>{p.text}</div>; } function Outer() { return <Inner />; }
//	valid:   const Hoisted = () => <div />; function Outer() { const C = Hoisted; return <C />; }
//	valid:   function Outer() { const c = makeIt(); return <c.x />; }
//	invalid: function Example() { const Component = createComponent(); return <Component />; }
//	invalid: function Example() { const Component = new ComponentFactory(); return <Component />; }
//	invalid: function Example(p) { let C; if (p.cond) { C = make(); } else { C = D; } return <C />; }
//
// A component created during render has a new identity on every render, so React unmounts the old
// one and mounts the new one instead of updating it. All of its state is discarded and its entire
// subtree re-renders, every time. The symptom is an input that loses what was typed in it, which
// reads as a state-management bug and is caused by the component's definition sitting inside the
// render rather than beside it.
//
// # Where this comes from
//
// React's own `ValidateStaticComponents.ts`, transcribed by oxc at
// `oxc_react_compiler/src/react_compiler_validation/validate_static_components.rs`, which is the
// version this follows. oxc's *linter* rule at `rules/react/static_components.rs` is a dispatcher:
// its whole body is `run_react_compiler_rule(ctx, ErrorCategory::StaticComponents)`. Twenty two of
// oxc's react rules share that body, so a grep for `run_react_compiler_rule` identifies the family
// at once; the judgment for every one of them lives in the compiler crate rather than in the rule
// file, and reading only the rule file makes a large port look like a ninety line one.
//
// # Why this is the first rule here built on the intermediate representation
//
// The decision is not syntactic and cannot be made syntactic. Three of the six upstream fixtures
// need a value to be followed across assignments, and one needs taint merged where two branches
// join. `let C; if (cond) { C = createComponent(); } else { C = Default; } return <C />;` reports,
// and it reports because the component is dynamic on *one* path. That is a phi node. Answering it
// from the syntax tree means re-deriving evaluation order and control flow per rule, which is the
// duplication `internal/utilities/hir` exists to end.
//
// So this lowers each function-like node, builds single-assignment form, and runs upstream's own
// algorithm over the result: a map from value to the span that made it dynamic, propagated forward
// through loads, stores and phis, consulted at every JSX tag. The transcription is close to
// line-for-line because the instruction set was deliberately built to upstream's shape, and that
// fidelity is what lets a disagreement be read as a defect rather than as a design difference.
//
// # Three behaviours that look wrong and are upstream's, verified by running it
//
// Each of these was measured against the release `oxlint` binary on the input named, because
// reading produced the wrong guess first in two of the three cases.
//
//   - **No component-name test anywhere.** The rule never asks whether a tag looks like a
//     component. `const Ábc = makeIt(); <Ábc />` reports; a lowercase binding used as a plain tag
//     reports too. What decides it is whether the tag is a value at all rather than a literal
//     element name, which the representation answers directly through `JsxTag.Place`. Neither
//     `react.IsLikelyComponentName` nor `react.EnclosingComponent` is consulted, and reaching for
//     one here would be a plausible-looking divergence.
//
//   - **A member-expression tag never reports**, however dynamic its base. `<c.x />`,
//     `<obj.Inner.Deep />` and `<T.x.y />` are all silent upstream. Such a tag is not a `Place`, so
//     no taint can attach to it. This is a real hole in the analysis rather than a considered
//     exemption, and it is reproduced rather than improved on.
//
//   - **A back edge is invisible.** `let C = Known; for (...) { C = makeIt(); } return <C />;` does
//     NOT report, while moving the dynamic assignment above the loop does. The pass walks blocks
//     once in reverse postorder, so taint arriving on an edge from a block it has already visited
//     is never seen. This looked like a defect in our lowering; the representation turned out to
//     place the loop phi correctly with its back edge, and running upstream showed identical
//     behaviour. Faithful, not broken, and recorded here so it is not helpfully repaired into a
//     divergence.
//
// # A JSX tag inside a nested function never reports, and the reason changed under this rule
//
// `const C = makeIt(); const render = () => <C />;` is silent, and upstream is silent on it too,
// measured on the exact input. The mechanism is the compilation gate below: the tag lives in a
// function that is not a unit, and the function that IS a unit does not contain the tag.
//
// This comment previously said the silence came from `FunctionExpression.Captures` being empty, so
// taint could not cross a function boundary at all. That was true when this rule was written and
// stopped being true one commit later, when lowering learned to resolve a closed-over variable to
// the binding it captures. Captures now populate and `LoadContext` is emitted. The BEHAVIOUR did
// not change and neither did the agreement with upstream; only the reason did. It is corrected
// here rather than left, because a stale reason is worse than no reason: the next reader would
// have gone looking for a substrate gap that no longer exists.
//
// Taint deliberately still does not follow a capture. Nothing in the corpus asks it to, upstream
// does not do it, and a rule-local approximation would make the two implementations agree on the
// cases both already handle while disagreeing about why.
//
// # One span where upstream has two
//
// Upstream reports at the JSX tag and attaches a secondary label at the creation site. Our
// `Diagnostic` carries a single `Range`, so the primary span is kept where upstream puts it and the
// creation site moves into the message text. No information is dropped, but a reader comparing
// against React's goldens should know the shape differs deliberately.
var StaticComponents = rule.Rule{
	Name:             "react-hooks/static-components",
	NeedsTypeChecker: true,
	// Reads other files only through shape readers (rule.ExportNameIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// The whole file is taken at once rather than listening on each function kind, because
			// lowering already descends into nested functions through the function arena. Listening
			// per function would lower every inner function twice, once as a child of its parent
			// and once standalone with no enclosing context.
			ast.KindSourceFile: func(node *ast.Node) {
				// A file that cannot hold a component or a hook is not lowered at all; see
				// high_level_intermediate_representation.MayHoldComponentOrHook for why that is exact.
				if !high_level_intermediate_representation.MayHoldComponentOrHook(ctx) {
					return
				}
				forEachCompiledFunction(node, func(functionNode *ast.Node) {
					// Shared with the other rules that lower this same function; see high_level_intermediate_representation.ForFunction.
					// Construct runs inside the cached computation, because it mutates in place and is
					// not idempotent.
					lowered := high_level_intermediate_representation.ForFunction(ctx, functionNode)
					if lowered == nil {
						return
					}
					analyzeCompiledFunction(ctx, lowered)
				})
			},
		}
	},
}

// forEachCompiledFunction visits the outermost function-like node on each branch of the tree.
//
// Descending past a function it has handed to the caller would lower that function's children a
// second time, and the standalone copy would be lowered without the bindings its parent supplies,
// so its references would resolve as globals and carry no taint.
//
// A node that turns out not to be a component or a hook is still visited: the caller's gate
// declines it, and lowering has already inlined whatever it contains into the graph of whichever
// enclosing function IS a unit. The one shape that needs care is a non-component wrapper around a
// component, `function outer() { function Widget() {...} }`, where the wrapper is not a unit and the
// component inside it is. That is handled by lowering rather than here: `Lower` builds the nested
// function into the parent's arena, and the gate walks that arena.
func forEachCompiledFunction(root *ast.Node, visit func(*ast.Node)) {
	if root == nil {
		return
	}
	root.ForEachChild(func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			visit(node)
			return false
		}
		forEachCompiledFunction(node, visit)
		return false
	})
}

// reportDynamicComponents runs upstream's forward taint pass over one lowered function and its
// nested functions.
//
// `Blocks` is in reverse postorder, which is what makes a single forward walk correct for
// everything except a back edge, and the back edge is upstream's limitation too.
func reportDynamicComponents(ctx rule.Context, function *high_level_intermediate_representation.Function) {
	if function == nil {
		return
	}

	// Value to the span of whatever made it dynamic. Upstream keys this by identifier and carries
	// an optional span; single-assignment form means one entry per value rather than per binding,
	// which is the property that makes the propagation below sound without a fixpoint.
	dynamic := map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{}

	for _, block := range function.Blocks {
		// A phi is dynamic when any operand reaching it is. Upstream stops at the first such
		// operand, and the choice of which span to keep is arbitrary in both, since a value merged
		// from two dynamic paths has two equally true creation sites.
		for _, phi := range block.Phis {
			for _, entry := range phi.Operands {
				if creator, ok := dynamic[entry.Place.Identifier]; ok {
					dynamic[phi.Place.Identifier] = creator
					break
				}
			}
		}

		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			target := instruction.LValue.Identifier

			switch value := instruction.Value.(type) {
			// Creating a function during render is the case the rule is named for. The other three
			// are values whose identity this pass cannot prove stable: a call, a construction, or a
			// method call could return anything, including a fresh component each time.
			case *high_level_intermediate_representation.FunctionExpression:
				dynamic[target] = target
			case *high_level_intermediate_representation.CallExpression:
				dynamic[target] = value.Callee.Identifier
			case *high_level_intermediate_representation.NewExpression:
				dynamic[target] = value.Callee.Identifier
			case *high_level_intermediate_representation.MethodCall:
				dynamic[target] = value.Property.Identifier

			// Reading and writing carry taint along, which is what makes an alias chain of any
			// length behave like the value at its head.
			case *high_level_intermediate_representation.LoadLocal:
				if creator, ok := dynamic[value.Place.Identifier]; ok {
					dynamic[target] = creator
				}
			case *high_level_intermediate_representation.StoreLocal:
				if creator, ok := dynamic[value.Value.Identifier]; ok {
					dynamic[target] = creator
					dynamic[value.LValue.Identifier] = creator
				}

			// A tag naming a value is the only shape that can carry taint. A host element such as
			// `div` and a member expression such as `obj.Inner` both arrive with no Place, so both
			// are silent regardless of what their names resolve to.
			case *high_level_intermediate_representation.JsxExpression:
				if value.Tag.Place == nil {
					continue
				}
				creation, ok := dynamic[value.Tag.Place.Identifier]
				if !ok {
					continue
				}
				message := staticComponentsMessage(ctx, function, creation)
				if tagNode := identifierNode(function, value.Tag.Place.Identifier); tagNode != nil {
					ctx.ReportNode(tagNode, message)
				} else {
					ctx.ReportRange(value.Tag.Place.Range, message)
				}
			}
		}
	}

}

// analyzeCompiledFunction runs the pass over the functions upstream would have compiled.
//
// # Why this gate exists, and what it cost to find
//
// The React Compiler does not analyse every function. `getReactFunctionType` in its `Program.ts`,
// transcribed by oxc at `react_compiler/entrypoint/program.rs`, admits a function only when its
// name looks like a component or a hook AND its body actually creates JSX or calls a hook. Anything
// else is never a compilation unit, so the validator never runs over it; when such a function is
// nested inside one that IS compiled, its body is analysed as part of the parent rather than on its
// own terms.
//
// Without this gate the rule analysed every function-like node in the file, and the difference is
// not theoretical. A dry run over 3,407 files produced seven findings, and ALL SEVEN were the same
// shape: a render callback defined inside a component, of the form
//
//	function Widget() {
//	  const render = function (item) { const Icon = iconFor(item); return <Icon />; };
//	  return <div>{render(x)}</div>;
//	}
//
// Upstream reports none of them, measured by bisecting a real file down to that shape. Every one
// would have been a false positive on Kirk's tree, and the fixture corpus could not have caught a
// single one, because upstream's six invalid cases all put the creation directly in the component
// where both implementations agree.
//
// # The intuitive reading is the wrong one, so it is recorded here
//
// These seven look like true positives, and by the rule's stated purpose they arguably are: a
// component built inside a callback that runs during render does get a fresh identity. That reading
// is what makes this divergence dangerous rather than obvious. Fidelity is the authority here, the
// gate is upstream's, and improving on it silently would make our findings and oxc's disagree on
// real code while both looked correct.
func analyzeCompiledFunction(ctx rule.Context, function *high_level_intermediate_representation.Function) {
	if function == nil {
		return
	}

	// A function upstream would not compile is not a subject, and the walk continues past it into
	// the functions it contains, because one of those may be a unit even when its wrapper is not.
	if !isCompilationUnit(function) {
		for _, nested := range function.Functions {
			analyzeCompiledFunction(ctx, nested)
		}
		return
	}
	// A unit found nested in a function that is not one is analyzed as React Compiler compiles it,
	// on its own; see AsCompilationUnit.
	function = high_level_intermediate_representation.AsCompilationUnit(ctx, function,
		high_level_intermediate_representation.ForFunction)

	// The OUTERMOST component or hook is the compilation unit, and the walk stops here. A function
	// nested inside it is not a second subject: upstream lowers it into this graph and analyses it
	// as part of this function, so descending would report its body twice.
	//
	// Measured, because the natural reading is the other one. For
	//
	//	function Outer() { function Inner() { const C = mk(); return <C />; } return <Inner />; }
	//
	// upstream reports ONCE, at `<Inner />`, and never at the inner `<C />`. Descending into
	// `Inner` as its own unit produced two findings where upstream gives one.
	//
	// The apparent counter-example resolves the same way. When the outer function returns `Inner`
	// rather than JSX it is not a component at all, so nothing admits it, and the walk continues
	// down to `Inner`, which is one. That is `forEachOutermostFunction` plus this gate agreeing,
	// not an exception to either.
	reportDynamicComponents(ctx, function)
}

// identifierNode returns the syntactic node a value came from, or nil for a pure temporary.
//
// This is the seam `Identifier.Node` exists for. Reporting through it rather than through
// `Place.Range` is not a stylistic choice: a Place's range is the range of the node that produced
// the INSTRUCTION, and for a JSX tag that node is the whole element, so `<Component />` yields a
// span starting at the `<`. Handing the node to `ctx.ReportNode` routes through `rule.TokenRange`,
// which trims leading trivia at the harness and is the mechanism every other rule here relies on.
//
// The defect this replaced was invisible in the obvious fixture: with no leading newline the
// element's raw start coincides with the tag identifier's start, so the span read correctly. oxc's
// tester writes its cases with a leading newline, which is the only reason it was caught.
func identifierNode(function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) *ast.Node {
	if function == nil || int(id) >= len(function.Identifiers) {
		return nil
	}
	identifier := function.Identifiers[id]
	if identifier == nil {
		return nil
	}
	return identifier.Node
}

// staticComponentsMessage renders the finding, naming the construct that created the component.
//
// The creation site is upstream's secondary span, which our single-Range Diagnostic has nowhere to
// put, so it is quoted into the message instead. Read from the source text rather than rebuilt from
// the representation, because the representation holds values rather than syntax and
// `props.foo.bar()` has no readable spelling there.
func staticComponentsMessage(ctx rule.Context, function *high_level_intermediate_representation.Function, creator static_single_assignment.IdentifierId) rule.Message {
	created := ""
	if node := identifierNode(function, creator); node != nil && ctx.SourceFile != nil {
		span := rule.TokenRange(ctx.SourceFile, node)
		sourceText := ctx.SourceFile.Text()
		if span.Pos() >= 0 && span.End() <= len(sourceText) && span.Pos() < span.End() {
			created = sourceText[span.Pos():span.End()]
		}
	}

	where := "in this render"
	if created != "" {
		where = "at `" + created + "`"
	}

	return rule.Message{
		Id: "staticComponents",
		Description: "This component is created during render, " + where + ", so it gets a new " +
			"identity every time this function runs. React compares components by identity, so it " +
			"unmounts the previous one and mounts this as a fresh component rather than updating " +
			"it, discarding all of its state and re-rendering its whole subtree on every render. " +
			"The usual symptom is an input that loses what was typed into it. Move the component " +
			"to module scope, or pass the value that varies as a prop.",
	}
}

// isCompilationUnit reports whether upstream would compile this function, and therefore whether the
// validator ever runs over it.
//
// Both halves of `getReactFunctionType` are required and the name alone is not enough. `hir` already
// classifies by name, which is the first half; the second is that the body must actually create JSX
// or call a hook. Measured: `function Widget() { const C = mk(); return C; }` is capitalized and is
// still not compiled, because it returns a value rather than an element.
//
// The name half is deliberately taken from `high_level_intermediate_representation.Function.Kind` rather than recomputed, so this rule
// and the representation cannot drift into disagreeing about what a component name is. Upstream's
// test is `is_ascii_uppercase`, and `classifyFunction` is ASCII too, so they agree today.
func isCompilationUnit(function *high_level_intermediate_representation.Function) bool {
	if function == nil || function.Kind == high_level_intermediate_representation.FunctionKindOther {
		return false
	}
	return createsJsx(function.Node)
}

// createsJsx reports whether a function body creates JSX, WITHOUT descending into nested functions.
//
// The non-descent is the whole point rather than an optimization. A component whose body is only a
// callback that returns JSX is not itself a component to upstream, which is what separates
//
//	function Widget() { const render = () => <div />; return render; }   not a unit
//	function Widget() { const render = () => <div />; return <div />; }  a unit
//
// and getting it wrong in the permissive direction is what produced seven false positives on the
// real tree before this gate existed.
//
// # Why the hook half of upstream's test is absent, rather than forgotten
//
// `getReactFunctionType` admits a function that creates JSX OR calls a hook. Only the first half is
// asked here, and a mutant deleting the hook half survived every fixture, which is the signal that
// sent this back to the code rather than to another test.
//
// The branch is unreachable for THIS rule. A finding requires a JSX tag naming a tainted value, so
// every reportable input contains JSX. The question is whether that JSX can be somewhere this walk
// does not look, and the only such place is inside a nested function. A tag there is never reported
// anyway: the enclosing function is not a compilation unit, so nothing analyses it, and the unit
// above it does not contain the tag. So no input can exist where the hook half changes the answer.
//
// Falsified rather than argued: a component calling `useEffect` whose only JSX sits inside a `.map`
// callback reports nothing upstream either, measured on that input.
//
// This argument originally rested on captures being untracked, which stopped being true one commit
// after this rule landed. It was re-run against the new substrate rather than assumed to survive:
// both probe inputs still report nothing and still agree with upstream, because the gate rather
// than the capture gap is what makes the branch unreachable. Any change to the compilation gate
// must revisit this.
func createsJsx(node *ast.Node) bool {
	body := functionBodyOf(node)
	if body == nil {
		return false
	}
	found := false
	var walk func(*ast.Node)
	walk = func(current *ast.Node) {
		if current == nil || found {
			return
		}
		switch current.Kind {
		case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
			found = true
			return
		// A nested function is a separate unit and its contents say nothing about this one.
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindConstructor, ast.KindClassDeclaration, ast.KindClassExpression:
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(body)
	return found
}

// functionBodyOf returns the body a function-like node executes, or nil when it has none.
//
// An arrow with an expression body is included: `() => <div />` creates JSX and upstream treats it
// as such, so returning nil for it would decline every concisely written component.
func functionBodyOf(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return node.Body()
}
